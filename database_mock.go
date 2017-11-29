package main

import _ "github.com/mattn/go-sqlite3"

// mock database used for testing
type mockDB struct {
	username     string
	password     string
	refreshToken string
}

func (db *mockDB) init() error {
	return nil
}

func (db *mockDB) open() error {
	return nil
}

func (db *mockDB) close() {
}

func (db *mockDB) updateAccountInfo(sid string, refreshToken string) error {
	return nil
}

func (db *mockDB) getCiphers(owner string) ([]Cipher, error) {
	return nil, nil
}

func (db *mockDB) newCipher(ciph Cipher, owner string) (Cipher, error) {
	return Cipher{}, nil

}

func (db *mockDB) updateCipher(newData Cipher, owner string, ciphID string) error {
	return nil
}

func (db *mockDB) deleteCipher(owner string, ciphID string) error {
	return nil
}

func (db *mockDB) addAccount(acc Account) error {
	return nil
}

func (db *mockDB) getAccount(username string, refreshtoken string) (Account, error) {
	return Account{Email: db.username, MasterPasswordHash: db.password, RefreshToken: db.refreshToken}, nil
}

func (db *mockDB) addFolder(name string, owner string) (Folder, error) {
	return Folder{}, nil
}
