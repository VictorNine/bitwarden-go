package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
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

// The data we get from the client. Only used to parse data
type newCipher struct {
	Type           int       `json:"type"`
	FolderId       string    `json:"folderId"`
	OrganizationId string    `json:"organizationId"`
	Name           string    `json:"name"`
	Notes          string    `json:"notes"`
	Favorite       bool      `json:"favorite"`
	Login          loginData `json:"login"`
}

type loginData struct {
	URI      string `json:"uri"`
	Username string `json:"username"`
	Password string `json:"password"`
	ToTp     string `json:"totp"`
}

func handleNewCipher(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)

	log.Println(email + " is trying to add data")

	acc, err := db.getAccount(email)
	if err != nil {
		log.Fatal("Account lookup " + err.Error())
	}

	rCiph, err := unmarshalCipher(req.Body)
	if err != nil {
		log.Fatal("Cipher decode error" + err.Error())
	}

	// Store the new cipher object in db
	newCiph, err := db.newCipher(rCiph, acc.Id)
	if err != nil {
		log.Fatal("newCipher error" + err.Error())
	}

	data, err := json.Marshal(&newCiph)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// This function handles updates and deleteing
func handleCipherUpdate(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)
	log.Println(email + " is trying to edit his data")

	// Get the cipher id
	id := req.URL.Path[len("/api/ciphers/"):]

	acc, err := db.getAccount(email)
	if err != nil {
		log.Fatal("Account lookup " + err.Error())
	}

	switch req.Method {
	case "PUT":
		rCiph, err := unmarshalCipher(req.Body)
		if err != nil {
			log.Fatal("Cipher decode error" + err.Error())
		}

		// Set correct ID
		rCiph.Id = id

		err = db.updateCipher(rCiph, acc.Id, id)
		if err != nil {
			w.Write([]byte("0"))
			log.Println(err)
			return
		}

		// Send response
		data, err := json.Marshal(&rCiph)
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		log.Println("Cipher " + id + " updated")
		return

	case "DELETE":
		err := db.deleteCipher(acc.Id, id)
		if err != nil {
			w.Write([]byte("0"))
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(""))
		log.Println("Cipher " + id + " deleted")
		return
	default:
		w.Write([]byte("0"))
		return
	}

}

func handleSync(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)

	log.Println(email + " is trying to sync")

	acc, err := db.getAccount(email)

	prof := Profile{
		Id:               acc.Id,
		Email:            acc.Email,
		EmailVerified:    false,
		Premium:          false,
		Culture:          "en-US",
		TwoFactorEnabled: false,
		Key:              acc.Key,
		SecurityStamp:    "123",
		Organizations:    nil,
		Object:           "profile",
	}

	ciphs, err := db.getCiphers(acc.Id)
	if err != nil {
		log.Println(err)
	}

	Domains := Domains{
		EquivalentDomains: nil,
		GlobalEquivalentDomains: []GlobalEquivalentDomains{
			GlobalEquivalentDomains{Type: 1, Domains: []string{"youtube.com", "google.com", "gmail.com"}, Excluded: false},
		},
	}

	data := SyncData{
		Profile: prof,
		Folders: make([]string, 0), // Needed or the android app will crash
		Domains: Domains,
		Object:  "sync",
		Ciphers: ciphs,
	}

	jdata, err := json.Marshal(&data)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jdata)
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

func main() {
	initDB := flag.Bool("init", false, "Initialize the database")
	flag.Parse()

	err := db.open()
	if err != nil {
		log.Fatal(err)
	}

	defer db.close()

	// Create a new database
	if *initDB {
		err := db.init()
		if err != nil {
			log.Fatal(err)
		}
	}

	http.HandleFunc("/api/accounts/register", handleRegister)
	http.HandleFunc("/identity/connect/token", handleLogin)

	http.Handle("/api/sync", jwtMiddleware(http.HandlerFunc(handleSync)))

	http.Handle("/api/ciphers", jwtMiddleware(http.HandlerFunc(handleNewCipher)))
	http.Handle("/api/ciphers/", jwtMiddleware(http.HandlerFunc(handleCipherUpdate)))

	log.Println("Starting server on " + serverAddr)
	http.ListenAndServe(serverAddr, nil)
}
