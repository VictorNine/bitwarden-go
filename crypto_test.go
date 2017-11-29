package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"golang.org/x/crypto/pbkdf2"
	"log"
	"strings"
	"testing"
)

func TestCrypto(t *testing.T) {
	var email = "test@example.com"
	var password = "password"

	var enckey = "0.yMeH5ypzRLcyJX69HAt6mQ==|H0mdMpoX1aguKIaCXOreL93JyCpo9ORiX8ZbK+taLXlGZfCb5TOs0eriKa7u1ocBp9gDHwYm5EUyobnbVfZ3uiP2suYWAXKmC4IO67b7ozc="

	dk := pbkdf2.Key([]byte(password), []byte(email), 5000, 256/8, sha256.New)
	log.Println(dk)

	// MasterPasswordHash
	hash := pbkdf2.Key(dk, []byte(password), 1, 256/8, sha256.New)
	log.Println(hash)
	log.Println(base64.StdEncoding.EncodeToString(hash))

	ke := strings.Split(enckey, ".")
	et := ke[0] // 0 = AesCbc256_B64

	if et != "0" {
		log.Println("ERROR, invalid et")
		return
	}
	ep := strings.Split(ke[1], "|")

	log.Println(ep[0])
	iv, _ := base64.StdEncoding.DecodeString(ep[0])
	log.Println(iv)

	log.Println(ep[1])
	ct, _ := base64.StdEncoding.DecodeString(ep[1])
	log.Println(ct)

	block, err := aes.NewCipher(dk)
	if err != nil {
		panic(err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks(ct, ct)
	log.Println(ct)
}
