package main

import (
	"encoding/json"
	"io"
	"time"
)

type Keys struct {
	EncryptedPrivateKey string `json:"encryptedPrivateKey"`
	PublicKey           string `json:"publicKey"`
}

type Account struct {
	Id                 string `json:"id"`
	Name               string `json:"name"`
	Email              string `json:"email"`
	MasterPasswordHash string `json:"masterPasswordHash"`
	MasterPasswordHint string `json:"masterPasswordHint"`
	Key                string `json:"key"`
	Keys               Keys   `json:"keys"`
	RefreshToken       string `json:"-"`
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
		PrivateKey:         "0.2o2tRZZ4twUsLId6tNU+eQ==|54SMNoQC4DdvJOpeG3uQHjAGtpy1H0DlXKFB+W6o1woA7EVp5TAb9d+4N3SoNyF/9nNdhVHmkIVIgjRpN8Wq0Gzb7TqqVwXbO7E7Gt73YZK+UMS9afyqpiNAGnSsyf/7R8KkbHov3PasFTOy9U75OLoeCzNkL0RlCtDlbblNZFGmE6MZtgCTTgGGrVGP82K6o3AZVpkvf+Y83IO9v1B8prqsNYlB8w8JRJ7fYV0yHrnbPfRzDsC9IcNtlXOs8+0omrgUUWiXQdDdhRkT1yHinSCeZNLc6JufJubOoROlux3DLMSpAu+Xgs8y3WThikIIZu/Oh60n9CEJrRJ/T79cKjPbFJhWlAaInliA2BT+yK1LFwgZoozgQ/ppX8vapt8yC6Cia66krZPE57N6WPeqnGwXXrT4L/SKA53Kcdknx3vEERebF5wton6eSejzMJSDye2hmqB8RXYelhYmHEFlcjNRUkDUplUBhQLMHXecZvh8/S6LBbyucy7R4QG1Z/EBVCtodpnerkt4STPS5nOLAUwxYv5HGkue8gIbl7ARXln+kGYlvV2HwnZefqOu+/44S3ELy5jIDlnCvgvk8A1voy7MhbdLCFPYUs+XnPPftiBw757GdsEfXAUku97qSvgy6Sd0uundfukEfBjH0Wrovi8POcaV61C+bvrSM7FYy30M8riEaH/9SfeWqVfciCSsU9mpSQ39RyoPwik89O1vfjjbsJMUp3HrZzkJmvr7rcm83Pc4/3ITyvEkLN3aYdEAgL/QIOGW8sSWxfotG+DtdEMULBi2qvX73XzMK893uqGR1CMtB/KcdDmlPpCEPYtGB6P2380fnyxLHD6EXtNHhl9v0ZWlwoFtUo+RNcn0XkXB0yVOLzQDAM2Ovxs1JHD/AxQRAnh/RaYm0miWQe6wgfIcavZodqLdhDf4lMXdpPRGL5YfOXua4z50ul0fl3erz3pv8T2YnP0uHHJVf0VLmzlL57jRCBHRldXSfbtYOXOlalcqTajIx8ONcaeEUvc40tJ+J43O5BjGQdt8dIM1zAgWDyWGWU6Im73IFD7EFsrbeCqneGqpvoChY6LirLCLgVlBSf4WlpLaGvDPzJHf/Ss+skyseUf3ahoeMlotqOeyRFCmi4zknfo1THgqUiKcuzKELTENkZmF7ra94f9VUOVMjacfD9EpweG/pLaZkdOxPf/MABIJ/gJ9Rv30XLSZdIvkKwm4LlCzLXjyBu3rr9HdknzUQN/GeeBKSMuu1f8P2BIV9ytS0bX0Ap/CsUiJeNXKXzLpwPPurKM8/7m4NGGLmod39K2H0vRdeZHJB1/TwCX6VAO6dGv/nAyq4kC40Pu3UCw/9BMIm8fK+vUMDBIbuieJxArLLuxo+gaQqhTL/ZYFA62CUFdDTWRwVHoJw9io8VxxxXcxqMK1QatDyHAWbBM0Y454gfDPYbq7IRGwc2P3xixX5gY730YCJMrQ35uzBQ2I4rkk9znH1BDAKynDQuUCdkcM+eCh6wPVq/SNL6DwWtqanbK8X7FGWUp8woKY5Twq0C95O57drB6fiRGinXBPm7WYWyV3ccZSiyFGNadsVC3HG/R+ExGTtwkC8LH+Ek1GgskwWY83D0Gv8yUbVTS6N3iXUkDmGjBswV0=",
		Object:             "profile",
	}
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
		Type: nciph.Type,
		Data: cdata,
	}

	return ciph, nil
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
