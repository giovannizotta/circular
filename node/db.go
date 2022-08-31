package node

import (
	"circular/graph"
	"encoding/json"
	badger "github.com/dgraph-io/badger/v3"
	"github.com/elementsproject/glightning/glightning"
	"log"
	"time"
)

const (
	FOURTEEN_DAYS = 14 * 24 * time.Hour
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

// Every key is allowed to stay in the db for at most 14 days
func (s *Store) Set(key string, value []byte) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry([]byte(key), value).WithTTL(FOURTEEN_DAYS))
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

func (s *Store) ListRoutes() ([]graph.PrettyRoute, error) {
	result := make([]graph.PrettyRoute, 0)
	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(ROUTE_PREFIX)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			v, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			var pr graph.PrettyRoute
			err = json.Unmarshal(v, &pr)
			if err != nil {
				return err
			}
			result = append(result, pr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (n *Node) GarbageCollect() {
	n.Logln(glightning.Info, "Garbage collecting")
	err := n.DB.db.RunValueLogGC(0.5)
	if err != nil {
		n.Logf(glightning.Unusual, "GC report: %+v", err)
	}
}

func (n *Node) SaveToDb(key string, value any) error {
	if !n.saveStats {
		return nil
	}

	b, err := json.Marshal(value)
	if err != nil {
		n.Logln(glightning.Unusual, err)
		return err
	}

	err = n.DB.Set(key, b)
	if err != nil {
		n.Logln(glightning.Unusual, err)
		return err
	}

	return nil
}
