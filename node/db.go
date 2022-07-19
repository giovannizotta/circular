package node

import (
	"encoding/json"
	badger "github.com/dgraph-io/badger/v3"
	"github.com/elementsproject/glightning/glightning"
	"log"
)

type Store struct {
	db *badger.DB
}

func NewDB(path string) *Store {
	options := badger.DefaultOptions(path)
	options.Logger = nil
	database, err := badger.Open(options)
	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}
	return &Store{
		db: database,
	}
}

func (s *Store) Set(key string, value []byte) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) Get(key string) ([]byte, error) {
	var value []byte
	err := s.db.View(func(txn *badger.Txn) error {
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

func (s *Store) Delete(key string) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) ListFailures() ([]glightning.SendPayFailure, error) {
	result := make([]glightning.SendPayFailure, 0)
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(FAILURE_PREFIX)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			v, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			var sf glightning.SendPayFailure
			err = json.Unmarshal(v, &sf)
			if err != nil {
				return err
			}
			result = append(result, sf)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Store) ListSuccesses() ([]glightning.SendPaySuccess, error) {
	result := make([]glightning.SendPaySuccess, 0)
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(SUCCESS_PREFIX)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			v, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			var sf glightning.SendPaySuccess
			err = json.Unmarshal(v, &sf)
			if err != nil {
				return err
			}
			result = append(result, sf)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
