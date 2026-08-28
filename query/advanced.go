package query

import (
	"example.com/knowledge-backend/domain"
	"fmt"
	"strconv"
	"strings"
)

type QueryPlan struct {
	Filter     Filter
	SortField  string
	Descending bool
	Offset     int
	Limit      int
}

func ParseQuery(values map[string]string) (QueryPlan, error) {
	plan := QueryPlan{Limit: 20}
	plan.Filter.Text = strings.TrimSpace(values["q"])
	if raw := strings.TrimSpace(values["status"]); raw != "" {
		status, err := domain.ParseStatus(raw)
		if err != nil {
			return QueryPlan{}, err
		}
		plan.Filter.Statuses = []domain.Status{status}
	}
	if raw := strings.TrimSpace(values["category"]); raw != "" {
		for _, value := range strings.Split(raw, ",") {
			if value = strings.TrimSpace(value); value != "" {
				plan.Filter.Categories = append(plan.Filter.Categories, value)
			}
		}
	}
	plan.SortField = values["sort"]
	if plan.SortField == "" {
		plan.SortField = "updated"
	}
	if values["desc"] == "1" || strings.EqualFold(values["desc"], "true") {
		plan.Descending = true
	}
	var err error
	if raw := values["offset"]; raw != "" {
		plan.Offset, err = strconv.Atoi(raw)
		if err != nil || plan.Offset < 0 {
			return QueryPlan{}, fmt.Errorf("invalid offset")
		}
	}
	if raw := values["limit"]; raw != "" {
		plan.Limit, err = strconv.Atoi(raw)
		if err != nil || plan.Limit < 1 {
			return QueryPlan{}, fmt.Errorf("invalid limit")
		}
	}
	if plan.Limit > 100 {
		plan.Limit = 100
	}
	return plan, nil
}

func Execute(records []domain.Record, plan QueryPlan) Page {
	filtered := Apply(records, plan.Filter)
	sorted := Sort(filtered, plan.SortField, plan.Descending)
	return Paginate(sorted, plan.Offset, plan.Limit)
}

type Facet struct {
	Value string
	Count int
}

func Facets(records []domain.Record) (statuses, categories []Facet) {
	statusCount := map[string]int{}
	categoryCount := map[string]int{}
	for _, record := range records {
		statusCount[record.Status.String()]++
		categoryCount[record.Category]++
	}
	for value, count := range statusCount {
		statuses = append(statuses, Facet{value, count})
	}
	for value, count := range categoryCount {
		categories = append(categories, Facet{value, count})
	}
	sortFacets(statuses)
	sortFacets(categories)
	return
}

func sortFacets(items []Facet) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Count > items[i].Count || (items[j].Count == items[i].Count && items[j].Value < items[i].Value) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}

func Explain(plan QueryPlan) []string {
	steps := []string{"load records"}
	if len(plan.Filter.Statuses) > 0 {
		steps = append(steps, "filter statuses")
	}
	if len(plan.Filter.Categories) > 0 {
		steps = append(steps, "filter categories")
	}
	if plan.Filter.Text != "" {
		steps = append(steps, "match text")
	}
	steps = append(steps, "sort by "+plan.SortField)
	steps = append(steps, fmt.Sprintf("page offset=%d limit=%d", plan.Offset, plan.Limit))
	return steps
}

func Cursor(records []domain.Record, after string, limit int) Page {
	start := 0
	if after != "" {
		for index, record := range records {
			if record.ID == after {
				start = index + 1
				break
			}
		}
	}
	return Paginate(records, start, limit)
}
