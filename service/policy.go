package service

import (
	"example.com/knowledge-backend/domain"
	"fmt"
	"strings"
)

type Policy struct {
	AllowedCategories map[string]bool
	ExportRoles       map[string]bool
	MaximumContent    int
}

func DefaultPolicy() Policy {
	return Policy{AllowedCategories: map[string]bool{"engine": true, "electrical": true, "brakes": true, "maintenance": true}, ExportRoles: map[string]bool{"manager": true, "technician": true}, MaximumContent: 12000}
}

func (p Policy) ValidateRegistration(input RegistrationInput) error {
	if !p.AllowedCategories[strings.ToLower(input.Category)] {
		return fmt.Errorf("category is not allowed")
	}
	if len(input.Content) > p.MaximumContent {
		return fmt.Errorf("content exceeds limit")
	}
	if !input.Actor.Active {
		return fmt.Errorf("actor is inactive")
	}
	return nil
}

func (p Policy) CanExport(user domain.User, record domain.Record) bool {
	if !user.Active {
		return false
	}
	if !p.ExportRoles[user.Role] {
		return false
	}
	return record.Status == domain.StatusApproved || record.Status == domain.StatusArchived
}

func (p Policy) Categories() []string {
	result := make([]string, 0, len(p.AllowedCategories))
	for category, allowed := range p.AllowedCategories {
		if allowed {
			result = append(result, category)
		}
	}
	return result
}
