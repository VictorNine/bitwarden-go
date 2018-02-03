package auth

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
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

type tfaObjectType struct {
	Enabled bool
	Type    int
	Object  string
}

func check2FA(w http.ResponseWriter, req *http.Request, secret string) error {
	code, ok := req.PostForm["twoFactorToken"]
	if !ok {
		code, ok = req.PostForm["TwoFactorToken"] // Android is different from web and browser
	}
	if !ok {
		resp := struct {
			Error               string `json:"error"`
			ErrorDescription    string `json:"error_description"`
			TwoFactorProviders  []int
			TwoFactorProviders2 map[string]*int
		}{
			Error:               "invalid_grant",
			ErrorDescription:    "Two factor required.",
			TwoFactorProviders:  []int{0},
			TwoFactorProviders2: make(map[string]*int),
		}
		resp.TwoFactorProviders2["0"] = nil

		data, _ := json.Marshal(&resp)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)
		return errors.New("Code not provided")
	}

	otpc := &dgoogauth.OTPConfig{
		Secret:      secret,
		WindowSize:  3,
		HotpCounter: 0,
	}

	authenticated, err := otpc.Authenticate(code[0])
	if err != nil || !authenticated {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(401)))
		return errors.New("Could not authenticat")
	}

	return nil
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
		Enabled: acc.GetProfile().TwoFactorEnabled,
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

func (auth *Auth) HandleTwoFactor(w http.ResponseWriter, req *http.Request) {
	email := GetEmail(req)

	acc, err := auth.db.GetAccount(email, "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		log.Println(err)
		return
	}

	tfadata := struct {
		Data              []tfaObjectType
		ContinuationToken *string
		Object            string
	}{
		ContinuationToken: nil,
		Object:            "list",
	}
	tfadata.Data = []tfaObjectType{tfaObjectType{
		Enabled: acc.GetProfile().TwoFactorEnabled,
		Type:    0,
		Object:  "twoFactorProvider",
	}}

	data, err := json.Marshal(&tfadata)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (auth *Auth) HandleDisableTwoFactor(w http.ResponseWriter, req *http.Request) {
	email := GetEmail(req)

	decoder := json.NewDecoder(req.Body)
	var reqData struct {
		Type               int    `json:"type"`
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

	err = auth.db.Update2FAsecret("", email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
		log.Println(err)
		return
	}

	tfaData := tfaObjectType{
		Enabled: false,
		Type:    0,
		Object:  "twoFactorProvider",
	}

	data, err := json.Marshal(&tfaData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)

}
