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

func createRefreshToken() string {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		log.Fatal(err)
	}

	tokenStr := base64.StdEncoding.EncodeToString(token)

	return tokenStr
}

type resToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Key          string `json:"key"`
}

func handleLogin(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	grantType, ok := req.PostForm["grant_type"]
	if !ok {
		w.Write([]byte("error"))
		log.Println("Login without grant_type")
		return
	}

	var acc Account
	var err error
	if grantType[0] == "refresh_token" {
		rrefreshToken := req.PostForm["refresh_token"][0]
		if len(rrefreshToken) < 4 {
			// Fatal to always catch this
			log.Fatal("fake refreshToken " + rrefreshToken)
		}

		acc, err = db.getAccount("", rrefreshToken)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(acc.Email + " is trying to refresh a token " + rrefreshToken)
		if acc.RefreshToken != rrefreshToken {
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

		acc, err = db.getAccount(username, "")
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
	refreshToken := createRefreshToken()
	err = db.updateAccountInfo(acc.Id, refreshToken)
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
		RefreshToken: refreshToken,
		Key:          acc.Key,
	}

	data, err := json.Marshal(&rtoken)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

type ctxKey string

func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var tokenString string

		tokens, ok := req.Header["Authorization"]
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
