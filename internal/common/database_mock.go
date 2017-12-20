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

func (db *mockDB) updateAccountInfo(acc Account) error {
	return nil
}

func (db *mockDB) getCipher(owner string, ciphID string) (Cipher, error) {
	return Cipher{}, nil
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

func (db *mockDB) getFolders(owner string) ([]Folder, error) {
	return nil, nil
}
