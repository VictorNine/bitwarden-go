package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
}

func (db *DB) init() error {
	query1 := "CREATE TABLE \"accounts\" ( `id` INTEGER, `name` TEXT, `email` TEXT UNIQUE, `masterPasswordHash` NUMERIC, `masterPasswordHint` TEXT, `key` TEXT, 'refreshtoken' TEXT, PRIMARY KEY(id) );"
	query2 := "CREATE TABLE \"ciphers\" ( `id` INTEGER PRIMARY KEY AUTOINCREMENT, `type` INTEGER, `revisiondate` INTEGER, `data` BLOB, `owner` INTEGER );"
	stmt1, err := db.db.Prepare(query1)
	if err != nil {
		return err
	}

	_, err = stmt1.Exec()
	if err != nil {
		return err
	}

	stmt2, err := db.db.Prepare(query2)
	if err != nil {
		return err
	}

	_, err = stmt2.Exec()
	if err != nil {
		return err
	}
	return err
}

func (db *DB) open() error {
	var err error
	db.db, err = sql.Open("sqlite3", "db")
	return err
}

func (db *DB) close() {
	db.db.Close()
}

func (db *DB) getCiphers(owner string) ([]Cipher, error) {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return nil, err
	}

	var ciphers []Cipher
	query := "SELECT id, type, revisiondate, data FROM ciphers WHERE owner = $1"
	rows, err := db.db.Query(query, iowner)

	var iid int
	var revDate int64
	var blob []byte
	for rows.Next() {
		ciph := Cipher{
			Favorite:            false,
			Edit:                true,
			OrganizationUseTotp: false,
			Object:              "cipher",
			Attachments:         nil,
		}

		err := rows.Scan(&iid, &ciph.Type, &revDate, &blob)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(blob, &ciph.Data)
		ciph.Id = strconv.Itoa(iid)
		ciph.RevisionDate = time.Unix(revDate, 0)

		if *ciph.Data.Notes == "" {
			ciph.Data.Notes = nil
		}

		ciphers = append(ciphers, ciph)
	}

	if len(ciphers) < 1 {
		ciphers = make([]Cipher, 0) // Make an empty slice if there are none or android app will crash
	}
	return ciphers, err
}

func (db *DB) newCipher(ciph Cipher, owner string) (Cipher, error) {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return Cipher{}, err
	}

	ciph.RevisionDate = time.Now()

	stmt, err := db.db.Prepare("INSERT INTO ciphers(type, revisiondate, data, owner) values(?,?,?, ?)")
	if err != nil {
		return ciph, err
	}

	data, err := ciph.Data.bytes()
	if err != nil {
		return ciph, err
	}

	res, err := stmt.Exec(ciph.Type, ciph.RevisionDate.Unix(), data, iowner)
	if err != nil {
		return ciph, err
	}

	lID, err := res.LastInsertId()
	ciph.Id = fmt.Sprintf("%v", lID)

	return ciph, nil

}

// Important to check that the owner is correct before an update!
func (db *DB) updateCipher(newData Cipher, owner string, ciphID string) error {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return err
	}

	iciphID, err := strconv.ParseInt(ciphID, 10, 64)
	if err != nil {
		return err
	}

	stmt, err := db.db.Prepare("UPDATE ciphers SET type=$1, revisiondate=$2, data=$3 WHERE id=$4 AND owner=$5")
	if err != nil {
		return err
	}

	bdata, err := newData.Data.bytes()
	if err != nil {
		return err
	}

	_, err = stmt.Exec(newData.Type, time.Now().Unix(), bdata, iciphID, iowner)
	if err != nil {
		return err
	}

	return nil
}

// Important to check that the owner is correct before an update!
func (db *DB) deleteCipher(owner string, ciphID string) error {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return err
	}

	iciphID, err := strconv.ParseInt(ciphID, 10, 64)
	if err != nil {
		return err
	}

	stmt, err := db.db.Prepare("DELETE from ciphers WHERE id=$1 AND owner=$2")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(iciphID, iowner)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) addAccount(acc Account) error {
	stmt, err := db.db.Prepare("INSERT INTO accounts(name, email, masterPasswordHash, masterPasswordHint, key, refreshtoken) values(?,?,?,?,?, ?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(acc.Name, acc.Email, acc.MasterPasswordHash, acc.MasterPasswordHint, acc.Key, "")
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) updateAccountInfo(sid string, refreshToken string) error {
	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		return err
	}

	stmt, err := db.db.Prepare("UPDATE accounts SET refreshtoken=$1 WHERE id=$2")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(refreshToken, id)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) getAccount(username string, refreshtoken string) (Account, error) {
	var row *sql.Row
	acc := Account{}
	if username != "" {
		query := "SELECT * FROM accounts WHERE email = $1"
		row = db.db.QueryRow(query, username)
	}
	if refreshtoken != "" {
		query := "SELECT * FROM accounts WHERE refreshtoken = $1"
		row = db.db.QueryRow(query, refreshtoken)
	}

	var iid int
	err := row.Scan(&iid, &acc.Name, &acc.Email, &acc.MasterPasswordHash, &acc.MasterPasswordHint, &acc.Key, &acc.RefreshToken)
	if err != nil {
		return acc, err
	}

	acc.Id = strconv.Itoa(iid)

	return acc, nil
}
