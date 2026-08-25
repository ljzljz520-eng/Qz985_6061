package query

import (
	"example.com/knowledge-backend/domain"
	"sort"
	"strings"
	"sync"
)

type Index struct {
	mu      sync.RWMutex
	records map[string]domain.Record
	tokens  map[string]map[string]bool
}

func NewIndex(records []domain.Record) *Index {
	index := &Index{records: map[string]domain.Record{}, tokens: map[string]map[string]bool{}}
	for _, record := range records {
		index.Upsert(record)
	}
	return index
}

func tokenize(value string) []string {
	fields := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool { return r == ' ' || r == '-' || r == '_' || r == '/' || r == ',' || r == '.' })
	unique := map[string]bool{}
	result := []string{}
	for _, field := range fields {
		if field != "" && !unique[field] {
			unique[field] = true
			result = append(result, field)
		}
	}
	return result
}

func (i *Index) Upsert(record domain.Record) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if previous, ok := i.records[record.ID]; ok {
		i.removeTokens(previous)
	}
	i.records[record.ID] = record
	for _, token := range tokenize(record.Title + " " + record.Content + " " + record.Category) {
		ids := i.tokens[token]
		if ids == nil {
			ids = map[string]bool{}
			i.tokens[token] = ids
		}
		ids[record.ID] = true
	}
}

func (i *Index) removeTokens(record domain.Record) {
	for _, token := range tokenize(record.Title + " " + record.Content + " " + record.Category) {
		delete(i.tokens[token], record.ID)
		if len(i.tokens[token]) == 0 {
			delete(i.tokens, token)
		}
	}
}

func (i *Index) Remove(id string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	record, ok := i.records[id]
	if !ok {
		return false
	}
	i.removeTokens(record)
	delete(i.records, id)
	return true
}

func (i *Index) Search(text string) []domain.Record {
	i.mu.RLock()
	defer i.mu.RUnlock()
	terms := tokenize(text)
	if len(terms) == 0 {
		return i.allLocked()
	}
	matches := map[string]int{}
	for _, term := range terms {
		for id := range i.tokens[term] {
			matches[id]++
		}
	}
	result := []domain.Record{}
	for id, count := range matches {
		if count == len(terms) {
			result = append(result, i.records[id])
		}
	}
	sort.Slice(result, func(a, b int) bool { return result[a].Title < result[b].Title })
	return result
}

func (i *Index) allLocked() []domain.Record {
	result := make([]domain.Record, 0, len(i.records))
	for _, record := range i.records {
		result = append(result, record)
	}
	sort.Slice(result, func(a, b int) bool { return result[a].ID < result[b].ID })
	return result
}

func (i *Index) Count() int { i.mu.RLock(); defer i.mu.RUnlock(); return len(i.records) }
