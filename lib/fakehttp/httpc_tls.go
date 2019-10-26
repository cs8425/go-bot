// +build !notls

package fakehttp

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"time"
	"strings"
)

type dialTLS struct {
	Transport     *http.Transport
	TLSConfig     *tls.Config
}

func (dl *dialTLS) GetProto() (string) {
	return "https://"
}

func (dl *dialTLS) Do(req *http.Request, timeout time.Duration) (*http.Response, error) {
	client := &http.Client{
		Timeout: timeout,
	}
	client.Transport = dl.Transport
	return client.Do(req)
}

func (dl *dialTLS) DialTimeout(host string, timeout time.Duration) (net.Conn, error) {
	tx, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		return nil, err
	}
	tx = tls.Client(tx, dl.TLSConfig)
	return tx, nil
}

func NewTLSClient(target string, caCrtByte []byte, skipVerify bool) (*Client) {
	cl := NewClient(target)

	colonPos := strings.LastIndex(target, ":")
	if colonPos == -1 {
		colonPos = len(target)
	}
	hostname := target[:colonPos]

	var caCrtPool *x509.CertPool
	if caCrtByte != nil {
		caCrtPool = x509.NewCertPool()
		caCrtPool.AppendCertsFromPEM(caCrtByte)
	}

	TLSConfig := &tls.Config{
		RootCAs: caCrtPool,
		InsecureSkipVerify: skipVerify,
		ServerName: hostname,
	}

	Transport := &http.Transport{
		TLSClientConfig: TLSConfig,
	}

	cl.Dialer = &dialTLS{
		TLSConfig: TLSConfig,
		Transport: Transport,
	}

	return cl
}

