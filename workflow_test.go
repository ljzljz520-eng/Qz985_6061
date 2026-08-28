package knowledge

import (
	"example.com/knowledge-backend/domain"
	"example.com/knowledge-backend/query"
	"example.com/knowledge-backend/service"
	"example.com/knowledge-backend/store"
	"path/filepath"
	"testing"
	"time"
)

func workflowService(t *testing.T) *service.Service {
	t.Helper()
	repository, err := store.Open(filepath.Join(t.TempDir(), "workflow.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { repository.Close() })
	return service.New(repository, service.FixedClock(time.Unix(2000, 0)))
}

func TestWorkflowOne(t *testing.T) {
	svc := workflowService(t)
	actor := domain.NewUser("front", "前台", "frontdesk", true)
	result, err := svc.RunRegistration(service.RegistrationInput{Actor: actor, ID: "w1", Title: "机油灯亮", Content: "检查机油液位和压力传感器", Category: "maintenance"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Record.ID != "w1" || len(result.Audits) != 1 {
		t.Fatal(result)
	}
}

func TestWorkflowTwo(t *testing.T) {
	svc := workflowService(t)
	_, err := svc.RunRegistration(service.RegistrationInput{Actor: domain.NewUser("front", "前台", "frontdesk", true), ID: "w2", Title: "制动抖动", Content: "检查刹车盘厚度和跳动值", Category: "brakes"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := svc.RunReview(service.ReviewInput{Actor: domain.NewUser("mgr", "经理", "manager", true), RecordID: "w2", Approve: true})
	if err != nil {
		t.Fatal(err)
	}
	records, err := svc.ListRecords()
	if err != nil {
		t.Fatal(err)
	}
	if len(query.FilterByStatus(records, domain.StatusApproved)) != 1 || result.Record.Status != domain.StatusApproved {
		t.Fatal(result, records)
	}
}

func TestWorkflowThree(t *testing.T) {
	svc := workflowService(t)
	_, err := svc.RunRegistration(service.RegistrationInput{Actor: domain.NewUser("front", "前台", "frontdesk", true), ID: "w3", Title: "故障码清除", Content: "使用诊断仪读取并记录故障码", Category: "electrical"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.RunReview(service.ReviewInput{Actor: domain.NewUser("mgr", "经理", "manager", true), RecordID: "w3", Approve: true})
	if err != nil {
		t.Fatal(err)
	}
	result, err := svc.RunExport(service.ExportInput{Actor: domain.NewUser("mgr", "经理", "manager", true), RecordID: "w3", Target: domain.StatusApproved})
	if err != nil {
		t.Fatal(err)
	}
	if result.Material.Status != domain.StatusApproved || len(result.Events) == 0 {
		t.Fatal(result)
	}
}
