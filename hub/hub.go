package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"sync/atomic"

	"crypto/tls"
	"net/http"

	"lib/fakehttp"
	"local/base"
	vlog "local/log"

	"gopkg.in/natefinch/lumberjack.v2"
)

// default config
const (
	fakeHttp = true // hub act as http server
	wsObf    = true // fake as websocket
	onlyWs   = false

	targetUrl    = "/"
	tokenCookieA = "cna"
	tokenCookieB = "_tb_token_"
	tokenCookieC = "_cna"

	defBindAddr = "wss://:8787/"
	defCrtFile  = "server.crt"
	defKeyFile  = "server.key"
	defWww      = "./www"

	Verbosity = 3
)

var (
	bindAddr = flag.String("l", "", "bind port")
	verb     = flag.Int("v", 6, "verbosity")

	//fakeHttp = flag.Bool("http", true, "act as http server")

	dir = flag.String("d", defWww, "web/file server root dir")

	//tokenCookieA = flag.String("ca", "cna", "token cookie name A")
	//tokenCookieB = flag.String("cb", "_tb_token_", "token cookie name B")
	//tokenCookieC = flag.String("cc", "_cna", "token cookie name C")
	//headerServer = flag.String("hdsrv", "nginx", "http header: Server") // not yet

	crtFile = flag.String("crt", "", "PEM encoded certificate file")
	keyFile = flag.String("key", "", "PEM encoded private key file")

	configJson = flag.String("c", "", "config.json")
)

type Config struct {
	HubPrivKey  []byte            `json:"hubkey,omitempty"`  // RSA private key for client check
	AdmPubKey   []byte            `json:"admkey,omitempty"`  // ECDSA public key for admin check
	HubPrivKeys map[string][]byte `json:"hubkeys,omitempty"` // RSA private key for client check
	AdmPubKeys  map[string][]byte `json:"admkeys,omitempty"` // ECDSA public key for admin check

	BindAddr     string `json:"bind,omitempty"` // raw, http, ws (https/wss by key/crt)
	OnlyWs       bool   `json:"onlyws,omitempty"`
	WwwRoot      string `json:"www,omitempty"` // web/file server root dir
	TokenCookieA string `json:"ca,omitempty"`  // token cookie name A
	TokenCookieB string `json:"cb,omitempty"`  // token cookie name B
	TokenCookieC string `json:"cc,omitempty"`  // token cookie name C
	CrtFile      string `json:"crt,omitempty"` // PEM encoded certificate file
	KeyFile      string `json:"key,omitempty"` // PEM encoded private key file
}

func parseJSONConfig(config *Config, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(config)
}

func main() {
	vlog.Std.SetFlags(log.LstdFlags | log.Lmicroseconds)
	flag.Parse()
	vlog.Verbosity = *verb

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(&lumberjack.Logger{
		Filename:   "./hub.log",
		MaxSize:    64,   // megabytes
		MaxBackups: 0,    // 0 for not to remove
		MaxAge:     0,    //days, 0 for not to remove
		Compress:   true, // disabled by default
	})

	ikey, _ := base64.StdEncoding.DecodeString("MIIEowIBAAKCAQEArogYEOHItjtm0wJOX+hSHjGTIPUsRo/TyLGYxWVk79pNWAhCSvH9nfvpx0skefcL/Nd++Qb/zb3c+o7ZI4zbMKZJLim3yaN8IDlgrjKG7wmjB5r49++LrvRzjIJCAoeFog2PfEn3qlQ+PA26TqLsbPNZi9nsaHlwTOqGljg82g23Zqj1o5JfitJvVlRLmhPqc8kO+4Dvf08MdVS6vBGZjzWFmGx9k3rrDoi7tem22MflFnOQhgLJ4/sbd4Y71ok98ChrQhb6SzZKVWN5v7VCuKqhFLmhZuK0z0f/xkBNcMeCplVLhs/gLIU3HBmvbBSYhmN4dDL19cAv1MkQ6lb1dwIDAQABAoIBAQCXlhqY5xGlvTgUg0dBI43XLaWlFWyMKLV/9UhEAknFzOwqTpoNb9qgUcD9WHVo/TpLM3vTnNGmh4YblOBhcSCbQ4IB9zfqiPTxJASlp7rseIlBvMcKyOKgZS7K1gOxILXfRzndcH0MUjjvfdjYHceM5VtcDT24i+kO1Q9p/5RSqfGu9wz56tqEQE4Z1OTzD+dD9tGeciiyZ9qDoDC/tb0oBKSFK+DlZZOrSBSpGk2Qur4BgVAgL3wunATzGpxxaCAf+9lBEUBCrZbUkeQIKoFbvjqee5Fb2tfdqquMG1FX3CuCovsW7aMKjpAK5TsKuZD88EWje42JV6wmJ/Q4nGvBAoGBAMs6Hs/UX60uZ10mTVKoHU/Mm6lr/FBDo4LF165SX/+sH87KbNlmOO9YBZGJBm1AnsxaNYLjT39EiGlZZbCYRwre/D/9z+hY9J0Yhz/eo8fGsee3f7SU8U9kRH0CFn5MI8Wf7YgNH97uky9i41rqYtkxf2GvqMYl5yzVpQk3fu0XAoGBANvaZQs9DuwFwekzncFcejLHv2CQEDDqtEybmh5PB9YHN+RyHRlxPmYC1d1ElvHO65Tfhgcd0fL0EkSHCXFHfmsIcpSHuUlBpFSrI6btygf+U/U8VLwzXI71cpoE5n+E7rR0J5hTvTo/FccdilV/CubgIZbQ6VSaAxw4HBA5JzahAn9Q+NdN91AnsFV+x8QHKvSC1wMufdgKIukDMdC9pBSbyfjia8Ty2cfVlTyiv/XPke+zfD3V6LvD+Ypgbz4VHpcvvajD1l0ANnFAJoW87PhUoNZBfNtlF/MNruWa6ToNGEkodJAvpQsNyADc4Im1r62y3AXk5hhY2sFBG96lzXbFAoGBAKhoBUhzj++ZhWz13dyU0wH84gq8r7pYvp2D/61BynXW96hlBQdNKIgJmfqxJJK7dteF1Ou0mvLopOmbKs97/UlNoj9GK9cCkjdNFLU0prIyzesnOJ0lFrxnJU73e/yoPhU6eG4FjwiD9FGevi05cIdjnjchdeoZQ1KlZFHFBdWhAoGBAMrwhd20ww6/VrVQShLVB0P3Zn3aKUqUvU9si616iyNSpuZ9dstXYNYAbPav02PL0NOPMDHC6/SERbJQQCnnBqbDBwmUHVmr0W3rvD+DUgihpgTTxArb0FfguJQlKN6whlHOLrf6sC1YebQWhFvPTNQqfSjfO9/g37usDNcskguf")
	akey, _ := base64.StdEncoding.DecodeString("MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEc8tgqZX82zrd6809YWzJkh4zhvaoCEbkU8yxBW+a9U1L+XgItJGRL33vYecv4lH9ovSNgJiyvnqdmqkJtwq52Q==")

	conf := &Config{
		BindAddr:     defBindAddr,
		HubPrivKey:   ikey,
		AdmPubKey:    akey,
		TokenCookieA: tokenCookieA,
		TokenCookieB: tokenCookieB,
		TokenCookieC: tokenCookieC,
		CrtFile:      defCrtFile,
		KeyFile:      defKeyFile,
	}
	if *configJson != "" {
		err := parseJSONConfig(conf, *configJson)
		if err != nil {
			fmt.Println("[config]parse error", err)
		}
	}
	if *bindAddr != "" {
		conf.BindAddr = *bindAddr
	}
	if *crtFile != "" {
		conf.CrtFile = *crtFile
	}
	if *keyFile != "" {
		conf.KeyFile = *keyFile
	}
	if *dir != "" || *dir != defWww {
		conf.WwwRoot = *dir
	}

	u, err := url.Parse(conf.BindAddr)
	if err != nil {
		fmt.Println("[config]parse addr error", err)
		return
	}
	useFakeHttp := fakeHttp
	useWs := wsObf
	useTLS := conf.CrtFile != "" && conf.KeyFile != ""
	TargetUrl := u.Path //targetUrl
	bind := u.Host
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

	hub := base.NewHubM()
	hub.DefIKey(2048, conf.HubPrivKey)
	hub.DefAKey(conf.AdmPubKey)
	hub.OnePerIP = false

	hub.SetEventCallback(handleEvent)

	var srv net.Listener
	if useFakeHttp {

		// simple http Handler setup
		fileHandler := http.FileServer(http.Dir(conf.WwwRoot))
		websrv := fakehttp.NewHandle(fileHandler) // bind handler
		websrv.UseWs = useWs
		websrv.OnlyWs = onlyWs
		switch TargetUrl {
		case "", "/":
			http.Handle("/", websrv) // now add to http.DefaultServeMux
		default:
			http.Handle(TargetUrl, websrv)
			http.Handle("/", fileHandler) // now add to http.DefaultServeMux
		}

		// start http server
		httpSrv := &http.Server{Addr: bind, Handler: nil}
		go startServer(httpSrv, useTLS, conf.CrtFile, conf.KeyFile)

		srv = websrv
	} else {
		lis, err := net.Listen("tcp", bind)
		if err != nil {
			vlog.Vln(2, "Error listening:", err.Error())
			os.Exit(1)
		}
		defer lis.Close()
		srv = lis
		vlog.Vln(2, "listening on:", lis.Addr())
	}
	vlog.Vln(2, "verbosity:", vlog.Verbosity)

	for {
		if conn, err := srv.Accept(); err == nil {
			//vlog.Vln(2, "remote address:", conn.RemoteAddr())

			go hub.HandleClient(conn)
		} else {
			vlog.Vln(2, "Accept err:", err)
		}
	}

}

func startServer(srv *http.Server, useTLS bool, crtFile string, keyFile string) {
	var err error

	// check tls
	if useTLS {
		cfg := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{

				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,

				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,

				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, // http/2 must
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,   // http/2 must

				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,

				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,

				tls.TLS_RSA_WITH_AES_256_GCM_SHA384, // weak
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,    // waek
			},
		}
		srv.TLSConfig = cfg
		//srv.TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0) // disable http/2

		vlog.Vf(2, "[server] HTTPS server Listen on: %v", srv.Addr)
		err = srv.ListenAndServeTLS(crtFile, keyFile)
	} else {
		vlog.Vf(2, "[server] HTTP server Listen on: %v", srv.Addr)
		err = srv.ListenAndServe()
	}

	if err != http.ErrServerClosed {
		vlog.Vf(2, "[server] ListenAndServe error: %v", err)
	}
}

type Conn struct {
	net.Conn
	Agent string

	tx int64 // write
	rx int64 // read
}

func (c *Conn) Read(data []byte) (n int, err error) {
	n, err = c.Conn.Read(data)
	atomic.AddInt64(&c.rx, int64(n))
	return n, err
}

func (c *Conn) Write(data []byte) (n int, err error) {
	n, err = c.Conn.Write(data)
	atomic.AddInt64(&c.tx, int64(n))
	return n, err
}

func (c *Conn) Rx() (n int64) {
	return atomic.LoadInt64(&c.rx)
}
func (c *Conn) Tx() (n int64) {
	return atomic.LoadInt64(&c.tx)
}

func (c *Conn) RxByteString() string {
	return Vsize(c.Rx())
}
func (c *Conn) TxByteString() string {
	return Vsize(c.Tx())
}

func NewConn(p1 net.Conn) *Conn {
	return &Conn{
		Conn: p1,
	}
}

func handleEvent(evType string, mainConn net.Conn, streamConn net.Conn, argv ...interface{}) (warpConn net.Conn, err error) {
	// TODO: log to file
	switch evType {
	case base.EV_admin_conn:
		Vf(2, "[ev][%v] %v <= %v", evType, mainConn.RemoteAddr(), argv)
		// Rx: admin => hub
		// Tx: hub => admin
		return NewConn(mainConn), nil

	case base.EV_admin_conn_cls:
		conn := mainConn.(*Conn)
		Vf(2, "[ev][%v] %v <= %v  Rx = %v / Tx = %v", evType, mainConn.RemoteAddr(), argv, conn.RxByteString(), conn.TxByteString())

	case base.EV_admin_stream:
		Vf(2, "[ev][%v] %v <- %v", evType, mainConn.RemoteAddr(), argv)
		// Rx: admin => hub => bot
		// Tx: bot => hub => admin
		return NewConn(streamConn), nil

	case base.EV_admin_stream_cls:
		conn := streamConn.(*Conn)
		Vf(2, "[ev][%v] %v <- %v  Rx = %v / Tx = %v", evType, mainConn.RemoteAddr(), argv, conn.RxByteString(), conn.TxByteString())

	}
	return nil, nil
}

func Vsize(byteCount int64) (ret string) {
	s := " "
	tmp := float64(byteCount)
	size := uint64(byteCount)

	switch {
	case size < uint64(2<<9):

	case size < uint64(2<<19):
		tmp = tmp / float64(2<<9)
		s = "K"

	case size < uint64(2<<29):
		tmp = tmp / float64(2<<19)
		s = "M"

	case size < uint64(2<<39):
		tmp = tmp / float64(2<<29)
		s = "G"

	case size < uint64(2<<49):
		tmp = tmp / float64(2<<39)
		s = "T"

	}
	ret = fmt.Sprintf("%06.2f %sB", tmp, s)
	return
}

func Vf(level int, format string, v ...interface{}) {
	if level <= Verbosity {
		log.Printf(format, v...)
	}
}
func Vln(level int, v ...interface{}) {
	if level <= Verbosity {
		log.Println(v...)
	}
}
