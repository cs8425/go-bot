// go build hub.go share.go

package main

import (
	"flag"
	"log"
	"net"
	"os"
	"encoding/base64"

	"./lib/base"
)

var bind = flag.String("l", ":8787", "bind port")
var verb = flag.Int("v", 6, "verbosity")

var verbosity int = 6

func main() {
	std.SetFlags(log.LstdFlags | log.Lmicroseconds)
	flag.Parse()
	verbosity = *verb

	lis, err := net.Listen("tcp", *bind)
	if err != nil {
		Vln(2, "Error listening:", err.Error())
		os.Exit(1)
	}
	defer lis.Close()


	Vln(2, "listening on:", lis.Addr())
	Vln(2, "verbosity:", verbosity)

	ikey, _ := base64.StdEncoding.DecodeString("MIIEowIBAAKCAQEArogYEOHItjtm0wJOX+hSHjGTIPUsRo/TyLGYxWVk79pNWAhCSvH9nfvpx0skefcL/Nd++Qb/zb3c+o7ZI4zbMKZJLim3yaN8IDlgrjKG7wmjB5r49++LrvRzjIJCAoeFog2PfEn3qlQ+PA26TqLsbPNZi9nsaHlwTOqGljg82g23Zqj1o5JfitJvVlRLmhPqc8kO+4Dvf08MdVS6vBGZjzWFmGx9k3rrDoi7tem22MflFnOQhgLJ4/sbd4Y71ok98ChrQhb6SzZKVWN5v7VCuKqhFLmhZuK0z0f/xkBNcMeCplVLhs/gLIU3HBmvbBSYhmN4dDL19cAv1MkQ6lb1dwIDAQABAoIBAQCXlhqY5xGlvTgUg0dBI43XLaWlFWyMKLV/9UhEAknFzOwqTpoNb9qgUcD9WHVo/TpLM3vTnNGmh4YblOBhcSCbQ4IB9zfqiPTxJASlp7rseIlBvMcKyOKgZS7K1gOxILXfRzndcH0MUjjvfdjYHceM5VtcDT24i+kO1Q9p/5RSqfGu9wz56tqEQE4Z1OTzD+dD9tGeciiyZ9qDoDC/tb0oBKSFK+DlZZOrSBSpGk2Qur4BgVAgL3wunATzGpxxaCAf+9lBEUBCrZbUkeQIKoFbvjqee5Fb2tfdqquMG1FX3CuCovsW7aMKjpAK5TsKuZD88EWje42JV6wmJ/Q4nGvBAoGBAMs6Hs/UX60uZ10mTVKoHU/Mm6lr/FBDo4LF165SX/+sH87KbNlmOO9YBZGJBm1AnsxaNYLjT39EiGlZZbCYRwre/D/9z+hY9J0Yhz/eo8fGsee3f7SU8U9kRH0CFn5MI8Wf7YgNH97uky9i41rqYtkxf2GvqMYl5yzVpQk3fu0XAoGBANvaZQs9DuwFwekzncFcejLHv2CQEDDqtEybmh5PB9YHN+RyHRlxPmYC1d1ElvHO65Tfhgcd0fL0EkSHCXFHfmsIcpSHuUlBpFSrI6btygf+U/U8VLwzXI71cpoE5n+E7rR0J5hTvTo/FccdilV/CubgIZbQ6VSaAxw4HBA5JzahAn9Q+NdN91AnsFV+x8QHKvSC1wMufdgKIukDMdC9pBSbyfjia8Ty2cfVlTyiv/XPke+zfD3V6LvD+Ypgbz4VHpcvvajD1l0ANnFAJoW87PhUoNZBfNtlF/MNruWa6ToNGEkodJAvpQsNyADc4Im1r62y3AXk5hhY2sFBG96lzXbFAoGBAKhoBUhzj++ZhWz13dyU0wH84gq8r7pYvp2D/61BynXW96hlBQdNKIgJmfqxJJK7dteF1Ou0mvLopOmbKs97/UlNoj9GK9cCkjdNFLU0prIyzesnOJ0lFrxnJU73e/yoPhU6eG4FjwiD9FGevi05cIdjnjchdeoZQ1KlZFHFBdWhAoGBAMrwhd20ww6/VrVQShLVB0P3Zn3aKUqUvU9si616iyNSpuZ9dstXYNYAbPav02PL0NOPMDHC6/SERbJQQCnnBqbDBwmUHVmr0W3rvD+DUgihpgTTxArb0FfguJQlKN6whlHOLrf6sC1YebQWhFvPTNQqfSjfO9/g37usDNcskguf")
	akey, _ := base64.StdEncoding.DecodeString("MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEc8tgqZX82zrd6809YWzJkh4zhvaoCEbkU8yxBW+a9U1L+XgItJGRL33vYecv4lH9ovSNgJiyvnqdmqkJtwq52Q==")

	hub := base.NewHubM()
	hub.DefIKey(2048, ikey)
	hub.DefAKey(akey)
	hub.OnePerIP = false

	var srv net.Listener
	srv = lis

	for {
		if conn, err := srv.Accept(); err == nil {
			//Vln(2, "remote address:", conn.RemoteAddr())

			go hub.HandleClient(conn)
		} else {
			Vln(2, "Accept err:", err)
		}
	}

}


