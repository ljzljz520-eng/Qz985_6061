package exporter

import (
	"example.com/knowledge-backend/domain"
	"fmt"
	"sort"
	"strings"
)

type Catalog struct{ bundles map[string]Bundle }

func NewCatalog() *Catalog { return &Catalog{bundles: make(map[string]Bundle)} }

func (c *Catalog) Add(bundle Bundle) error {
	name := strings.TrimSpace(bundle.Name)
	if name == "" {
		return fmt.Errorf("bundle name is required")
	}
	if len(bundle.Materials) == 0 {
		return fmt.Errorf("bundle is empty")
	}
	for _, item := range bundle.Materials {
		if err := ValidateMaterial(item); err != nil {
			return err
		}
	}
	bundle.Name = name
	bundle.Count = len(bundle.Materials)
	c.bundles[name] = bundle
	return nil
}

func (c *Catalog) Get(name string) (Bundle, bool) { bundle, ok := c.bundles[name]; return bundle, ok }

func (c *Catalog) Names() []string {
	names := make([]string, 0, len(c.bundles))
	for name := range c.bundles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c *Catalog) Remove(name string) bool {
	if _, ok := c.bundles[name]; !ok {
		return false
	}
	delete(c.bundles, name)
	return true
}

func (c *Catalog) CountByStatus() map[domain.Status]int {
	result := map[domain.Status]int{}
	for _, bundle := range c.bundles {
		for _, item := range bundle.Materials {
			result[item.Status]++
		}
	}
	return result
}

func (c *Catalog) Search(term string) []Material {
	needle := strings.ToLower(strings.TrimSpace(term))
	result := []Material{}
	for _, bundle := range c.bundles {
		for _, item := range bundle.Materials {
			if needle == "" || strings.Contains(strings.ToLower(item.Title), needle) || strings.Contains(strings.ToLower(item.Content), needle) {
				result = append(result, item)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Category == result[j].Category {
			return result[i].Title < result[j].Title
		}
		return result[i].Category < result[j].Category
	})
	return result
}
