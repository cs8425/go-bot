package base

import (
	"net"
	//"os"

	"lib/smux"
	kit "local/toolkit"
)

func init() {
	RegOps(B_bind, bindSrv)
}

var bindSrv = func(op string, p1 net.Conn, c *Client, mux *smux.Session) {
	// get port
	addr, err := kit.ReadTagStr(p1)
	if err != nil {
		Vln(3, "[bind]Error get binding addr:", err)
		return
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		Vln(3, "[bind]Error listening:", err)
		return
	}
	defer lis.Close()

	// ret stats & addr
	kit.WriteVLen(p1, int64(0))
	kit.WriteTagStr(p1, lis.Addr().String())

	// stream multiplex
	smuxConfig := smux.DefaultConfig()
	mux2, err := smux.Server(p1, smuxConfig)
	if err != nil {
		Vln(3, "[bind]mux init err", err)
		return
	}

	go bindWroker(lis, mux2)

	// hack for OnClose
	for {
		_, err := mux2.AcceptStream()
		if err != nil {
			mux2.Close()
			p1.Close()
			Vln(3, "[bind][cls]", addr)
			break
		}
	}
}

var bindWroker = func(lis net.Listener, mux *smux.Session) {
	for {
		if conn, err := lis.Accept(); err == nil {
			//Vln(2, "[bind][new]", conn.RemoteAddr())
			go bindHandle(conn, mux)
		} else {
			Vln(3, "[bind]Accept err", err)
			return
		}
	}
}

var bindHandle = func(p1 net.Conn, mux *smux.Session) {
	defer p1.Close()

	conn, err := mux.OpenStream()
	if err != nil {
		return
	}
	defer conn.Close()

	kit.Cp(conn, p1)
}
