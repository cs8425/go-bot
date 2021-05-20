package toolkit

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

var ErrLength = errors.New("wrong length of format")

// key = 16 or 24 or 32 Bytes
// nonce = 12 Bytes
func Encrypt(plaintext []byte, nonce []byte, key []byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		//		panic(err.Error())
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		//		panic(err.Error())
		return
	}

	ciphertext = aesgcm.Seal(nil, nonce, plaintext, nil)
	//	fmt.Printf("%d [%x]\n", len(ciphertext), ciphertext)

	return
}

func Decrypt(ciphertext []byte, nonce []byte, key []byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		//		panic(err.Error())
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		//		panic(err.Error())
		return
	}

	plaintext, err = aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		//		panic(err.Error())
		return
	}

	//	fmt.Printf("%s\n", plaintext)
	return
}

func UnPackLineByte(b []byte) (int64, []byte, []byte, error) {
	if len(b) < 4 {
		return 0, nil, nil, ErrLength
	}

	addr := int64(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24)
	pair := b[4:]
	pairlen := len(pair) / 2
	buf := pair[:pairlen]
	rbuf := pair[pairlen:]
	if len(buf) != len(rbuf) {
		return 0, nil, nil, ErrLength
	}
	return addr, buf, rbuf, nil
}

func PackLineByte(addr int64, buf []byte, rbuf []byte) (line []byte) {
	if len(buf) != len(rbuf) {
		return
	}

	size := len(buf) + len(rbuf) + 4
	b := make([]byte, size, size)
	b[0] = byte(addr)
	b[1] = byte(addr >> 8)
	b[2] = byte(addr >> 16)
	b[3] = byte(addr >> 24)
	copy(b[4:], buf)
	copy(b[len(rbuf)+4:size], rbuf)

	return b
}
