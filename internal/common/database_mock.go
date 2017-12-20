package common

import _ "github.com/mattn/go-sqlite3"

// mock database used for testing
type MockDB struct {
	Username     string
	Password     string
	RefreshToken string
}

func (db *MockDB) Init() error {
	return nil
}

func (db *MockDB) Open() error {
	return nil
}

func (db *MockDB) Close() {
}

func (db *MockDB) UpdateAccountInfo(acc Account) error {
	return nil
}

func (db *MockDB) GetCipher(owner string, ciphID string) (Cipher, error) {
	return Cipher{}, nil
}

func (db *MockDB) GetCiphers(owner string) ([]Cipher, error) {
	return nil, nil
}

func (db *MockDB) NewCipher(ciph Cipher, owner string) (Cipher, error) {
	return Cipher{}, nil

}

func (db *MockDB) UpdateCipher(newData Cipher, owner string, ciphID string) error {
	return nil
}

func (db *MockDB) DeleteCipher(owner string, ciphID string) error {
	return nil
}

func (db *MockDB) AddAccount(acc Account) error {
	return nil
}

func (db *MockDB) GetAccount(username string, refreshtoken string) (Account, error) {
	return Account{Email: db.Username, MasterPasswordHash: db.Password, RefreshToken: db.RefreshToken}, nil
}

func (db *MockDB) AddFolder(name string, owner string) (Folder, error) {
	return Folder{}, nil
}

func (db *MockDB) GetFolders(owner string) ([]Folder, error) {
	return nil, nil
}
