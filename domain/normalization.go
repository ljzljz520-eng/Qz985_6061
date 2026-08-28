package domain

import (
	"regexp"
	"sort"
	"strings"
	"time"
)

var whitespace = regexp.MustCompile(`\s+`)

func NormalizeRecord(record Record) Record {
	record.ID = strings.TrimSpace(record.ID)
	record.Title = normalizeText(record.Title)
	record.Content = normalizeText(record.Content)
	record.Category = strings.ToLower(strings.TrimSpace(record.Category))
	if record.Status == "" {
		record.Status = StatusDraft
	}
	if record.Version < 1 {
		record.Version = 1
	}
	return record
}

func normalizeText(value string) string {
	return strings.TrimSpace(whitespace.ReplaceAllString(value, " "))
}

func MergeRecord(base, update Record, now time.Time) Record {
	merged := base.Clone()
	if strings.TrimSpace(update.Title) != "" {
		merged.Title = normalizeText(update.Title)
	}
	if strings.TrimSpace(update.Content) != "" {
		merged.Content = normalizeText(update.Content)
	}
	if strings.TrimSpace(update.Category) != "" {
		merged.Category = strings.ToLower(strings.TrimSpace(update.Category))
	}
	if update.Status != "" {
		merged.Status = update.Status
	}
	merged.UpdatedAt = now.UTC()
	merged.Version++
	return merged
}

type RecordChange struct {
	Field  string
	Before string
	After  string
}

func DiffRecord(before, after Record) []RecordChange {
	changes := make([]RecordChange, 0, 5)
	if before.Title != after.Title {
		changes = append(changes, RecordChange{"title", before.Title, after.Title})
	}
	if before.Content != after.Content {
		changes = append(changes, RecordChange{"content", before.Content, after.Content})
	}
	if before.Category != after.Category {
		changes = append(changes, RecordChange{"category", before.Category, after.Category})
	}
	if before.Status != after.Status {
		changes = append(changes, RecordChange{"status", before.Status.String(), after.Status.String()})
	}
	if before.Version != after.Version {
		changes = append(changes, RecordChange{"version", formatInt(before.Version), formatInt(after.Version)})
	}
	return changes
}

func formatInt(value int) string {
	if value == 0 {
		return "0"
	}
	digits := ""
	for value > 0 {
		digits = string(rune('0'+value%10)) + digits
		value /= 10
	}
	return digits
}

func CompletionScore(record Record) int {
	score := 0
	if record.ID != "" {
		score += 20
	}
	if record.Title != "" {
		score += 20
	}
	if len(record.Content) >= 8 {
		score += 30
	} else if len(record.Content) > 0 {
		score += 10
	}
	if record.Category != "" {
		score += 15
	}
	if record.IsVisible() {
		score += 15
	}
	return score
}

func ValidateBatch(records []Record) map[string]error {
	result := make(map[string]error)
	seen := make(map[string]bool)
	for _, record := range records {
		if seen[record.ID] {
			result[record.ID] = NewError(CodeConflict, "batch", "duplicate id")
			continue
		}
		seen[record.ID] = true
		if err := ValidateRecord(record); err != nil {
			result[record.ID] = err
		}
	}
	return result
}

func GroupByCategory(records []Record) map[string][]Record {
	result := make(map[string][]Record)
	for _, record := range records {
		key := strings.ToLower(strings.TrimSpace(record.Category))
		result[key] = append(result[key], record)
	}
	for key := range result {
		sort.Slice(result[key], func(i, j int) bool { return result[key][i].Title < result[key][j].Title })
	}
	return result
}

func StatusRank(status Status) int {
	switch status {
	case StatusDraft:
		return 1
	case StatusPending:
		return 2
	case StatusApproved:
		return 3
	case StatusArchived:
		return 4
	case StatusRejected:
		return 5
	default:
		return 0
	}
}

func CompareRecords(left, right Record) int {
	if StatusRank(left.Status) != StatusRank(right.Status) {
		if StatusRank(left.Status) < StatusRank(right.Status) {
			return -1
		}
		return 1
	}
	if left.UpdatedAt.Before(right.UpdatedAt) {
		return -1
	}
	if left.UpdatedAt.After(right.UpdatedAt) {
		return 1
	}
	return strings.Compare(left.ID, right.ID)
}

func HydrateTimes(record Record, now time.Time) Record {
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now.UTC()
	}
	if record.UpdatedAt.IsZero() {
		record.UpdatedAt = record.CreatedAt
	}
	return record
}

func ExtractKeywords(record Record) []string {
	words := strings.Fields(strings.ToLower(record.Title + " " + record.Category + " " + record.Content))
	unique := map[string]bool{}
	result := []string{}
	for _, word := range words {
		word = strings.Trim(word, "，。,.!?！？:：")
		if len(word) >= 2 && !unique[word] {
			unique[word] = true
			result = append(result, word)
		}
	}
	sort.Strings(result)
	return result
}
