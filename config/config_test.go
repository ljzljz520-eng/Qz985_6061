package config

import (
	"testing"
	"time"
)

func TestConfigDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.Address != "127.0.0.1:8080" || cfg.DatabasePath != "knowledge.db" {
		t.Fatal(cfg)
	}
	built, err := Build(WithAddress("localhost:9000"), WithTimeouts(time.Second, 2*time.Second, 3*time.Second))
	if err != nil || built.Address != "localhost:9000" {
		t.Fatalf("%v %#v", err, built)
	}
}
