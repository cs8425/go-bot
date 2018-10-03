package toolkit

import (
	"io"
	"net"
	"sync"
	"time"
	"math/rand"

)

const (
	disconnectMin = 15 * 1000
	disconnectRange = 30 * 1000

	copyPipeSize = 2048
)

// global recycle buffer
var copyBuf sync.Pool

func init() {
	rand.Seed(int64(time.Now().Nanosecond()))

	copyBuf.New = func() interface{} {
		return make([]byte, copyPipeSize)
	}
}

var TrollConn = func (p1 net.Conn) {
	sec := time.Duration(disconnectMin + rand.Intn(disconnectRange)) * time.Millisecond
	time.Sleep(sec)
	p1.Close()
}

var SleepRand = func () {
	sec := time.Duration(disconnectMin + rand.Intn(disconnectRange)) * time.Millisecond
	time.Sleep(sec)
}

var Cp = func (p1, p2 net.Conn) {
	// start tunnel
	p1die := make(chan struct{})
	go func() {
		buf := copyBuf.Get().([]byte)
		io.CopyBuffer(p1, p2, buf)
		close(p1die)
		copyBuf.Put(buf)
	}()

	p2die := make(chan struct{})
	go func() {
		buf := copyBuf.Get().([]byte)
		io.CopyBuffer(p2, p1, buf)
		close(p2die)
		copyBuf.Put(buf)
	}()

	// wait for tunnel termination
	select {
	case <-p1die:
	case <-p2die:
	}
}

var Cp1 = func (p1 io.Reader, p2 io.Writer) {
	buf := copyBuf.Get().([]byte)
	io.CopyBuffer(p2, p1, buf) // p2 << p1
	copyBuf.Put(buf)
}

// p1 >> p0 >> p2
var Cp3 = func (p1 io.Reader, p0 net.Conn, p2 io.Writer) {
	p1die := make(chan struct{})
	go func() {
		buf := copyBuf.Get().([]byte)
		io.CopyBuffer(p0, p1, buf) // p0 << p1
		close(p1die)
		copyBuf.Put(buf)
	}()

	p2die := make(chan struct{})
	go func() {
		buf := copyBuf.Get().([]byte)
		io.CopyBuffer(p2, p0, buf) // p2 << p0
		close(p2die)
		copyBuf.Put(buf)
	}()

	// wait for tunnel termination
	select {
	case <-p1die:
	case <-p2die:
	}
}


