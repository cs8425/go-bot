package base

import (
	"crypto/rand"
	"crypto/rsa"
	"net"
	"runtime"

	"lib/smux"
	"local/streamcoder"
	kit "local/toolkit"
)

type RSAPrivKey struct {
	Key    *rsa.PrivateKey
	Bytes  []byte // raw key
	KeyLen int
}

func newRSAPrivKeyBase64(keylen int, rawbyte []byte) *RSAPrivKey {
	key, err := kit.ParseRSAPriv(rawbyte)
	if err != nil {
		return nil
	}
	return &RSAPrivKey{
		Key:    key,
		Bytes:  rawbyte,
		KeyLen: keylen / 8,
	}
}

type Hub struct {
	Proc      int
	HubKeyTag string
	OnePerIP  bool

	Pool  *Pool
	IKeys map[string]*RSAPrivKey
	AKeys map[string][]byte
	CTags map[string]bool

	evFn EventCallbackFunc
}

func NewHub() *Hub {
	h := &Hub{
		HubKeyTag: initKeyTag,
		Proc:      1,
		OnePerIP:  true,
		Pool:      NewPool(),
		IKeys:     make(map[string]*RSAPrivKey),
		AKeys:     make(map[string][]byte),
		CTags:     make(map[string]bool),
	}

	h.CTags[clientAgentTag] = true

	return h
}

func NewHubM() *Hub {
	h := &Hub{
		HubKeyTag: initKeyTag,
		Proc:      runtime.NumCPU(),
		Pool:      NewPool(),
		OnePerIP:  true,
		IKeys:     make(map[string]*RSAPrivKey),
		AKeys:     make(map[string][]byte),
		CTags:     make(map[string]bool),
	}

	h.CTags[clientAgentTag] = true

	return h
}

// Warring: lock it yourself
func (h *Hub) DefIKey(keylen int, keybyte []byte) {
	h.IKeys[h.HubKeyTag] = newRSAPrivKeyBase64(keylen, keybyte)
}

// Warring: lock it yourself
func (h *Hub) AddIKey(tag string, keylen int, keybyte []byte) {
	h.IKeys[tag] = newRSAPrivKeyBase64(keylen, keybyte)
}

// Warring: lock it yourself
func (h *Hub) DefAKey(key []byte) {
	h.AKeys[adminAgentTag] = key
}

// Warring: lock it yourself
func (h *Hub) AddAKey(tag string, key []byte) {
	h.AKeys[tag] = key
}

// Warring: lock it yourself
func (h *Hub) SetCTag(tag string) {
	h.CTags[tag] = true
}

// Warring: lock it yourself
func (h *Hub) DelCTag(tag string) {
	delete(h.CTags, tag)
}

// Warring: lock it yourself
func (h *Hub) SetEventCallback(fn EventCallbackFunc) {
	h.evFn = fn
}

func (h *Hub) HandleClient(p1 net.Conn) {
	defer func() {
		kit.TrollConn(p1)
	}()

	// do handshake
	tag, err := kit.ReadTagStr(p1)
	if err != nil {
		Vln(3, "Read Tag err:", err)
		return
	}
	//Vln(2, "Tag:", len(tag), tag)

	// do decode
	privKey, ok := h.IKeys[tag]
	if !ok {
		Vln(3, "tag not found!")
		return
	}

	ciphertext := make([]byte, privKey.KeyLen, privKey.KeyLen)
	n, err := p1.Read(ciphertext)
	if err != nil || n != privKey.KeyLen {
		Vln(3, "read ciphertext err:", err, n)
		return
	}

	encKey, err := kit.DecRSA(privKey.Key, ciphertext)
	if err != nil {
		Vln(3, "RSA decode err:", err)
		return
	}
	//Vln(5, "encKey = ", encKey)

	// do encrypt
	enccon, _ := streamcoder.NewCoder(p1, encKey[0:64], encKey[64:88], false)

	// send nonce
	nonce := make([]byte, 32, 32)
	rand.Read(nonce)
	kit.WriteTagByte(enccon, nonce)

	// rekey
	pass := append(encKey[0:64], nonce...)
	enccon.ReKey(pass)

	// check agent
	agent, err := kit.ReadTagStr(enccon)
	if err != nil {
		Vln(3, "Read agent err:", err)
		return
	}
	//Vln(5, "agent:", agent)

	// client
	cok, ok := h.CTags[agent]
	if ok && cok {
		// get UUID
		UUID, err := kit.ReadTagByte(enccon)
		if err != nil {
			Vln(3, "Read UUID err:", err)
			return
		}

		// uuidStr := kit.Hex(UUID)
		h.addClient(p1, enccon, agent, UUID)
	}

	// admin
	aKey, ok := h.AKeys[agent]
	if ok {
		// check Signature
		signature, _ := kit.ReadTagByte(enccon)
		hashed := kit.HashBytes256(pass)
		ok = kit.VerifyECDSA(aKey, hashed, signature)
		if !ok {
			Vln(5, "agent Verify error!", agent)
			return
		}

		// ACK
		kit.WriteTagStr(enccon, agent)

		if h.evFn != nil {
			// add connect event callback
			warpConn, err := h.evFn(EV_admin_conn, enccon, nil, agent)
			if err != nil { // auth failed
				return
			}

			h.doREPL(warpConn)

			// add disconnect event callback
			h.evFn(EV_admin_conn_cls, warpConn, nil)
			return
		}

		h.doREPL(enccon)
	}

	return
}

func (h *Hub) RunEmbed(c *Client) {
	p0, p1 := net.Pipe()
	go c.TakeOver(p0)
	h.addClient(p1, p1, "embed", []byte("embed"))
}

func (h *Hub) addClient(p1 net.Conn, enccon net.Conn, agent string, UUID []byte) {
	// stream multiplex
	smuxConfig := smux.DefaultConfig()
	mux, err := smux.Server(enccon, smuxConfig)
	if err != nil {
		Vln(3, "mux init err", err)
		return
	}

	// add to pool
	uuidStr := kit.Hex(UUID)
	// uuidStr := string(UUID)
	addr := p1.RemoteAddr().String()
	peer := NewPeer(p1, mux, UUID)

	if h.OnePerIP {
		// get old clients & send signal
		oldlist := h.Pool.CheckOld(UUID, addr)
		// Vf(3, "[client][oldlist]%v\n", len(oldlist))
		for _, peer := range oldlist {

			go func(item *Peer) {
				// Vf(3, "[client][old peer]%v\n", item.id, item)
				p1, err := item.Mux.OpenStream()
				if err != nil {
					return
				}
				defer p1.Close()
				Vf(3, "[client][old peer]kill %v\n", item.id)
				kit.WriteTagStr(p1, B_kill)
			}(peer)

		}
	}

	id, ok := h.Pool.AddPear(peer)
	if !ok {
		Vf(2, "[client][new][same ID & Addr]%v %v %v %v %v\n", id, uuidStr, addr, agent, peer)
		return
	}
	Vf(2, "[client][new]%v %v %v %v %v\n", id, uuidStr, addr, agent, peer)

	// hack for OnClose
	for {
		// TODO: random pull info
		_, err := mux.AcceptStream()
		if err != nil {
			mux.Close()
			p1.Close()
			Vf(2, "[client][cls]%v %v %v %v\n", id, uuidStr, addr, peer)
			h.Pool.DelPear(id)
			break
		}
	}
}

func (h *Hub) doREPL(conn net.Conn) {
	// stream multiplex
	smuxConfig := smux.DefaultConfig()
	mux, err := smux.Server(conn, smuxConfig)
	if err != nil {
		Vln(3, "mux init err", err)
		return
	}

	Vln(3, "[admin]connect start")
	for {
		p1, err := mux.AcceptStream()
		if err != nil {
			mux.Close()
			break
		}

		go h.doOP(conn, p1)
	}
	Vln(3, "[admin]connect end")
}

func (h *Hub) doOP(mainConn net.Conn, p1 net.Conn) {
	// read OP
	op, err := kit.ReadTagStr(p1)
	if err != nil {
		Vln(3, "Read OP err:", err)
		p1.Close()
	}
	defer p1.Close()

	// TODO: add event callback
	switch op {
	case H_ls:
		// list all
		h.Pool.WriteListTo(p1)

	case H_fetch:
		// pull select
		item, ok := doSelect(p1, h.Pool)
		if !ok {
			return
		}

		// force pull info
		conn, err := item.Mux.OpenStream()
		if err != nil {
			return
		}
		defer conn.Close()

		// op to client
		kit.WriteTagStr(conn, B_info)
		ret64, err := kit.ReadVLen(conn)
		if err != nil || int(ret64) != 0 {
			Vln(2, "[pull]err", ret64, err)
			kit.WriteVLen(p1, int64(-1)) // ret
			return
		}

		newinfo := NewInfo()
		n, err := newinfo.ReadFrom(conn)
		if err != nil {
			kit.WriteVLen(p1, int64(-1)) // ret
			return
		}
		item.Info = newinfo
		Vln(3, "[pull]Info:", n, item.Info)

		// ret
		kit.WriteVLen(p1, int64(0))

	case H_sync:
		// pull select
		item, ok := doSelect(p1, h.Pool)
		if !ok {
			return
		}
		kit.WriteVLen(p1, int64(0))

		item.Info.WriteTo(p1)

	case H_select:
		// select & send
		item, ok := doSelect(p1, h.Pool)
		if !ok {
			kit.WriteVLen(p1, int64(-1)) // ret
			return
		}

		// pipe client & admin
		conn, err := item.Mux.OpenStream()
		if err != nil {
			return
		}
		defer conn.Close()

		if h.evFn != nil {
			// add new_stream event callback
			newConn, err := h.evFn(EV_admin_stream, mainConn, p1)
			if err != nil { // auth failed
				return
			}

			kit.Cp(conn, newConn)

			// add close_stream event callback
			h.evFn(EV_admin_stream_cls, mainConn, newConn)
			return
		}

		kit.Cp(conn, p1)
	}
}

func doSelect(p1 net.Conn, pool *Pool) (item *Peer, ok bool) {
	id, err := kit.ReadTagStr(p1)
	if err != nil {
		Vln(3, "Read ID err:", err)
		return nil, false
	}

	item, ok = pool.GetByUTag(id)
	if !ok {
		kit.WriteVLen(p1, int64(-1))
		return
	}

	return
}
