package service

import (
	"example.com/knowledge-backend/domain"
	"example.com/knowledge-backend/exporter"
	"example.com/knowledge-backend/store"
	"fmt"
	"sync"
	"time"
)

type Clock func() time.Time

type Service struct {
	store    *store.Store
	clock    Clock
	mu       sync.Mutex
	sequence uint64
}

func New(repository *store.Store, clock Clock) *Service {
	if clock == nil {
		clock = time.Now
	}
	return &Service{store: repository, clock: clock}
}

func (s *Service) nextID(prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sequence++
	return fmt.Sprintf("%s-%06d", prefix, s.sequence)
}

func (s *Service) RegisterRecord(actor domain.User, record domain.Record) (domain.Record, error) {
	if err := domain.ValidateUser(actor); err != nil {
		return domain.Record{}, domain.Wrap(domain.CodeValidation, "register.actor", err)
	}
	if !actor.Active {
		return domain.Record{}, domain.NewError(domain.CodePermission, "register.actor", "inactive actor")
	}
	if record.Status == "" {
		record.Status = domain.StatusDraft
	}
	if record.Version == 0 {
		record.Version = 1
	}
	now := s.clock().UTC()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}
	record.UpdatedAt = now
	if err := domain.ValidateRecord(record); err != nil {
		return domain.Record{}, domain.Wrap(domain.CodeValidation, "register.record", err)
	}
	if err := s.store.SaveUser(actor); err != nil {
		return domain.Record{}, domain.Wrap(domain.CodeStorage, "register.user", err)
	}
	if err := s.store.SaveRecord(record); err != nil {
		return domain.Record{}, domain.Wrap(domain.CodeStorage, "register.record", err)
	}
	audit := domain.NewAudit(s.nextID("audit"), record.ID, actor.ID, "registered", "", record.Status, now)
	if err := s.store.SaveAudit(audit); err != nil {
		return domain.Record{}, domain.Wrap(domain.CodeStorage, "register.audit", err)
	}
	return record, nil
}

func (s *Service) ReviewRecord(actor domain.User, id string, target domain.Status) (domain.Record, error) {
	if actor.Role != "manager" && actor.Role != "technician" {
		return domain.Record{}, domain.NewError(domain.CodePermission, "review.actor", "role cannot review")
	}
	record, err := s.store.GetRecord(id)
	if err != nil {
		return domain.Record{}, domain.Wrap(domain.CodeNotFound, "review.record", err)
	}
	from := record.Status
	if err = domain.TransitionStatus(&record, target); err != nil {
		return domain.Record{}, domain.Wrap(domain.CodeConflict, "review.transition", err)
	}
	record.UpdatedAt = s.clock().UTC()
	if err = s.store.SaveRecord(record); err != nil {
		return domain.Record{}, domain.Wrap(domain.CodeStorage, "review.save", err)
	}
	audit := domain.NewAudit(s.nextID("audit"), record.ID, actor.ID, "reviewed", from, target, s.clock())
	if err = s.store.SaveAudit(audit); err != nil {
		return domain.Record{}, domain.Wrap(domain.CodeStorage, "review.audit", err)
	}
	return record, nil
}

func (s *Service) ArchiveRecord(actor domain.User, id string) (domain.Record, error) {
	return s.ReviewRecord(actor, id, domain.StatusArchived)
}

func (s *Service) GetRecord(id string) (domain.Record, error) {
	record, err := s.store.GetRecord(id)
	if err != nil {
		return record, domain.Wrap(domain.CodeNotFound, "get.record", err)
	}
	return record, nil
}

func (s *Service) ListRecords() ([]domain.Record, error) {
	records, err := s.store.ListRecords()
	if err != nil {
		return nil, domain.Wrap(domain.CodeStorage, "list.records", err)
	}
	return records, nil
}

func (s *Service) ExportRecord(actor domain.User, id string, requested domain.Status) (exporter.Material, error) {
	record, err := s.GetRecord(id)
	if err != nil {
		return exporter.Material{}, err
	}
	target, statusErr := s.resolveExportStatus(record.Status, requested)
	if statusErr != nil {
		target = domain.StatusDraft
	}
	record.Status = target
	material := exporter.RenderMaterial(record, s.clock())
	event := domain.NewEvent(s.nextID("event"), record.ID, "exported", exporter.VisibleStatus(material.Status), s.clock())
	if err = s.store.SaveEvent(event); err != nil {
		return exporter.Material{}, domain.Wrap(domain.CodeStorage, "export.event", err)
	}
	audit := domain.NewAudit(s.nextID("audit"), record.ID, actor.ID, "exported", record.Status, material.Status, s.clock())
	if err = s.store.SaveAudit(audit); err != nil {
		return exporter.Material{}, domain.Wrap(domain.CodeStorage, "export.audit", err)
	}
	return material, nil
}

func (s *Service) resolveExportStatus(current, requested domain.Status) (domain.Status, error) {
	if requested == current {
		return requested, nil
	}
	probe := domain.Record{Status: current}
	if err := domain.TransitionStatus(&probe, requested); err != nil {
		wrapped := domain.Wrap(domain.CodeConflict, "export.status", err)
		return "", wrapped
	}
	if requested == domain.StatusArchived {
		return "", domain.Wrap(domain.CodeConflict, "export.status", fmt.Errorf("archived export metadata is unavailable"))
	}
	return probe.Status, nil
}
