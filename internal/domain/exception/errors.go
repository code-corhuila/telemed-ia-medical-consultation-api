// Package exception defines the typed errors of the medical-consultation domain.
//
// The application layer translates these errors to HTTP status codes at the
// adapter boundary. The domain itself does not know what an HTTP status is.
package exception

import "errors"

var (
	// ErrConsultationNotFound is returned when a consultation cannot be located.
	ErrConsultationNotFound = errors.New("consultation not found")

	// ErrAttentionSummaryNotFound is returned when an attention summary cannot be located.
	ErrAttentionSummaryNotFound = errors.New("attention summary not found")

	// ErrPostSummaryNotFound is returned when a post-consultation summary cannot be located.
	ErrPostSummaryNotFound = errors.New("post summary not found")

	// ErrConsultationAlreadyCompleted is returned when trying to complete a consultation
	// that is already completed.
	ErrConsultationAlreadyCompleted = errors.New("consultation already completed")

	// ErrConsultationNotCompleted is returned when trying to generate a post-summary
	// before the consultation is completed.
	ErrConsultationNotCompleted = errors.New("consultation is not completed")

	// ErrInvalidAttentionSummary is returned when the attention summary payload
	// violates a domain invariant.
	ErrInvalidAttentionSummary = errors.New("invalid attention summary")

	// ErrInvalidPostSummary is returned when the post-summary payload violates
	// a domain invariant.
	ErrInvalidPostSummary = errors.New("invalid post summary")

	// ErrInvalidConsultation is returned when the consultation payload violates
	// a domain invariant.
	ErrInvalidConsultation = errors.New("invalid consultation")

	// ErrUnauthorized is returned when the authenticated subject is not allowed
	// to perform the requested operation on the resource.
	ErrUnauthorized = errors.New("unauthorized")
)
