package main

import (
	"crypto/sha256"
	"crypto/sha512"
	"fmt"

	"time"
	"flag"

	"encoding/base64"

	"./lib/toolkit"
)

var verbosity = flag.Int("v", 3, "verbosity")

func main() {

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

