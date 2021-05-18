package main

import (
	"crypto/sha256"
	"crypto/sha512"
	"fmt"

	"time"
	"flag"

	"encoding/base64"
	"encoding/json"
	"os"

	"local/toolkit"
)

var (
	verbosity = flag.Int("v", 3, "verbosity")

	genMaster = flag.Bool("m", false, "generate Master Key")

	clientJson = flag.String("o1", "", "bot.json")
	hubJson = flag.String("o2", "", "hub.json")
	adminJson = flag.String("o3", "", "admin.json")
)

type ClientConfig struct {
	Name      string `json:"name,omitempty"`
	HubAddr   string `json:"addr,omitempty"`
	HubPubKey []byte `json:"hubkey,omitempty"` // RSA public key for connect to hub
	MasterKey []byte `json:"masterkey,omitempty"` // ECDSA public key for access
	UserAgent      string `json:"useragent,omitempty"`
}

type HubConfig struct {
	HubPrivKey  []byte `json:"hubkey,omitempty"` // RSA private key for client check
	AdmPubKey   []byte `json:"admkey,omitempty"` // ECDSA public key for admin check
	HubPrivKeys map[string][]byte `json:"hubkeys,omitempty"` // RSA private key for client check
	AdmPubKeys  map[string][]byte `json:"admkeys,omitempty"` // ECDSA public key for admin check

	BindAddr     string `json:"bind,omitempty"` // raw, http, ws (https/wss by key/crt)
	OnlyWs       bool   `json:"onlyws,omitempty"`
	WwwRoot      string `json:"www,omitempty"` // web/file server root dir
	TokenCookieA string `json:"ca,omitempty"` // token cookie name A
	TokenCookieB string `json:"cb,omitempty"` // token cookie name B
	TokenCookieC string `json:"cc,omitempty"` // token cookie name C
	CrtFile      string `json:"crt,omitempty"` // PEM encoded certificate file
	KeyFile      string `json:"key,omitempty"` // PEM encoded private key file
}

type AdminConfig struct {
	HubAddr      string `json:"addr,omitempty"`
	HubPubKey    []byte `json:"hubkey,omitempty"` // RSA public key for connect to hub
	AdmPrivKey   []byte `json:"admkey,omitempty"` // ECDSA private key for access hub
	MasterKey    []byte `json:"masterkey,omitempty"` // ECDSA private key for access bot
	UserAgent    string `json:"useragent,omitempty"`
	TokenCookieA string `json:"ca,omitempty"` // token cookie name A
	TokenCookieB string `json:"cb,omitempty"` // token cookie name B
	TokenCookieC string `json:"cc,omitempty"` // token cookie name C
}

func main() {
	flag.Parse()

	start := time.Now()

	private_key, public_key := toolkit.GenRSAKeys(2048)

	Vlogf(3, "RSA GerRSAKeys took %s\n", time.Since(start))

	Vlogln(2, "RSA private key:", len(private_key), base64.StdEncoding.EncodeToString(private_key))
	Vlogln(2, "RSA public key:", len(public_key), base64.StdEncoding.EncodeToString(public_key))


	fmt.Println()


	start = time.Now()
	private_ECDSA, public_ECDSA := toolkit.GenECDSAKeys()
	Vlogf(3, "GenECDSAKeys took %s\n", time.Since(start))

	Vlogln(2, "ECDSA private key:", len(private_key), base64.StdEncoding.EncodeToString(private_ECDSA))
	Vlogln(2, "ECDSA public key:", len(public_key), base64.StdEncoding.EncodeToString(public_ECDSA))

	var private_ECDSA2, public_ECDSA2 []byte
	if *genMaster {
		start = time.Now()
		private_ECDSA2, public_ECDSA2 = toolkit.GenECDSAKeys()
		Vlogf(3, "GenECDSAKeys took %s\n", time.Since(start))

		Vlogln(2, "ECDSA private key2:", len(private_key), base64.StdEncoding.EncodeToString(private_ECDSA2))
		Vlogln(2, "ECDSA public key2:", len(public_key), base64.StdEncoding.EncodeToString(public_ECDSA2))
	}

	if *clientJson != "" {
		cli := &ClientConfig{
			Name: fmt.Sprintf("client-%v", time.Now().Format(time.RFC3339)),
			HubAddr: "wss://127.0.0.1:8787",
			HubPubKey: public_key,
		}
		if *genMaster {
			cli.MasterKey = public_ECDSA2
		}

		json, err := json.MarshalIndent(cli, "", "\t")
		if err != nil {
			Vlogln(2, "Marshal client json error", err)
		}
		if err := os.WriteFile(*clientJson, json, 0600); err != nil {
			Vlogln(2, "wrtie client json error", *clientJson, err)
		}
		Vlogln(2, "client config.json:", *clientJson)
	}

	if *hubJson != "" {
		hub := &HubConfig{
			BindAddr: "wss://:8787",
			HubPrivKey: private_key,
			AdmPubKey: public_ECDSA,
			WwwRoot: "./www",
		}

		json, err := json.MarshalIndent(hub, "", "\t")
		if err != nil {
			Vlogln(2, "Marshal client json error", err)
		}
		if err := os.WriteFile(*hubJson, json, 0600); err != nil {
			Vlogln(2, "wrtie client json error", *hubJson, err)
		}
		Vlogln(2, "hub config.json:", *clientJson)
	}

	if *adminJson != "" {
		adm := &AdminConfig{
			HubAddr: "wss://127.0.0.1:8787",
			HubPubKey: public_key,
			AdmPrivKey: private_ECDSA,
		}
		if *genMaster {
			adm.MasterKey = private_ECDSA2
		}

		json, err := json.MarshalIndent(adm, "", "\t")
		if err != nil {
			Vlogln(2, "Marshal client json error", err)
		}
		if err := os.WriteFile(*adminJson, json, 0600); err != nil {
			Vlogln(2, "wrtie client json error", *adminJson, err)
		}
		Vlogln(2, "admin config.json:", *clientJson)
	}
}

func hashBytes(a []byte) []byte {
	sha1h := sha256.New()
	sha1h.Write(a)
	return sha1h.Sum([]byte(""))
}

func hashBytes512(a []byte) []byte {
	sha1h := sha512.New()
	sha1h.Write(a)
	return sha1h.Sum([]byte(""))
}

func Vlogf(level int, format string, v ...interface{}) {
	if level <= *verbosity {
		fmt.Printf(format, v...)
	}
}
func Vlog(level int, v ...interface{}) {
	if level <= *verbosity {
		fmt.Print(v...)
	}
}
func Vlogln(level int, v ...interface{}) {
	if level <= *verbosity {
		fmt.Println(v...)
	}
}

