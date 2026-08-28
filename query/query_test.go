package query

import (
	"example.com/knowledge-backend/domain"
	"testing"
	"time"
)

func TestQueryByStatus(t *testing.T) {
	records := []domain.Record{{ID: "1", Title: "a", Content: "engine content", Category: "engine", Status: domain.StatusApproved, Version: 2, UpdatedAt: time.Unix(1, 0)}, {ID: "2", Title: "b", Content: "brake content", Category: "brakes", Status: domain.StatusDraft, Version: 1, UpdatedAt: time.Unix(2, 0)}}
	filtered := FilterByStatus(records, domain.StatusApproved)
	if len(filtered) != 1 {
		t.Fatal(len(filtered))
	}
	page := Paginate(records, 0, 1)
	if !page.HasMore || len(page.Items) != 1 {
		t.Fatal(page)
	}
	summary := Summarize(records)
	if summary.Visible != 1 || summary.ByCategory["brakes"] != 1 {
		t.Fatal(summary)
	}
}

func TestIndexSearch(t *testing.T) {
	index := NewIndex(nil)
	index.Upsert(domain.Record{ID: "1", Title: "火花塞", Content: "发动机 点火", Category: "engine"})
	if index.Count() != 1 || len(index.Search("发动机")) != 1 {
		t.Fatal("index")
	}
	if !index.Remove("1") || index.Count() != 0 {
		t.Fatal("remove")
	}
}
