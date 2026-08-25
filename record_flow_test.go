package knowledge

import (
	"example.com/knowledge-backend/domain"
	"example.com/knowledge-backend/service"
	"example.com/knowledge-backend/store"
	"path/filepath"
	"testing"
)

func TestRecordFlow12(t *testing.T) {
	repository, err := store.Open(filepath.Join(t.TempDir(), "flow12.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	svc := service.New(repository, nil)
	actor := domain.NewUser("manager-12", "店长", "manager", true)
	_, err = svc.RunRegistration(service.RegistrationInput{Actor: actor, ID: "record-12", Title: "汽修12导出资料", Content: "记录目标状态并导出知识资料", Category: "maintenance"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.RunReview(service.ReviewInput{Actor: actor, RecordID: "record-12", Approve: true})
	if err != nil {
		t.Fatal(err)
	}
	material, err := svc.ExportRecord(actor, "record-12", domain.StatusArchived)
	if err != nil {
		t.Fatal(err)
	}
	if material.Status != domain.StatusArchived {
		t.Fatalf("expected archived status, got %s", material.Status)
	}
}
