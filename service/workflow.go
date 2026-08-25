package service

import (
	"example.com/knowledge-backend/domain"
	"example.com/knowledge-backend/exporter"
	"fmt"
	"time"
)

type RegistrationInput struct {
	Actor    domain.User
	ID       string
	Title    string
	Content  string
	Category string
}
type RegistrationResult struct {
	Record domain.Record
	Audits []domain.Audit
}
type ReviewInput struct {
	Actor    domain.User
	RecordID string
	Approve  bool
}
type ReviewResult struct {
	Record domain.Record
	Events []domain.Event
}
type ExportInput struct {
	Actor    domain.User
	RecordID string
	Target   domain.Status
}
type ExportResult struct {
	Material exporter.Material
	Audits   []domain.Audit
	Events   []domain.Event
}

func (s *Service) RunRegistration(input RegistrationInput) (RegistrationResult, error) {
	record := domain.NewRecord(input.ID, input.Title, input.Content, input.Category, s.clock())
	saved, err := s.RegisterRecord(input.Actor, record)
	if err != nil {
		return RegistrationResult{}, err
	}
	audits, err := s.store.ListAudits(saved.ID)
	if err != nil {
		return RegistrationResult{}, domain.Wrap(domain.CodeStorage, "workflow.registration.audits", err)
	}
	return RegistrationResult{Record: saved, Audits: audits}, nil
}

func (s *Service) RunReview(input ReviewInput) (ReviewResult, error) {
	requested := domain.StatusRejected
	if input.Approve {
		requested = domain.StatusApproved
	}
	record, err := s.GetRecord(input.RecordID)
	if err != nil {
		return ReviewResult{}, err
	}
	if record.Status == domain.StatusDraft {
		record, err = s.ReviewRecord(input.Actor, input.RecordID, domain.StatusPending)
		if err != nil {
			return ReviewResult{}, err
		}
	}
	record, err = s.ReviewRecord(input.Actor, input.RecordID, requested)
	if err != nil {
		return ReviewResult{}, err
	}
	event := domain.NewEvent(s.nextID("event"), record.ID, "reviewed", fmt.Sprintf("status=%s", record.Status), s.clock())
	if err = s.store.SaveEvent(event); err != nil {
		return ReviewResult{}, domain.Wrap(domain.CodeStorage, "workflow.review.event", err)
	}
	events, err := s.store.ListEvents(record.ID)
	if err != nil {
		return ReviewResult{}, err
	}
	return ReviewResult{Record: record, Events: events}, nil
}

func (s *Service) RunExport(input ExportInput) (ExportResult, error) {
	material, err := s.ExportRecord(input.Actor, input.RecordID, input.Target)
	if err != nil {
		return ExportResult{}, err
	}
	audits, err := s.store.ListAudits(input.RecordID)
	if err != nil {
		return ExportResult{}, err
	}
	events, err := s.store.ListEvents(input.RecordID)
	if err != nil {
		return ExportResult{}, err
	}
	return ExportResult{Material: material, Audits: audits, Events: events}, nil
}

func FixedClock(value time.Time) Clock {
	fixed := value.UTC()
	return func() time.Time { return fixed }
}
