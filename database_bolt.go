package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/leonelquinteros/gorand"

	bolt "github.com/coreos/bbolt"
)

const BUCKET_ACCOUNTS = "Accounts"
const BUCKET_CIPHERS = "Ciphers"
const BUCKET_FOLDERS = "Folders"

type DBBolt struct {
	db *bolt.DB
}

func (db *DBBolt) init() error {
	db.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BUCKET_ACCOUNTS))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte(BUCKET_CIPHERS))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte(BUCKET_FOLDERS))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	return nil
}

func (db *DBBolt) open() error {
	var err error
	db.db, err = bolt.Open("my.db", 0600, nil)
	db.init()
	return err
}

func (db *DBBolt) close() {
	db.db.Close()
}

func (db *DBBolt) getCipher(owner string, ciphID string) (Cipher, error) {
	cipher := Cipher{}
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Ciphers"))
		b = b.Bucket([]byte(owner))

		data := b.Get([]byte(ciphID))
		json.Unmarshal(data, &cipher)
		return nil
	})
	return cipher, err
}

func (db *DBBolt) getCiphers(owner string) ([]Cipher, error) {
	var ciphers []Cipher
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_CIPHERS))
		b = b.Bucket([]byte(owner))
		if b != nil {
			c := b.Cursor()

			for k, v := c.First(); k != nil; k, v = c.Next() {
				var ciph Cipher
				json.Unmarshal(v, &ciph)
				ciph.Object = "cipher"
				ciph.Id = string(k)
				ciphers = append(ciphers, ciph)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(ciphers) < 1 {
		ciphers = make([]Cipher, 0) // Make an empty slice if there are none or android app will crash
	}
	return ciphers, nil
}

func (db *DBBolt) newCipher(ciph Cipher, owner string) (Cipher, error) {
	uuid, err := gorand.UUID()
	if err != nil {
		return ciph, err
	}

	//ciph.RevisionDate = time.Now()
	ciph.Id = uuid
	ciph.Object = "cipher"

	db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Ciphers"))
		b, err := b.CreateBucketIfNotExists([]byte(owner))
		if err != nil {
			return err
		}

		encoded, err := json.Marshal(ciph)
		if err != nil {
			return err
		}
		err = b.Put([]byte(uuid), encoded)
		return err
	})

	return ciph, nil
}

// Important to check that the owner is correct before an update!
func (db *DBBolt) updateCipher(newData Cipher, owner string, ciphID string) error {
	//newData.RevisionDate = time.Now()

	db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Ciphers"))
		b, err := b.CreateBucketIfNotExists([]byte(owner))
		if err != nil {
			return err
		}

		encoded, err := json.Marshal(newData)
		if err != nil {
			return err
		}
		err = b.Put([]byte(ciphID), []byte(encoded))
		return err
	})
	return nil
}

// Important to check that the owner is correct before an update!
func (db *DBBolt) deleteCipher(owner string, ciphID string) error {
	db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Ciphers"))
		b, err := b.CreateBucketIfNotExists([]byte(owner))
		if err != nil {
			return err
		}
		err = b.Delete([]byte(ciphID))
		return err
	})
	return nil
}

func (db *DBBolt) addAccount(acc Account) error {
	if acc.Id == "" {
		uuid, err := gorand.UUID()
		if err != nil {
			return err
		}
		acc.Id = uuid
	}

	db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Accounts"))
		encoded, err := json.Marshal(acc)
		if err != nil {
			return err
		}
		err = b.Put([]byte(acc.Email), []byte(encoded))
		return err
	})
	return nil
}

func (db *DBBolt) updateAccountInfo(acc Account) error {
	return db.addAccount(acc)
}

func (db *DBBolt) getAccount(username string) (Account, error) {
	acc := Account{}
	db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Accounts"))
		v := b.Get([]byte(username))
		json.Unmarshal(v, &acc)
		log.Println(acc)
		return nil
	})
	return acc, nil
}

func (db *DBBolt) addFolder(name string, owner string) (Folder, error) {
	folder := Folder{}
	folder.Name = name

	if folder.Id == "" {
		uuid, err := gorand.UUID()
		if err != nil {
			return folder, err
		}
		folder.Id = uuid
	}

	db.db.Update(func(tx *bolt.Tx) error {
		fmt.Println(BUCKET_FOLDERS)
		b := tx.Bucket([]byte(BUCKET_FOLDERS))
		fmt.Println(owner)
		b, err := b.CreateBucketIfNotExists([]byte(owner))
		if err != nil {
			return err
		}
		encoded, err := json.Marshal(folder)
		if err != nil {
			return err
		}
		err = b.Put([]byte(folder.Id), []byte(encoded))
		return err
	})
	return folder, nil
}

func (db *DBBolt) getFolders(owner string) ([]Folder, error) {
	var folders []Folder
	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKET_FOLDERS))
		b = b.Bucket([]byte(owner))
		if b != nil {
			c := b.Cursor()

			for k, v := c.First(); k != nil; k, v = c.Next() {
				var folder Folder
				json.Unmarshal(v, &folder)
				folder.Object = "folder"
				folder.Id = string(k)
				folders = append(folders, folder)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(folders) < 1 {
		folders = make([]Folder, 0)
	}
	return folders, nil
}
