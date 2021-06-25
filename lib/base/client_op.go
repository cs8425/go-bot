package base

import (
	"net"
	"os"

	//kit "local/toolkit"
	"lib/smux"
)

func init() {
	RegOps(B_info, pullInfo)

	RegOps(B_fast0, fastC)
	RegOps(B_fast1, fastC)
	RegOps(B_fast2, fastC)

	RegOps(B_mux, muxC)

	RegOps(B_reconn, cc1)
	RegOps(B_kill, cc1)
}

var pullInfo = func(op string, p1 net.Conn, c *Client, mux *smux.Session) {
	c.Info.WriteTo(p1)
}

var fastC = func(op string, p1 net.Conn, c *Client, mux *smux.Session) {
	handleFastS(p1)
}

var cc1 = func(op string, p1 net.Conn, c *Client, mux *smux.Session) {
	switch op {
	case B_reconn:
		mux.Close()

	case B_kill:
		os.Exit(0)
	}
}

var muxC = func(op string, p1 net.Conn, c *Client, mux *smux.Session) {
	smuxConfig := smux.DefaultConfig()
	mux2, err := smux.Client(p1, smuxConfig) // client here
	if err != nil {
		return
	}

	for {
		conn, err := mux2.AcceptStream()
		if err != nil {
			mux2.Close()
			return
		}

		go c.handle1(conn, mux2, true)
	}
}
