package toolkit

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
)

// RSA
var Label = []byte("")

func GenRSAKeys(length int) (private_key_bytes, public_key_bytes []byte) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, length)
	publicKey := &privateKey.PublicKey

	private_key_bytes = x509.MarshalPKCS1PrivateKey(privateKey)
	public_key_bytes, _ = x509.MarshalPKIXPublicKey(publicKey)
	return private_key_bytes, public_key_bytes
}

func EncRSA(publicKey *rsa.PublicKey, message []byte) ([]byte, error) {
//	Label := []byte("")
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, publicKey, message, Label)

	return ciphertext, err
}

func DecRSA(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
//	Label := []byte("")
	hash := sha256.New()

	plainText, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, ciphertext, Label)

	return plainText, err
}

//func SignRSA(privateKey *rsa.PrivateKey, message []byte) ([]byte, error) {
func SignRSA(privateKey *rsa.PrivateKey, sha256hashed []byte) ([]byte, error) {
	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto // for simple example
	newhash := crypto.SHA256
//	pssh := newhash.New()
//	pssh.Write(message)
//	sha256hashed := pssh.Sum(nil)

	signature, err := rsa.SignPSS(rand.Reader, privateKey, newhash, sha256hashed, &opts)
	return signature, err
}

//func VerifyRSA(publicKey *rsa.PublicKey, message []byte, signature []byte) (result bool) {
func VerifyRSA(publicKey *rsa.PublicKey, sha256hashed []byte, signature []byte) (result bool) {
	result = false

	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto // for simple example
	newhash := crypto.SHA256
//	pssh := newhash.New()
//	pssh.Write(message)
//	sha256hashed := pssh.Sum(nil)

	err := rsa.VerifyPSS(publicKey, newhash, sha256hashed, signature, &opts)
	if err == nil {
		result = true
	}
	return
}

func ParseRSAPub(public_key_bytes []byte) (*rsa.PublicKey, error) {
	public_key, err := x509.ParsePKIXPublicKey(public_key_bytes)
	if err != nil {
		return nil, err
	}

	key, ok := public_key.(*rsa.PublicKey)
	if !ok {
		return nil, rsa.ErrVerification
	}
	return key, nil
}

func ParseRSAPriv(private_key_bytes []byte) (*rsa.PrivateKey, error) {
	return x509.ParsePKCS1PrivateKey(private_key_bytes)
}

