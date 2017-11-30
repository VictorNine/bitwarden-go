package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/andreburgaud/crypt2go/padding"
	"golang.org/x/crypto/pbkdf2"
)

type CipherString struct {
	encryptedString      string
	encryptionType       int
	decryptedValue       string
	cipherText           string
	initializationVector string
	mac                  string
}

const (
	AesCbc256_B64                     = iota
	AesCbc128_HmacSha256_B64          = iota
	AesCbc256_HmacSha256_B64          = iota
	Rsa2048_OaepSha256_B64            = iota
	Rsa2048_OaepSha1_B64              = iota
	Rsa2048_OaepSha256_HmacSha256_B64 = iota
	Rsa2048_OaepSha1_HmacSha256_B64   = iota
)

func NewCipherString(encryptedString string) (*CipherString, error) {
	cs := CipherString{}
	cs.encryptedString = encryptedString
	if encryptedString == "" {
		return nil, errors.New("empty key")
	}
	headerPieces := strings.Split(cs.encryptedString, ".")
	var encPieces []string
	if len(headerPieces) == 2 {
		cs.encryptionType, _ = strconv.Atoi(headerPieces[0])
		encPieces = strings.Split(headerPieces[1], "|")
	} else {
		return nil, errors.New("invalid key header")
	}

	switch cs.encryptionType {
	case AesCbc256_B64:
		if len(encPieces) != 2 {
			return nil, fmt.Errorf("invalid key body len %d", len(encPieces))
		}
		cs.initializationVector = encPieces[0]
		cs.cipherText = encPieces[1]
	case AesCbc256_HmacSha256_B64:
		if len(encPieces) != 3 {

		}
		cs.initializationVector = encPieces[0]
		cs.cipherText = encPieces[1]
		cs.mac = encPieces[2]
	default:
		return nil, errors.New("unknown algorithm")
	}
	return &cs, nil
}

func NewCipherStringRaw(encryptionType int, ct string, iv string, mac string) (*CipherString, error) {
	cs := CipherString{encryptionType: encryptionType, cipherText: ct, initializationVector: iv, mac: mac}
	return &cs, nil
}

func (cs *CipherString) Decrypt(key []byte) []byte {
	iv, _ := base64.StdEncoding.DecodeString(cs.initializationVector)
	ct, _ := base64.StdEncoding.DecodeString(cs.cipherText)
	//mac, _ := base64.StdEncoding.DecodeString(cs.mac)

	//var hmac []byte
	switch cs.encryptionType {
	case AesCbc256_HmacSha256_B64:
		//	hmac = key[32:]
		key = key[:32]
	default:
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks(ct, ct)
	ct, _ = padding.NewPkcs7Padding(16).Unpad(ct) //TODO, configurable size
	return ct
}

func Encrypt(pt []byte, key []byte) (*CipherString, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure.
	iv := make([]byte, aes.BlockSize)

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	pt, _ = padding.NewPkcs7Padding(16).Pad(pt) //TODO, configurable size
	ct := make([]byte, len(pt))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ct, pt)

	cs := CipherString{encryptionType: 0, cipherText: base64.StdEncoding.EncodeToString(ct), initializationVector: base64.StdEncoding.EncodeToString(iv)}
	return &cs, nil
}

func MakeEncKey(key []byte) (*CipherString, error) {
	b := make([]byte, 512/8)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(err)
	}
	return Encrypt(b, key)

}

func MakeKey(password string, salt string) []byte {
	dk := pbkdf2.Key([]byte(password), []byte(salt), 5000, 256/8, sha256.New)
	return dk
}

func HashPassword(password string, key []byte) string {
	hash := pbkdf2.Key(key, []byte(password), 1, 256/8, sha256.New)
	return base64.StdEncoding.EncodeToString(hash)
}
