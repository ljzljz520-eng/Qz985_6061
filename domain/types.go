package domain

import (
	"strings"
	"time"
)

type Status string

const (
	StatusDraft    Status = "draft"
	StatusPending  Status = "pending_review"
	StatusApproved Status = "approved"
	StatusArchived Status = "archived"
	StatusRejected Status = "rejected"
)

type Record struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    Status    `json:"status"`
	Category  string    `json:"category"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
	Version   int       `json:"version"`
}

type User struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Active bool   `json:"active"`
}

type Event struct {
	ID        string    `json:"id"`
	RecordID  string    `json:"record_id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type Audit struct {
	ID        string    `json:"id"`
	RecordID  string    `json:"record_id"`
	ActorID   string    `json:"actor_id"`
	Action    string    `json:"action"`
	From      Status    `json:"from"`
	To        Status    `json:"to"`
	CreatedAt time.Time `json:"created_at"`
}

func NewRecord(id, title, content, category string, now time.Time) Record {
	return Record{ID: strings.TrimSpace(id), Title: strings.TrimSpace(title), Content: strings.TrimSpace(content), Category: strings.TrimSpace(category), Status: StatusDraft, CreatedAt: now.UTC(), UpdatedAt: now.UTC(), Version: 1}
}

func NewUser(id, name, role string, active bool) User {
	return User{ID: strings.TrimSpace(id), Name: strings.TrimSpace(name), Role: strings.TrimSpace(role), Active: active}
}

func NewEvent(id, recordID, eventType, message string, now time.Time) Event {
	return Event{ID: id, RecordID: recordID, Type: eventType, Message: message, CreatedAt: now.UTC()}
}

func NewAudit(id, recordID, actorID, action string, from, to Status, now time.Time) Audit {
	return Audit{ID: id, RecordID: recordID, ActorID: actorID, Action: action, From: from, To: to, CreatedAt: now.UTC()}
}

func (r Record) Clone() Record {
	return Record{ID: r.ID, Title: r.Title, Content: r.Content, Status: r.Status, Category: r.Category, UpdatedAt: r.UpdatedAt, CreatedAt: r.CreatedAt, Version: r.Version}
}

func (r Record) IsVisible() bool {
	return r.Status == StatusApproved || r.Status == StatusArchived
}

func (r Record) Summary() string {
	text := strings.TrimSpace(r.Content)
	if len(text) > 80 {
		return text[:80]
	}
	return text
}

func (s Status) String() string { return string(s) }

func (s Status) IsTerminal() bool { return s == StatusArchived || s == StatusRejected }

func AllStatuses() []Status {
	return []Status{StatusDraft, StatusPending, StatusApproved, StatusArchived, StatusRejected}
}
