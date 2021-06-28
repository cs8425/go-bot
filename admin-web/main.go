package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"lib/fakehttp"
	"local/base"
	vlog "local/log"
)

// default config
const (
	fakeHttp = true // hub act as http server
	tls      = true // via https
	wsObf    = true // fake as websocket

	targetUrl    = "/"
	tokenCookieA = "cna"
	tokenCookieB = "_tb_token_"
	tokenCookieC = "_cna"
	userAgent    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36"

	defHubAddr = "wss://cs8425.noip.me:8787"
)

var (
	webAddr = flag.String("web", ":8001", "port for UI")

	verbosity = flag.Int("v", 4, "verbosity")

	hubPubKey, _     = base64.StdEncoding.DecodeString("MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArogYEOHItjtm0wJOX+hSHjGTIPUsRo/TyLGYxWVk79pNWAhCSvH9nfvpx0skefcL/Nd++Qb/zb3c+o7ZI4zbMKZJLim3yaN8IDlgrjKG7wmjB5r49++LrvRzjIJCAoeFog2PfEn3qlQ+PA26TqLsbPNZi9nsaHlwTOqGljg82g23Zqj1o5JfitJvVlRLmhPqc8kO+4Dvf08MdVS6vBGZjzWFmGx9k3rrDoi7tem22MflFnOQhgLJ4/sbd4Y71ok98ChrQhb6SzZKVWN5v7VCuKqhFLmhZuK0z0f/xkBNcMeCplVLhs/gLIU3HBmvbBSYhmN4dDL19cAv1MkQ6lb1dwIDAQAB")
	public_ECDSA, _  = base64.StdEncoding.DecodeString("MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEc8tgqZX82zrd6809YWzJkh4zhvaoCEbkU8yxBW+a9U1L+XgItJGRL33vYecv4lH9ovSNgJiyvnqdmqkJtwq52Q==")
	private_ECDSA, _ = base64.StdEncoding.DecodeString("MHcCAQEEIFABqR2iqeprQ5Mu3236NGFryXU+J8pUlC14ijvhuSBgoAoGCCqGSM49AwEHoUQDQgAEc8tgqZX82zrd6809YWzJkh4zhvaoCEbkU8yxBW+a9U1L+XgItJGRL33vYecv4lH9ovSNgJiyvnqdmqkJtwq52Q==")

	// ECDSA private key for access bot
	//masterKey, _ = base64.StdEncoding.DecodeString("MHcCAQEEIExJp+W5/4LtO4TPl+dtfStzS2o5x+HEknHMfSLNcI4loAoGCCqGSM49AwEHoUQDQgAEm9tjBE8e0jYIcXUkB19q88RVNkuzqle2vJIB9wc4grM4txyn6WpBOFG17QqajSemJarrQ+FFmPEAlVIXEMDt4g==")

	crtFile = flag.String("crt", "", "PEM encoded certificate file for self sign hub")

	//tokenCookieA = flag.String("ca", "cna", "token cookie name A")
	//tokenCookieB = flag.String("cb", "_tb_token_", "token cookie name B")
	//tokenCookieC = flag.String("cc", "_cna", "token cookie name C")

	//userAgent = flag.String("ua", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36", "User-Agent (default: Chrome)")

	hubAddr    = flag.String("t", "", "hub url")
	configJson = flag.String("c", "", "config.json")

	tlsVerify = flag.Bool("k", true, "InsecureSkipVerify")
)

type Config struct {
	HubAddr      string `json:"addr,omitempty"`
	HubPubKey    []byte `json:"hubkey,omitempty"`    // RSA public key for connect to hub
	AdmPrivKey   []byte `json:"admkey,omitempty"`    // ECDSA private key for access hub
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

func main() {
	flag.Parse()
	vlog.Verbosity = *verbosity
	if vlog.Verbosity > 3 {
		vlog.Std.SetFlags(log.LstdFlags | log.Lmicroseconds)
	}

	// TODO: set by web
	conf := &Config{
		HubAddr:    defHubAddr,
		HubPubKey:  hubPubKey,
		AdmPrivKey: private_ECDSA,
		//MasterKey: masterKey,
		UserAgent:    userAgent,
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

	admin := base.NewAuth()
	admin.HubPubKey = conf.HubPubKey
	admin.Private_ECDSA = conf.AdmPrivKey
	admin.Public_ECDSA = public_ECDSA // not used
	admin.MasterKey = conf.MasterKey

	var dialFn DialFn
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

		dialFn = func() (net.Conn, error) {
			vlog.Vln(1, "connect", huburl)
			return cl.Dial()
		}
	} else {
		dialFn = func() (net.Conn, error) {
			vlog.Vln(1, "connect", huburl)
			return net.Dial("tcp", huburl)
		}
	}

	api := NewWebAPI(admin)
	go api.Start(dialFn)
	webStart(api, *webAddr)
}

const (
	wwwPath = "./www"
)

var (
	//go:embed www
	wwwRoot embed.FS
)

type WebFS []http.FileSystem

func (fs WebFS) Open(name string) (file http.File, err error) {
	for _, i := range fs {
		file, err = i.Open(path.Join(wwwPath, name))
		if err == nil {
			return
		}
	}
	return
}

func webStart(api *WebAPI, addr string) {
	if addr == "" {
		return
	}
	vlog.Vf(2, "Web Listening: %v\n\n", addr)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/node/", api.Node)
	mux.HandleFunc("/api/local/", api.Local)
	mux.HandleFunc("/api/rev/", api.Reverse)
	mux.HandleFunc("/api/key/", api.Keys)
	// mux.HandleFunc("/api/cmd", api.Cmd)

	// SPA
	mux.Handle("/", http.FileServer(WebFS{
		http.Dir("."),
		http.FS(wwwRoot),
	}))

	srv := &http.Server{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		Addr:         addr,
		Handler:      mux,
	}
	err := srv.ListenAndServe()
	vlog.Vln(2, "[web]state:", err)
}
