// go build proxy.go share.go
package main

import (
	"flag"
	"log"
//	"io"
	"net"
	"net/http"
	"fmt"
//	"os"
//	"os/signal"
//	"syscall"
//	"strings"
	"strconv"

	"time"
	"sync"
	"errors"

	"io/ioutil"
	"encoding/base64"

	"lib/fakehttp"
//	kit "local/toolkit"
	"local/base"
	vlog "local/log"
)


var hubPubKey, _ = base64.StdEncoding.DecodeString("MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArogYEOHItjtm0wJOX+hSHjGTIPUsRo/TyLGYxWVk79pNWAhCSvH9nfvpx0skefcL/Nd++Qb/zb3c+o7ZI4zbMKZJLim3yaN8IDlgrjKG7wmjB5r49++LrvRzjIJCAoeFog2PfEn3qlQ+PA26TqLsbPNZi9nsaHlwTOqGljg82g23Zqj1o5JfitJvVlRLmhPqc8kO+4Dvf08MdVS6vBGZjzWFmGx9k3rrDoi7tem22MflFnOQhgLJ4/sbd4Y71ok98ChrQhb6SzZKVWN5v7VCuKqhFLmhZuK0z0f/xkBNcMeCplVLhs/gLIU3HBmvbBSYhmN4dDL19cAv1MkQ6lb1dwIDAQAB")
var public_ECDSA, _ = base64.StdEncoding.DecodeString("MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEc8tgqZX82zrd6809YWzJkh4zhvaoCEbkU8yxBW+a9U1L+XgItJGRL33vYecv4lH9ovSNgJiyvnqdmqkJtwq52Q==")
var private_ECDSA, _ = base64.StdEncoding.DecodeString("MHcCAQEEIFABqR2iqeprQ5Mu3236NGFryXU+J8pUlC14ijvhuSBgoAoGCCqGSM49AwEHoUQDQgAEc8tgqZX82zrd6809YWzJkh4zhvaoCEbkU8yxBW+a9U1L+XgItJGRL33vYecv4lH9ovSNgJiyvnqdmqkJtwq52Q==")

var verb = flag.Int("v", 6, "Verbosity")
var huburl = flag.String("t", "cs8425.noip.me:8787", "hub url")
var port = flag.String("p", ":8181", "proxy listen on")
var webPort = flag.String("l", ":8888", "web UI listen on")

var (
	fakeHttp = flag.Bool("http", true, "hub act as http server")
	httpTLS = flag.Bool("tls", true, "via https")

	crtFile    = flag.String("crt", "", "PEM encoded certificate file")

	tokenCookieA = flag.String("ca", "cna", "token cookie name A")
	tokenCookieB = flag.String("cb", "_tb_token_", "token cookie name B")
	tokenCookieC = flag.String("cc", "_cna", "token cookie name C")

	userAgent = flag.String("ua", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36", "User-Agent (default: Chrome)")

	wsObf = flag.Bool("usews", true, "fake as websocket")
	tlsVerify = flag.Bool("k", true, "InsecureSkipVerify")
)

var localservers SrvList
var hubs HubList

type HubList struct {
	lock    sync.RWMutex
	hubs    []*hubLink
	nextid  int
}

func (hl *HubList) WriteList(w http.ResponseWriter) {
	hl.lock.RLock()

	fmt.Fprintf(w, "i\thid\taddr\n")
	for i, hub := range hl.hubs {
		fmt.Fprintf(w, "%v\t%v\t%v\n", i, hub.Id, hub.HubAddr)
	}

	hl.lock.RUnlock()
}

func (hl *HubList) Conn(hubaddr string) (*hubLink, error) {
	hub, err := NewHubLinkConn(hubaddr)
	if err != nil {
		return nil, err
	}

	hl.lock.Lock()
	hl.hubs = append(hl.hubs, hub)
	hub.Id = hl.nextid
	hl.nextid += 1

	hl.lock.Unlock()
	return hub, nil
}

func (hl *HubList) StopId(id int) (int, string) {
	hl.lock.Lock()
	defer hl.lock.Unlock()

	for i, hub := range hl.hubs {
		if id == hub.Id {
			hub.Admin.Raw.Close()
			hl.hubs = append(hl.hubs[:i], hl.hubs[i+1:]...)
			return i, hub.HubAddr
		}
	}
	return -1, "not found"
}

func (hl *HubList) GetId(id int) (*hubLink) {
	hl.lock.RLock()
	defer hl.lock.RUnlock()

	for _, hub := range hl.hubs {
		if id == hub.Id {
			return hub
		}
	}
	return nil
}

type SrvList struct {
	lock    sync.RWMutex
	srvs    []*loSrv
}

func (sl *SrvList) WriteList(w http.ResponseWriter) {
	sl.lock.RLock()

	fmt.Fprintf(w, "sid\tbind\tstarted\n")
	for i, srv := range sl.srvs {
		fmt.Fprintf(w, "%v\t%v\t%v\n", i, srv.BindAddr, srv.Lis != nil)
	}

	sl.lock.RUnlock()
}

func (sl *SrvList) Add(srv *loSrv) {
	sl.lock.Lock()
	sl.srvs = append(sl.srvs, srv)
	sl.lock.Unlock()
}

func (sl *SrvList) Get(addr string) (*loSrv) {
	sl.lock.RLock()
	defer sl.lock.RUnlock()

	for _, srv := range sl.srvs {
		if addr == srv.BindAddr {
			return srv
		}
	}
	return nil
}

func (sl *SrvList) Stop(addr string) (int, string) {
	sl.lock.Lock()
	defer sl.lock.Unlock()

	for i, srv := range sl.srvs {
		if addr == srv.BindAddr {
			srv.Close()
			sl.srvs = append(sl.srvs[:i], sl.srvs[i+1:]...)
			return i, srv.BindAddr
		}
	}
	return -1, "not found"
}

func hubOP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "HEAD":
		return
	default:
	}

	var hid int
	var op string

	err := r.ParseForm()
	if err != nil {
		goto BAD_REQ
	}

	hid, _ = strconv.Atoi(r.Form.Get("hid"))
	op = r.Form.Get("op")

	switch op {
	case "":
		fallthrough
	case "ls":
		hubs.WriteList(w)

	case "conn":
		addr := r.Form.Get("bind")
		if addr == "" {
			goto BAD_REQ
		}

		hub, err := hubs.Conn(addr)
		if err != nil {
			goto BAD_REQ
		}
		fmt.Fprintf(w, "hid:%v", hub.Id)

	case "stop":
		idx, addr := hubs.StopId(hid)
		fmt.Fprintf(w, "[stop]%v\t%v\t%v\n", idx, hid, addr)

	case "flush":
		hub := hubs.GetId(hid)
		if hub == nil {
			goto BAD_REQ
		}
		hub.FlushClients()

	case "lsc":
		hub := hubs.GetId(hid)
		if hub == nil {
			goto BAD_REQ
		}
		hub.WriteList(w)

	default:
		goto BAD_REQ
	}
	return

BAD_REQ:
	http.Error(w, "bad request", http.StatusBadRequest)
	return
}

func srvOP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "HEAD":
		return
	default:
	}

	var addr string
	var op string

	err := r.ParseForm()
	if err != nil {
		goto BAD_REQ
	}

	op = r.Form.Get("op")
	switch op {
	case "":
		fallthrough
	case "ls":
		localservers.WriteList(w)
		return
	}

	addr = r.Form.Get("bind")
	if addr == "" {
		goto BAD_REQ
	}

	switch op {
	case "start":
		hid, err := strconv.Atoi(r.Form.Get("hid"))
		if err != nil {
			goto BAD_REQ
		}
		hublink := hubs.GetId(hid)
		if hublink == nil {
			goto BAD_REQ
		}

		// TODO: select clients
		go startSrv(hublink, addr)

		fmt.Fprintf(w, "server start")

	case "stop":
		idx, addr := localservers.Stop(addr)
		fmt.Fprintf(w, "[stop]%v\t%v\n", idx, addr)

	case "mode":
	case "flush":
		lo := localservers.Get(addr)
		if lo == nil {
			goto BAD_REQ
		}
		lo.FlushClients()

	case "lsc":
		lo := localservers.Get(addr)
		if lo == nil {
			goto BAD_REQ
		}
		lo.list.WriteList(w)

	default:
		goto BAD_REQ
	}
	return

BAD_REQ:
	http.Error(w, "bad request", http.StatusBadRequest)
	return
}

func main() {
	flag.Parse()
	vlog.Verbosity = *verb
	if vlog.Verbosity > 3 {
		vlog.Std.SetFlags(log.LstdFlags | log.Lmicroseconds)
	}

	hublink, err := hubs.Conn(*huburl)
	if err != nil {
		vlog.Vln(1, "connect err", err)
		return
	}
	go startSrv(hublink, *port)

	mux := http.NewServeMux()
	mux.HandleFunc("/jquery.min.js", func (w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "www/jquery.min.js")
	})
	mux.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "www/index.html")
	})
	//mux.Handle("/r/", http.StripPrefix("/r/", http.FileServer(http.Dir(config.ResDir))))
	mux.HandleFunc("/api/hub", hubOP)
	mux.HandleFunc("/api/srv", srvOP)
	//mux.HandleFunc("/api/bot", botOP)

	srv := &http.Server{Addr: *webPort, Handler: mux}

	log.Printf("[server] HTTP server Listen on: %v", *webPort)
	err = srv.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Printf("[server] ListenAndServe error: %v", err)
	}
}


var ErrNoClient = errors.New("no available client")

type ClientList interface {
	SetList(list []*base.PeerInfo)
	GetClientId() (string, error)
	RmClientId(id string)

	WriteList(w http.ResponseWriter) // web api
}

type roundList struct {
	lock           sync.RWMutex
	Clients        []*base.PeerInfo
	nextId         int
}

func (cl *roundList) SetList(list []*base.PeerInfo) {
	cl.lock.Lock()
	cl.Clients = list
	cl.lock.Unlock()
}

func (cl *roundList) GetClientId() (string, error) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	if len(cl.Clients) == 0 {
		return "", ErrNoClient
	}

	if cl.nextId >= len(cl.Clients) {
		cl.nextId = 0
	}
	id := cl.Clients[cl.nextId]
	cl.nextId += 1

	return id.UTag, nil
}

func (cl *roundList) RmClientId(id string) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	for i, c := range cl.Clients {
		if c.UTag == id {
			cl.Clients = append(cl.Clients[:i], cl.Clients[i+1:]...)
			break
		}
	}
}

func (cl *roundList) WriteList(w http.ResponseWriter) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	fmt.Fprintf(w, "count:%d\n", len(cl.Clients))
	for _, c := range cl.Clients {
		fmt.Fprintf(w, "%s\n", c.String())
	}
}

type hubLink struct {
	lock           sync.RWMutex
	HubAddr        string
	Admin          *base.Auth
	Clients        []*base.PeerInfo
	Id             int
}

func (hl *hubLink) FlushClients() ([]*base.PeerInfo, error) {
	hl.lock.Lock()
	defer hl.lock.Unlock()

	p1, err := hl.Admin.GetConn(base.H_ls)
	if err != nil {
		return nil, err
	}

	list := base.PeerList{}
	_, err = list.ReadFrom(p1)
	if err != nil {
		return nil, err
	}
	hl.Clients = list.GetListByRTT()

	return hl.Clients, nil
}

func (hl *hubLink) WriteList(w http.ResponseWriter) {
	hl.lock.RLock()
	defer hl.lock.RUnlock()

	fmt.Fprintf(w, "count:%d\n", len(hl.Clients))
	for _, c := range hl.Clients {
		fmt.Fprintf(w, "%s\n", c.String())
	}
}

func NewHubLinkConn(hubaddr string) (*hubLink, error) {
	admin := base.NewAuth()
	admin.HubPubKey = hubPubKey
	admin.Private_ECDSA = private_ECDSA

	// TODO: other transport layer
	var conn net.Conn
	var err error
	if *fakeHttp {
		var cl *fakehttp.Client
		if *httpTLS {
			var caCert []byte
			if *crtFile != "" {
				var err error
				caCert, err = ioutil.ReadFile(*crtFile)
				if err != nil {
					vlog.Vln(2, "Reading certificate error:", err)
					return nil, err
				}
			}
			cl = fakehttp.NewTLSClient(hubaddr, caCert, true)
		} else {
			cl = fakehttp.NewClient(hubaddr)
		}
		cl.TokenCookieA = *tokenCookieA
		cl.TokenCookieB = *tokenCookieB
		cl.TokenCookieC = *tokenCookieC
		cl.UseWs = *wsObf
		cl.UserAgent = *userAgent

		conn, err = cl.Dial()
	} else {
		conn, err = net.Dial("tcp", hubaddr)
	}
	if err != nil {
		return nil, err
	}

	_, err = admin.InitConn(conn)
	if err != nil {
		return nil, err
	}

	h := NewHubLink(hubaddr, admin)
	return h, nil
}

func NewHubLink(hubaddr string, admin *base.Auth) (*hubLink) {
	h := hubLink{
		HubAddr: hubaddr,
		Admin: admin,
	}
	return &h
}

type loSrv struct {
	lock           sync.RWMutex
	BindAddr       string
	Link           *hubLink
	Lis            net.Listener

	list           ClientList // for selected clients
}

func NewLoSrv(hub *hubLink) (*loSrv) {
	lo := loSrv{
		Link: hub,
		list: &roundList{},
	}
	return &lo
}

func (lo *loSrv) Close() (error) {
	lo.lock.Lock()
	defer lo.lock.Unlock()

	err := lo.Lis.Close()
	if err == nil {
		lo.Lis = nil
	}

	return err
}

func (lo *loSrv) FlushClients() ([]*base.PeerInfo, error) {
	lo.lock.Lock()
	defer lo.lock.Unlock()

	pl, err := lo.Link.FlushClients()
	if err != nil {
		return nil, err
	}

	lst := make([]*base.PeerInfo, len(pl), len(pl))
	copy(lst, pl)
	lo.list.SetList(lst)

	return lst, nil
}

func (lo *loSrv) GetClient() (p1 net.Conn, err error) {
	lo.lock.RLock()
	defer lo.lock.RUnlock()

	for {
		utag, err := lo.list.GetClientId()
		if err != nil {
			break
		}
		vlog.Vln(5, "[GetClient]utag:", utag)

		p1, err = lo.Link.Admin.GetConn2Client(utag, base.B_fast0)
		if err == nil {
			return p1, nil
		}
		vlog.Vln(5, "[GetClient]err:", err)

		// remove error client
		lo.list.RmClientId(utag)
	}

	return nil, ErrNoClient
}

func (lo *loSrv) StartSocks() {
	lis, err := net.Listen("tcp", lo.BindAddr)
	if err != nil {
		vlog.Vln(2, "[local]Error listening:", err.Error())
		return
	}
	defer lis.Close()

	lo.lock.Lock()
	lo.Lis = lis
	lo.lock.Unlock()

	for {
		if conn, err := lis.Accept(); err == nil {
			vlog.Vln(2, "[local][new]", conn.RemoteAddr())
			go lo.handleClient(conn)
		} else {
			vlog.Vln(2, "[local]Accept err", err)
			return
		}
	}
}

func (lo *loSrv) handleClient(p0 net.Conn) {
	defer p0.Close()
	//vlog.Vln(2, "socksv5")

	// select client
	p1, err := lo.GetClient()
	if err != nil {
		vlog.Vln(3, "socks5 err", err)
		return
	}
	defer p1.Close()

	// do socks5
	base.HandleSocksF(p0, p1)
}

func startSrv(hublink *hubLink, localport string) {
	srv := NewLoSrv(hublink)
	srv.BindAddr = localport
	mux := srv.Link.Admin.Sess

	localservers.Add(srv)

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

	go srv.StartSocks()

	// TODO: better update & exit check
	for {
		_, err := srv.FlushClients()
		if err != nil {
			return
		}
		time.Sleep(30 * time.Second)
	}
}

