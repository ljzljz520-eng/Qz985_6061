package store

import (
	"encoding/json"
	"example.com/knowledge-backend/domain"
	"fmt"
	"go.etcd.io/bbolt"
	"time"
)

type Snapshot struct {
	Records   []domain.Record `json:"records"`
	Users     []domain.User   `json:"users"`
	Events    []domain.Event  `json:"events"`
	Audits    []domain.Audit  `json:"audits"`
	CreatedAt time.Time       `json:"created_at"`
}

func (s *Store) Snapshot(now time.Time) (Snapshot, error) {
	snapshot := Snapshot{Records: []domain.Record{}, Users: []domain.User{}, Events: []domain.Event{}, Audits: []domain.Audit{}, CreatedAt: now.UTC()}
	err := s.View(func(tx *bbolt.Tx) error {
		if err := forEachJSON(tx, recordBucket, &snapshot.Records); err != nil {
			return err
		}
		if err := forEachJSON(tx, userBucket, &snapshot.Users); err != nil {
			return err
		}
		if err := forEachJSON(tx, eventBucket, &snapshot.Events); err != nil {
			return err
		}
		return forEachJSON(tx, auditBucket, &snapshot.Audits)
	})
	return snapshot, err
}

func forEachJSON[T any](tx *bbolt.Tx, bucket []byte, target *[]T) error {
	b := tx.Bucket(bucket)
	if b == nil {
		return fmt.Errorf("missing bucket")
	}
	return b.ForEach(func(_, value []byte) error {
		var item T
		if err := json.Unmarshal(value, &item); err != nil {
			return err
		}
		*target = append(*target, item)
		return nil
	})
}

func (s *Store) Restore(snapshot Snapshot) error {
	return s.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range [][]byte{recordBucket, userBucket, eventBucket, auditBucket} {
			if err := tx.Bucket(bucket).ForEach(func(key, _ []byte) error { return tx.Bucket(bucket).Delete(key) }); err != nil {
				return err
			}
		}
		for _, record := range snapshot.Records {
			if err := putJSON(tx, recordBucket, []byte(record.ID), record); err != nil {
				return err
			}
		}
		for _, user := range snapshot.Users {
			if err := putJSON(tx, userBucket, []byte(user.ID), user); err != nil {
				return err
			}
		}
		for _, event := range snapshot.Events {
			if err := putJSON(tx, eventBucket, []byte(event.ID), event); err != nil {
				return err
			}
		}
		for _, audit := range snapshot.Audits {
			if err := putJSON(tx, auditBucket, []byte(audit.ID), audit); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) Count(bucket string) (int, error) {
	var count int
	err := s.View(func(tx *bbolt.Tx) error {
		var name []byte
		switch bucket {
		case "records":
			name = recordBucket
		case "users":
			name = userBucket
		case "events":
			name = eventBucket
		case "audits":
			name = auditBucket
		default:
			return fmt.Errorf("unknown bucket")
		}
		count = int(tx.Bucket(name).Stats().KeyN)
		return nil
	})
	return count, err
}

func (s *Store) PurgeArchivedBefore(cutoff time.Time) (int, error) {
	removed := 0
	err := s.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(recordBucket)
		return b.ForEach(func(key, value []byte) error {
			var record domain.Record
			if err := json.Unmarshal(value, &record); err != nil {
				return err
			}
			if record.Status == domain.StatusArchived && record.UpdatedAt.Before(cutoff) {
				if err := b.Delete(key); err != nil {
					return err
				}
				removed++
			}
			return nil
		})
	})
	return removed, err
}

func (s *Store) VerifyEntities() error {
	snapshot, err := s.Snapshot(time.Now())
	if err != nil {
		return err
	}
	for _, record := range snapshot.Records {
		if err := domain.ValidateRecord(record); err != nil {
			return err
		}
	}
	for _, user := range snapshot.Users {
		if err := domain.ValidateUser(user); err != nil {
			return err
		}
	}
	return nil
}
