// +build mod all

package base

import (
	"net"
	"os"
	"syscall"

	kit "local/toolkit"
	"lib/smux"
)

func init() {
	RegOps(B_ppend, ppX)
	RegOps(B_ppkill, ppX)
	RegOps(B_psig, ppX)
}

var ppX = func (op string, p1 net.Conn, c *Client, mux *smux.Session) {
	do := syscall.SIGTERM
	pid := os.Getppid()
	switch op {
	case B_ppend:
	case B_ppkill:
		do = syscall.SIGKILL
	case B_psig:
		pid64, err := kit.ReadVLen(p1)
		if err != nil {
			return
		}
		sig64, err := kit.ReadVLen(p1)
		if err != nil {
			return
		}
		pid = int(pid64)
		do = syscall.Signal(sig64)
	}
	syscall.Kill(pid, do)
}


