package domain

import (
	"fmt"
	"strings"
)

func ValidateRecord(r Record) error {
	if strings.TrimSpace(r.ID) == "" {
		return NewError(CodeValidation, "record.id", "id is required")
	}
	if strings.TrimSpace(r.Title) == "" {
		return NewError(CodeValidation, "record.title", "title is required")
	}
	if strings.TrimSpace(r.Content) == "" {
		return NewError(CodeValidation, "record.content", "content is required")
	}
	if strings.TrimSpace(r.Category) == "" {
		return NewError(CodeValidation, "record.category", "category is required")
	}
	if len(r.Title) > 160 {
		return NewError(CodeValidation, "record.title", "title is too long")
	}
	if len(r.Content) < 8 {
		return NewError(CodeValidation, "record.content", "content is too short")
	}
	if !IsKnownStatus(r.Status) {
		return NewError(CodeValidation, "record.status", "unknown status")
	}
	return nil
}

func ValidateUser(u User) error {
	if u.ID == "" || u.Name == "" {
		return NewError(CodeValidation, "user", "identity is required")
	}
	if u.Role != "manager" && u.Role != "technician" && u.Role != "frontdesk" {
		return NewError(CodeValidation, "user.role", "unsupported role")
	}
	return nil
}

func ValidateEvent(e Event) error {
	if e.ID == "" || e.RecordID == "" || e.Type == "" {
		return NewError(CodeValidation, "event", "event fields are required")
	}
	if e.Message == "" {
		return NewError(CodeValidation, "event.message", "message is required")
	}
	return nil
}

func ValidateAudit(a Audit) error {
	if a.ID == "" || a.RecordID == "" || a.ActorID == "" {
		return NewError(CodeValidation, "audit", "audit identity is required")
	}
	if a.Action == "" {
		return NewError(CodeValidation, "audit.action", "action is required")
	}
	return nil
}

func IsKnownStatus(s Status) bool {
	for _, known := range AllStatuses() {
		if s == known {
			return true
		}
	}
	return false
}

func CanTransition(from, to Status) bool {
	switch from {
	case StatusDraft:
		return to == StatusPending || to == StatusRejected
	case StatusPending:
		return to == StatusApproved || to == StatusRejected
	case StatusApproved:
		return to == StatusArchived
	case StatusArchived, StatusRejected:
		return false
	default:
		return false
	}
}

func TransitionStatus(r *Record, target Status) error {
	if r == nil {
		return NewError(CodeValidation, "transition", "record is nil")
	}
	if !IsKnownStatus(target) {
		return NewError(CodeValidation, "transition", "target status is unknown")
	}
	if !CanTransition(r.Status, target) {
		return NewError(CodeConflict, "transition", fmt.Sprintf("cannot move %s to %s", r.Status, target))
	}
	r.Status = target
	r.Version++
	return nil
}

func ParseStatus(value string) (Status, error) {
	s := Status(strings.ToLower(strings.TrimSpace(value)))
	if !IsKnownStatus(s) {
		return "", NewError(CodeValidation, "status", "unsupported status")
	}
	return s, nil
}
