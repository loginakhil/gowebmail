package api

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"gowebmail/internal/email"
	"gowebmail/internal/storage"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// handleListEmails handles GET /api/emails
func (s *Server) handleListEmails(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit := parseIntParam(r, "limit", 50, 1, 100)
	offset := parseIntParam(r, "offset", 0, 0, math.MaxInt)

	// Build filter
	filter := &storage.EmailFilter{
		From:    r.URL.Query().Get("from"),
		To:      r.URL.Query().Get("to"),
		Subject: r.URL.Query().Get("subject"),
	}

	// Parse date filters
	if since := r.URL.Query().Get("since"); since != "" {
		if t, err := time.Parse(time.RFC3339, since); err == nil {
			filter.Since = &t
		}
	}
	if until := r.URL.Query().Get("until"); until != "" {
		if t, err := time.Parse(time.RFC3339, until); err == nil {
			filter.Until = &t
		}
	}

	// Get emails
	result, err := s.storage.ListEmails(filter, limit, offset)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "STORAGE_ERROR", err.Error())
		return
	}

	s.sendSuccess(w, map[string]interface{}{
		"emails": result.Emails,
		"total":  result.Total,
		"limit":  limit,
		"offset": offset,
	})
}

// handleGetEmail handles GET /api/emails/{id}
func (s *Server) handleGetEmail(w http.ResponseWriter, r *http.Request) {
	id := parseIDParam(r)
	if id == 0 {
		s.sendError(w, http.StatusBadRequest, "INVALID_ID", "Invalid email ID")
		return
	}

	email, err := s.storage.GetEmail(id)
	if err != nil {
		if err == storage.ErrNotFound {
			s.sendError(w, http.StatusNotFound, "NOT_FOUND", "Email not found")
		} else {
			s.sendError(w, http.StatusInternalServerError, "STORAGE_ERROR", err.Error())
		}
		return
	}

	s.sendSuccess(w, email)
}

// handleDeleteEmail handles DELETE /api/emails/{id}
func (s *Server) handleDeleteEmail(w http.ResponseWriter, r *http.Request) {
	id := parseIDParam(r)
	if id == 0 {
		s.sendError(w, http.StatusBadRequest, "INVALID_ID", "Invalid email ID")
		return
	}

	err := s.storage.DeleteEmail(id)
	if err != nil {
		if err == storage.ErrNotFound {
			s.sendError(w, http.StatusNotFound, "NOT_FOUND", "Email not found")
		} else {
			s.sendError(w, http.StatusInternalServerError, "STORAGE_ERROR", err.Error())
		}
		return
	}

	// Notify WebSocket clients
	s.wsHub.Broadcast(&WebSocketMessage{
		Type: "email.deleted",
		Data: map[string]interface{}{"id": id},
	})

	s.sendSuccess(w, map[string]interface{}{"deleted": id})
}

// handleDeleteAllEmails handles DELETE /api/emails
func (s *Server) handleDeleteAllEmails(w http.ResponseWriter, r *http.Request) {
	err := s.storage.DeleteAllEmails()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "STORAGE_ERROR", err.Error())
		return
	}

	// Notify WebSocket clients
	s.wsHub.Broadcast(&WebSocketMessage{
		Type: "emails.cleared",
		Data: map[string]interface{}{},
	})

	s.sendSuccess(w, map[string]interface{}{"message": "All emails deleted"})
}

// handleSearchEmails handles GET /api/emails/search
func (s *Server) handleSearchEmails(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		s.sendError(w, http.StatusBadRequest, "INVALID_REQUEST", "Search query is required")
		return
	}

	limit := parseIntParam(r, "limit", 50, 1, 100)
	offset := parseIntParam(r, "offset", 0, 0, math.MaxInt)

	result, err := s.storage.SearchEmails(query, limit, offset)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "STORAGE_ERROR", err.Error())
		return
	}

	s.sendSuccess(w, map[string]interface{}{
		"emails": result.Emails,
		"total":  result.Total,
		"limit":  limit,
		"offset": offset,
	})
}

// handleGetEmailRaw handles GET /api/emails/{id}/raw
func (s *Server) handleGetEmailRaw(w http.ResponseWriter, r *http.Request) {
	id := parseIDParam(r)
	if id == 0 {
		s.sendError(w, http.StatusBadRequest, "INVALID_ID", "Invalid email ID")
		return
	}

	email, err := s.storage.GetEmail(id)
	if err != nil {
		if err == storage.ErrNotFound {
			s.sendError(w, http.StatusNotFound, "NOT_FOUND", "Email not found")
		} else {
			s.sendError(w, http.StatusInternalServerError, "STORAGE_ERROR", err.Error())
		}
		return
	}

	// Build raw email
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	
	// Write headers
	for key, values := range email.Headers {
		for _, value := range values {
			fmt.Fprintf(w, "%s: %s\r\n", key, value)
		}
	}
	
	fmt.Fprintf(w, "\r\n")
	
	// Write body (prefer plain text)
	if email.BodyPlain != "" {
		fmt.Fprint(w, email.BodyPlain)
	} else if email.BodyHTML != "" {
		fmt.Fprint(w, email.BodyHTML)
	}
}

// handleGetEmailHTML handles GET /api/emails/{id}/html
func (s *Server) handleGetEmailHTML(w http.ResponseWriter, r *http.Request) {
	id := parseIDParam(r)
	if id == 0 {
		s.sendError(w, http.StatusBadRequest, "INVALID_ID", "Invalid email ID")
		return
	}

	emailData, err := s.storage.GetEmail(id)
	if err != nil {
		if err == storage.ErrNotFound {
			s.sendError(w, http.StatusNotFound, "NOT_FOUND", "Email not found")
		} else {
			s.sendError(w, http.StatusInternalServerError, "STORAGE_ERROR", err.Error())
		}
		return
	}

	if emailData.BodyHTML == "" {
		s.sendError(w, http.StatusNotFound, "NOT_FOUND", "No HTML body available")
		return
	}

	// Sanitize HTML
	sanitizer := email.NewSanitizer()
	sanitized := sanitizer.Sanitize(emailData.BodyHTML)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; img-src data:")
	fmt.Fprint(w, sanitized)
}

// handleGetAttachment handles GET /api/emails/{id}/attachments/{aid}
func (s *Server) handleGetAttachment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	
	aid, err := strconv.ParseInt(vars["aid"], 10, 64)
	if err != nil || aid <= 0 {
		s.sendError(w, http.StatusBadRequest, "INVALID_ID", "Invalid attachment ID")
		return
	}

	attachment, err := s.storage.GetAttachment(aid)
	if err != nil {
		if err == storage.ErrNotFound {
			s.sendError(w, http.StatusNotFound, "NOT_FOUND", "Attachment not found")
		} else {
			s.sendError(w, http.StatusInternalServerError, "STORAGE_ERROR", err.Error())
		}
		return
	}

	// Set headers
	w.Header().Set("Content-Type", attachment.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", attachment.Filename))
	w.Header().Set("Content-Length", strconv.FormatInt(attachment.Size, 10))

	// Write data
	w.Write(attachment.Data)
}

// handleGetStats handles GET /api/stats
func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	count, err := s.storage.GetEmailCount()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "STORAGE_ERROR", err.Error())
		return
	}

	// Get today's count
	today := time.Now().Truncate(24 * time.Hour)
	filter := &storage.EmailFilter{Since: &today}
	todayResult, _ := s.storage.ListEmails(filter, 1, 0)
	todayCount := int64(0)
	if todayResult != nil {
		todayCount = todayResult.Total
	}

	s.sendSuccess(w, map[string]interface{}{
		"totalEmails": count,
		"todayCount":  todayCount,
	})
}

// handleHealth handles GET /api/health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.sendSuccess(w, map[string]interface{}{
		"status":  "healthy",
		"version": "1.0.0",
	})
}

// sendSuccess sends a successful API response
func (s *Server) sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

// sendError sends an error API response
func (s *Server) sendError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

// parseIntParam parses an integer query parameter with default and bounds
func parseIntParam(r *http.Request, name string, defaultValue, min, max int) int {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	if parsed < min {
		return min
	}
	if parsed > max {
		return max
	}

	return parsed
}

// parseIDParam parses the ID parameter from the URL
func parseIDParam(r *http.Request) int64 {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return 0
	}
	return id
}
