package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/rs/cors"
)

func handleKeysUpdate(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)

	acc, err := db.getAccount(email, "")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Adding key pair")

	decoder := json.NewDecoder(req.Body)
	var kp KeyPair
	err = decoder.Decode(&kp)
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()

	acc.KeyPair = kp

	db.updateAccountInfo(acc)
}

func handleProfile(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)
	log.Println("Profile requested")

	acc, err := db.getAccount(email, "")
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

func handleCollections(w http.ResponseWriter, req *http.Request) {

	collections := Data{Object: "list", Data: []string{}}
	data, err := json.Marshal(collections)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func handleCipher(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)

	log.Println(email + " is trying to add data")

	acc, err := db.getAccount(email, "")
	if err != nil {
		log.Fatal("Account lookup " + err.Error())
	}

	var data []byte

	if req.Method == "POST" {
		rCiph, err := unmarshalCipher(req.Body)
		if err != nil {
			log.Fatal("Cipher decode error" + err.Error())
		}

		// Store the new cipher object in db
		newCiph, err := db.newCipher(rCiph, acc.Id)
		if err != nil {
			log.Fatal("newCipher error" + err.Error())
		}
		data, err = json.Marshal(&newCiph)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		ciphs, err := db.getCiphers(acc.Id)
		if err != nil {
			log.Println(err)
		}
		for i, _ := range ciphs {
			ciphs[i].CollectionIds = make([]string, 0)
			ciphs[i].Object = "cipherDetails"
		}
		list := Data{Object: "list", Data: ciphs}
		data, err = json.Marshal(&list)
		if err != nil {
			log.Fatal(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// This function handles updates and deleteing
func handleCipherUpdate(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)
	log.Println(email + " is trying to edit his data")

	// Get the cipher id
	id := strings.TrimPrefix(req.URL.Path, "/api/ciphers/")

	acc, err := db.getAccount(email, "")
	if err != nil {
		log.Fatal("Account lookup " + err.Error())
	}

	switch req.Method {
	case "GET":
		log.Println("GET Ciphers for " + acc.Id)
		var data []byte
		ciph, err := db.getCipher(acc.Id, id)
		if err != nil {
			log.Fatal(err)
		}
		data, err = json.Marshal(&ciph)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	case "POST":
		fallthrough // Do same as PUT. Web Vault want's to post
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

	acc, err := db.getAccount(email, "")

	prof := Profile{
		Id:               acc.Id,
		Email:            acc.Email,
		EmailVerified:    false,
		Premium:          false,
		Culture:          "en-US",
		TwoFactorEnabled: false,
		Key:              acc.Key,
		SecurityStamp:    nil,
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

// Only handles ciphers
// TODO: handle folders and folderRelationships
func handleImport(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)

	log.Println(email + " is trying to import data")

	acc, err := db.getAccount(email, "")
	if err != nil {
		log.Fatal("Account lookup " + err.Error())
	}

	decoder := json.NewDecoder(req.Body)
	data := struct {
		Ciphers             []newCipher `json:"ciphers"`
		Foders              []string    `json:"folders"`
		FolderRelationships []string    `json:"folderRelationships"`
	}{}

	err = decoder.Decode(&data)
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()

	for _, nc := range data.Ciphers {
		c, err := nc.toCipher()
		if err != nil {
			log.Fatal(err.Error())
		}

		_, err = db.newCipher(c, acc.Id)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	w.Write([]byte{0x00})
}

func handleFolder(w http.ResponseWriter, req *http.Request) {
	email := req.Context().Value(ctxKey("email")).(string)

	log.Println(email + " is trying to add a new folder")

	acc, err := db.getAccount(email, "")
	if err != nil {
		log.Fatal("Account lookup " + err.Error())
	}

	var data []byte
	if req.Method == "POST" {
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

		data, err = json.Marshal(&folder)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		folders, err := db.getFolders(acc.Id)
		if err != nil {
			log.Println(err)
		}
		list := Data{Object: "list", Data: folders}
		data, err = json.Marshal(list)
		if err != nil {
			log.Fatal(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// Interface to make testing easier
type database interface {
	init() error
	addAccount(acc Account) error
	getAccount(username string, refreshtoken string) (Account, error)
	updateAccountInfo(acc Account) error
	getCipher(owner string, ciphID string) (Cipher, error)
	getCiphers(owner string) ([]Cipher, error)
	newCipher(ciph Cipher, owner string) (Cipher, error)
	updateCipher(newData Cipher, owner string, ciphID string) error
	deleteCipher(owner string, ciphID string) error
	open() error
	close()
	addFolder(name string, owner string) (Folder, error)
	getFolders(owner string) ([]Folder, error)
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

	mux := http.NewServeMux()

	mux.HandleFunc("/api/accounts/register", handleRegister)
	mux.HandleFunc("/identity/connect/token", handleLogin)

	mux.Handle("/api/accounts/keys", jwtMiddleware(http.HandlerFunc(handleKeysUpdate)))
	mux.Handle("/api/accounts/profile", jwtMiddleware(http.HandlerFunc(handleProfile)))
	mux.Handle("/api/collections", jwtMiddleware(http.HandlerFunc(handleCollections)))
	mux.Handle("/api/folders", jwtMiddleware(http.HandlerFunc(handleFolder)))
	mux.Handle("/apifolders", jwtMiddleware(http.HandlerFunc(handleFolder))) // The android app want's the address like this, will be fixed in the next version. Issue #174
	mux.Handle("/api/sync", jwtMiddleware(http.HandlerFunc(handleSync)))

	mux.Handle("/api/ciphers/import", jwtMiddleware(http.HandlerFunc(handleImport)))
	mux.Handle("/api/ciphers", jwtMiddleware(http.HandlerFunc(handleCipher)))
	mux.Handle("/api/ciphers/", jwtMiddleware(http.HandlerFunc(handleCipherUpdate)))

	log.Println("Starting server on " + serverAddr)
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
	}).Handler(mux)
	http.ListenAndServe(serverAddr, handler)
}
