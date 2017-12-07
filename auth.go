package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"

	jwt "github.com/dgrijalva/jwt-go"
)

func reHashPassword(key, salt string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}

	hash := pbkdf2.Key(b, []byte(salt), 5000, 256/8, sha256.New)

	return base64.StdEncoding.EncodeToString(hash), nil
}

func handleKeysUpdate(w http.ResponseWriter, req *http.Request) {

}

func handleProfile(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)
	log.Println("Profile requested")

	acc, err := db.getAccount(email)
	if err != nil {
		log.Fatal(err)
	}

	prof := acc.getProfile()

	data, err := json.Marshal(&prof)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func handleRegister(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var acc Account
	err := decoder.Decode(&acc)
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()

	log.Println(acc.Email + " is trying to register")

	acc.MasterPasswordHash, err = reHashPassword(acc.MasterPasswordHash, acc.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
		log.Println(err)
		return
	}

	err = db.addAccount(acc)
	if err != nil {
		log.Fatal(err)
	}

	w.Write([]byte{0x00})
}

func createRefreshToken(id string) string {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		log.Fatal(err)
	}

	tokenStr := base64.StdEncoding.EncodeToString(token)

	return id + ":" + tokenStr
}

type resToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Key          string `json:"Key"`
}

// PrivateKey is needed by the web vault. But android will crash if it's included
type resTokenWPK struct {
	resToken
	PrivateKey string `json:"PrivateKey"`
}

func handleLogin(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	grantType, ok := req.PostForm["grant_type"]
	if !ok {
		w.Write([]byte("error"))
		log.Println("Login without grant_type")
		return
	}

	clientID := req.PostForm["client_id"][0]

	var acc Account
	var err error
	if grantType[0] == "refresh_token" {
		rrefreshToken := req.PostForm["refresh_token"][0]
		rt := strings.Split(rrefreshToken, ":")
		if len(rt) != 2 {
			// Fatal to always catch this
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(401)))
			log.Println("fake refreshToken " + rrefreshToken)
			return
		}

		acc, err = db.getAccount(rt[0])
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(401)))
			log.Println("Account not found")
			return
		}
		log.Println(acc.Email + " is trying to refresh a token " + rt[1])
		if acc.RefreshToken != rt[1] {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(401)))
			log.Println("Login attempt failed")
			return
		}
	} else {
		// Login with username
		username := req.PostForm["username"][0]
		passwordHash := req.PostForm["password"][0]

		log.Println(username + " is trying to login")

		acc, err = db.getAccount(username)
		if err != nil {
			log.Fatal(err)
		}

		reHash, _ := reHashPassword(passwordHash, acc.Email)

		if acc.MasterPasswordHash != reHash {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(401)))
			log.Println("Login attempt failed")
			return
		}
	}

	// Create refreshtoken and store in db
	acc.RefreshToken = createRefreshToken(acc.Id)
	err = db.updateAccountInfo(acc)
	if err != nil {
		log.Fatal(err)
	}

	// Create the token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["nbf"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Second * time.Duration(jwtExpire)).Unix()
	claims["iss"] = "NA"
	claims["sub"] = "NA"
	claims["email"] = acc.Email
	claims["name"] = acc.Name
	claims["premium"] = false
	tokenString, _ := token.SignedString(mySigningKey)

	rtoken := resToken{AccessToken: tokenString,
		ExpiresIn:    jwtExpire,
		TokenType:    "Bearer",
		RefreshToken: acc.RefreshToken,
		Key:          acc.Key,
	}

	var data []byte
	// Login from web vault add priv key
	if clientID == "web" {
		// temporary priv key to make tesing easier
		tempPrivKey := "0.2o2tRZZ4twUsLId6tNU+eQ==|54SMNoQC4DdvJOpeG3uQHjAGtpy1H0DlXKFB+W6o1woA7EVp5TAb9d+4N3SoNyF/9nNdhVHmkIVIgjRpN8Wq0Gzb7TqqVwXbO7E7Gt73YZK+UMS9afyqpiNAGnSsyf/7R8KkbHov3PasFTOy9U75OLoeCzNkL0RlCtDlbblNZFGmE6MZtgCTTgGGrVGP82K6o3AZVpkvf+Y83IO9v1B8prqsNYlB8w8JRJ7fYV0yHrnbPfRzDsC9IcNtlXOs8+0omrgUUWiXQdDdhRkT1yHinSCeZNLc6JufJubOoROlux3DLMSpAu+Xgs8y3WThikIIZu/Oh60n9CEJrRJ/T79cKjPbFJhWlAaInliA2BT+yK1LFwgZoozgQ/ppX8vapt8yC6Cia66krZPE57N6WPeqnGwXXrT4L/SKA53Kcdknx3vEERebF5wton6eSejzMJSDye2hmqB8RXYelhYmHEFlcjNRUkDUplUBhQLMHXecZvh8/S6LBbyucy7R4QG1Z/EBVCtodpnerkt4STPS5nOLAUwxYv5HGkue8gIbl7ARXln+kGYlvV2HwnZefqOu+/44S3ELy5jIDlnCvgvk8A1voy7MhbdLCFPYUs+XnPPftiBw757GdsEfXAUku97qSvgy6Sd0uundfukEfBjH0Wrovi8POcaV61C+bvrSM7FYy30M8riEaH/9SfeWqVfciCSsU9mpSQ39RyoPwik89O1vfjjbsJMUp3HrZzkJmvr7rcm83Pc4/3ITyvEkLN3aYdEAgL/QIOGW8sSWxfotG+DtdEMULBi2qvX73XzMK893uqGR1CMtB/KcdDmlPpCEPYtGB6P2380fnyxLHD6EXtNHhl9v0ZWlwoFtUo+RNcn0XkXB0yVOLzQDAM2Ovxs1JHD/AxQRAnh/RaYm0miWQe6wgfIcavZodqLdhDf4lMXdpPRGL5YfOXua4z50ul0fl3erz3pv8T2YnP0uHHJVf0VLmzlL57jRCBHRldXSfbtYOXOlalcqTajIx8ONcaeEUvc40tJ+J43O5BjGQdt8dIM1zAgWDyWGWU6Im73IFD7EFsrbeCqneGqpvoChY6LirLCLgVlBSf4WlpLaGvDPzJHf/Ss+skyseUf3ahoeMlotqOeyRFCmi4zknfo1THgqUiKcuzKELTENkZmF7ra94f9VUOVMjacfD9EpweG/pLaZkdOxPf/MABIJ/gJ9Rv30XLSZdIvkKwm4LlCzLXjyBu3rr9HdknzUQN/GeeBKSMuu1f8P2BIV9ytS0bX0Ap/CsUiJeNXKXzLpwPPurKM8/7m4NGGLmod39K2H0vRdeZHJB1/TwCX6VAO6dGv/nAyq4kC40Pu3UCw/9BMIm8fK+vUMDBIbuieJxArLLuxo+gaQqhTL/ZYFA62CUFdDTWRwVHoJw9io8VxxxXcxqMK1QatDyHAWbBM0Y454gfDPYbq7IRGwc2P3xixX5gY730YCJMrQ35uzBQ2I4rkk9znH1BDAKynDQuUCdkcM+eCh6wPVq/SNL6DwWtqanbK8X7FGWUp8woKY5Twq0C95O57drB6fiRGinXBPm7WYWyV3ccZSiyFGNadsVC3HG/R+ExGTtwkC8LH+Ek1GgskwWY83D0Gv8yUbVTS6N3iXUkDmGjBswV0="
		rtokenWPK := resTokenWPK{resToken: rtoken, PrivateKey: tempPrivKey}
		//rtokenWPK := resTokenWPK{resToken: rtoken, PrivateKey: acc.Keys.EncryptedPrivateKey}
		data, err = json.Marshal(&rtokenWPK)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		data, err = json.Marshal(&rtoken)
		if err != nil {
			log.Fatal(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

type ctxKey string

func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var tokenString string

		tokens, ok := req.Header["Authorization"]
		if !ok {
			// hack in web-app to use Content-Language
			tokens, ok = req.Header["Content-Language"]
		}
		if ok && len(tokens) >= 1 {
			tokenString = tokens[0]
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			return mySigningKey, nil
		})

		if err != nil {
			log.Println("JWT: " + err.Error()) // Fatal for now to catch all errors here

			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(401)))
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			email, ok := claims["email"].(string)
			if ok {
				ctx := context.WithValue(req.Context(), ctxKey("email"), email)
				next.ServeHTTP(w, req.WithContext(ctx))
				return
			}
		}

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(401)))
	})
}
