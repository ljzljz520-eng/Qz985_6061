package domain

import (
	"errors"
	"testing"
	"time"
)

func TestValidateRecord(t *testing.T) {
	record := NewRecord("r1", "刹车异响", "检查刹车片和卡钳间隙", "brakes", time.Unix(100, 0))
	if err := ValidateRecord(record); err != nil {
		t.Fatal(err)
	}
	record.Title = ""
	if !IsCode(ValidateRecord(record), CodeValidation) {
		t.Fatal("expected validation")
	}
}

func TestTransitionStatus(t *testing.T) {
	record := NewRecord("r1", "title", "sufficient content", "engine", time.Now())
	if err := TransitionStatus(&record, StatusPending); err != nil {
		t.Fatal(err)
	}
	if err := TransitionStatus(&record, StatusArchived); err == nil {
		t.Fatal("expected conflict")
	}
	wrapped := Wrap(CodeConflict, "x", errors.New("y"))
	if wrapped == nil || !IsCode(wrapped, CodeConflict) {
		t.Fatal("wrap")
	}
}
