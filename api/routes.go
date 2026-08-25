package api

import (
	"encoding/json"
	"errors"
	"example.com/knowledge-backend/domain"
	"io"
	"net/http"
	"strings"
)

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal"
	switch {
	case domain.IsCode(err, domain.CodeValidation):
		status = http.StatusBadRequest
		code = "validation"
	case domain.IsCode(err, domain.CodeNotFound):
		status = http.StatusNotFound
		code = "not_found"
	case domain.IsCode(err, domain.CodeConflict):
		status = http.StatusConflict
		code = "conflict"
	case domain.IsCode(err, domain.CodePermission):
		status = http.StatusForbidden
		code = "permission"
	}
	writeJSON(w, status, errorResponse{Error: err.Error(), Code: code})
}

func decodeJSON(r *http.Request, target any, max int64) error {
	reader := io.Reader(r.Body)
	if max > 0 {
		reader = http.MaxBytesReader(nil, r.Body, max)
	}
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("request must contain one JSON value")
	}
	return nil
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func actorFromRequest(r *http.Request) domain.User {
	id := strings.TrimSpace(r.Header.Get("X-Actor-ID"))
	if id == "" {
		id = "api-user"
	}
	role := strings.TrimSpace(r.Header.Get("X-Actor-Role"))
	if role == "" {
		role = "manager"
	}
	return domain.NewUser(id, id, role, true)
}
