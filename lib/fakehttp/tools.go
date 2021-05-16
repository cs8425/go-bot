package fakehttp

import (
	"bytes"
	"math/rand"
	"net"
	"io"
	"log"
	"time"
)

const verbosity int = 0

type Conn struct {
	R     io.ReadCloser
	W     net.Conn //io.WriteCloser
}
func (c Conn) Read(data []byte) (n int, err error)  { return c.R.Read(data) }
func (c Conn) Write(data []byte) (n int, err error) { return c.W.Write(data) }

func (c Conn) Close() error {
	if err := c.W.Close(); err != nil {
		return err
	}
	if err := c.R.Close(); err != nil {
		return err
	}
	return nil
}

func (c Conn) LocalAddr() net.Addr {
	if ts, ok := c.W.(interface {
		LocalAddr() net.Addr
	}); ok {
		return ts.LocalAddr()
	}
	return nil
}

func (c Conn) RemoteAddr() net.Addr {
	if ts, ok := c.W.(interface {
		RemoteAddr() net.Addr
	}); ok {
		return ts.RemoteAddr()
	}
	return nil
}

func (c Conn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c Conn) SetWriteDeadline(t time.Time) error {
	return c.W.SetWriteDeadline(t)
}

func (c Conn) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	if err := c.SetWriteDeadline(t); err != nil {
		return err
	}
	return nil
}

type CloseableReader struct {
	io.Reader
	r0     io.ReadCloser
}
func (c CloseableReader) Close() error {
	return c.r0.Close()
}

func mkconn(p1 net.Conn, p2 net.Conn, rbuf []byte) (net.Conn){
	rem := bytes.NewReader(rbuf)
	r := io.MultiReader(rem, p1)
	rc := CloseableReader{ r, p1 }

	pipe := Conn {
		R: rc,
		W: p2,
	}
	return pipe
}

type ConnAddr struct {
	net.Conn //io.WriteCloser
	Addr string
}
func (c *ConnAddr) RemoteAddr() net.Addr {
	return (*StrAddr)(c)
}

type StrAddr ConnAddr
func (c *StrAddr) Network() string {
	return c.Conn.RemoteAddr().Network()
}
func (c *StrAddr) String() string {
	if c == nil {
		return "<nil>"
	}
	if c.Addr == "" {
		return c.Conn.RemoteAddr().String()
	}
	return c.Addr
}

func mkConnAddr(p1 net.Conn, address string) (net.Conn) {
	if address != "" {
		conn := &ConnAddr{
			Conn: p1,
			Addr: address,
		}
		return conn
	}
	return p1
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/-_"
func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func Vlogf(level int, format string, v ...interface{}) {
	if level <= verbosity {
		log.Printf(format, v...)
	}
}
func Vlog(level int, v ...interface{}) {
	if level <= verbosity {
		log.Print(v...)
	}
}
func Vlogln(level int, v ...interface{}) {
	if level <= verbosity {
		log.Println(v...)
	}
}


