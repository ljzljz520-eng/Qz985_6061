package service

import (
	"example.com/knowledge-backend/domain"
	"example.com/knowledge-backend/exporter"
	"fmt"
	"sort"
	"strings"
	"time"
)

type RecordReport struct {
	Record   domain.Record
	Score    int
	Keywords []string
	Changes  []domain.RecordChange
	Audits   []domain.Audit
	Events   []domain.Event
}

func (s *Service) BuildReport(id string) (RecordReport, error) {
	record, err := s.GetRecord(id)
	if err != nil {
		return RecordReport{}, err
	}
	audits, err := s.store.ListAudits(id)
	if err != nil {
		return RecordReport{}, err
	}
	events, err := s.store.ListEvents(id)
	if err != nil {
		return RecordReport{}, err
	}
	return RecordReport{Record: record, Score: domain.CompletionScore(record), Keywords: domain.ExtractKeywords(record), Audits: audits, Events: events}, nil
}

func (s *Service) Timeline(id string) ([]string, error) {
	audits, err := s.store.ListAudits(id)
	if err != nil {
		return nil, err
	}
	events, err := s.store.ListEvents(id)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(audits)+len(events))
	for _, audit := range audits {
		result = append(result, fmt.Sprintf("%s %s %s->%s", audit.CreatedAt.Format(time.RFC3339), audit.Action, audit.From, audit.To))
	}
	for _, event := range events {
		result = append(result, fmt.Sprintf("%s %s %s", event.CreatedAt.Format(time.RFC3339), event.Type, event.Message))
	}
	sort.Strings(result)
	return result, nil
}

func (s *Service) Notify(id, message string) (domain.Event, error) {
	if strings.TrimSpace(message) == "" {
		return domain.Event{}, fmt.Errorf("message is required")
	}
	if _, err := s.GetRecord(id); err != nil {
		return domain.Event{}, err
	}
	event := domain.NewEvent(s.nextID("event"), id, "notification", message, s.clock())
	if err := s.store.SaveEvent(event); err != nil {
		return domain.Event{}, domain.Wrap(domain.CodeStorage, "notify", err)
	}
	return event, nil
}

func (s *Service) BulkReview(actor domain.User, ids []string, approve bool) ([]domain.Record, []error) {
	result := make([]domain.Record, 0, len(ids))
	errs := make([]error, 0)
	for _, id := range ids {
		review, err := s.RunReview(ReviewInput{Actor: actor, RecordID: id, Approve: approve})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		result = append(result, review.Record)
	}
	return result, errs
}

func (s *Service) ExportBundle(actor domain.User, ids []string, target domain.Status) (exporter.Bundle, error) {
	records := make([]domain.Record, 0, len(ids))
	for _, id := range ids {
		record, err := s.GetRecord(id)
		if err != nil {
			return exporter.Bundle{}, err
		}
		record.Status = target
		records = append(records, record)
	}
	bundle := exporter.RenderBundle("knowledge-export", records, s.clock())
	if len(bundle.Materials) == 0 {
		return exporter.Bundle{}, fmt.Errorf("no records to export")
	}
	return bundle, nil
}

func (s *Service) StatusCounts() (map[domain.Status]int, error) {
	records, err := s.ListRecords()
	if err != nil {
		return nil, err
	}
	counts := map[domain.Status]int{}
	for _, record := range records {
		counts[record.Status]++
	}
	return counts, nil
}

func (s *Service) UpdateContent(actor domain.User, id, title, content, category string) (domain.Record, error) {
	record, err := s.GetRecord(id)
	if err != nil {
		return domain.Record{}, err
	}
	if actor.Role != "manager" && actor.Role != "technician" {
		return domain.Record{}, domain.NewError(domain.CodePermission, "update", "role cannot update")
	}
	before := record
	record = domain.MergeRecord(record, domain.Record{Title: title, Content: content, Category: category}, s.clock())
	if err = domain.ValidateRecord(record); err != nil {
		return domain.Record{}, err
	}
	if err = s.store.SaveRecord(record); err != nil {
		return domain.Record{}, domain.Wrap(domain.CodeStorage, "update.save", err)
	}
	changes := domain.DiffRecord(before, record)
	for _, change := range changes {
		event := domain.NewEvent(s.nextID("event"), id, "field_changed", fmt.Sprintf("%s:%s->%s", change.Field, change.Before, change.After), s.clock())
		if err = s.store.SaveEvent(event); err != nil {
			return domain.Record{}, err
		}
	}
	return record, nil
}
