package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/VictorNine/bitwarden-go/internal/auth"
	bw "github.com/VictorNine/bitwarden-go/internal/common"
)

type APIHandler struct {
	db bw.Database
}

// TODO: Rewrite to new when moved to sep pkg
func New(db bw.Database) APIHandler {
	h := APIHandler{
		db: db,
	}

	return h
}

func (h *APIHandler) HandleKeysUpdate(w http.ResponseWriter, req *http.Request) {
	email := auth.GetEmail(req)

	acc, err := h.db.GetAccount(email, "")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Adding key pair")

	decoder := json.NewDecoder(req.Body)
	var kp bw.KeyPair
	err = decoder.Decode(&kp)
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()

	acc.KeyPair = kp

	h.db.UpdateAccountInfo(acc)
}

func (h *APIHandler) HandleProfile(w http.ResponseWriter, req *http.Request) {
	email := auth.GetEmail(req)
	log.Println("Profile requested")

	acc, err := h.db.GetAccount(email, "")
	if err != nil {
		log.Fatal(err)
	}

	prof := acc.GetProfile()

	data, err := json.Marshal(&prof)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (h *APIHandler) HandleCollections(w http.ResponseWriter, req *http.Request) {

	collections := bw.Data{Object: "list", Data: []string{}}
	data, err := json.Marshal(collections)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (h *APIHandler) HandleCipher(w http.ResponseWriter, req *http.Request) {
	email := auth.GetEmail(req)

	log.Println(email + " is trying to add data")

	acc, err := h.db.GetAccount(email, "")
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
		newCiph, err := h.db.NewCipher(rCiph, acc.Id)
		if err != nil {
			log.Fatal("newCipher error" + err.Error())
		}
		data, err = json.Marshal(&newCiph)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		ciphs, err := h.db.GetCiphers(acc.Id)
		if err != nil {
			log.Println(err)
		}
		for i, _ := range ciphs {
			ciphs[i].CollectionIds = make([]string, 0)
			ciphs[i].Object = "cipherDetails"
		}
		list := bw.Data{Object: "list", Data: ciphs}
		data, err = json.Marshal(&list)
		if err != nil {
			log.Fatal(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// This function handles updates and deleteing
func (h *APIHandler) HandleCipherUpdate(w http.ResponseWriter, req *http.Request) {
	email := auth.GetEmail(req)
	log.Println(email + " is trying to edit his data")

	// Get the cipher id
	id := strings.TrimPrefix(req.URL.Path, "/api/ciphers/")

	acc, err := h.db.GetAccount(email, "")
	if err != nil {
		log.Fatal("Account lookup " + err.Error())
	}

	switch req.Method {
	case "GET":
		log.Println("GET Ciphers for " + acc.Id)
		var data []byte
		ciph, err := h.db.GetCipher(acc.Id, id)
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

		err = h.db.UpdateCipher(rCiph, acc.Id, id)
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
		err := h.db.DeleteCipher(acc.Id, id)
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

func (h *APIHandler) HandleSync(w http.ResponseWriter, req *http.Request) {
	email := auth.GetEmail(req)

	log.Println(email + " is trying to sync")

	acc, err := h.db.GetAccount(email, "")

	prof := bw.Profile{
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

	ciphs, err := h.db.GetCiphers(acc.Id)
	if err != nil {
		log.Println(err)
	}

	folders, err := h.db.GetFolders(acc.Id)
	if err != nil {
		log.Println(err)
	}

	Domains := bw.Domains{
		Object:            "domains",
		EquivalentDomains: nil,
		GlobalEquivalentDomains: []bw.GlobalEquivalentDomains{
			bw.GlobalEquivalentDomains{Type: 1, Domains: []string{"youtube.com", "google.com", "gmail.com"}, Excluded: false},
		},
	}

	data := bw.SyncData{
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
func (h *APIHandler) HandleImport(w http.ResponseWriter, req *http.Request) {
	email := auth.GetEmail(req)

	log.Println(email + " is trying to import data")

	acc, err := h.db.GetAccount(email, "")
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

		_, err = h.db.NewCipher(c, acc.Id)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	w.Write([]byte{0x00})
}

func (h *APIHandler) HandleFolder(w http.ResponseWriter, req *http.Request) {
	email := auth.GetEmail(req)

	log.Println(email + " is trying to add a new folder")

	acc, err := h.db.GetAccount(email, "")
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

		folder, err := h.db.AddFolder(folderData.Name, acc.Id)
		if err != nil {
			log.Fatal("newFolder error" + err.Error())
		}

		data, err = json.Marshal(&folder)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		folders, err := h.db.GetFolders(acc.Id)
		if err != nil {
			log.Println(err)
		}
		list := bw.Data{Object: "list", Data: folders}
		data, err = json.Marshal(list)
		if err != nil {
			log.Fatal(err)
		}
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
	Fields         string    `json:"fields"`
}

type loginData struct {
	URI      string `json:"uri"`
	Username string `json:"username"`
	Password string `json:"password"`
	ToTp     string `json:"totp"`
}

func (nciph *newCipher) toCipher() (bw.Cipher, error) {
	// Create new
	cdata := bw.CipherData{
		Uri:      nciph.Login.URI,
		Username: nciph.Login.Username,
		Password: nciph.Login.Password,
		Totp:     nil,
		Name:     nciph.Name,
		Notes:    new(string),
		Fields:   nil,
	}

	(*cdata.Notes) = nciph.Notes

	if *cdata.Notes == "" {
		cdata.Notes = nil
	}

	ciph := bw.Cipher{ // Only including the data we use when we store it
		Type:     nciph.Type,
		Data:     cdata,
		Favorite: nciph.Favorite,
	}

	if nciph.FolderId != "" {
		ciph.FolderId = &nciph.FolderId
	}

	return ciph, nil
}

// unmarshalCipher Take the recived bytes and make it a Cipher struct
func unmarshalCipher(data io.ReadCloser) (bw.Cipher, error) {
	decoder := json.NewDecoder(data)
	var nciph newCipher
	err := decoder.Decode(&nciph)
	if err != nil {
		return bw.Cipher{}, err
	}

	defer data.Close()

	return nciph.toCipher()
}
