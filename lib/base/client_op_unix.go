// +build !windows

package base

import (
	"net"
	//"os"
	"os/exec"
	"syscall"

	kit "local/toolkit"
	"lib/smux"
)

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

func init() {
	RegOps(B_shk, sh)
	RegOps(B_csh, sh)
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

