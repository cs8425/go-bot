package base

import (
	"io"
	"net"
	//	"os"
	"os/exec"
	"sync"
	//	"time"
	"crypto/rand"
	"runtime"

	"lib/smux"
	"local/streamcoder"
	kit "local/toolkit"
)

var ops = make(map[string](func(string, net.Conn, *Client, *smux.Session)))
var inits = make([](func(*Client)), 0)

type Client struct {
	UUID      []byte
	cmd       *exec.Cmd
	cmdOut    io.ReadCloser
	cmdIn     io.WriteCloser
	cmdMx     sync.Mutex
	Proc      int
	AgentTag  string
	HubPubKey []byte
	HubKeyTag string
	Daemon    bool
	AutoClean bool
	Info      *Info
	MasterKey []byte

	Dial func(addr string) (net.Conn, error)

	binMx     sync.Mutex
	selfbyte  []byte
	selfhex   []byte
	selfbyte1 []byte
	selfhex1  []byte
}

func NewClient() *Client {
	return &Client{
		AgentTag:  clientAgentTag,
		HubKeyTag: initKeyTag,
		Proc:      1,
		Info:      NewInfo(),
		Daemon:    false,
	}
}

func NewClientM() *Client {
	return &Client{
		AgentTag:  clientAgentTag,
		HubKeyTag: initKeyTag,
		Proc:      runtime.NumCPU(),
		Info:      NewInfo(),
		Daemon:    false,
	}
}

var RegOps = func(tag string, fn func(string, net.Conn, *Client, *smux.Session)) {
	if ops == nil {
		ops = make(map[string](func(string, net.Conn, *Client, *smux.Session)))
	}

	if fn == nil {
		delete(ops, tag)
	} else {
		ops[tag] = fn
	}
}

var RegInit = func(fn func(*Client)) {
	if inits == nil {
		inits = make([](func(*Client)), 0)
	}

	inits = append(inits, fn)
}

func (c *Client) Start(addr string) {

	for _, f := range inits {
		f(c)
	}

	runtime.GOMAXPROCS(c.Proc)

	createConn := func() (session *smux.Session, err error) {

		var conn net.Conn

		if c.Dial == nil {
			conn, err = net.Dial("tcp", addr)
		} else {
			conn, err = c.Dial(addr)
		}
		if err != nil {
			return
		}

		// do handshake
		encKey := make([]byte, 88, 88)
		rand.Read(encKey)

		//Vln(6, "encKey = ", encKey)

		publicKey, _ := kit.ParseRSAPub(c.HubPubKey)
		ciphertext, err := kit.EncRSA(publicKey, encKey)
		if err != nil {
			return
		}

		kit.WriteTagStr(conn, c.HubKeyTag)
		conn.Write(ciphertext)

		// do encode
		// key = 32 bytes x 2
		// nonce = 12 bytes x 2
		enccon, _ := streamcoder.NewCoder(conn, encKey[0:64], encKey[64:88], true)

		// read nonce && rekey
		nonce, err := kit.ReadTagByte(enccon)
		if err != nil {
			return
		}
		pass := append(encKey[0:64], nonce...)
		enccon.ReKey(pass)

		// send agent
		kit.WriteTagStr(enccon, c.AgentTag)
		kit.WriteTagByte(enccon, c.UUID)

		// stream multiplex
		smuxConfig := smux.DefaultConfig()
		session, err = smux.Client(enccon, smuxConfig)
		if err != nil {
			return
		}

		//Vln(2, "connect to:", conn.RemoteAddr())

		return session, nil
	}

	// wait until a connection is ready
	waitConn := func() *smux.Session {
		for {
			if session, err := createConn(); err == nil {
				return session
			} else {
				kit.SleepRand()
			}
		}
	}

	for {
		mux := waitConn()
		for {
			p1, err := mux.AcceptStream()
			if err != nil {
				mux.Close()
				break
			}

			go c.handle1(p1, mux)
		}
		//Vln(2, "connect end")
		kit.SleepRand()
	}
}

func (c *Client) handle1(p1 net.Conn, mux *smux.Session) {

	if c.MasterKey != nil {
		pass := make([]byte, 32, 32)
		rand.Read(pass)
		kit.WriteTagByte(p1, pass)

		// check Signature
		signature, err := kit.ReadTagByte(p1)
		if err != nil {
			Vln(5, "can not read master signature!")
			return
		}
		hashed := kit.HashBytes256(pass)
		ok := kit.VerifyECDSA(c.MasterKey, hashed, signature)
		if !ok {
			Vln(5, "master key Verify error!")
			return
		}
		// ret
		kit.WriteVLen(p1, int64(0))
	}

	// get mode
	mode, err := kit.ReadTagStr(p1)
	if err != nil {
		kit.TrollConn(p1)
	}
	//Vln(3, "Mode:", mode)
	defer p1.Close()

	fn, ok := ops[mode]
	if !ok {
		kit.WriteVLen(p1, int64(9))
		//kit.TrollConn(p1)
		return
	}

	kit.WriteVLen(p1, int64(0))
	fn(mode, p1, c, mux)
}
