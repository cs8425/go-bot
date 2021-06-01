package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"lib/smux"
	"local/base"
	vlog "local/log"
	kit "local/toolkit"
)

type AtomicBool struct {
	atomic.Value
}

func (c *AtomicBool) MarshalJSON() ([]byte, error) {
	if val, ok := c.Load().(bool); ok {
		if val {
			return []byte("true"), nil
		} else {
			return []byte("false"), nil
		}
	}
	return []byte("null"), nil
}
func (c *AtomicBool) Set(val bool) {
	c.Store(val)
}
func (c *AtomicBool) Get() bool {
	if val, ok := c.Load().(bool); ok {
		return val
	}
	return false
}

type ConnPool struct {
	Mx    sync.RWMutex
	Conns map[net.Conn]net.Conn
}

func (cp *ConnPool) Add(conn net.Conn) {
	cp.Mx.Lock()
	cp.Conns[conn] = conn
	cp.Mx.Unlock()
}
func (cp *ConnPool) Del(conn net.Conn) {
	cp.Mx.Lock()
	delete(cp.Conns, conn)
	cp.Mx.Unlock()
}
func (cp *ConnPool) KillAll() {
	cp.Mx.RLock()
	for _, conn := range cp.Conns {
		conn.Close()
	}
	cp.Mx.RUnlock()
}
func NewConnPool() *ConnPool {
	return &ConnPool{
		Conns: make(map[net.Conn]net.Conn, 8),
	}
}

type loSrv struct {
	ID    string       `json:"id"` // node id
	Addr  string       `json:"addr"`
	Args  []string     `json:"args,omitempty"`
	Pause AtomicBool   `json:"pause,omitempty"` // atomic bool
	Admin *base.Auth   `json:"-"`
	Lis   net.Listener `json:"-"`
	Conns *ConnPool    `json:"-"`
}

func (srv *loSrv) Init() error {
	lis, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		vlog.Vln(2, "[local]Error listening:", err.Error())
		return err
	}
	srv.Lis = lis
	return nil
}

func (srv *loSrv) Start() {
	defer srv.Lis.Close()

	for {
		if conn, err := srv.Lis.Accept(); err == nil {
			//vlog.Vln(2, "[local][new]", conn.RemoteAddr())

			// pause, close connection
			if srv.Pause.Get() {
				conn.Close()
				continue
			}

			// TODO: check client still online
			go srv.handleClient(conn)
		} else {
			vlog.Vln(2, "[local]Accept err", err)
			return
		}
	}
}

func (srv *loSrv) handleClient(p0 net.Conn) {
	defer p0.Close()

	cp := srv.Conns
	cp.Add(p0)
	defer cp.Del(p0)

	admin := srv.Admin
	id := srv.ID
	argv := srv.Args

	mode := argv[0]
	switch mode {
	case "socks":
		//vlog.Vln(2, "socksv5")
		p1, err := admin.GetConn2Client(id, base.B_fast0)
		if err != nil {
			vlog.Vln(2, "[socks]init err", err)
			return
		}
		defer p1.Close()

		// do socks5
		base.HandleSocksF(p0, p1)

	case "http":
		//vlog.Vln(2, "http")
		p1, err := admin.GetConn2Client(id, base.B_fast0)
		if err != nil {
			vlog.Vln(2, "[http]init err", err)
			return
		}
		defer p1.Close()

		// do http
		handleHttp(p0, p1)

	case "raw":
		//vlog.Vln(2, "raw")
		if len(argv) < 2 {
			vlog.Vln(2, "[raw]need target url!!")
			return
		}

		p1, err := admin.GetConn2Client(id, base.B_fast0)
		if err != nil {
			vlog.Vln(2, "[raw]init err", err)
			return
		}
		defer p1.Close()

		raw2fast(p0, p1, argv[1])

	default:
	}
}

type revSrv struct {
	CID    int        `json:"cid"` // connection id
	ID     string     `json:"id"`  // node id
	Addr   string     `json:"addr"`
	Target string     `json:"target"`
	Args   []string   `json:"args,omitempty"`
	Pause  AtomicBool `json:"pause,omitempty"` // atomic bool
	Admin  *base.Auth `json:"-"`
	Conn   net.Conn   `json:"-"`
	Conns  *ConnPool  `json:"-"`
}

func (srv *revSrv) Init(p1 net.Conn) (string, error) {
	srv.Conn = p1
	kit.WriteTagStr(p1, srv.Addr)

	ret64, err := kit.ReadVLen(p1)
	if err != nil {
		vlog.Vln(3, "[rev]bind err0", err)
		return "", err
	}
	if int(ret64) != 0 {
		vlog.Vln(3, "[rev]bind err", ret64)
		return "", errors.New("bind ret code != 0")
	}

	bindAddr, err := kit.ReadTagStr(p1)
	if err != nil {
		vlog.Vln(3, "[rev]Error get binding addr:", err)
		return "", errors.New("get bind addr error")
	}
	srv.Addr = bindAddr
	vlog.Vln(1, "[rev]bind on:", bindAddr)
	return bindAddr, nil
}

func (srv *revSrv) Start() {
	p1 := srv.Conn
	defer p1.Close()

	cp := srv.Conns
	target := srv.Target
	handleFn := func(p1 net.Conn) {
		defer p1.Close()

		cp.Add(p1)
		defer cp.Del(p1)

		p2, err := net.DialTimeout("tcp", target, 10*time.Second)
		if err != nil {
			vlog.Vln(3, "[rev][err]target", target, err)
			return
		}
		defer p2.Close()

		vlog.Vln(3, "[rev][got]", target)
		kit.Cp(p1, p2)
	}

	// stream multiplex
	smuxConfig := smux.DefaultConfig()
	mux, err := smux.Server(p1, smuxConfig)
	if err != nil {
		vlog.Vln(3, "[rev]mux init err", err)
		return
	}
	for {
		conn, err := mux.AcceptStream()
		if err != nil {
			mux.Close()
			break
		}

		// pause, close connection
		if srv.Pause.Get() {
			conn.Close()
			continue
		}

		go handleFn(conn)
	}
}

// thanks: http://www.golangnote.com/topic/141.html
// p1 = http client, p2 = fast server
func handleHttp(client net.Conn, p2 net.Conn) {
	var b [1024]byte
	n, err := client.Read(b[:])
	if err != nil {
		vlog.Vln(3, "[http]read err", client, err)
		return
	}
	var method, host, address string
	idx := bytes.IndexByte(b[:], '\n')
	if idx == -1 {
		vlog.Vln(3, "[http]parse err", idx, client.RemoteAddr())
		return
	}
	fmt.Sscanf(string(b[:idx]), "%s%s", &method, &host)

	if strings.Index(host, "://") == -1 {
		host = "//" + host
	}
	hostPortURL, err := url.Parse(host)
	if err != nil {
		vlog.Vln(3, "[http]parse hostPortURL err", client, hostPortURL, err)
		return
	}
	if strings.Index(hostPortURL.Host, ":") == -1 { // no port, default 80
		address = hostPortURL.Host + ":80"
	} else {
		address = hostPortURL.Host
	}

	vlog.Vln(3, "[http]dial to:", method, address)
	var target = append([]byte{0, 0, 0, 0x05}, []byte(address)...)
	p2.Write(target)

	var b2 [10]byte
	n2, err := p2.Read(b2[:10])
	if n2 < 10 {
		vlog.Vln(2, "[http]dial err replay:", address, n2)
		return
	}
	if err != nil || b2[1] != 0x00 {
		vlog.Vln(2, "[http]dial err:", address, n2, b2[1], err)
		return
	}

	if method == "CONNECT" {
		client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	} else {
		p2.Write(b[:n])
	}

	kit.Cp(client, p2)
}

func raw2fast(p0, p1 net.Conn, targetAddr string) {
	host, portStr, err := net.SplitHostPort(targetAddr)
	if err != nil {
		vlog.Vln(2, "[raw]SplitHostPort err:", targetAddr, err)
		return
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		vlog.Vln(2, "[raw]failed to parse port number:", portStr, err)
		return
	}
	if port < 1 || port > 0xffff {
		vlog.Vln(2, "[raw]port number out of range:", portStr, err)
		return
	}

	socksReq := []byte{0x05, 0x01, 0x00, 0x03}
	socksReq = append(socksReq, byte(len(host)))
	socksReq = append(socksReq, host...)
	socksReq = append(socksReq, byte(port>>8), byte(port))

	var b [10]byte

	// send server addr
	p1.Write(socksReq)

	// read reply
	n, err := p1.Read(b[:10])
	if n < 10 {
		vlog.Vln(2, "[raw]Dial err replay:", targetAddr, n)
		return
	}
	if err != nil || b[1] != 0x00 {
		vlog.Vln(2, "Dial err:", targetAddr, n, b[1], err)
		return
	}

	kit.Cp(p0, p1)
}
