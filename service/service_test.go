package service

import (
	"example.com/knowledge-backend/domain"
	"example.com/knowledge-backend/store"
	"path/filepath"
	"testing"
	"time"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "service.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return New(s, FixedClock(time.Unix(1000, 0)))
}

func TestRegisterAndReview(t *testing.T) {
	svc := newTestService(t)
	actor := domain.NewUser("u1", "技师", "technician", true)
	result, err := svc.RunRegistration(RegistrationInput{Actor: actor, ID: "r1", Title: "发动机抖动", Content: "检查火花塞与点火线圈", Category: "engine"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Record.Status != domain.StatusDraft {
		t.Fatal(result.Record.Status)
	}
	review, err := svc.RunReview(ReviewInput{Actor: domain.NewUser("m1", "主管", "manager", true), RecordID: "r1", Approve: true})
	if err != nil {
		t.Fatal(err)
	}
	if review.Record.Status != domain.StatusApproved {
		t.Fatal(review.Record.Status)
	}
}
