package exporter

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"example.com/knowledge-backend/domain"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Material struct {
	RecordID    string        `json:"record_id"`
	Title       string        `json:"title"`
	Category    string        `json:"category"`
	Status      domain.Status `json:"status"`
	Content     string        `json:"content"`
	Revision    int           `json:"revision"`
	GeneratedAt time.Time     `json:"generated_at"`
	Labels      []string      `json:"labels"`
}

type Bundle struct {
	Name      string     `json:"name"`
	Materials []Material `json:"materials"`
	Count     int        `json:"count"`
}

func RenderMaterial(record domain.Record, now time.Time) Material {
	labels := []string{record.Category, record.Status.String()}
	sort.Strings(labels)
	return Material{RecordID: record.ID, Title: record.Title, Category: record.Category, Status: record.Status, Content: record.Content, Revision: record.Version, GeneratedAt: now.UTC(), Labels: labels}
}

func RenderBundle(name string, records []domain.Record, now time.Time) Bundle {
	items := make([]Material, 0, len(records))
	for _, record := range records {
		items = append(items, RenderMaterial(record, now))
	}
	return Bundle{Name: strings.TrimSpace(name), Materials: items, Count: len(items)}
}

func EncodeJSON(bundle Bundle) ([]byte, error) { return json.MarshalIndent(bundle, "", "  ") }

func EncodeCSV(bundle Bundle) ([]byte, error) {
	var output bytes.Buffer
	w := csv.NewWriter(&output)
	if err := w.Write([]string{"record_id", "title", "category", "status", "revision", "content"}); err != nil {
		return nil, err
	}
	for _, item := range bundle.Materials {
		if err := w.Write([]string{item.RecordID, item.Title, item.Category, item.Status.String(), fmt.Sprintf("%d", item.Revision), item.Content}); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func VisibleStatus(status domain.Status) string {
	switch status {
	case domain.StatusApproved:
		return "target"
	case domain.StatusArchived:
		return "archived"
	case domain.StatusPending:
		return "reviewing"
	case domain.StatusRejected:
		return "rejected"
	default:
		return "default"
	}
}

func ValidateMaterial(material Material) error {
	if material.RecordID == "" || material.Title == "" {
		return fmt.Errorf("material identity is incomplete")
	}
	if !domain.IsKnownStatus(material.Status) {
		return fmt.Errorf("material status is invalid")
	}
	if material.Revision < 1 {
		return fmt.Errorf("material revision is invalid")
	}
	return nil
}
