package base

import (
	"net"
	"strconv"
	"time"

	kit "../toolkit"

//	"fmt"
	"log"
)
/*
func Vf(level int, format string, v ...interface{}) { }
func V(level int, v ...interface{}) { }
func Vln(level int, v ...interface{}) { }
*/
func Vf(level int, format string, v ...interface{}) {
	if level <= 6 {
		log.Printf(format, v...)
	}
}
func V(level int, v ...interface{}) {
	if level <= 6 {
		log.Print(v...)
	}
}
func Vln(level int, v ...interface{}) {
	if level <= 6 {
		log.Println(v...)
	}
}

func replyAndClose(p1 net.Conn, rpy int) {
	p1.Write([]byte{0x05, byte(rpy), 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	p1.Close()
}

func handleFastS(p1 net.Conn) {
	var b [320]byte
	n, err := p1.Read(b[:])
	if err != nil {
		//Vln(1, "[fast client read]", p1, err)
		return
	}
	// b[0:2] // ignore

	var host, port, backend string
	switch b[3] {
	case 0x01: //IP V4
		host = net.IPv4(b[4], b[5], b[6], b[7]).String()
	case 0x03: //DOMAINNAME
		host = string(b[5 : n-2]) //b[4] domain name length
	case 0x04: //IP V6
		host = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
	case 0x05: //DOMAINNAME + PORT
		backend = string(b[4 : n])
		goto CONN
	default:
		replyAndClose(p1, 0x08) // X'08' Address type not supported
		return
	}
	port = strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
	backend = net.JoinHostPort(host, port)

CONN:
	p2, err := net.DialTimeout("tcp", backend, 10*time.Second)
	if err != nil {
		//Vln(2, "[err]", backend, err)

		switch t := err.(type) {
		case *net.AddrError:
			replyAndClose(p1, 0x03) // X'03' Network unreachable

		case *net.OpError:
			if t.Timeout() {
				replyAndClose(p1, 0x06) // X'06' TTL expired
			} else if t.Op == "dial" {
				replyAndClose(p1, 0x05) // X'05' Connection refused
			}

		default:
			//replyAndClose(p1, 0x03) // X'03' Network unreachable
			//replyAndClose(p1, 0x04) // X'04' Host unreachable
			replyAndClose(p1, 0x05) // X'05' Connection refused
			//replyAndClose(p1, 0x06) // X'06' TTL expired
		}
		return
	}
	defer p2.Close()

	//Vln(3, "[got]", backend)
	reply := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	p1.Write(reply) // reply OK

	kit.Cp(p1, p2)
	//Vln(3, "[cls]", backend)
}

// p1 = socks5 client, p2 = fast server
func HandleSocksF(p1, p2 net.Conn) {
	var b [320]byte
	n, err := p1.Read(b[:])
	if err != nil {
		//Vln(3, "socks client read", p1, err)
		return
	}
	if b[0] != 0x05 { //only Socket5
		return
	}

	//reply: NO AUTHENTICATION REQUIRED
	p1.Write([]byte{0x05, 0x00})

	n, err = p1.Read(b[:])
	if b[1] != 0x01 { // 0x01: CONNECT
		replyAndClose(p1, 0x07) // X'07' Command not supported
		return
	}

//	var backend string
	switch b[3] {
	case 0x01: //IP V4
//		backend = net.IPv4(b[4], b[5], b[6], b[7]).String()
		if n != 10 {
			replyAndClose(p1, 0x07) // X'07' Command not supported
			return
		}
	case 0x03: //DOMAINNAME
//		backend = string(b[5 : n-2]) //b[4] domain name length
	case 0x04: //IP V6
//		backend = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
		if n != 22 {
			replyAndClose(p1, 0x07) // X'07' Command not supported
			return
		}
	default:
		replyAndClose(p1, 0x08) // X'08' Address type not supported
		return
	}


	// send to proxy
	p2.Write(b[0:n])

	var b2 [10]byte
	n2, err := p2.Read(b2[:10])
	if n2 < 10 {
//		Vln(2, "Dial err replay:", backend, n2)
		replyAndClose(p1, 0x03)
		return
	}
	if err != nil || b2[1] != 0x00 {
//		Vln(2, "socks err to:", backend, n2, b2[1], err)
		replyAndClose(p1, int(b2[1]))
		return
	}

//	Vln(3, "[got]", backend, p1.RemoteAddr())
	reply := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	p1.Write(reply) // reply OK
	kit.Cp(p1, p2)
//	Vln(3, "[cls]", backend, p1.RemoteAddr())
}

