package fakehttp

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"io/ioutil"
	"time"
)

var (
	ErrNotServer       = errors.New("may not tunnel server")
	ErrTokenTimeout    = errors.New("token may timeout")
)

type NetDialer interface {
	GetProto() (string)
	Do(req *http.Request, timeout time.Duration) (*http.Response, error) // http.Client
	DialTimeout(host string, timeout time.Duration) (net.Conn, error) // net.DialTimeout("tcp", Host, Timeout)
}

type dialNonTLS Client
func (dl dialNonTLS) GetProto() (string) {
	return "http://"
}
func (dl dialNonTLS) Do(req *http.Request, timeout time.Duration) (*http.Response, error) {
	client := &http.Client{
		Timeout: timeout,
	}
	return client.Do(req)
}
func (dl dialNonTLS) DialTimeout(host string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", host, timeout)
}

type Client struct {
	TxMethod      string
	RxMethod      string
	TxFlag        string
	RxFlag        string
	TokenCookieA  string
	TokenCookieB  string
	TokenCookieC  string
	UserAgent     string
	Url           string
	Timeout       time.Duration
	Host          string
	UseWs         bool

	Dialer        NetDialer
}

func (cl *Client) getURL() (string) {
	url := cl.Host + cl.Url
	return cl.Dialer.GetProto() + url
}

func (cl *Client) getToken() (string, error) {
	req, err := http.NewRequest("GET", cl.getURL(), nil)
	if err != nil {
		Vlogln(2, "getToken() NewRequest err:", err)
		return "", err
	}

	req.Header.Set("User-Agent", cl.UserAgent)
	res, err := cl.Dialer.Do(req, cl.Timeout)
	if err != nil {
		Vlogln(2, "getToken() send Request err:", err)
		return "", err
	}
	defer res.Body.Close()

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		Vlogln(2, "getToken() ReadAll err:", err)
	}

	Vlogln(3, "getToken() http version:", res.Proto)

	return cl.checkToken(res)
}

func (cl *Client) checkToken(res *http.Response) (string, error) {
	cookies := res.Cookies()
	Vlogln(3, "checkToken()", cookies)

	for _, cookie := range cookies {
		Vlogln(4, "cookie:", cookie.Name, cookie.Value)
		if cookie.Name == cl.TokenCookieA {
			return cookie.Value, nil
		}
	}

	return  "", ErrNotServer
}

func (cl *Client) getTx(token string) (net.Conn, []byte, error) { //io.WriteCloser

	req, err := http.NewRequest(cl.TxMethod, cl.getURL(), nil)
	if err != nil {
		Vlogln(2, "getTx() NewRequest err:", err)
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "private, no-store, no-cache, max-age=0")
	req.Header.Set("User-Agent", cl.UserAgent)
	req.Header.Set("Cookie", cl.TokenCookieB + "=" + token + "; " + cl.TokenCookieC + "=" + cl.TxFlag)

	tx, err := cl.Dialer.DialTimeout(cl.Host, cl.Timeout)
	if err != nil {
		Vlogln(2, "Tx connect to:", cl.Host, err)
		return nil, nil, err
	}

	Vlogln(3, "Tx connect ok:", cl.Host)
	req.Write(tx)

	txbuf := bufio.NewReaderSize(tx, 1024)
//	Vlogln(2, "Tx Reader", txbuf)
	res, err := http.ReadResponse(txbuf, req)
	if err != nil {
		Vlogln(2, "Tx ReadResponse", err, res)
		tx.Close()
		return nil, nil, err
	}
	Vlogln(3, "Tx http version:", res.Proto)

	_, err = cl.checkToken(res)
	if err == nil {
		tx.Close()
		return nil, nil, ErrTokenTimeout
	}

	n := txbuf.Buffered()
	Vlogln(3, "Tx Response", n)

	return tx, nil, nil
}

func (cl *Client) getRx(token string) (net.Conn, []byte, error) { //io.ReadCloser

	req, err := http.NewRequest(cl.RxMethod, cl.getURL(), nil)
	if err != nil {
		Vlogln(2, "getRx() NewRequest err:", err)
		return nil, nil, err
	}

	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "private, no-store, no-cache, max-age=0")
	req.Header.Set("User-Agent", cl.UserAgent)
	req.Header.Set("Cookie", cl.TokenCookieB + "=" + token + "; " + cl.TokenCookieC + "=" + cl.RxFlag)


	rx, err := cl.Dialer.DialTimeout(cl.Host, cl.Timeout)
	if err != nil {
		Vlogln(2, "Rx connect to:", cl.Host, err)
		return nil, nil, err
	}
	Vlogln(3, "Rx connect ok:", cl.Host)
	req.Write(rx)

	rxbuf := bufio.NewReaderSize(rx, 1024)
//	Vlogln(2, "Rx Reader", rxbuf)
	res, err := http.ReadResponse(rxbuf, req)
	if err != nil {
		Vlogln(2, "Rx ReadResponse", err, res, rxbuf)
		rx.Close()
		return nil, nil, err
	}
	Vlogln(3, "Rx http version:", res.Proto)

	_, err = cl.checkToken(res)
	if err == nil {
		rx.Close()
		return nil, nil, ErrTokenTimeout
	}

	n := rxbuf.Buffered()
	Vlogln(3, "Rx Response", n)
	if n > 0 {
		buf := make([]byte, n)
		rxbuf.Read(buf[:n])
		return rx, buf[:n], nil
	} else {
		return rx, nil, nil
	}
}


func NewClient(target string) (*Client) {
	cl := &Client {
		TxMethod:     txMethod,
		RxMethod:     rxMethod,
		TxFlag:       txFlag,
		RxFlag:       rxFlag,
		TokenCookieA: tokenCookieA,
		TokenCookieB: tokenCookieB,
		TokenCookieC: tokenCookieC,
		UserAgent:    userAgent,
		Url:          targetUrl,
		Timeout:      timeout,
		Host:         target,
		UseWs:        false,
	}
	cl.Dialer = dialNonTLS(*cl)
	return cl
}

func Dial(target string) (net.Conn, error) {
	cl := NewClient(target)
	return cl.Dial()
}

func (cl *Client) Dial() (net.Conn, error) {
	token, err := cl.getToken()
	if token == "" || err != nil {
		return nil, err
	}
	Vlogln(2, "token:", token)

	if cl.UseWs {
		return cl.dialWs(token)
	}

	return cl.dialNonWs(token)
}

func (cl *Client) dialWs(token string) (net.Conn, error) {
	req, err := http.NewRequest(cl.RxMethod, cl.getURL(), nil)
	if err != nil {
		Vlogln(2, "dialWs() NewRequest err:", err)
		return nil, err
	}

	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "private, no-store, no-cache, max-age=0")
	req.Header.Set("User-Agent", cl.UserAgent)
	req.Header.Set("Cookie", cl.TokenCookieB + "=" + token + "; " + cl.TokenCookieC + "=" + cl.RxFlag)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Key", token)
	req.Header.Set("Sec-WebSocket-Version", "13")

	rx, err := cl.Dialer.DialTimeout(cl.Host, cl.Timeout)
	if err != nil {
		Vlogln(2, "WS connect to:", cl.Host, err)
		return nil, err
	}
	Vlogln(3, "WS connect ok:", cl.Host)
	req.Write(rx)

	rxbuf := bufio.NewReaderSize(rx, 1024)
//	Vlogln(2, "Rx Reader", rxbuf)
	res, err := http.ReadResponse(rxbuf, req)
	if err != nil {
		Vlogln(2, "WS ReadResponse", err, res, rxbuf)
		rx.Close()
		return nil, err
	}
	Vlogln(3, "WS http version:", res.Proto)

	_, err = cl.checkToken(res)
	if err == nil {
		rx.Close()
		return nil, ErrTokenTimeout
	}

	n := rxbuf.Buffered()
	Vlogln(3, "WS Response", n)
	if n > 0 {
		buf := make([]byte, n)
		rxbuf.Read(buf[:n])
		return mkconn(rx, rx, buf[:n]), nil
	}

	return rx, nil
}

func (cl *Client) dialNonWs(token string) (net.Conn, error) {
	type ret struct {
		conn  net.Conn
		buf   []byte
		err   error
	}
	txRetCh := make(chan ret, 1)
	rxRetCh := make(chan ret, 1)

	go func () {
		tx, _, err := cl.getTx(token)
		Vlogln(4, "tx:", tx)
		txRetCh <- ret{tx, nil, err}
	}()
	go func () {
		rx, rxbuf, err := cl.getRx(token)
		Vlogln(4, "rx:", rx, rxbuf)
		rxRetCh <- ret{rx, rxbuf, err}
	}()

	txRet := <-txRetCh
	tx, _, txErr := txRet.conn, txRet.buf, txRet.err

	rxRet := <-rxRetCh
	rx, rxbuf, rxErr := rxRet.conn, rxRet.buf, rxRet.err

	if txErr != nil {
		if rx != nil { // close other side, no half open
			rx.Close()
		}
		return nil, txErr
	}

	if rxErr != nil {
		if tx != nil {
			tx.Close()
		}
		return nil, rxErr
	}

	return mkconn(rx, tx, rxbuf), nil
}


