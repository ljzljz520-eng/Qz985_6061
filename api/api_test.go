package api

import (
	"bytes"
	"example.com/knowledge-backend/service"
	"example.com/knowledge-backend/store"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestHTTPWorkflow(t *testing.T) {
	repository, err := store.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	server := NewServer(service.New(repository, nil), 1<<20)
	request := httptest.NewRequest(http.MethodPost, "/records", bytes.NewBufferString(`{"id":"r1","title":"空调异味","content":"更换滤芯并清洁风道","category":"maintenance"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("%d %s", response.Code, response.Body.String())
	}
	get := httptest.NewRequest(http.MethodGet, "/records/r1", nil)
	getResponse := httptest.NewRecorder()
	server.ServeHTTP(getResponse, get)
	if getResponse.Code != http.StatusOK {
		t.Fatal(getResponse.Code)
	}
}
