package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
    "github.com/mattn/go-sqlite3"
)

ErrNotFound := errors.New("Not found in index")

type Index struct {
	DB *sql.DB
}

func (this *Index) Transaction() (*Transaction, error) {
	tx, err := this.DB.Begin()
	return &Transaction{tx}
}

func (this *Index) Get(key string) (value string, err error) {
	err = this.DB.QueryRow("SELECT v FROM rows WHERE k=?", key).Scan(&value)
	if err == sql.ErrNoRows {
		err = ErrNotFound
	}
	return
}

func (this *Index) Set(key, value string) error {
	_, err := this.DB.Exec("REPLACE INTO rows (k, v) VALUES (?, ?)", key, value)
	return err
}

func (this *Index) Delete(key string) error {
	_, err := this.DB.Exec("DELETE FROM rows WHERE k=?", key)
	return err
}

func (this *Index) Find(key string, ch chan KeyValue) {
    defer close(ch)
    const batchSize = 50
    rows, err := this.DB.Query(
        "SELECT k, v FROM rows WHERE k >= ? ORDER BY k LIMIT ?",
        key, batchSize
    )
    if err != nil {
        ch <- KeyValue{"", "", err}
        return
    }
    for rows.Next() {
        var kv KeyValue
        kv.Err = rows.Scan(&kv.Key, &kv.Value)
        ch <- kv
    }
    return
}

func NewIndex(file string) (*Index, error) {
	db, err := sql.Open("sqlite3", file)
    return &Index{db}, err
}

type KeyValue struct {
    Key string
    Value string
    Err error
}


type Transaction struct {
	tx  *sql.Tx
}

func (this *Transaction) Set(key, value string) error {
	_, err := this.tx.Exec("REPLACE INTO rows (k, v) VALUES (?, ?)", key, value)
    return err
}

func (this *Transaction) Delete(key string) {
	_, err := this.tx.Exec("DELETE FROM rows WHERE k=?", key)
    return err
}

func (this *Transaction) Commit() error {
	return this.tx.Commit()
}
