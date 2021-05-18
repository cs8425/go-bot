package main

import (
	"bufio"
	"fmt"
	"flag"
	"log"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	//"strconv"
	"sync"
	"encoding/json"
	"bytes"

	"io/ioutil"
	"encoding/base64"

	"lib/fakehttp"
	kit "local/toolkit"
	"local/base"
	vlog "local/log"
)


// default config
const (
	fakeHttp = true // hub act as http server
	tls = true // via https
	wsObf = true // fake as websocket

	targetUrl = "/"
	tokenCookieA = "cna"
	tokenCookieB = "_tb_token_"
	tokenCookieC = "_cna"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36"

	defHubAddr = "wss://cs8425.noip.me:8787"
)

var (
	hubPubKey, _ = base64.StdEncoding.DecodeString("MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArogYEOHItjtm0wJOX+hSHjGTIPUsRo/TyLGYxWVk79pNWAhCSvH9nfvpx0skefcL/Nd++Qb/zb3c+o7ZI4zbMKZJLim3yaN8IDlgrjKG7wmjB5r49++LrvRzjIJCAoeFog2PfEn3qlQ+PA26TqLsbPNZi9nsaHlwTOqGljg82g23Zqj1o5JfitJvVlRLmhPqc8kO+4Dvf08MdVS6vBGZjzWFmGx9k3rrDoi7tem22MflFnOQhgLJ4/sbd4Y71ok98ChrQhb6SzZKVWN5v7VCuKqhFLmhZuK0z0f/xkBNcMeCplVLhs/gLIU3HBmvbBSYhmN4dDL19cAv1MkQ6lb1dwIDAQAB")
	public_ECDSA, _ = base64.StdEncoding.DecodeString("MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEc8tgqZX82zrd6809YWzJkh4zhvaoCEbkU8yxBW+a9U1L+XgItJGRL33vYecv4lH9ovSNgJiyvnqdmqkJtwq52Q==")
	private_ECDSA, _ = base64.StdEncoding.DecodeString("MHcCAQEEIFABqR2iqeprQ5Mu3236NGFryXU+J8pUlC14ijvhuSBgoAoGCCqGSM49AwEHoUQDQgAEc8tgqZX82zrd6809YWzJkh4zhvaoCEbkU8yxBW+a9U1L+XgItJGRL33vYecv4lH9ovSNgJiyvnqdmqkJtwq52Q==")

	// ECDSA private key for access bot
	//masterKey, _ = base64.StdEncoding.DecodeString("MHcCAQEEIExJp+W5/4LtO4TPl+dtfStzS2o5x+HEknHMfSLNcI4loAoGCCqGSM49AwEHoUQDQgAEm9tjBE8e0jYIcXUkB19q88RVNkuzqle2vJIB9wc4grM4txyn6WpBOFG17QqajSemJarrQ+FFmPEAlVIXEMDt4g==")

	verb = flag.Int("v", 6, "Verbosity")

	crtFile    = flag.String("crt", "", "PEM encoded certificate file for self sign hub")

	//tokenCookieA = flag.String("ca", "cna", "token cookie name A")
	//tokenCookieB = flag.String("cb", "_tb_token_", "token cookie name B")
	//tokenCookieC = flag.String("cc", "_cna", "token cookie name C")

	//userAgent = flag.String("ua", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36", "User-Agent (default: Chrome)")

	hubAddr = flag.String("t", "", "hub url")
	configJson = flag.String("c", "", "config.json")

	tlsVerify = flag.Bool("k", true, "InsecureSkipVerify")
)

type Config struct {
	HubAddr      string `json:"addr,omitempty"`
	HubPubKey    []byte `json:"hubkey,omitempty"` // RSA public key for connect to hub
	AdmPrivKey   []byte `json:"admkey,omitempty"` // ECDSA private key for access hub
	MasterKey    []byte `json:"masterkey,omitempty"` // ECDSA private key for access bot
	UserAgent    string `json:"useragent,omitempty"`
	TokenCookieA string `json:"ca,omitempty"` // token cookie name A
	TokenCookieB string `json:"cb,omitempty"` // token cookie name B
	TokenCookieC string `json:"cc,omitempty"` // token cookie name C
}

func parseJSONConfig(config *Config, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(config)
}

var admin *base.Auth
var localserver []*loSrv

type loSrv struct {
	ID             string
	Addr           string
	Args           []string
	Admin          *base.Auth
	Lis            net.Listener
}

func main() {
	flag.Parse()
	vlog.Verbosity = *verb
	if vlog.Verbosity > 3 {
		vlog.Std.SetFlags(log.LstdFlags | log.Lmicroseconds)
	}

	conf := &Config{
		HubAddr: defHubAddr,
		HubPubKey: hubPubKey,
		AdmPrivKey: private_ECDSA,
		//MasterKey: masterKey,
		UserAgent: userAgent,
		TokenCookieA: tokenCookieA,
		TokenCookieB: tokenCookieB,
		TokenCookieC: tokenCookieC,
	}
	if *configJson != "" {
		err := parseJSONConfig(conf, *configJson)
		if err != nil {
			fmt.Println("[config]parse error", err)
		}
	}
	if *hubAddr != "" {
		conf.HubAddr = *hubAddr
	}

	u, err := url.Parse(conf.HubAddr)
	if err != nil {
		fmt.Println("[config]parse addr error", err)
		return
	}
	useFakeHttp := fakeHttp
	useWs := wsObf
	useTLS := tls
	TargetUrl := u.Path //targetUrl
	huburl := u.Host
	switch u.Scheme {
	case "http":
		useFakeHttp = true
		useWs = false
		useTLS = false
	case "https":
		useFakeHttp = true
		useWs = false
		useTLS = true
	case "ws":
		useFakeHttp = true
		useWs = true
		useTLS = false
	case "wss":
		useFakeHttp = true
		useWs = true
		useTLS = true
	case "tcp":
		useFakeHttp = false
	default:
	}

	admin = base.NewAuth()
	admin.HubPubKey = conf.HubPubKey
	admin.Private_ECDSA = conf.AdmPrivKey
	admin.Public_ECDSA = public_ECDSA // not used
	admin.MasterKey = conf.MasterKey

	var conn net.Conn
	if useFakeHttp {
		var cl *fakehttp.Client
		if useTLS {
			var caCert []byte
			if *crtFile != "" {
				var err error
				caCert, err = ioutil.ReadFile(*crtFile)
				if err != nil {
					vlog.Vln(2, "Reading certificate error:", err)
					os.Exit(1)
				}
			}
			cl = fakehttp.NewTLSClient(huburl, caCert, true)
		} else {
			cl = fakehttp.NewClient(huburl)
		}
		cl.TokenCookieA = conf.TokenCookieA
		cl.TokenCookieB = conf.TokenCookieB
		cl.TokenCookieC = conf.TokenCookieC
		cl.UseWs = useWs
		cl.UserAgent = conf.UserAgent
		cl.Url = TargetUrl
		vlog.Vln(1, "connect", conf.HubAddr)

		conn, err = cl.Dial()
	} else {
		conn, err = net.Dial("tcp", huburl)
	}
	if err != nil {
		vlog.Vln(1, "connect err", err)
		return
	}

	mux, err := admin.InitConn(conn)
	if err != nil {
		vlog.Vln(1, "connect err", err)
		return
	}

	// check connection to hub
	go func(){
		for {
			_, err := mux.AcceptStream()
			if err != nil {
				mux.Close()
				vlog.Vln(2, "connection to hub reset!!")
				break
			}
		}
	}()

	var wg sync.WaitGroup
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		text, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			vlog.Vln(3, "[ReadLine]", err)
			break
		}
		if err == io.EOF {
			break
		}

		wg.Add(1)
		go func (line string, wg *sync.WaitGroup) {
			defer wg.Done()

			cmd := strings.Fields(line)
			found, exit, hasout, out := execute(cmd)
			if found {
				if hasout {
					fmt.Printf("<%v>%v\n", cmd[0], out)
					fmt.Print(">")
				}
			} else {
				fmt.Println("command not found!")
			}
			if exit {
				//break
			}
		}(text, &wg)
	}
	wg.Wait()
}

func execute(args []string) (bool, bool, bool, string) {
	if len(args) == 0 {
		return true, false, false, ""
	}

	if op := opFunc[args[0]]; op != nil {
		exit, hasout, out := op(args[1:])
		return true, exit, hasout, out
	}

	return false, false, false, ""
}


// exit?, hasout?, what?
//var opFunc map[string](func([]string) (bool, bool, string))
var opFunc = map[string](func([]string) (bool, bool, string)){
	"bye": opBye,
	"exit": opBye,
	"quit": opBye,
	"ls": opBot,
	"local": opLocal,
}

func opBye(args []string) (bool, bool, string) {
	return true, true, "bye~"
}

func opBot(args []string) (exit bool, hasout bool, out string) {
	exit , hasout , out  = false, false, "\n"

	by := "id"
	if len(args) >= 1 {
		by = args[0]
	}

	p1, err := admin.GetConn(base.H_ls)
	if err != nil {
		return
	}

	list := base.PeerList{}
	n, err := list.ReadFrom(p1)
	if err != nil {
		return
	}

	var pl []*base.PeerInfo
	switch by {
	case "addr":
		pl = list.GetListByAddr()

	case "time":
		pl = list.GetListByTime()

	case "id":
		pl = list.GetListByID()

	case "rtt":
		fallthrough
	default:
		pl = list.GetListByRTT()
	}
	out += fmt.Sprintf("total=%v\n", n)
	for _, v := range pl {
		out += v.String() + "\n"
	}
	out += fmt.Sprintf("total=%v\n", n)

	return false, true, out
}

func opLocal(args []string) (exit bool, hasout bool, out string) {
	// local list
	// local bind bot_id bind_addr [socks|http]
	exit, hasout, out = false, true, `
local ls
local bind $bot_id $bind_addr [socks|http] [mode_argv...]
local stop $bind_addr`

	if len(args) < 1 {
		return
	}

	switch args[0] {
	case "ls":
		hasout, out = true, ""
		for i, srv := range localserver {
			out += fmt.Sprintf("[%v][%v]%v\t%v\n", i, srv.Addr, srv.ID, srv.Args)
		}

	case "stop":
		if len(args) < 2 {
			return
		}
		hasout, out = false, ""
		for i, srv := range localserver {
			if args[1] == srv.Addr {
				out += fmt.Sprintf("[local][stop][%v][%v]%v\t%v\n", i, srv.Addr, srv.ID, srv.Args)
				srv.Lis.Close()
				localserver = append(localserver[:i], localserver[i+1:]...)
				hasout = true
				break
			}
		}

	default:
		args = append([]string{"bind"}, args...)
		fallthrough
	case "bind":
		if len(args) < 3 {
			return
		}

		if len(args) == 3 {
			args = append(args, "socks")
		}

		srv := &loSrv {
			ID: args[1],
			Addr: args[2],
			Args: args[3:],
			Admin: admin,
		}

		go startLocal(srv)

		hasout, out = true, "ok...\n"
	}
	return
}

func startLocal(srv *loSrv) {

	lis, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		vlog.Vln(2, "[local]Error listening:", err.Error())
		return
	}
	defer lis.Close()
	srv.Lis = lis
	localserver = append(localserver, srv)

	for {
		if conn, err := lis.Accept(); err == nil {
			//vlog.Vln(2, "[local][new]", conn.RemoteAddr())

			// TODO: check client still online
			go handleClient(srv.Admin, conn, srv.ID, srv.Args)
		} else {
			vlog.Vln(2, "[local]Accept err", err)
			return
		}
	}

}

func handleClient(admin *base.Auth, p0 net.Conn, id string, argv []string) {
	defer p0.Close()

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

	default:
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

