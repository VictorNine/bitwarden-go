package server

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/rs/cors"
)

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

	folders, err := db.getFolders(acc.Id)
	if err != nil {
		log.Println(err)
	}

	Domains := Domains{
		Object:            "domains",
		EquivalentDomains: nil,
		GlobalEquivalentDomains: []GlobalEquivalentDomains{
			GlobalEquivalentDomains{Type: 1, Domains: []string{"youtube.com", "google.com", "gmail.com"}, Excluded: false},
		},
	}

	data := SyncData{
		Profile: prof,
		Folders: folders,
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

func handleNewFolder(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)

	log.Println(email + " is trying to add a new folder")

	acc, err := db.getAccount(email)
	if err != nil {
		log.Fatal("Account lookup " + err.Error())
	}

	decoder := json.NewDecoder(req.Body)

	var folderData struct {
		Name string `json:"name"`
	}

	err = decoder.Decode(&folderData)
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()

	folder, err := db.addFolder(folderData.Name, acc.Id)
	if err != nil {
		log.Fatal("newFolder error" + err.Error())
	}

	data, err := json.Marshal(&folder)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// Interface to make testing easier
type database interface {
	init() error
	addAccount(acc Account) error
	getAccount(username string) (Account, error)
	updateAccountInfo(acc Account) error
	getCiphers(owner string) ([]Cipher, error)
	newCipher(ciph Cipher, owner string) (Cipher, error)
	updateCipher(newData Cipher, owner string, ciphID string) error
	deleteCipher(owner string, ciphID string) error
	open() error
	close()
	addFolder(name string, owner string) (Folder, error)
	getFolders(owner string) ([]Folder, error)
}

func Start() {
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

	mux := http.NewServeMux()

	mux.HandleFunc("/api/accounts/register", handleRegister)
	// TODO /api/accounts/keys ?
	// TODO /api/accounts/profile ?
	mux.HandleFunc("/identity/connect/token", handleLogin)

	mux.Handle("/api/folders", jwtMiddleware(http.HandlerFunc(handleNewFolder)))
	mux.Handle("/apifolders", jwtMiddleware(http.HandlerFunc(handleNewFolder))) // The android app want's the address like this, will be fixed in the next version. Issue #174
	mux.Handle("/api/sync", jwtMiddleware(http.HandlerFunc(handleSync)))

	mux.Handle("/api/ciphers", jwtMiddleware(http.HandlerFunc(handleNewCipher)))
	mux.Handle("/api/ciphers/", jwtMiddleware(http.HandlerFunc(handleCipherUpdate)))

	log.Println("Starting server on " + serverAddr)
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
		Debug:            true,
	}).Handler(mux)
	http.ListenAndServe(serverAddr, handler)
}
