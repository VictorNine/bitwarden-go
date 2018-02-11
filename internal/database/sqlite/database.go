package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"
	"time"

	bw "github.com/VictorNine/bitwarden-go/internal/common"
	_ "github.com/mattn/go-sqlite3"
	uuid "github.com/satori/go.uuid"
)

type DB struct {
	db  *sql.DB
	dir string
}

const acctTbl = `
CREATE TABLE IF NOT EXISTS "accounts" (
  id                  INTEGER,
  name                TEXT,
  email               TEXT UNIQUE,
  masterPasswordHash  NUMERIC,
  masterPasswordHint  TEXT,
  key                 TEXT,
  refreshtoken        TEXT,
  privatekey          TEXT NOT NULL,
  pubkey              TEXT NOT NULL,
  tfasecret           TEXT NOT NULL,
PRIMARY KEY(id)
)`

const ciphersTbl = `
CREATE TABLE IF NOT EXISTS "ciphers" (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  type         INT,
  revisiondate INT,
  data         REAL,
  owner        INT,
  folderid     TEXT,
  favorite     INT NOT NULL
)
`
const foldersTbl = `
CREATE TABLE IF NOT EXISTS "folders" (
  id           TEXT,
  name         TEXT,
  revisiondate INTEGER,
  owner        INTEGER,
PRIMARY KEY(id)
)
`

func (db *DB) Init() error {
	for _, sql := range []string{acctTbl, ciphersTbl, foldersTbl} {
		if _, err := db.db.Exec(sql); err != nil {
			return errors.New(fmt.Sprintf("SQL error with %s\n%s", sql, err.Error()))
		}
	}
	return nil
}

func (db *DB) SetDir(d string) {
	db.dir = d
}

func (db *DB) Open() error {
	var err error
	if db.dir != "" {
		db.db, err = sql.Open("sqlite3", path.Join(db.dir, "db"))
	} else {
		db.db, err = sql.Open("sqlite3", "db")
	}
	return err
}

func (db *DB) Close() {
	db.db.Close()
}

func sqlRowToCipher(row interface {
	Scan(dest ...interface{}) error
}) (bw.Cipher, error) {
	ciph := bw.Cipher{
		Favorite:            false,
		Edit:                true,
		OrganizationUseTotp: false,
		Object:              "cipher",
		Attachments:         nil,
		FolderId:            nil,
	}

	var iid, favorite int
	var revDate int64
	var blob []byte
	var folderid sql.NullString
	err := row.Scan(&iid, &ciph.Type, &revDate, &blob, &folderid, &favorite)
	if err != nil {
		return ciph, err
	}

	err = json.Unmarshal(blob, &ciph.Data)
	if err != nil {
		return ciph, err
	}

	if favorite == 1 {
		ciph.Favorite = true
	}

	ciph.Id = strconv.Itoa(iid)
	ciph.RevisionDate = time.Unix(revDate, 0)
	if folderid.Valid {
		ciph.FolderId = &folderid.String
	}
	return ciph, nil
}

func (db *DB) GetCipher(owner string, ciphID string) (bw.Cipher, error) {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return bw.Cipher{}, err
	}
	iciphID, err := strconv.ParseInt(ciphID, 10, 64)
	if err != nil {
		return bw.Cipher{}, err
	}

	query := "SELECT id, type, revisiondate, data, folderid, favorite FROM ciphers WHERE owner = $1 AND id = $2"
	row := db.db.QueryRow(query, iowner, iciphID)

	return sqlRowToCipher(row)
}

func (db *DB) GetCiphers(owner string) ([]bw.Cipher, error) {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return nil, err
	}

	var ciphers []bw.Cipher
	query := "SELECT id, type, revisiondate, data, folderid, favorite FROM ciphers WHERE owner = $1"
	rows, err := db.db.Query(query, iowner)

	for rows.Next() {
		ciph, err := sqlRowToCipher(rows)
		if err != nil {
			return nil, err
		}

		ciphers = append(ciphers, ciph)
	}

	if len(ciphers) < 1 {
		ciphers = make([]bw.Cipher, 0) // Make an empty slice if there are none or android app will crash
	}
	return ciphers, err
}

func (db *DB) NewCipher(ciph bw.Cipher, owner string) (bw.Cipher, error) {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return bw.Cipher{}, err
	}

	ciph.RevisionDate = time.Now()

	stmt, err := db.db.Prepare("INSERT INTO ciphers(type, revisiondate, data, owner,folderid, favorite) values(?,?,?, ?, ?, ?)")
	if err != nil {
		return ciph, err
	}

	data, err := ciph.Data.Bytes()
	if err != nil {
		return ciph, err
	}

	res, err := stmt.Exec(ciph.Type, ciph.RevisionDate.Unix(), data, iowner, ciph.FolderId, 0)
	if err != nil {
		return ciph, err
	}

	lID, err := res.LastInsertId()
	ciph.Id = fmt.Sprintf("%v", lID)

	return ciph, nil

}

// Important to check that the owner is correct before an update!
func (db *DB) UpdateCipher(newData bw.Cipher, owner string, ciphID string) error {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return err
	}

	iciphID, err := strconv.ParseInt(ciphID, 10, 64)
	if err != nil {
		return err
	}

	favorite := 0
	if newData.Favorite {
		favorite = 1
	}

	stmt, err := db.db.Prepare("UPDATE ciphers SET type=$1, revisiondate=$2, data=$3, folderid=$4, favorite=$5 WHERE id=$6 AND owner=$7")
	if err != nil {
		return err
	}

	bdata, err := newData.Data.Bytes()
	if err != nil {
		return err
	}

	_, err = stmt.Exec(newData.Type, time.Now().Unix(), bdata, newData.FolderId, favorite, iciphID, iowner)
	if err != nil {
		return err
	}

	return nil
}

// Important to check that the owner is correct before an update!
func (db *DB) DeleteCipher(owner string, ciphID string) error {
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

func (db *DB) AddAccount(acc bw.Account) error {
	stmt, err := db.db.Prepare("INSERT INTO accounts(name, email, masterPasswordHash, masterPasswordHint, key, refreshtoken, privatekey, pubkey, tfasecret) values(?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(acc.Name, acc.Email, acc.MasterPasswordHash, acc.MasterPasswordHint, acc.Key, "", "", "", "")
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) UpdateAccountInfo(acc bw.Account) error {
	id, err := strconv.ParseInt(acc.Id, 10, 64)
	if err != nil {
		return err
	}

	stmt, err := db.db.Prepare("UPDATE accounts SET refreshtoken=$1, privatekey=$2, pubkey=$3 WHERE id=$4")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(acc.RefreshToken, acc.KeyPair.EncryptedPrivateKey, acc.KeyPair.PublicKey, id)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetAccount(username string, refreshtoken string) (bw.Account, error) {
	var row *sql.Row
	acc := bw.Account{}
	acc.KeyPair = bw.KeyPair{}
	if username != "" {
		query := "SELECT * FROM accounts WHERE email = $1"
		row = db.db.QueryRow(query, username)
	}

	if refreshtoken != "" {
		query := "SELECT * FROM accounts WHERE refreshtoken = $1"
		row = db.db.QueryRow(query, refreshtoken)
	}

	var iid int
	err := row.Scan(&iid, &acc.Name, &acc.Email, &acc.MasterPasswordHash, &acc.MasterPasswordHint, &acc.Key, &acc.RefreshToken, &acc.KeyPair.EncryptedPrivateKey, &acc.KeyPair.PublicKey, &acc.TwoFactorSecret)
	if err != nil {
		return acc, err
	}

	acc.Id = strconv.Itoa(iid)

	return acc, nil
}

func (db *DB) AddFolder(name string, owner string) (bw.Folder, error) {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return bw.Folder{}, err
	}

	newFolderID, err := uuid.NewV4()
	if err != nil {
		return bw.Folder{}, err
	}

	folder := bw.Folder{
		Id:           newFolderID.String(),
		Name:         name,
		Object:       "folder",
		RevisionDate: time.Now(),
	}

	stmt, err := db.db.Prepare("INSERT INTO folders(id, name, revisiondate, owner) values(?,?,?, ?)")
	if err != nil {
		return bw.Folder{}, err
	}

	_, err = stmt.Exec(folder.Id, folder.Name, folder.RevisionDate.Unix(), iowner)
	if err != nil {
		return bw.Folder{}, err
	}

	return folder, nil
}

func (db *DB) UpdateFolder(newFolder bw.Folder, owner string) error {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return err
	}

	stmt, err := db.db.Prepare("UPDATE folders SET name=$1, revisiondate=$2 WHERE id=$3 AND owner=$4")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(newFolder.Name, newFolder.RevisionDate.Unix(), newFolder.Id, iowner)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetFolders(owner string) ([]bw.Folder, error) {
	iowner, err := strconv.ParseInt(owner, 10, 64)
	if err != nil {
		return nil, err
	}

	var folders []bw.Folder
	query := "SELECT id, name, revisiondate FROM folders WHERE owner = $1"
	rows, err := db.db.Query(query, iowner)
	if err != nil {
		return nil, err
	}

	var revDate int64
	for rows.Next() {
		f := bw.Folder{}
		err := rows.Scan(&f.Id, &f.Name, &revDate)
		if err != nil {
			return nil, err
		}
		f.RevisionDate = time.Unix(revDate, 0)

		folders = append(folders, f)
	}

	if len(folders) < 1 {
		folders = make([]bw.Folder, 0) // Make an empty slice if there are none or android app will crash
	}
	return folders, err
}

func (db *DB) Update2FAsecret(secret string, email string) error {
	stmt, err := db.db.Prepare("UPDATE accounts SET tfasecret=$1 WHERE email=$2")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(secret, email)
	if err != nil {
		return err
	}

	return nil
}
