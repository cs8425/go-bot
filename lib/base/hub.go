package base

import (
	"net"
	"runtime"
	"crypto/rand"
	"crypto/rsa"

	kit "../toolkit"
	"../streamcoder"
	"../smux"
)


type RSAPrivKey struct {
	Key *rsa.PrivateKey
	Bytes []byte // raw key
	KeyLen int
}

func newRSAPrivKeyBase64(keylen int, rawbyte []byte) (*RSAPrivKey) {
	key, err := kit.ParseRSAPriv(rawbyte)
	if err != nil {
		return nil
	}
	return &RSAPrivKey{
		Key: key,
		Bytes: rawbyte,
		KeyLen: keylen / 8,
	}
}

type Hub struct {
	Proc       int
	HubKeyTag  string
	OnePerIP   bool

	Pool       *Pool
	IKeys      map[string]*RSAPrivKey
	AKeys      map[string][]byte
	CTags      map[string]bool
}

func NewHub() (*Hub) {
	h := &Hub{
		HubKeyTag: initKeyTag,
		Proc: 1,
		OnePerIP: true,
		Pool: NewPool(),
		IKeys: make(map[string]*RSAPrivKey),
		AKeys: make(map[string][]byte),
		CTags: make(map[string]bool),
	}

	h.CTags[clientAgentTag] = true

	return h
}

func NewHubM() (*Hub) {
	h := &Hub{
		HubKeyTag: initKeyTag,
		Proc: runtime.NumCPU(),
		Pool: NewPool(),
		OnePerIP: true,
		IKeys: make(map[string]*RSAPrivKey),
		AKeys: make(map[string][]byte),
		CTags: make(map[string]bool),
	}

	h.CTags[clientAgentTag] = true

	return h
}

func (h *Hub) DefIKey(keylen int, keybyte []byte) {
	h.IKeys[initKeyTag] = newRSAPrivKeyBase64(keylen, keybyte)
}
func (h *Hub) AddIKey(tag string, keylen int, keybyte []byte) {
	h.IKeys[tag] = newRSAPrivKeyBase64(keylen, keybyte)
}

func (h *Hub) DefAKey(key []byte) {
	h.AKeys[adminAgentTag] = key
}
func (h *Hub) AddAKey(tag string, key []byte) {
	h.AKeys[tag] = key
}

func (h *Hub) SetCTag(tag string) {
	h.CTags[tag] = true
}
func (h *Hub) DelCTag(tag string) {
	delete(h.CTags, tag)
}

func (h *Hub) HandleClient(p1 net.Conn) {
	defer func(){
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

	// client
	cok, ok := h.CTags[agent]
	if ok && cok {
		// get UUID
		UUID, _ := kit.ReadTagByte(enccon)

		// stream multiplex
		smuxConfig := smux.DefaultConfig()
		mux, err := smux.Server(enccon, smuxConfig)
		if err != nil {
			Vln(3, "mux init err", err)
			return
		}

		// add to pool
		uuidStr := kit.Hex(UUID)
		addr := p1.RemoteAddr().String()
		peer := NewPeer(p1, mux, UUID)

		if h.OnePerIP {
			// TODO: get old clients & send signal
		}

		id, ok := h.Pool.AddPear(peer)
		if !ok {
			Vf(2, "[client][new][same ID & Addr]%v %v %v %v %v\n", id, uuidStr, addr, agent, peer)
			return
		}
		Vf(2, "[client][new]%v %v %v %v %v\n", id, uuidStr, addr, agent, peer)

		// hack for OnClose
		for {
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

		// stream multiplex
		smuxConfig := smux.DefaultConfig()
		mux, err := smux.Server(enccon, smuxConfig)
		if err != nil {
			Vln(3, "mux init err", err)
			return
		}

		h.doREPL(mux)
	}

	return
}

func (h *Hub) doREPL(mux *smux.Session) {
	Vln(3, "[admin]connect start")
	for {
		p1, err := mux.AcceptStream()
		if err != nil {
			mux.Close()
			break
		}

		go h.doOP(p1)
	}
	Vln(3, "[admin]connect end")
}

func (h *Hub) doOP(p1 net.Conn) {
	// read OP
	op, err := kit.ReadTagStr(p1)
	if err != nil {
		Vln(3, "Read OP err:", err)
		p1.Close()
	}
	defer p1.Close()

	switch op {
	case H_ls:
		// list all
		by, err := kit.ReadTagStr(p1)
		if err != nil {
			return
		}
		var list []string
		switch by {
		case "addr":
			list = h.Pool.GetListByAddr()

		case "time":
			list = h.Pool.GetListByTime()

		case "id":
			fallthrough
		default:
			list = h.Pool.GetListByID()
		}
		count := len(list)
		kit.WriteVLen(p1, int64(count))
		for _, v := range list {
			kit.WriteTagStr(p1, v)
		}

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

