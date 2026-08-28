package knowledge

import (
	"example.com/knowledge-backend/domain"
	"example.com/knowledge-backend/store"
	"path/filepath"
	"testing"
	"time"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reopen.db")
	first, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	record := domain.NewRecord("persist-1", "冷却液", "检查液位、泄漏和水泵", "maintenance", time.Now())
	if err = first.SaveRecord(record); err != nil {
		t.Fatal(err)
	}
	if err = first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	got, err := second.GetRecord("persist-1")
	if err != nil || got.Content != record.Content {
		t.Fatalf("%v %#v", err, got)
	}
}
