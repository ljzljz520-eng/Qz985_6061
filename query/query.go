package query

import (
	"example.com/knowledge-backend/domain"
	"sort"
	"strings"
)

type Filter struct {
	Statuses       []domain.Status
	Categories     []string
	Text           string
	VisibleOnly    bool
	MinimumVersion int
}
type Page struct {
	Items   []domain.Record
	Offset  int
	Limit   int
	Total   int
	HasMore bool
}
type Summary struct {
	Total      int
	Visible    int
	ByStatus   map[domain.Status]int
	ByCategory map[string]int
}

func Apply(records []domain.Record, filter Filter) []domain.Record {
	statuses := map[domain.Status]bool{}
	for _, status := range filter.Statuses {
		statuses[status] = true
	}
	categories := map[string]bool{}
	for _, category := range filter.Categories {
		categories[strings.ToLower(strings.TrimSpace(category))] = true
	}
	text := strings.ToLower(strings.TrimSpace(filter.Text))
	result := make([]domain.Record, 0, len(records))
	for _, record := range records {
		if len(statuses) > 0 && !statuses[record.Status] {
			continue
		}
		if len(categories) > 0 && !categories[strings.ToLower(record.Category)] {
			continue
		}
		if filter.VisibleOnly && !record.IsVisible() {
			continue
		}
		if filter.MinimumVersion > 0 && record.Version < filter.MinimumVersion {
			continue
		}
		if text != "" && !strings.Contains(strings.ToLower(record.Title+" "+record.Content), text) {
			continue
		}
		result = append(result, record)
	}
	return result
}

func Paginate(records []domain.Record, offset, limit int) Page {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset > len(records) {
		offset = len(records)
	}
	end := offset + limit
	if end > len(records) {
		end = len(records)
	}
	items := append([]domain.Record(nil), records[offset:end]...)
	return Page{Items: items, Offset: offset, Limit: limit, Total: len(records), HasMore: end < len(records)}
}

func Summarize(records []domain.Record) Summary {
	result := Summary{Total: len(records), ByStatus: map[domain.Status]int{}, ByCategory: map[string]int{}}
	for _, record := range records {
		result.ByStatus[record.Status]++
		result.ByCategory[record.Category]++
		if record.IsVisible() {
			result.Visible++
		}
	}
	return result
}

func Sort(records []domain.Record, field string, descending bool) []domain.Record {
	result := append([]domain.Record(nil), records...)
	sort.SliceStable(result, func(i, j int) bool {
		var less bool
		switch field {
		case "title":
			less = result[i].Title < result[j].Title
		case "category":
			less = result[i].Category < result[j].Category
		case "status":
			less = result[i].Status < result[j].Status
		case "version":
			less = result[i].Version < result[j].Version
		default:
			less = result[i].UpdatedAt.Before(result[j].UpdatedAt)
		}
		if descending {
			return !less
		}
		return less
	})
	return result
}

func FilterByStatus(records []domain.Record, status domain.Status) []domain.Record {
	return Apply(records, Filter{Statuses: []domain.Status{status}})
}

func FilterByCategory(records []domain.Record, category string) []domain.Record {
	return Apply(records, Filter{Categories: []string{category}})
}
