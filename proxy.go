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

	"strings"

	"time"
	"sync"
	"errors"

	"encoding/base64"

	kit "./lib/toolkit"
	"./lib/base"
)


var hubPubKey, _ = base64.StdEncoding.DecodeString("MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArogYEOHItjtm0wJOX+hSHjGTIPUsRo/TyLGYxWVk79pNWAhCSvH9nfvpx0skefcL/Nd++Qb/zb3c+o7ZI4zbMKZJLim3yaN8IDlgrjKG7wmjB5r49++LrvRzjIJCAoeFog2PfEn3qlQ+PA26TqLsbPNZi9nsaHlwTOqGljg82g23Zqj1o5JfitJvVlRLmhPqc8kO+4Dvf08MdVS6vBGZjzWFmGx9k3rrDoi7tem22MflFnOQhgLJ4/sbd4Y71ok98ChrQhb6SzZKVWN5v7VCuKqhFLmhZuK0z0f/xkBNcMeCplVLhs/gLIU3HBmvbBSYhmN4dDL19cAv1MkQ6lb1dwIDAQAB")
var public_ECDSA, _ = base64.StdEncoding.DecodeString("MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEc8tgqZX82zrd6809YWzJkh4zhvaoCEbkU8yxBW+a9U1L+XgItJGRL33vYecv4lH9ovSNgJiyvnqdmqkJtwq52Q==")
var private_ECDSA, _ = base64.StdEncoding.DecodeString("MHcCAQEEIFABqR2iqeprQ5Mu3236NGFryXU+J8pUlC14ijvhuSBgoAoGCCqGSM49AwEHoUQDQgAEc8tgqZX82zrd6809YWzJkh4zhvaoCEbkU8yxBW+a9U1L+XgItJGRL33vYecv4lH9ovSNgJiyvnqdmqkJtwq52Q==")

var verb = flag.Int("v", 6, "Verbosity")
var huburl = flag.String("t", "cs8425.noip.me:8787", "hub url")
var port = flag.String("p", ":8181", "proxy listen on")
var webPort = flag.String("l", ":8888", "web UI listen on")

var verbosity int = 2

var local *loSrv

func srvOP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "HEAD":
		return
	default:
	}

	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	op := r.Form.Get("op")
	switch op {
	case "conn":
		//addr := r.Form.Get("addr")


	case "start":
		//addr := r.Form.Get("addr")


	case "stop":
		if local == nil {
			return
		}
		local.Close()
		local = nil


	case "mode":
	case "flush":
		if local == nil {
			return
		}
		local.FlushClients()
	default:
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
}

func main() {
	flag.Parse()
	verbosity = *verb
	if verbosity > 3 {
		std.SetFlags(log.LstdFlags | log.Lmicroseconds)
	}

	go startBG(*huburl, *port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		if local == nil {
			return
		}

		local.lock.RLock()
		list := local.list.(*roundList)
		local.lock.RUnlock()

		list.lock.RLock()
		defer list.lock.RUnlock()

		fmt.Fprintf(w, "count:%d\n", len(list.Clients))
		for _, utag := range list.Clients {
			fmt.Fprintf(w, "%s\n", utag)
		}
	})
	mux.HandleFunc("/api/srv", srvOP)
	//mux.HandleFunc("/api/bot", botCtrl)


	srv := &http.Server{Addr: *webPort, Handler: mux}

	log.Printf("[server] HTTP server Listen on: %v", *webPort)
	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Printf("[server] ListenAndServe error: %v", err)
	}
}



var ErrNoClient = errors.New("no available client")

type ClientList interface {
	SetList(list []string)
	GetClientId() (string, error)
	RmClientId(id string)
}

type roundList struct {
	lock           sync.RWMutex
	Clients        []string
	nextId         int
}

func (cl *roundList) SetList(list []string) {
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

	return id, nil
}

func (cl *roundList) RmClientId(id string) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	for i, cid := range cl.Clients {
		if cid == id {
			cl.Clients = append(cl.Clients[:i], cl.Clients[i+1:]...)
			break
		}
	}
}

type loSrv struct {
	lock           sync.RWMutex
	Admin          *base.Auth
	Lis            net.Listener

	list           ClientList
}

func NewLoSrv(admin *base.Auth) (*loSrv) {
	lo := loSrv{
		Admin: admin,
		list: &roundList{},
	}
	return &lo
}

func (lo *loSrv) Close() (error) {
	lo.lock.Lock()
	defer lo.lock.Unlock()

	return lo.Admin.Raw.Close()
}

func (lo *loSrv) FlushClients() (error) {
	lo.lock.RLock()
	defer lo.lock.RUnlock()

	p1, err := lo.Admin.GetConn(base.H_ls)
	if err != nil {
		return err
	}
	kit.WriteTagStr(p1, "rtt")

	n, _ := kit.ReadVLen(p1)
	list := make([]string, 0, n)
	for i := 0; i < int(n); i++ {
		id, err := kit.ReadTagStr(p1)
		if err != nil {
			Vln(3, "Read ID err:", err)
			break
		}
		list = append(list, id)
	}

	lo.list.SetList(list)

	return nil
}

func (lo *loSrv) GetClient() (p1 net.Conn, err error) {
	lo.lock.RLock()
	defer lo.lock.RUnlock()

	for {
		utag, err := lo.list.GetClientId()
		if err != nil {
			break
		}

		Vln(5, "[GetClient]utag:", utag)

		arg := strings.Split(utag, " ")
		if len(arg) >= 2 {
			p1, err = lo.Admin.GetConn2Client(arg[0], base.B_fast0)
			if err == nil {
				return p1, nil
			}
			Vln(5, "[GetClient]err:", err)
		}

		// remove error client
		lo.list.RmClientId(utag)
	}

	return nil, ErrNoClient
}

func (lo *loSrv) StartSocks(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		Vln(2, "[local]Error listening:", err.Error())
		return
	}
	defer lis.Close()

	local.lock.Lock()
	local.Lis = lis
	local.lock.Unlock()

	for {
		if conn, err := lis.Accept(); err == nil {
			Vln(2, "[local][new]", conn.RemoteAddr())
			go lo.handleClient(conn)
		} else {
			Vln(2, "[local]Accept err", err)
			return
		}
	}
}

func (lo *loSrv) handleClient(p0 net.Conn) {
	defer p0.Close()
	//Vln(2, "socksv5")

	// select client
	p1, err := lo.GetClient()
	if err != nil {
		Vln(3, "socks5 err", err)
		return
	}
	defer p1.Close()

	// do socks5
	base.HandleSocksF(p0, p1)

	return
}



func startBG(hubaddr string, localport string) {
	admin := base.NewAuth()
	admin.HubPubKey = hubPubKey
	admin.Private_ECDSA = private_ECDSA
	admin.Public_ECDSA = public_ECDSA // not used

	mux, err := admin.CreateConn(hubaddr)
	if err != nil {
		Vln(1, "connect err", err)
		return
	}

	local = NewLoSrv(admin)

	// check connection to hub
	go func(){
		for {
			_, err := mux.AcceptStream()
			if err != nil {
				mux.Close()
				Vln(2, "connection to hub reset!!")
				break
			}
		}
	}()

	go local.StartSocks(localport)

	for {
		local.FlushClients()
		/*for i, utag := range local.Clients {
			Vln(5, "[FlushClients]utag:", i, utag)
		}*/

		time.Sleep(30 * time.Second)
	}
}

