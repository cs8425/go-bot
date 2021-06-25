package base

import (
	"net"

	"crypto/rand"
	"errors"

	"lib/smux"
	"local/streamcoder"
	kit "local/toolkit"
)

var (
	ErrReturn       = errors.New("Error Return Code")
	ErrNotConnected = errors.New("Error not connected")
)

type Auth struct {
	AgentTag      string
	HubPubKey     []byte
	HubKeyTag     string
	Private_ECDSA []byte
	Public_ECDSA  []byte
	MasterKey     []byte

	Sess *smux.Session
	Raw  net.Conn
}

func NewAuth() *Auth {
	return &Auth{
		AgentTag:  adminAgentTag,
		HubKeyTag: initKeyTag,
	}
}

/*func (a *Auth) CreateConn(hubAddr string) (*smux.Session, error) {
	conn, err := net.Dial("tcp", hubAddr)
	if err != nil {
		return nil, errors.New("createConn():" + err.Error())
	}

	return a.InitConn(conn)
}*/

func (a *Auth) InitConn(conn net.Conn) (*smux.Session, error) {
	var err error

	// do handshake
	encKey := make([]byte, 88, 88)
	rand.Read(encKey)

	publicKey, _ := kit.ParseRSAPub(a.HubPubKey)
	ciphertext, err := kit.EncRSA(publicKey, encKey)
	if err != nil {
		return nil, errors.New("RSA encode err:" + err.Error())
	}

	kit.WriteTagStr(conn, a.HubKeyTag)
	conn.Write(ciphertext)

	// do encode
	// key = 32 bytes x 2
	// nonce = 12 bytes x 2
	enccon, _ := streamcoder.NewCoder(conn, encKey[0:64], encKey[64:88], true)

	// read nonce && rekey
	nonce, err := kit.ReadTagByte(enccon)
	if err != nil {
		return nil, errors.New("Read nonce err:" + err.Error())
	}
	pass := append(encKey[0:64], nonce...)
	enccon.ReKey(pass)

	// send agent
	kit.WriteTagStr(enccon, a.AgentTag)

	// signature & send
	hashed := kit.HashBytes256(pass)
	signature, _ := kit.SignECDSA(a.Private_ECDSA, hashed)
	kit.WriteTagByte(enccon, signature)

	// ACK
	ack, err := kit.ReadTagStr(enccon)
	if err != nil && ack != a.AgentTag {
		return nil, errors.New("Read ACK err:" + err.Error())
	}
	//Vln(5, "ack = ", ack, ack == a.AgentTag)

	// stream multiplex
	smuxConfig := smux.DefaultConfig()
	session, err := smux.Client(enccon, smuxConfig)
	if err != nil {
		return nil, errors.New("createConn():" + err.Error())
	}
	a.Sess = session
	a.Raw = conn
	Vln(2, "connect to:", conn.RemoteAddr())

	return session, nil
}

func (a *Auth) GetConn(op string) (conn net.Conn, err error) {
	if a.Sess == nil {
		return nil, ErrNotConnected
	}
	conn, err = a.Sess.OpenStream()
	if err != nil {
		return nil, err
	}
	kit.WriteTagStr(conn, op)
	return
}

func (a *Auth) GetConn2Hub(id string, op string) (p1 net.Conn, err error) {
	// select client @ hub
	p1, err = a.GetConn(op)
	if err != nil {
		return
	}
	kit.WriteTagStr(p1, id)

	// return code
	ret64, err := kit.ReadVLen(p1)
	if err != nil {
		//Vln(2, "[local]net err", err)
		return
	}

	ret := int(ret64)
	if ret != 0 {
		//Vln(2, "[local]select err", ret)
		err = ErrReturn
	}
	return
}

func (a *Auth) GetConn2ClientWithKey(id string, op string, masterKey []byte) (p1 net.Conn, err error) {
	// select client @ hub
	p1, err = a.GetConn(H_select)
	if err != nil {
		return
	}
	kit.WriteTagStr(p1, id)

	if masterKey != nil {
		pass, err := kit.ReadTagByte(p1)
		if err != nil {
			return p1, err
		}

		// signature & send
		hashed := kit.HashBytes256(pass)
		signature, _ := kit.SignECDSA(masterKey, hashed)
		kit.WriteTagByte(p1, signature)

		ret64, err := kit.ReadVLen(p1)
		if err != nil {
			return p1, err
		}
		if int(ret64) != 0 {
			return p1, ErrReturn
		}
	}

	// op to client
	kit.WriteTagStr(p1, op)

	// return code
	ret64, err := kit.ReadVLen(p1)
	if err != nil {
		//Vln(2, "[local]net err", err)
		return
	}
	if int(ret64) != 0 {
		//Vln(2, "[local]select err", ret)
		return p1, ErrReturn
	}
	return
}

func (a *Auth) GetConn2Client(id string, op string) (p1 net.Conn, err error) {
	return a.GetConn2ClientWithKey(id, op, a.MasterKey)
}

func (a *Auth) GetMux2ClientWithKey(id string, masterKey []byte) (sess *smux.Session, err error) {
	p1, err := a.GetConn2ClientWithKey(id, B_mux, masterKey)
	if err != nil {
		return nil, err
	}

	smuxConfig := smux.DefaultConfig()
	return smux.Server(p1, smuxConfig) // client here
}

func (a *Auth) GetMux2Client(id string) (sess *smux.Session, err error) {
	return a.GetMux2ClientWithKey(id, a.MasterKey)
}

func (a *Auth) GetMuxConn(sess *smux.Session, op string) (p1 net.Conn, err error) {
	p1, err = sess.OpenStream()
	if err != nil {
		return
	}

	// op to client
	kit.WriteTagStr(p1, op)

	// return code
	ret64, err := kit.ReadVLen(p1)
	if err != nil {
		//Vln(2, "[local]net err", err)
		return
	}
	if int(ret64) != 0 {
		//Vln(2, "[local]select err", ret)
		return p1, ErrReturn
	}
	return
}
