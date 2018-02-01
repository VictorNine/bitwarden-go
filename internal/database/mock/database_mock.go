package mock

import (
	bw "github.com/VictorNine/bitwarden-go/internal/common"
	_ "github.com/mattn/go-sqlite3"
)

// mock database used for testing
type MockDB struct {
	Username        string
	Password        string
	RefreshToken    string
	TwoFactorSecret string
}

func (db *MockDB) Init() error {
	return nil
}

func (db *MockDB) SetDir(d string) {
}

func (db *MockDB) Open() error {
	return nil
}

func (db *MockDB) Close() {
}

func (db *MockDB) UpdateAccountInfo(acc bw.Account) error {
	return nil
}

func (db *MockDB) GetCipher(owner string, ciphID string) (bw.Cipher, error) {
	return bw.Cipher{}, nil
}

func (db *MockDB) GetCiphers(owner string) ([]bw.Cipher, error) {
	return nil, nil
}

func (db *MockDB) NewCipher(ciph bw.Cipher, owner string) (bw.Cipher, error) {
	return bw.Cipher{}, nil

}

func (db *MockDB) UpdateCipher(newData bw.Cipher, owner string, ciphID string) error {
	return nil
}

func (db *MockDB) DeleteCipher(owner string, ciphID string) error {
	return nil
}

func (db *MockDB) AddAccount(acc bw.Account) error {
	return nil
}

func (db *MockDB) GetAccount(username string, refreshtoken string) (bw.Account, error) {
	return bw.Account{Email: db.Username, MasterPasswordHash: db.Password, RefreshToken: db.RefreshToken, TwoFactorSecret: db.TwoFactorSecret}, nil
}

func (db *MockDB) AddFolder(name string, owner string) (bw.Folder, error) {
	return bw.Folder{}, nil
}

func (db *MockDB) GetFolders(owner string) ([]bw.Folder, error) {
	return nil, nil
}

func (db *MockDB) Update2FAsecret(secret string, email string) error {
	return nil
}
