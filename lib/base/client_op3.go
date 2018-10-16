// +build mod

package base

import (
	"net"
	"os"
	"syscall"

//	kit "../toolkit"
	"../smux"
)

func init() {
	RegOps(B_ppend, ppX)
	RegOps(B_ppkill, ppX)
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


