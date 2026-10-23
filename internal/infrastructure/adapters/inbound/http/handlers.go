package http

import (
	"errors"
	nethttp "net/http"
	"regexp"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/dto"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/in"
)

// Handlers holds the use cases the router invokes.
type Handlers struct {
	recordAttention  in.RecordAttentionPort
	getAttention     in.GetAttentionPort
	getPostSummary   in.GetPostSummaryPort
	generatePostSumm in.GeneratePostSummaryPort
}

// NewHandlers wires the use cases.
func NewHandlers(
	recordAttention in.RecordAttentionPort,
	getAttention in.GetAttentionPort,
	getPostSummary in.GetPostSummaryPort,
	generatePostSumm in.GeneratePostSummaryPort,
) *Handlers {
	return &Handlers{
		recordAttention:  recordAttention,
		getAttention:     getAttention,
		getPostSummary:   getPostSummary,
		generatePostSumm: generatePostSumm,
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

	resp, err := h.recordAttention.Execute(r.Context(), body)
	if err != nil {
		status, code := mapError(err)
		WriteError(w, status, code, err.Error(), trace)
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
		WriteError(w, status, code, err.Error(), trace)
		return
	}
	WriteData(w, nethttp.StatusOK, resp)
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
		WriteError(w, status, code, err.Error(), trace)
		return
	}
	WriteData(w, nethttp.StatusOK, resp)
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

	resp, err := h.generatePostSumm.Execute(r.Context(), body)
	if err != nil {
		status, code := mapError(err)
		WriteError(w, status, code, err.Error(), trace)
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
