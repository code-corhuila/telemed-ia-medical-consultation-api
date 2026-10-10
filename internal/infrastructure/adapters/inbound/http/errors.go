package http

import (
	"errors"
	nethttp "net/http"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
)

// mapError translates a domain error to an HTTP status + a stable code.
func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, exception.ErrConsultationNotFound),
		errors.Is(err, exception.ErrAttentionSummaryNotFound),
		errors.Is(err, exception.ErrPostSummaryNotFound):
		return nethttp.StatusNotFound, "NOT_FOUND"

	case errors.Is(err, exception.ErrUnauthorized):
		return nethttp.StatusUnauthorized, "UNAUTHORIZED"

	case errors.Is(err, exception.ErrInvalidConsultation),
		errors.Is(err, exception.ErrInvalidAttentionSummary),
		errors.Is(err, exception.ErrInvalidPostSummary):
		return nethttp.StatusBadRequest, "INVALID_PAYLOAD"

	case errors.Is(err, exception.ErrConsultationAlreadyCompleted),
		errors.Is(err, exception.ErrConsultationNotCompleted):
		return nethttp.StatusConflict, "CONFLICT"

	default:
		return nethttp.StatusInternalServerError, "INTERNAL_ERROR"
	}
}
