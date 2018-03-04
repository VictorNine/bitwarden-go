package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gitlab.com/Odysseus16/bitwarden-go/internal/auth"
	bw "gitlab.com/Odysseus16/bitwarden-go/internal/common"
)

type APIHandler struct {
	db database
}

func New(db database) APIHandler {
	h := APIHandler{
		db: db,
	}

	return h
}

// Interface to make testing easier
type database interface {
	GetAccount(username string, refreshtoken string) (bw.Account, error)
	UpdateAccountInfo(acc bw.Account) error
	GetCipher(owner string, ciphID string) (bw.Cipher, error)
	GetCiphers(owner string) ([]bw.Cipher, error)
	NewCipher(ciph bw.Cipher, owner string) (bw.Cipher, error)
	UpdateCipher(newData bw.Cipher, owner string, ciphID string) error
	DeleteCipher(owner string, ciphID string) error
	AddFolder(name string, owner string) (bw.Folder, error)
	UpdateFolder(newFolder bw.Folder, owner string) error
	GetFolders(owner string) ([]bw.Folder, error)
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
func HandleAttachments(w http.ResponseWriter, req *http.Request) {

	path := strings.TrimPrefix(req.URL.Path, "/")
	Openfile, err := os.Open(path)
	defer Openfile.Close()
	if err != nil {
		log.Println(err)
	}
	Openfile.Seek(0, 0)
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, Openfile)
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
		var attachments = bw.Attachments{}
		rCiph.Attachments = attachments
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
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

// This function handles updates and deleteing
func (h *APIHandler) HandleCipherUpdate(w http.ResponseWriter, req *http.Request) {
	//type Attachments []bw.AttachmentData
	var attachment bool
	email := auth.GetEmail(req)
	log.Println(email + " is trying to edit his data")

	// Get the cipher id
	id := strings.TrimPrefix(req.URL.Path, "/api/ciphers/")
	attachment = strings.Contains(req.URL.Path, "attachment") //check if attachment is added
	acc, err := h.db.GetAccount(email, "")
	if err != nil {
		log.Fatal("Account lookup " + err.Error())
	}
	switch attachment {
	case true:
		id2 := strings.SplitAfter(id, "/")
		log.Println(id2)
		id = strings.TrimSuffix(id2[0], "/")
		log.Println(id)
		ciph, err2 := h.db.GetCipher(acc.Id, id)
		if err2 != nil {
			log.Fatal(err)
		}
		switch req.Method {
		case "DELETE":
			var attachments = bw.Attachments{}
			ciph2 := bw.Cipher{
				Type:                ciph.Type,
				FolderId:            ciph.FolderId,
				OrganizationId:      ciph.OrganizationId,
				Favorite:            ciph.Favorite,
				Id:                  id,
				Data:                ciph.Data,
				Attachments:         attachments,
				OrganizationUseTotp: ciph.OrganizationUseTotp,
				RevisionDate:        ciph.RevisionDate,
				Object:              "cipher",
			}
			err = h.db.UpdateCipher(ciph2, acc.Id, id)
			if err != nil {
				w.Write([]byte("0"))
				log.Println(err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("0"))
			log.Println("Attachment deleted")
		default:
			file, header, err := req.FormFile("data")
			if err != nil {
				log.Println("[-] Error in req.FormFile ", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "{'error': %s}", err)
				return
			}
			defer file.Close()
			token, err1 := GenerateRandomString(15)
			if err1 != nil {
			}
			log.Println(token)
			log.Println(header.Filename)
			//String value: http://localhost:4567/attachments/4f3e0842-210c-4549-b7fc-4986ffd7d031/ece4f584289eec7f13d74f41498dd8b5
			os.MkdirAll("./attachments/"+id, os.ModePerm)
			f, err3 := os.OpenFile("./attachments/"+id+"/"+token, os.O_WRONLY|os.O_CREATE, 0666)
			if err3 != nil {
				fmt.Println(err)
				return
			}
			defer f.Close()
			io.Copy(f, file)
			attach := bw.AttachmentData{
				Id:       token,
				Url:      "http://localhost:8000/attachments/" + id + "/" + token,
				FileName: header.Filename,
				Size:     header.Size,
				SizeName: strconv.FormatInt(header.Size, 10) + " Bytes",
				Object:   "attachment",
			}
			var attachments = bw.Attachments{
				attach,
			}
			chip2 := bw.Cipher{
				Type:                ciph.Type,
				FolderId:            ciph.FolderId,
				OrganizationId:      ciph.OrganizationId,
				Favorite:            ciph.Favorite,
				Id:                  id,
				Data:                ciph.Data,
				Attachments:         attachments,
				OrganizationUseTotp: ciph.OrganizationUseTotp,
				RevisionDate:        ciph.RevisionDate,
				Object:              "cipher",
			}
			data, err := json.Marshal(&chip2)
			if err != nil {
				log.Fatal(err)
			}
			err = h.db.UpdateCipher(chip2, acc.Id, id)
			if err != nil {
				w.Write([]byte("0"))
				log.Println(err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		}
	case false:
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
			/**err = h.db.UpdateCipher(rCiph, acc.Id, id)
			if err != nil {
				w.Write([]byte("0"))
				log.Println(err)
				return
			}**/
			ciph3, err := h.db.GetCipher(acc.Id, id)
			if err != nil {
				log.Fatal(err)
			}

			// Send response
			data, err := json.Marshal(&ciph3)
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
	default:
	}

}

func (h *APIHandler) HandleSync(w http.ResponseWriter, req *http.Request) {
	email := auth.GetEmail(req)

	log.Println(email + " is trying to sync")

	acc, err := h.db.GetAccount(email, "")

	prof := bw.Profile{
		Id:               acc.Id,
		Email:            acc.Email,
		EmailVerified:    true,
		Premium:          true,
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

func (h *APIHandler) HandleFolderUpdate(w http.ResponseWriter, req *http.Request) {
	email := auth.GetEmail(req)

	log.Println(email + " is trying to update a folder")

	acc, err := h.db.GetAccount(email, "")
	if err != nil {
		log.Fatal("Account lookup " + err.Error())
	}

	switch req.Method {
	case "POST":
		fallthrough // Do same as PUT. Web Vault wants to post
	case "PUT":
		// Get the folder id
		folderID := strings.TrimPrefix(req.URL.Path, "/api/folders/")

		decoder := json.NewDecoder(req.Body)

		var folderData struct {
			Name string `json:"name"`
		}

		err := decoder.Decode(&folderData)
		if err != nil {
			log.Fatal(err)
		}
		defer req.Body.Close()

		newFolder := bw.Folder{
			Id:           folderID,
			Name:         folderData.Name,
			RevisionDate: time.Now().UTC(),
			Object:       "folder",
		}

		err = h.db.UpdateFolder(newFolder, acc.Id)
		if err != nil {
			w.Write([]byte("0"))
			log.Println(err)
			return
		}

		// Send response
		data, err := json.Marshal(&newFolder)
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		log.Println("Folder " + folderID + " updated")
		return
	}
	w.Header().Set("Content-Type", "application/json")
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
		Uri:      &nciph.Login.URI,
		Username: &nciph.Login.Username,
		Password: &nciph.Login.Password,
		Totp:     nil,
		Name:     &nciph.Name,
		Notes:    new(string),
		Fields:   nil,
	}

	(*cdata.Notes) = nciph.Notes

	if *cdata.Notes == "" {
		cdata.Notes = nil
	}

	if *cdata.Uri == "" {
		cdata.Uri = nil
	}

	if *cdata.Username == "" {
		cdata.Username = nil
	}

	if *cdata.Password == "" {
		cdata.Password = nil
	}

	if *cdata.Name == "" {
		cdata.Name = nil
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

//87f11661-f334-4043-af8a-a884017f0a11
//7dce1c01-110e-11e8-9445-a0999b1c5a79
