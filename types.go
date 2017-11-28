package main

import (
	"encoding/json"
	"io"
	"time"
)

type Account struct {
	Id                 string `json:"-"`
	Name               string `json:"name"`
	Email              string `json:"email"`
	MasterPasswordHash string `json:"masterPasswordHash"`
	MasterPasswordHint string `json:"masterPasswordHint"`
	Key                string `json:"key"`
	RefreshToken       string `json:"-"`
}

// The data we store and send to the client
type Cipher struct {
	Type                int
	FolderId            *string // Must be pointer to output null in json. Android app will crash if not null
	OrganizationId      *string
	Favorite            bool
	Edit                bool
	Id                  string
	Data                CipherData
	Attachments         []string
	OrganizationUseTotp bool
	RevisionDate        time.Time
	Object              string
}

type CipherData struct {
	Uri      string
	Username string
	Password string
	Totp     *string // Must be pointer to output null in json. Android app will crash if not null
	Name     string
	Notes    *string // Must be pointer to output null in json. Android app will crash if not null
	Fields   []string
}

func (data *CipherData) bytes() ([]byte, error) {
	b, err := json.Marshal(&data)
	return b, err
}

// unmarshalCipher Take the recived bytes and make it a Cipher struct
func unmarshalCipher(data io.ReadCloser) (Cipher, error) {
	decoder := json.NewDecoder(data)
	var nciph newCipher
	err := decoder.Decode(&nciph)
	if err != nil {
		return Cipher{}, err
	}

	defer data.Close()

	// Create new
	cdata := CipherData{
		Uri:      nciph.Login.URI,
		Username: nciph.Login.Username,
		Password: nciph.Login.Password,
		Totp:     nil,
		Name:     nciph.Name,
		Notes:    new(string),
		Fields:   nil,
	}

	(*cdata.Notes) = nciph.Notes

	ciph := Cipher{ // Only including the data we use when we store it
		Type: nciph.Type,
		Data: cdata,
	}

	return ciph, nil
}

type Profile struct {
	Id                 string
	Name               string
	Email              string
	EmailVerified      bool
	Premium            bool
	MasterPasswordHint string
	Culture            string
	TwoFactorEnabled   bool
	Key                string
	PrivateKey         string
	SecurityStamp      string
	Organizations      []string
	Object             string
}

type SyncData struct {
	Profile Profile
	Folders []string
	Ciphers []Cipher
	Domains Domains
	Object  string
}

type Domains struct {
	EquivalentDomains       []string
	GlobalEquivalentDomains []GlobalEquivalentDomains
}

type GlobalEquivalentDomains struct {
	Type     int
	Domains  []string
	Excluded bool
}
