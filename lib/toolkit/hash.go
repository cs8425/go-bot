package toolkit

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"

	"io"
)

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// hash function
func HashBytes256(a []byte) []byte {
	sha1h := sha256.New()
	sha1h.Write(a)
	return sha1h.Sum([]byte(""))
}

func HashBytes512(a []byte) []byte {
	sha1h := sha512.New()
	sha1h.Write(a)
	return sha1h.Sum([]byte(""))
}

func IOHash(in io.Reader, out io.Writer) ([]byte, error) {
	quit := false
	shah := sha256.New()
	buf := make([]byte, 8192)
	for {
		n, err := in.Read(buf)
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			quit = true
		}
		shah.Write(buf[:n])

		_, err = out.Write(buf[:n])
		if err != nil {
			return nil, err
		}

		if quit {
			break
		}
	}
	hash := shah.Sum([]byte(""))
	return hash, nil
}

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
func XORBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
	return n
}

// string encode
func Hex(a []byte) string {
	return hex.EncodeToString(a)
}

func Base64URL(a []byte) string {
	return base64.URLEncoding.EncodeToString(a)
}
