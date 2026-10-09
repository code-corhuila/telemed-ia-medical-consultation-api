// Package http contains the inbound HTTP adapter.
package http

import (
	"encoding/json"
	"net/http"
)

// Envelope is the common response shape of every service in the platform
// (norm 5.3.5). Success responses carry `data`; error responses carry
// `error` and `message`.
type Envelope struct {
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	TraceID string `json:"traceId,omitempty"`
}

// WriteJSON writes an Envelope with the given status.
func WriteJSON(w http.ResponseWriter, status int, payload Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// WriteData writes a success envelope.
func WriteData(w http.ResponseWriter, status int, data any) {
	WriteJSON(w, status, Envelope{Data: data})
}

// WriteError writes an error envelope with the given code and message.
func WriteError(w http.ResponseWriter, status int, code, message, traceID string) {
	WriteJSON(w, status, Envelope{
		Error:   code,
		Message: message,
		TraceID: traceID,
	})
}

// DecodeJSON reads the body into v, rejecting unknown fields and trailing
// data to keep the contract tight.
func DecodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
