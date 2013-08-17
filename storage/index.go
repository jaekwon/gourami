package storage

import (
    "os"
	"database/sql"
	"errors"
	"fmt"
    "strconv"
    _ "github.com/mattn/go-sqlite3"
    "github.com/jaekwon/gourami/types"
    "github.com/jaekwon/go-prelude/colors"
)

var ErrNotFound error = errors.New("Not found in index")
var ErrSchemaUnknown error = errors.New("Schema unknown")

var CurrentSchemaVersion = 1

var MetaSchemaVersion string = "meta:schema_version"

type Index struct {
	DB *sql.DB
}

func (this *Index) SchemaVersion() (int, error) {
    value, err := this.Get(MetaSchemaVersion)
	if err != nil {
        return -1, ErrSchemaUnknown
	}
	return strconv.Atoi(value)
}

// called by NewIndex
func (this *Index) initialize() error {

    // nothing to do if schema version is 1
    schemaVersion, err := this.SchemaVersion()
    if err == nil {
        if schemaVersion == CurrentSchemaVersion {
            return nil
        } else {
            // include upgrade path here.
            return errors.New("Schema version unrecognized")
        }
    }
    if err != nil && err != ErrSchemaUnknown {
        return err
    }

    // create new index table...

    fmt.Println(colors.Blue("Initializing Index"))
    _, err = this.DB.Exec(
    `CREATE TABLE kv (
        k VARCHAR(255) NOT NULL PRIMARY KEY,
        v VARCHAR(255)
    )`)
    if err != nil { return err }
    _, err = this.DB.Exec(
    `CREATE TABLE items (
        counter INTEGER PRIMARY KEY AUTOINCREMENT,
        id VARCHAR(44) NOT NULL
    )`)
    if err != nil { return err }

    // set schema version
    err = this.Set(MetaSchemaVersion, strconv.Itoa(CurrentSchemaVersion))

    return err
}

func (this *Index) Transaction() (*Transaction, error) {
	tx, err := this.DB.Begin()
	return &Transaction{tx}, err
}

func (this *Index) Get(key string) (value string, err error) {
	err = this.DB.QueryRow("SELECT v FROM kv WHERE k=?", key).Scan(&value)
	if err == sql.ErrNoRows {
		err = ErrNotFound
	}
	return
}

func (this *Index) Set(key, value string) error {
	_, err := this.DB.Exec("REPLACE INTO kv (k, v) VALUES (?, ?)", key, value)
	return err
}

func (this *Index) Delete(key string) error {
	_, err := this.DB.Exec("DELETE FROM kv WHERE k=?", key)
	return err
}

func (this *Index) Find(key string, limit int, ch chan KeyValueErr) {
    defer close(ch)
    rows, err := this.DB.Query("SELECT k, v FROM kv WHERE k >= ? ORDER BY k LIMIT ?", key, limit)
    if err != nil {
        ch <- KeyValueErr{"", "", err}
        return
    }
    for rows.Next() {
        var kvErr KeyValueErr
        kvErr.Err = rows.Scan(&kvErr.Key, &kvErr.Value)
        ch <- kvErr
    }
    return
}

type KeyValueErr struct {
    Key string
    Value string
    Err error
}

func (this *Index) AddItem(id types.Id) (lastInsertId int64, err error) {
    idString, err := id.ToString()
    if err != nil { return -1, err }
    result, err := this.DB.Exec("INSERT INTO items (id) VALUES (?)", idString)
    if err != nil { return -1, err }
    return result.LastInsertId()
}

func (this *Index) FindItems(start int64, limit int, ch chan IdErr) {
    defer close(ch)
    rows, err := this.DB.Query("SELECT item FROM items WHERE id >= ? ORDER BY k LIMIT ?", start, limit)
    if err != nil {
        ch <- IdErr{nil, err}
        return
    }
    for rows.Next() {
        var idErr IdErr
        var idString string
        err = rows.Scan(&idString)
        if err != nil {
            ch <- IdErr{nil, err}
        }
        idErr.Id, idErr.Err = types.StringToId(idString)
        ch <- idErr
    }
    return
}

type IdErr struct {
    Id types.Id
    Err error
}

func NewIndex(file string) (*Index, error) {
    _, err := os.Create(file)
    if err != nil { return nil, err }
	db, err := sql.Open("sqlite3", file)
    if err != nil { return nil, err }
    index := &Index{db}
    err = index.initialize()
    if err != nil { return nil, err }
    return index, nil
}



type Transaction struct {
	tx  *sql.Tx
}

func (this *Transaction) Set(key, value string) error {
	_, err := this.tx.Exec("REPLACE INTO kv (k, v) VALUES (?, ?)", key, value)
    return err
}

func (this *Transaction) Delete(key string) error {
	_, err := this.tx.Exec("DELETE FROM kv WHERE k=?", key)
    return err
}

func (this *Transaction) AddItem(id types.Id) error {
    idString, err := id.ToString()
    if err != nil { return err }
    _, err = this.tx.Exec("INSERT INTO items (id) VALUES (?)", idString)
    return err
}

func (this *Transaction) Commit() error {
	return this.tx.Commit()
}
