package main

import (
	"encoding/json"
	"io"
	"time"
)

type KeyPair struct {
	EncryptedPrivateKey string `json:"encryptedPrivateKey"`
	PublicKey           string `json:"publicKey"`
}

type Account struct {
	Id                 string  `json:"id"`
	Name               string  `json:"name"`
	Email              string  `json:"email"`
	MasterPasswordHash string  `json:"masterPasswordHash"`
	MasterPasswordHint string  `json:"masterPasswordHint"`
	Key                string  `json:"key"`
	KeyPair            KeyPair `json:"keys"`
	RefreshToken       string  `json:"-"`
}

func (acc Account) getProfile() Profile {
	return Profile{
		Id:                 acc.Id,
		Name:               nil,
		Email:              acc.Email,
		EmailVerified:      false,
		Premium:            false,
		Culture:            "en-US",
		TwoFactorEnabled:   false,
		Key:                acc.Key,
		SecurityStamp:      nil,
		Organizations:      make([]string, 0),
		MasterPasswordHint: nil,
		PrivateKey:         acc.KeyPair.EncryptedPrivateKey,
		Object:             "profile",
	}
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

func (nciph *newCipher) toCipher() (Cipher, error) {
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

	if *cdata.Notes == "" {
		cdata.Notes = nil
	}

	ciph := Cipher{ // Only including the data we use when we store it
		Type:     nciph.Type,
		Data:     cdata,
		Favorite: nciph.Favorite,
	}

	if nciph.FolderId != "" {
		ciph.FolderId = &nciph.FolderId
	}

	return ciph, nil
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
	CollectionIds       []string
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

	return nciph.toCipher()
}

type Profile struct {
	Id                 string
	Name               *string
	Email              string
	EmailVerified      bool
	Premium            bool
	MasterPasswordHint *string
	Culture            string
	TwoFactorEnabled   bool
	Key                string
	PrivateKey         string
	SecurityStamp      *string
	Organizations      []string
	Object             string
}

type SyncData struct {
	Profile Profile
	Folders []Folder
	Ciphers []Cipher
	Domains Domains
	Object  string
}

type Domains struct {
	EquivalentDomains       []string
	GlobalEquivalentDomains []GlobalEquivalentDomains
	Object                  string
}

type GlobalEquivalentDomains struct {
	Type     int
	Domains  []string
	Excluded bool
}

type Folder struct {
	Id           string
	Name         string
	Object       string
	RevisionDate time.Time
}

type Data struct {
	Object string
	Data   interface{}
}
