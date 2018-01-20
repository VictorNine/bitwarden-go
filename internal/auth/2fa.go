package auth

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"log"
	"net/http"

	bw "github.com/VictorNine/bitwarden-go/internal/common"
)

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

	authData := struct {
		Enabled bool
		Key     string
		Object  string
	}{
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
