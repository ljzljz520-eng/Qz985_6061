package api

import (
	"example.com/knowledge-backend/domain"
	"example.com/knowledge-backend/query"
	"example.com/knowledge-backend/service"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	service *service.Service
	maxBody int64
}

func NewServer(svc *service.Service, maxBody int64) *Server {
	if maxBody <= 0 {
		maxBody = 1 << 20
	}
	return &Server{service: svc, maxBody: maxBody}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := splitPath(r.URL.Path)
	if len(parts) == 0 {
		writeJSON(w, http.StatusOK, map[string]string{"service": "knowledge-backend", "status": "ready"})
		return
	}
	switch parts[0] {
	case "records":
		s.handleRecords(w, r, parts[1:])
	case "export":
		s.handleExport(w, r, parts[1:])
	case "health":
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleRecords(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 0 {
		if r.Method == http.MethodPost {
			s.createRecord(w, r)
			return
		}
		if r.Method == http.MethodGet {
			s.listRecords(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := parts[0]
	if len(parts) == 1 && r.Method == http.MethodGet {
		record, err := s.service.GetRecord(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, record)
		return
	}
	if len(parts) == 2 && parts[1] == "review" && r.Method == http.MethodPost {
		s.reviewRecord(w, r, id)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type createRequest struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
}

func (s *Server) createRecord(w http.ResponseWriter, r *http.Request) {
	var input createRequest
	if err := decodeJSON(r, &input, s.maxBody); err != nil {
		writeError(w, domain.Wrap(domain.CodeValidation, "api.create", err))
		return
	}
	result, err := s.service.RunRegistration(service.RegistrationInput{Actor: actorFromRequest(r), ID: input.ID, Title: input.Title, Content: input.Content, Category: input.Category})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

type reviewRequest struct {
	Approve bool `json:"approve"`
}

func (s *Server) reviewRecord(w http.ResponseWriter, r *http.Request, id string) {
	var input reviewRequest
	if err := decodeJSON(r, &input, s.maxBody); err != nil {
		writeError(w, domain.Wrap(domain.CodeValidation, "api.review", err))
		return
	}
	result, err := s.service.RunReview(service.ReviewInput{Actor: actorFromRequest(r), RecordID: id, Approve: input.Approve})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) listRecords(w http.ResponseWriter, r *http.Request) {
	records, err := s.service.ListRecords()
	if err != nil {
		writeError(w, err)
		return
	}
	filter := query.Filter{Text: r.URL.Query().Get("q")}
	if status := r.URL.Query().Get("status"); status != "" {
		parsed, parseErr := domain.ParseStatus(status)
		if parseErr != nil {
			writeError(w, parseErr)
			return
		}
		filter.Statuses = []domain.Status{parsed}
	}
	records = query.Apply(records, filter)
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	writeJSON(w, http.StatusOK, query.Paginate(records, offset, limit))
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request, parts []string) {
	if r.Method != http.MethodGet || len(parts) != 1 {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	target := domain.StatusApproved
	if value := r.URL.Query().Get("status"); value != "" {
		parsed, err := domain.ParseStatus(value)
		if err != nil {
			writeError(w, err)
			return
		}
		target = parsed
	}
	result, err := s.service.RunExport(service.ExportInput{Actor: actorFromRequest(r), RecordID: parts[0], Target: target})
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	writeJSON(w, http.StatusOK, result)
}
