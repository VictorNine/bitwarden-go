package common

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

type Auth struct {
	db         Database
	signingKey []byte
	jwtExpire  int
}

// TODO: rename to New() when when moved to sep pkg
func NewAuth(db Database, signingKey []byte, jwtExpire int) Auth {
	auth := Auth{
		db:         db,
		signingKey: signingKey,
		jwtExpire:  jwtExpire,
	}

	return auth
}

func reHashPassword(key, salt string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}

	hash := pbkdf2.Key(b, []byte(salt), 5000, 256/8, sha256.New)

	return base64.StdEncoding.EncodeToString(hash), nil
}

func (auth *Auth) HandleRegister(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var acc Account
	err := decoder.Decode(&acc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		log.Println(err)
		return
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

	err = auth.db.AddAccount(acc)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(500)))
		log.Println(err)
		return
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
	Key          string `json:"Key"`
}

// PrivateKey is needed by the web vault. But android will crash if it's included
type resTokenWPK struct {
	resToken
	PrivateKey string `json:"PrivateKey"`
}

func (auth *Auth) HandleLogin(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	grantType, ok := req.PostForm["grant_type"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		log.Println("Login without grant_type")
		return
	}

	clientID := req.PostForm["client_id"][0]

	var acc Account
	var err error
	if grantType[0] == "refresh_token" {
		rrefreshToken := req.PostForm["refresh_token"][0]
		if len(rrefreshToken) < 4 {
			// Fatal to always catch this
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(401)))
			log.Println("fake refreshToken " + rrefreshToken)
			return
		}

		acc, err = auth.db.GetAccount("", rrefreshToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(401)))
			log.Println("Account not found")
			return
		}
		log.Println(acc.Email + " is trying to refresh a token ")
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

		acc, err = auth.db.GetAccount(username, "")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(401)))
			log.Println("Account not found")
			return
		}

		reHash, _ := reHashPassword(passwordHash, acc.Email)

		if acc.MasterPasswordHash != reHash {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(401)))
			log.Println("Login attempt failed")
			return
		}
	}

	// Don't change refresh token every time or the other clients will be logged out
	if acc.RefreshToken == "" {
		// Create refreshtoken and store in db
		acc.RefreshToken = createRefreshToken()
		err = auth.db.UpdateAccountInfo(acc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(http.StatusText(401)))
			log.Println(err)
			return
		}
	}

	// Create the token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["nbf"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Second * time.Duration(auth.jwtExpire)).Unix()
	claims["iss"] = "NA"
	claims["sub"] = "NA"
	claims["email"] = acc.Email
	claims["name"] = acc.Name
	claims["premium"] = false
	tokenString, _ := token.SignedString(auth.signingKey)

	rtoken := resToken{AccessToken: tokenString,
		ExpiresIn:    auth.jwtExpire,
		TokenType:    "Bearer",
		RefreshToken: acc.RefreshToken,
		Key:          acc.Key,
	}

	var data []byte
	// Login from web vault add priv key
	if clientID == "web" {
		rtokenWPK := resTokenWPK{resToken: rtoken, PrivateKey: acc.KeyPair.EncryptedPrivateKey}
		data, err = json.Marshal(&rtokenWPK)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			log.Println("Account not found")
			return
		}
	} else {
		data, err = json.Marshal(&rtoken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			log.Println("Account not found")
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

type ctxKey string

func (auth *Auth) JwtMiddleware(next http.Handler) http.Handler {
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

			return auth.signingKey, nil
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
