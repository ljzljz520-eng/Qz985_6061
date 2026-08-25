package store

import (
	"go.etcd.io/bbolt"
	"path/filepath"
	"time"
)

var (
	recordBucket = []byte("records")
	userBucket   = []byte("users")
	eventBucket  = []byte("events")
	auditBucket  = []byte("audits")
)

type Store struct {
	db   *bbolt.DB
	path string
}

func Open(path string) (*Store, error) {
	if path == "" {
		path = filepath.Join(".", "knowledge.db")
	}
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 2 * time.Second, NoSync: true})
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, path: path}
	if err = s.initialize(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) initialize() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range [][]byte{recordBucket, userBucket, eventBucket, auditBucket} {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	err := s.db.Sync()
	closeErr := s.db.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

func (s *Store) Health() error {
	if s == nil || s.db == nil {
		return bbolt.ErrDatabaseNotOpen
	}
	return s.db.View(func(tx *bbolt.Tx) error {
		if tx.Bucket(recordBucket) == nil {
			return bbolt.ErrBucketNotFound
		}
		return nil
	})
}

func (s *Store) Update(fn func(*bbolt.Tx) error) error {
	if s == nil || s.db == nil {
		return bbolt.ErrDatabaseNotOpen
	}
	return s.db.Update(fn)
}

func (s *Store) View(fn func(*bbolt.Tx) error) error {
	if s == nil || s.db == nil {
		return bbolt.ErrDatabaseNotOpen
	}
	return s.db.View(fn)
}
