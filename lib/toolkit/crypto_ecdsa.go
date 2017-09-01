package toolkit

import (
	"crypto/rand"
 	"crypto/ecdsa"
 	"crypto/elliptic"
	"crypto/x509"
 	"math/big"
)

// ECDSA function
func GenECDSAKeys() (private_key_bytes, public_key_bytes []byte) {
	private_key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	private_key_bytes, _ = x509.MarshalECPrivateKey(private_key)
	public_key_bytes, _ = x509.MarshalPKIXPublicKey(&private_key.PublicKey)
	return private_key_bytes, public_key_bytes
}

func SignECDSA(private_key_bytes []byte, hash []byte) ([]byte, error) {
	private_key, err := x509.ParseECPrivateKey(private_key_bytes)
	if err != nil {
		return nil, err
	}

	r, s, err := ecdsa.Sign(rand.Reader, private_key, hash)
	if err != nil {
		return nil, err
	}
//	v.Vlogln(5, "r: ", len(r.Bytes()), r.Bytes())
//	v.Vlogln(5, "s: ", len(s.Bytes()), s.Bytes())

	return append(r.Bytes(), s.Bytes()...), nil
}

func SignECDSA2(private_key *ecdsa.PrivateKey, hash []byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, private_key, hash)
	if err != nil {
		return nil, err
	}
//	v.Vlogln(5, "r: ", len(r.Bytes()), r.Bytes())
//	v.Vlogln(5, "s: ", len(s.Bytes()), s.Bytes())

	return append(r.Bytes(), s.Bytes()...), nil
}

func VerifyECDSA(public_key_bytes []byte, hash []byte, signature []byte) (result bool) {
	public_key, err := x509.ParsePKIXPublicKey(public_key_bytes)
	if err != nil {
		return false
	}

	var r big.Int
	r.SetBytes(signature[0:32])
	var s big.Int
	s.SetBytes(signature[32:64])

	switch public_key := public_key.(type) {
	case *ecdsa.PublicKey:
		return ecdsa.Verify(public_key, hash, &r, &s)
	default:
		return false
	}
}


