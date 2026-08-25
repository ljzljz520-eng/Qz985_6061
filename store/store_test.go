package store

import (
	"example.com/knowledge-backend/domain"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "records.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	record := domain.NewRecord("r1", "点火线圈", "检查线圈和插头", "engine", time.Now())
	if err = s.SaveRecord(record); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetRecord("r1")
	if err != nil || got.Title != record.Title {
		t.Fatalf("%v %#v", err, got)
	}
	if err = s.SaveEvent(domain.NewEvent("e1", "r1", "notice", "sent", time.Now())); err != nil {
		t.Fatal(err)
	}
	events, err := s.ListEvents("r1")
	if err != nil || len(events) != 1 {
		t.Fatalf("%v %d", err, len(events))
	}
}
