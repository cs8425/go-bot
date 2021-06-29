package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	vlog "local/log"
)

type KeyStore struct {
	keys map[string][]byte
	iv   []byte
	key  []byte
}

func (ks *KeyStore) Keys() []string {
	list := make([]string, 0, len(ks.keys))
	for uuid, _ := range ks.keys {
		list = append(list, uuid)
	}
	return list
}

func (ks *KeyStore) Get(uuid string) ([]byte, bool) {
	buf, ok := ks.keys[uuid]
	if !ok {
		return nil, false
	}
	dec, err := Decrypt(buf, ks.iv, ks.key)
	if err != nil {
		return nil, false
	}
	return dec, true
}

func (ks *KeyStore) Set(uuid string, key []byte) {
	if buf, ok := ks.keys[uuid]; ok {
		ks.clear(buf) // clear up
	}
	enc, err := Encrypt(key, ks.iv, ks.key)
	if err != nil {
		return
	}
	ks.clear(key) // clear up
	ks.keys[uuid] = enc
}

func (ks *KeyStore) Del(uuid string) bool {
	buf, ok := ks.keys[uuid]
	if !ok {
		return false
	}
	ks.clear(buf) // clear up
	delete(ks.keys, uuid)
	return true
}

func (ks *KeyStore) Clear() {
	for uuid, buf := range ks.keys {
		ks.clear(buf) // clear up
		delete(ks.keys, uuid)
	}
}

func (ks *KeyStore) clear(buf []byte) {
	for i, _ := range buf {
		buf[i] = 0x00
	}
}

func NewKeyStore() *KeyStore {
	iv := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		vlog.Vln(2, "[keystore]init:", err)
	}

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		vlog.Vln(2, "[keystore]init:", err)
	}

	return &KeyStore{
		keys: make(map[string][]byte),
		iv:   iv,
		key:  key,
	}
}

// key = 16 or 24 or 32 Bytes
// nonce = 12 Bytes
func Encrypt(plaintext []byte, nonce []byte, key []byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		// panic(err.Error())
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		// panic(err.Error())
		return
	}

	ciphertext = aesgcm.Seal(nil, nonce, plaintext, nil)
	// fmt.Printf("%d [%x]\n", len(ciphertext), ciphertext)
	return
}

func Decrypt(ciphertext []byte, nonce []byte, key []byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		// panic(err.Error())
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		// panic(err.Error())
		return
	}

	plaintext, err = aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		//panic(err.Error())
		return
	}

	// fmt.Printf("%s\n", plaintext)
	return
}
