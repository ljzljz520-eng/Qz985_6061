package store

import (
	"encoding/json"
	"errors"
	"example.com/knowledge-backend/domain"
	"go.etcd.io/bbolt"
	"sort"
)

func (s *Store) SaveRecord(record domain.Record) error {
	if err := domain.ValidateRecord(record); err != nil {
		return err
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return s.Update(func(tx *bbolt.Tx) error { return tx.Bucket(recordBucket).Put([]byte(record.ID), data) })
}

func (s *Store) GetRecord(id string) (domain.Record, error) {
	var record domain.Record
	err := s.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket(recordBucket).Get([]byte(id))
		if value == nil {
			return errors.New("record not found")
		}
		return json.Unmarshal(append([]byte(nil), value...), &record)
	})
	return record, err
}

func (s *Store) ListRecords() ([]domain.Record, error) {
	items := make([]domain.Record, 0)
	err := s.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(recordBucket).ForEach(func(_, value []byte) error {
			var record domain.Record
			if err := json.Unmarshal(value, &record); err != nil {
				return err
			}
			items = append(items, record)
			return nil
		})
	})
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.Before(items[j].UpdatedAt) })
	return items, err
}

func (s *Store) DeleteRecord(id string) error {
	return s.Update(func(tx *bbolt.Tx) error { return tx.Bucket(recordBucket).Delete([]byte(id)) })
}

func (s *Store) SaveUser(user domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.Update(func(tx *bbolt.Tx) error { return tx.Bucket(userBucket).Put([]byte(user.ID), data) })
}

func (s *Store) GetUser(id string) (domain.User, error) {
	var user domain.User
	err := s.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket(userBucket).Get([]byte(id))
		if value == nil {
			return errors.New("user not found")
		}
		return json.Unmarshal(append([]byte(nil), value...), &user)
	})
	return user, err
}

func (s *Store) SaveEvent(event domain.Event) error {
	if err := domain.ValidateEvent(event); err != nil {
		return err
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return s.Update(func(tx *bbolt.Tx) error { return tx.Bucket(eventBucket).Put([]byte(event.ID), data) })
}

func (s *Store) ListEvents(recordID string) ([]domain.Event, error) {
	result := []domain.Event{}
	err := s.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(eventBucket).ForEach(func(_, value []byte) error {
			var event domain.Event
			if err := json.Unmarshal(value, &event); err != nil {
				return err
			}
			if recordID == "" || event.RecordID == recordID {
				result = append(result, event)
			}
			return nil
		})
	})
	return result, err
}

func (s *Store) SaveAudit(audit domain.Audit) error {
	if err := domain.ValidateAudit(audit); err != nil {
		return err
	}
	data, err := json.Marshal(audit)
	if err != nil {
		return err
	}
	return s.Update(func(tx *bbolt.Tx) error { return tx.Bucket(auditBucket).Put([]byte(audit.ID), data) })
}

func (s *Store) ListAudits(recordID string) ([]domain.Audit, error) {
	result := []domain.Audit{}
	err := s.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(auditBucket).ForEach(func(_, value []byte) error {
			var audit domain.Audit
			if err := json.Unmarshal(value, &audit); err != nil {
				return err
			}
			if recordID == "" || audit.RecordID == recordID {
				result = append(result, audit)
			}
			return nil
		})
	})
	return result, err
}
