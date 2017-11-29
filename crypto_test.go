package main

import (
	"log"
	"testing"
)

func TestCrypto(t *testing.T) {
	var email = "test@example.com"
	var password = "password"
	var passwordHash = "Q4zw5LmXHMJDJYBPfeFYtW8+dxbcCHTFmzE04OXS6Ic="

	var enckey = "0.yMeH5ypzRLcyJX69HAt6mQ==|H0mdMpoX1aguKIaCXOreL93JyCpo9ORiX8ZbK+taLXlGZfCb5TOs0eriKa7u1ocBp9gDHwYm5EUyobnbVfZ3uiP2suYWAXKmC4IO67b7ozc="

	var encdata = "2.eWiu5v/7OWt5EiuypCP9nQ==|8vxfq3AsARNjPE8rWcDLSg==|TKN0DmdhK8qjIqLe7WPpjVcAoUghGDxnpWUb4WS0jHQ="

	var encTest = "TESTING ENCRYPTN"

	cs, err := NewCipherString(enckey)
	if err != nil {
		t.Error(err)
	}
	log.Println(cs)

	dk := MakeKey(password, email)
	log.Println(dk)

	// MasterPasswordHash
	hash := HashPassword(password, dk)
	if hash != passwordHash {
		t.Errorf("Expected %v got %v", passwordHash, hash)
	}

	ct, err := Encrypt([]byte(encTest), dk)
	if err != nil {
		t.Error(err)
	}
	pt := ct.Decrypt(dk)
	if string(pt) != encTest {
		t.Errorf("Expected %v got %v", encTest, string(pt))
	}

	cs, err = NewCipherString(enckey)
	if err != nil {
		t.Error(err)
	}
	mk := cs.Decrypt(dk)
	log.Println(mk)
	//mk = mk[:32]
	log.Println(mk)

	ds, err := NewCipherString(encdata)
	if err != nil {
		t.Error(err)
	}
	d := ds.Decrypt(mk)
	log.Println(d)
	log.Println(string(d))

	/*
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
	*/
}
