package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	configFile = flag.String("c", "config.json", "json file")

	bufPool = NewBuffPool(1024 * 1024) // 1 MB

	dialer = net.Dialer{
		Timeout: 15 * time.Second,
	}
)

func main() {
	flag.Parse()

	buf, err := os.ReadFile(*configFile)
	if err != nil {
		Vln(1, "[err]read config", err)
		return
	}

	conf := &Conf{}
	err = json.Unmarshal(buf, conf)
	if err != nil {
		Vln(1, "[err]parse config", err)
		return
	}

	Vln(2, "[local]", len(conf.Local))
	var wg sync.WaitGroup
	for i, srv := range conf.Local {
		err := srv.Init()
		if err != nil {
			continue
		}
		wg.Add(1)
		go srv.Start(&wg)
		Vln(2, "[local]", i, srv.Addr, srv.Args)
	}

	wg.Wait()
}

type Conf struct {
	Local []*loSrv `json:"local"`
	Rev   []*Param `json:"rev"`
}

type Param struct {
	Addr string   `json:"addr"`
	Args []string `json:"args"`
	Tag  string   `json:"tag"`
}

type loSrv struct {
	ID   string       `json:"id"` // node id
	Addr string       `json:"addr"`
	Args []string     `json:"args,omitempty"`
	Tag  string       `json:"tag"`
	Lis  net.Listener `json:"-"`
}

func (srv *loSrv) Init() error {
	lis, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		Vln(2, "[local]Error listening:", err.Error())
		return err
	}
	srv.Lis = lis
	return nil
}

func (srv *loSrv) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	defer srv.Lis.Close()

	for {
		if conn, err := srv.Lis.Accept(); err == nil {
			//vlog.Vln(2, "[local][new]", conn.RemoteAddr())

			// TODO: check client still online
			go srv.handleClient(conn)
		} else {
			Vln(2, "[local]Accept err", err)
			return
		}
	}
}

func (srv *loSrv) handleClient(p0 net.Conn) {
	defer p0.Close()

	argv := srv.Args
	mode := argv[0]
	switch mode {
	case "socks": // do socks5
		// Vln(2, "socksv5")
		handleSocks(p0, dialer)

	case "http": // do http
		// Vln(2, "http")
		handleHttp(p0)

	case "raw":
		// Vln(2, "raw")
		if len(argv) < 2 {
			Vln(2, "[raw]need target url!!")
			return
		}
		backend := argv[1]

		p1, err := dialer.Dial("tcp", backend)
		if err != nil {
			Vln(2, "[raw]init err", backend, err)
			return
		}
		defer p1.Close()

		Vln(3, "[raw]to:", backend)
		Vln(6, "[dbg]conn", p0.RemoteAddr(), "=>", p1.RemoteAddr())
		cp(p0, p1)

	default:
	}
}

var (
	Std = log.New(os.Stdout, "", log.LstdFlags)

	Verbosity int = 2
)

func Vf(level int, format string, v ...interface{}) {
	if level <= Verbosity {
		Std.Printf(format, v...)
	}
}
func V(level int, v ...interface{}) {
	if level <= Verbosity {
		Std.Print(v...)
	}
}
func Vln(level int, v ...interface{}) {
	if level <= Verbosity {
		Std.Println(v...)
	}
}

func replyAndClose(p1 net.Conn, rpy int) {
	p1.Write([]byte{0x05, byte(rpy), 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	p1.Close()
}

// thanks: http://www.golangnote.com/topic/141.html
func handleSocks(p1 net.Conn, dialer net.Dialer) {
	var b [1024]byte
	n, err := p1.Read(b[:])
	if err != nil {
		Vln(3, "[socks]client read", p1, err)
		p1.Close()
		return
	}
	if b[0] != 0x05 { //only Socket5
		p1.Close()
		return
	}

	//reply: NO AUTHENTICATION REQUIRED
	p1.Write([]byte{0x05, 0x00})

	n, err = p1.Read(b[:])
	if b[1] != 0x01 { // 0x01: CONNECT
		replyAndClose(p1, 0x07) // X'07' Command not supported
		return
	}

	var host, port string
	switch b[3] {
	case 0x01: //IP V4
		host = net.IPv4(b[4], b[5], b[6], b[7]).String()
	case 0x03: //DOMAINNAME
		host = string(b[5 : n-2]) //b[4] domain name length
	case 0x04: //IP V6
		host = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
	default:
		replyAndClose(p1, 0x08) // X'08' Address type not supported
		return
	}
	port = strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
	backend := net.JoinHostPort(host, port)
	p2, err := dialer.Dial("tcp", backend)
	if err != nil {
		Vln(2, backend, err)
		replyAndClose(p1, 0x05) // X'05'
		return
	}
	defer p2.Close()

	reply := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	Vln(3, "[socks]to:", backend)
	Vln(6, "[dbg]conn", p1.RemoteAddr(), "=>", p2.RemoteAddr())
	p1.Write(reply) // reply OK

	cp(p1, p2)
}

// thanks: http://www.golangnote.com/topic/141.html
func handleHttp(client net.Conn) {
	defer client.Close()

	var b [1024]byte
	n, err := client.Read(b[:])
	if err != nil {
		Vln(3, "[http]client read err", client, err)
		return
	}
	var method, host, address string
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)

	if strings.Index(host, "://") == -1 {
		host = "//" + host
	}
	hostPortURL, err := url.Parse(host)
	if err != nil {
		Vln(3, "[http]Parse hostPortURL err:", client, hostPortURL, err)
		return
	}
	if strings.Index(hostPortURL.Host, ":") == -1 { // no port, default 80
		address = hostPortURL.Host + ":80"
	} else {
		address = hostPortURL.Host
	}

	Vln(3, "[http]to:", method, address)
	server, err := net.Dial("tcp", address)
	if err != nil {
		Vln(2, "[http]Dial err:", address, err)
		return
	}
	defer server.Close()

	if method == "CONNECT" {
		client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	} else {
		server.Write(b[:n])
	}

	cp(client, server)
}

func cp(p1, p2 net.Conn) {
	// start tunnel
	p1die := make(chan struct{})
	go func() {
		buf := bufPool.New()
		io.CopyBuffer(p1, p2, *buf)
		close(p1die)
		bufPool.Put(buf)
	}()

	p2die := make(chan struct{})
	go func() {
		buf := bufPool.New()
		io.CopyBuffer(p2, p1, *buf)
		close(p2die)
		bufPool.Put(buf)
	}()

	// wait for tunnel termination
	select {
	case <-p1die:
	case <-p2die:
	}
}

type BuffPool struct {
	sz int
	pl *sync.Pool
}

func (pl *BuffPool) New() *[]byte {
	return pl.pl.Get().(*[]byte)
}

func (pl *BuffPool) Put(b *[]byte) {
	*b = (*b)[:pl.sz]
	pl.pl.Put(b)
}

func NewBuffPool(maxSize int) *BuffPool {
	return &BuffPool{
		sz: maxSize,
		pl: &sync.Pool{
			New: func() interface{} {
				b := make([]byte, maxSize)
				return &b
			},
		},
	}
}
