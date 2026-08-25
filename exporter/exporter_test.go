package exporter

import (
	"example.com/knowledge-backend/domain"
	"testing"
	"time"
)

func TestRenderMaterial(t *testing.T) {
	record := domain.Record{ID: "r1", Title: "轮胎", Content: "检查胎压", Category: "maintenance", Status: domain.StatusApproved, Version: 3}
	material := RenderMaterial(record, time.Unix(1, 0))
	if material.Status != domain.StatusApproved || VisibleStatus(material.Status) != "target" {
		t.Fatal(material)
	}
	bundle := RenderBundle("daily", []domain.Record{record}, time.Now())
	data, err := EncodeCSV(bundle)
	if err != nil || len(data) == 0 {
		t.Fatal(err)
	}
}
