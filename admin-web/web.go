package main

import (
	//"encoding/binary"
	"encoding/json"

	//"fmt"
	//"io"
	//"os"

	//"errors"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"

	"local/base"
	vlog "local/log"
)

type loSrv struct {
	ID    string       `json:"id"` // node id
	Addr  string       `json:"addr"`
	Args  []string     `json:"args,omitempty"`
	Admin *base.Auth   `json:"-"`
	Lis   net.Listener `json:"-"`
}

type revSrv struct {
	CID    int        `json:"cid"` // connection id
	ID     string     `json:"id"`  // node id
	Addr   string     `json:"addr"`
	Target string     `json:"targer"`
	Args   []string   `json:"args,omitempty"`
	Admin  *base.Auth `json:"-"`
	Conn   net.Conn   `json:"-"`
}

type WebAPI struct {
	ln net.Listener // for client connect

	// worker
	//die chan struct{}
	//watch map[uint32]*todo

	mx        sync.RWMutex // lock for runtime data
	adm       *base.Auth   // TODO: multiple
	srvInfo   []*loSrv     // atomic.Value
	revNextID uint32       // atomic add
	revInfo   []*revSrv    // atomic.Value
}

func NewWebAPI(admin *base.Auth) *WebAPI {
	api := &WebAPI{
		adm:     admin,
		srvInfo: make([]*loSrv, 0),

		revNextID: 1,
		revInfo:   make([]*revSrv, 0),
	}
	// TODO: worker
	return api
}

func (api *WebAPI) Start(conn net.Conn) {
	mux, err := api.adm.InitConn(conn)
	if err != nil {
		vlog.Vln(1, "[core]connect err", err)
		return
	}
	// check connection to hub
	for {
		_, err := mux.AcceptStream()
		if err != nil {
			mux.Close()
			vlog.Vln(2, "[core]connection to hub reset!!")
			break
		}
	}
}

func (api *WebAPI) Node(w http.ResponseWriter, r *http.Request) {
	if ok := checkReqType(w, r, "R", false); !ok {
		return
	}

	api.mx.RLock()
	defer api.mx.RUnlock()

	p1, err := api.adm.GetConn(base.H_ls)
	if err != nil {
		return
	}

	list := base.PeerList{}
	if _, err := list.ReadFrom(p1); err != nil {
		return
	}

	// vlog.Vln(3, "[Node]", list)
	JsonRes(w, list)
}

func (api *WebAPI) Local(w http.ResponseWriter, r *http.Request) {
	if ok := checkReqType(w, r, "RW", true); !ok {
		return
	}
	op := r.Form.Get("op")
	// uuid := r.Form.Get("uuid")
	// addr := r.Form.Get("addr")
	// args := r.Form.Get("args")
	if r.Method == "GET" {
		goto RETOK
	}

	switch op {
	case "bind": // bind: uuid, bind_addr, type, argv
		srv, err := api.localBind(r)
		if err != nil {
			goto ERR400
		}
		api.mx.Lock()
		api.srvInfo = append(api.srvInfo, srv)
		api.mx.Unlock()
		go startLocal(srv)
		goto RETOK
	case "stop": // stop
		addr := r.Form.Get("addr")
		var found *loSrv
		idx := -1
		for i, srv := range api.srvInfo {
			if addr == srv.Addr {
				idx = i
				found = srv
				break
			}
		}
		if found == nil {
			goto ERR404
		}

		vlog.Vln(3, "[local][stop]", idx, found.Addr, found.ID, found.Args)
		found.Lis.Close()
		api.srvInfo = append(api.srvInfo[:idx], api.srvInfo[idx+1:]...)
		goto RETOK

	default:
		goto ERR400
	}

RETOK:
	api.mx.RLock()
	defer api.mx.RUnlock()
	JsonRes(w, api.srvInfo)
	return
ERR400:
	http.Error(w, "bad request", http.StatusBadRequest)
	return
ERR404:
	http.Error(w, "not found", http.StatusNotFound)
	return
}

func (api *WebAPI) localBind(r *http.Request) (*loSrv, error) {
	type param struct {
		ID   string   `json:"uuid"`
		Addr string   `json:"bind_addr"`
		Argv []string `json:"argv"`
	}
	p := param{
		Argv: []string{"socks"},
	}

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		return nil, err
	}

	srv := &loSrv{
		ID:    p.ID,
		Addr:  p.Addr,
		Args:  p.Argv,
		Admin: api.adm,
	}
	return srv, nil
}

func (api *WebAPI) Reverse(w http.ResponseWriter, r *http.Request) {
	if ok := checkReqType(w, r, "RW", true); !ok {
		return
	}
	op := r.Form.Get("op")
	if r.Method == "GET" {
		goto RETOK
	}

	switch op {
	case "bind": // uuid, remote_addr, target_addr
		srv, err := api.reverseBind(r)
		if err != nil {
			goto ERR400
		}

		p1, err := api.adm.GetConn2Client(srv.ID, base.B_bind)
		if err != nil {
			vlog.Vln(2, "[rev]init err", err)
			goto ERR500
		}
		srv.Conn = p1
		go handleReverse(p1, srv.Addr, srv.Target)

		api.mx.Lock()
		api.revInfo = append(api.revInfo, srv)
		api.mx.Unlock()

		goto RETOK
	case "stop": // stop
		addr := r.Form.Get("addr")
		var found *revSrv
		idx := -1
		for i, srv := range api.revInfo {
			if addr == srv.Addr {
				idx = i
				found = srv
				break
			}
		}
		if found == nil {
			goto ERR404
		}

		vlog.Vln(3, "[rev][stop]", idx, found.Addr, found.ID, found.Args)
		found.Conn.Close()
		api.revInfo = append(api.revInfo[:idx], api.revInfo[idx+1:]...)
		goto RETOK

	default:
		goto ERR400
	}

RETOK:
	api.mx.RLock()
	defer api.mx.RUnlock()
	JsonRes(w, api.revInfo)
	return
ERR400:
	http.Error(w, "bad request", http.StatusBadRequest)
	return
ERR404:
	http.Error(w, "not found", http.StatusNotFound)
	return
ERR500:
	http.Error(w, "not found", http.StatusInternalServerError)
	return
}

func (api *WebAPI) reverseBind(r *http.Request) (*revSrv, error) {
	type param struct {
		ID     string   `json:"uuid"`
		Addr   string   `json:"remote"`
		Target string   `json:"target"`
		Argv   []string `json:"argv"`
	}
	var p param

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		return nil, err
	}

	revID := atomic.AddUint32(&api.revNextID, 1)
	srv := &revSrv{
		CID:    int(revID),
		ID:     p.ID,
		Addr:   p.Addr,
		Target: p.Target,
		Args:   p.Argv,
		Admin:  api.adm,
	}
	return srv, nil
}

// func (api *WebAPI) Cmd(w http.ResponseWriter, r *http.Request) {
// 	if ok := checkReqType(w, r, "W", true); !ok {
// 		return
// 	}

// RETOK:
// 	return
// ERR400:
// 	http.Error(w, "bad request", http.StatusBadRequest)
// 	return
// ERR404:
// 	http.Error(w, "not found", http.StatusNotFound)
// 	return
// ERRWIP:
// 	http.Error(w, "not implemented", http.StatusNotImplemented)
// 	return
// }

// basic tool function
func checkReqType(w http.ResponseWriter, r *http.Request, typeLim string, parse bool) bool {
	lut := map[string]map[string]int{
		"R": map[string]int{
			"HEAD": 0,
			"GET":  0,
		},
		"W": map[string]int{
			"HEAD": 0,
			"POST": 1,
			"PUT":  1,
		},
		"RW": map[string]int{
			"HEAD": 0,
			"GET":  0,
			"POST": 1,
			"PUT":  1,
		},
	}
	tab, ok := lut[typeLim]
	if !ok {
		return false // not found
	}
	_, ok = tab[r.Method]
	if !ok {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}

	// only parse query for our usage
	if parse {
		//err := r.ParseMultipartForm(16*1024*1024)
		//err := r.ParseForm()
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return false
		}
		r.Form = values
	}
	return true
}

func SetHeader(w http.ResponseWriter) {
	header := w.Header()
	header.Set("Cache-Control", "public, max-age=0, must-revalidate")
	header.Set("Content-Type", "application/json")
}

func JsonRes(w http.ResponseWriter, value interface{}) {
	SetHeader(w)
	enc := json.NewEncoder(w)
	err := enc.Encode(value)
	if err != nil {
		vlog.Vln(3, "[web] json response err:", err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

// func cleanPath(fp string) string {
// 	return filepath.FromSlash(path.Clean("/" + fp))
// }
