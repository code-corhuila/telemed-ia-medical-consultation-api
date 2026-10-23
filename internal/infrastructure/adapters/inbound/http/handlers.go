package http

import (
	"encoding/json"
	"errors"
	"log"
	nethttp "net/http"
	"regexp"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/in"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/out"
)

// Handlers holds the use cases the router invokes.
type Handlers struct {
	recordAttention  in.RecordAttentionPort
	getAttention     in.GetAttentionPort
	getPostSummary   in.GetPostSummaryPort
	generatePostSumm in.GeneratePostSummaryPort
	consultations    out.ConsultationRepositoryPort
}

// NewHandlers wires the use cases.
func NewHandlers(
	recordAttention in.RecordAttentionPort,
	getAttention in.GetAttentionPort,
	getPostSummary in.GetPostSummaryPort,
	generatePostSumm in.GeneratePostSummaryPort,
	consultations out.ConsultationRepositoryPort,
) *Handlers {
	return &Handlers{
		recordAttention:  recordAttention,
		getAttention:     getAttention,
		getPostSummary:   getPostSummary,
		generatePostSumm: generatePostSumm,
		consultations:    consultations,
	}
}

// Health is the liveness endpoint. It is intentionally unauthenticated.
func (h *Handlers) Health(w nethttp.ResponseWriter, _ *nethttp.Request) {
	WriteData(w, nethttp.StatusOK, map[string]string{
		"status":  "UP",
		"service": "medical-consultation-api",
	})
}

// RecordAttention handles POST /api/consultations/{id}/attention
func (h *Handlers) RecordAttention(w nethttp.ResponseWriter, r *nethttp.Request) {
	trace := traceFrom(r.Context())

	id, err := pathUUID(r)
	if err != nil {
		WriteError(w, nethttp.StatusBadRequest, "INVALID_PATH",
			"consultation id must be a UUID", trace)
		return
	}

	var body dto.RecordAttentionCommand
	if err := DecodeJSON(r, &body); err != nil {
		WriteError(w, nethttp.StatusBadRequest, "INVALID_PAYLOAD",
			"malformed JSON body", trace)
		return
	}
	body.ConsultationID = id

	claims := FromContext(r.Context())
	if claims == nil || claims.Role != "PROFESSIONAL" {
		WriteError(w, nethttp.StatusForbidden, "FORBIDDEN", "only professionals can record attention", trace)
		return
	}

	c, err := h.consultations.FindByID(r.Context(), body.ConsultationID)
	if err != nil || c == nil {
		WriteError(w, nethttp.StatusNotFound, "NOT_FOUND", "consultation not found", trace)
		return
	}
	if c.ProfessionalID != claims.Subject {
		WriteError(w, nethttp.StatusForbidden, "FORBIDDEN", "you are not the assigned professional", trace)
		return
	}

	if len(body.FollowUp) > 0 && string(body.FollowUp) != "null" {
		var obj map[string]interface{}
		if err := json.Unmarshal(body.FollowUp, &obj); err != nil {
			WriteError(w, nethttp.StatusBadRequest, "INVALID_PAYLOAD", "followUp must be a JSON object", trace)
			return
		}
	}
	resp, err := h.recordAttention.Execute(r.Context(), body)
	if err != nil {
		status, code := mapError(err)
		log.Printf("[%s] ERROR: %v", trace, err)

		WriteError(w, status, code, safeErrorMessage(err, status), trace)
		return
	}
	WriteData(w, nethttp.StatusCreated, resp)
}

// GetAttention handles GET /api/consultations/{id}/attention
func (h *Handlers) GetAttention(w nethttp.ResponseWriter, r *nethttp.Request) {
	trace := traceFrom(r.Context())

	id, err := pathUUID(r)
	if err != nil {
		WriteError(w, nethttp.StatusBadRequest, "INVALID_PATH",
			"consultation id must be a UUID", trace)
		return
	}

	resp, err := h.getAttention.Execute(r.Context(), id)
	if err != nil {
		status, code := mapError(err)
		log.Printf("[%s] ERROR: %v", trace, err)

		WriteError(w, status, code, safeErrorMessage(err, status), trace)
		return
	}

	// RBAC Check
	claims := FromContext(r.Context())
	if claims == nil {
		WriteError(w, nethttp.StatusUnauthorized, "UNAUTHORIZED", "missing claims", trace)
		return
	}
	if claims.Role == "SERVICE" && claims.Subject == "document-generation" {
		WriteData(w, nethttp.StatusOK, resp)
		return
	}

	c, err := h.consultations.FindByID(r.Context(), resp.ConsultationID)
	if err == nil && c != nil {
		if claims.Role == "PATIENT" && c.PatientID == claims.Subject {
			WriteData(w, nethttp.StatusOK, resp)
			return
		}
		if claims.Role == "PROFESSIONAL" && c.ProfessionalID == claims.Subject {
			WriteData(w, nethttp.StatusOK, resp)
			return
		}
	}

	WriteError(w, nethttp.StatusForbidden, "FORBIDDEN", "insufficient permissions", trace)
}

// GetPostSummary handles GET /api/post-summaries/{id}
func (h *Handlers) GetPostSummary(w nethttp.ResponseWriter, r *nethttp.Request) {
	trace := traceFrom(r.Context())

	id, err := pathUUID(r)
	if err != nil {
		WriteError(w, nethttp.StatusBadRequest, "INVALID_PATH",
			"post summary id must be a UUID", trace)
		return
	}

	resp, err := h.getPostSummary.Execute(r.Context(), id)
	if err != nil {
		status, code := mapError(err)
		log.Printf("[%s] ERROR: %v", trace, err)

		WriteError(w, status, code, safeErrorMessage(err, status), trace)
		return
	}

	// RBAC Check
	claims := FromContext(r.Context())
	if claims == nil {
		WriteError(w, nethttp.StatusUnauthorized, "UNAUTHORIZED", "missing claims", trace)
		return
	}
	if claims.Role == "SERVICE" && claims.Subject == "document-generation" {
		WriteData(w, nethttp.StatusOK, resp)
		return
	}

	c, err := h.consultations.FindByID(r.Context(), resp.ConsultationID)
	if err == nil && c != nil {
		if claims.Role == "PATIENT" && c.PatientID == claims.Subject {
			WriteData(w, nethttp.StatusOK, resp)
			return
		}
		if claims.Role == "PROFESSIONAL" && c.ProfessionalID == claims.Subject {
			WriteData(w, nethttp.StatusOK, resp)
			return
		}
	}

	WriteError(w, nethttp.StatusForbidden, "FORBIDDEN", "insufficient permissions", trace)
}

// GeneratePostSummary handles POST /api/consultations/{id}/post-summary
func (h *Handlers) GeneratePostSummary(w nethttp.ResponseWriter, r *nethttp.Request) {
	trace := traceFrom(r.Context())

	id, err := pathUUID(r)
	if err != nil {
		WriteError(w, nethttp.StatusBadRequest, "INVALID_PATH",
			"consultation id must be a UUID", trace)
		return
	}

	var body dto.GeneratePostSummaryCommand
	if err := DecodeJSON(r, &body); err != nil && !errors.Is(err, errors.New("")) {
		// Empty body is allowed for this endpoint (no fields required).
		body = dto.GeneratePostSummaryCommand{}
	}
	body.ConsultationID = id

	claims := FromContext(r.Context())
	if claims == nil || claims.Role != "PROFESSIONAL" {
		WriteError(w, nethttp.StatusForbidden, "FORBIDDEN", "only professionals can generate post-summaries", trace)
		return
	}

	c, err := h.consultations.FindByID(r.Context(), body.ConsultationID)
	if err != nil || c == nil {
		WriteError(w, nethttp.StatusNotFound, "NOT_FOUND", "consultation not found", trace)
		return
	}
	if c.ProfessionalID != claims.Subject {
		WriteError(w, nethttp.StatusForbidden, "FORBIDDEN", "you are not the assigned professional", trace)
		return
	}

	resp, err := h.generatePostSumm.Execute(r.Context(), body)
	if err != nil {
		status, code := mapError(err)
		log.Printf("[%s] ERROR: %v", trace, err)

		WriteError(w, status, code, safeErrorMessage(err, status), trace)
		return
	}
	WriteData(w, nethttp.StatusCreated, resp)
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// pathUUID extracts {id} from the request path and validates it is a UUID.
func pathUUID(r *nethttp.Request) (string, error) {
	raw := r.PathValue("id")
	if !uuidPattern.MatchString(raw) {
		return "", errors.New("invalid id")
	}
	return raw, nil
}

func safeErrorMessage(err error, status int) string {
	if status == nethttp.StatusInternalServerError {
		return "An unexpected error occurred. Please contact support."
	}
	return err.Error()
}
