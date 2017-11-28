package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func handleRegister(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var acc Account
	err := decoder.Decode(&acc)
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()

	log.Println(acc.Email + " is trying to register")

	err = db.addAccount(acc)
	if err != nil {
		log.Fatal(err)
	}

	w.Write([]byte{0x00})
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

	if grantType[0] == "refresh_token" {
		log.Println("Trying to refresh token, not implemented")
		w.Write([]byte("error"))
		return // Not implemented just return
	}

	username := req.PostForm["username"][0]
	passwordHash := req.PostForm["password"][0]

	log.Println(username + " is trying to login")

	acc, err := db.getAccount(username)
	if acc.MasterPasswordHash != passwordHash {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(401)))
		log.Println("Login attempt failed")
		return
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
		RefreshToken: "NA",
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
