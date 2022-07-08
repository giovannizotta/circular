package node

import (
	badger "github.com/dgraph-io/badger/v3"
	"log"
)

type DB struct {
	db *badger.DB
}

func NewDB(path string) *DB {
	database, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}
	return &DB{
		db: database,
	}
}

func (d *DB) Set(key, value string) error {
	err := d.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), []byte(value))
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) Get(key string) (string, error) {
	var value string
	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		result, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		value = string(result)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}
