package node

import (
	badger "github.com/dgraph-io/badger/v3"
	"log"
)

type DB struct {
	db *badger.DB
}

func NewDB(path string) *DB {
	database, err := badger.Open(badger.DefaultOptions(path + "/db"))
	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}
	return &DB{
		db: database,
	}
}

func (d *DB) Set(secret PreimageHashPair) error {
	err := d.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(secret.Hash), []byte(secret.Preimage))
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) Get(hash string) (string, error) {
	var preimage string
	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(hash))
		if err != nil {
			return err
		}
		result, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		preimage = string(result)
		return nil
	})
	if err != nil {
		return "", err
	}
	return preimage, nil
}

func (d *DB) PrintAllPreimageHashPairs() error {
	err := d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				log.Printf("key=%s, value=%s\n", k, v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
