package base

import (
	"net"
	"os"

	kit "local/toolkit"
	"lib/smux"
)

func init() {
	RegOps(B_info, pullInfo)

	RegOps(B_fast0, fastC)
	RegOps(B_fast1, fastC)
	RegOps(B_fast2, fastC)

	RegOps(B_shk, sh)
	RegOps(B_csh, sh)

	RegOps(B_reconn, cc1)
	RegOps(B_kill, cc1)
}

var pullInfo = func (op string, p1 net.Conn, c *Client, mux *smux.Session) {
	c.Info.WriteTo(p1)
}

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

var cc1 = func (op string, p1 net.Conn, c *Client, mux *smux.Session) {
	switch op {
	case B_reconn:
		mux.Close()

	case B_kill:
		os.Exit(0)
	}
}



