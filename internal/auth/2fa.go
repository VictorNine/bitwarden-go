package auth

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"log"
	"net/http"

	bw "github.com/VictorNine/bitwarden-go/internal/common"
	"github.com/dgryski/dgoogauth"
)

type tfaObject struct {
	Enabled bool
	Key     string
	Object  string
}

func (auth *Auth) GetAuthenticator(w http.ResponseWriter, req *http.Request) {
	email := GetEmail(req)

	decoder := json.NewDecoder(req.Body)
	var acc bw.Account
	err := decoder.Decode(&acc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		log.Println(err)
		return
	}
	defer req.Body.Close()

	acc, err = checkPassword(auth.db, email, acc.MasterPasswordHash)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(401)))
		log.Println(err)
		return
	}

	// Generate secret
	random := make([]byte, 20)
	rand.Read(random)
	secret := base32.StdEncoding.EncodeToString(random)

	authData := tfaObject{
		Enabled: false,
		Key:     secret,
		Object:  "twoFactorAuthenticator",
	}

	data, err := json.Marshal(&authData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (auth *Auth) VerifyAuthenticatorSecret(w http.ResponseWriter, req *http.Request) {
	email := GetEmail(req)

	decoder := json.NewDecoder(req.Body)
	var reqData struct {
		Token              string `json:"token"`
		Key                string `json:"key"`
		MasterPasswordHash string `json:"masterPasswordHash"`
	}
	err := decoder.Decode(&reqData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		log.Println(err)
		return
	}
	defer req.Body.Close()

	_, err = checkPassword(auth.db, email, reqData.MasterPasswordHash)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(401)))
		log.Println(err)
		return
	}

	otpc := &dgoogauth.OTPConfig{
		Secret:      reqData.Key,
		WindowSize:  3,
		HotpCounter: 0,
	}

	authenticated, err := otpc.Authenticate(reqData.Token)
	if err != nil || !authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(401)))
		log.Println(err)
		return
	}

	err = auth.db.Update2FAsecret(reqData.Key, email)
	if err != nil || !authenticated {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
		log.Println(err)
		return
	}

	authData := tfaObject{
		Enabled: true,
		Key:     reqData.Key,
		Object:  "twoFactorAuthenticator",
	}

	data, err := json.Marshal(&authData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
