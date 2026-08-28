package exporter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"example.com/knowledge-backend/domain"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Manifest struct {
	Name       string                `json:"name"`
	CreatedAt  time.Time             `json:"created_at"`
	Hash       string                `json:"hash"`
	Records    int                   `json:"records"`
	Statuses   map[domain.Status]int `json:"statuses"`
	Categories []string              `json:"categories"`
}

func BuildManifest(bundle Bundle, now time.Time) Manifest {
	payload, _ := json.Marshal(bundle)
	sum := sha256.Sum256(payload)
	statuses := map[domain.Status]int{}
	categories := map[string]bool{}
	for _, item := range bundle.Materials {
		statuses[item.Status]++
		categories[item.Category] = true
	}
	categoryList := make([]string, 0, len(categories))
	for category := range categories {
		categoryList = append(categoryList, category)
	}
	sort.Strings(categoryList)
	return Manifest{Name: bundle.Name, CreatedAt: now.UTC(), Hash: hex.EncodeToString(sum[:]), Records: len(bundle.Materials), Statuses: statuses, Categories: categoryList}
}

func (m Manifest) Validate(bundle Bundle) error {
	if strings.TrimSpace(m.Name) == "" || m.Name != bundle.Name {
		return fmt.Errorf("manifest name mismatch")
	}
	if m.Records != len(bundle.Materials) {
		return fmt.Errorf("manifest count mismatch")
	}
	if m.Hash == "" {
		return fmt.Errorf("manifest hash missing")
	}
	return nil
}

func (m Manifest) Label() string {
	return fmt.Sprintf("%s/%d/%s", m.Name, m.Records, m.Hash[:minInt(12, len(m.Hash))])
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Reconcile(bundle Bundle, manifest Manifest) (bool, error) {
	expected := BuildManifest(bundle, manifest.CreatedAt)
	if err := manifest.Validate(bundle); err != nil {
		return false, err
	}
	return expected.Hash == manifest.Hash, nil
}
