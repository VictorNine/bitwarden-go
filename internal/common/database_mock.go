package common

import _ "github.com/mattn/go-sqlite3"

// mock database used for testing
type mockDB struct {
	username     string
	password     string
	refreshToken string
}

func (db *mockDB) Init() error {
	return nil
}

func (db *mockDB) Open() error {
	return nil
}

func (db *mockDB) Close() {
}

func (db *mockDB) UpdateAccountInfo(acc Account) error {
	return nil
}

func (db *mockDB) GetCipher(owner string, ciphID string) (Cipher, error) {
	return Cipher{}, nil
}

func (db *mockDB) GetCiphers(owner string) ([]Cipher, error) {
	return nil, nil
}

func (db *mockDB) NewCipher(ciph Cipher, owner string) (Cipher, error) {
	return Cipher{}, nil

}

func (db *mockDB) UpdateCipher(newData Cipher, owner string, ciphID string) error {
	return nil
}

func (db *mockDB) DeleteCipher(owner string, ciphID string) error {
	return nil
}

func (db *mockDB) AddAccount(acc Account) error {
	return nil
}

func (db *mockDB) GetAccount(username string, refreshtoken string) (Account, error) {
	return Account{Email: db.username, MasterPasswordHash: db.password, RefreshToken: db.refreshToken}, nil
}

func (db *mockDB) AddFolder(name string, owner string) (Folder, error) {
	return Folder{}, nil
}

func (db *mockDB) GetFolders(owner string) ([]Folder, error) {
	return nil, nil
}
