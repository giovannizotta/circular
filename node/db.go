package node

import (
	badger "github.com/dgraph-io/badger/v3"
	"log"
)

type PreimageStore struct {
	db *badger.DB
}

func NewDB(path string) *PreimageStore {
	options := badger.DefaultOptions(path)
	options.Logger = nil
	database, err := badger.Open(options)
	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}
	return &PreimageStore{
		db: database,
	}
}

func (d *PreimageStore) Set(key string, value []byte) error {
	err := d.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *PreimageStore) Get(key string) ([]byte, error) {
	var value []byte
	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		result, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		value = result
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (d *PreimageStore) Delete(key string) error {
	err := d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	if err != nil {
		return err
	}
	return nil
}
