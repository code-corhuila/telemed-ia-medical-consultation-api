// Package appointments reads appointments from appointment-scheduling over HTTP (DEC-006).
package appointments

import (
	"context"
	"errors"
	nethttp "net/http"
	"time"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/out"
)

// Client implements out.AppointmentReaderPort.
type Client struct {
	baseURL string
	http    *nethttp.Client
}

// NewClient builds a client for baseURL (e.g. http://appointment-scheduling-api:8080).
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{baseURL: baseURL, http: &nethttp.Client{Timeout: timeout}}
}

func (c *Client) FindAppointment(_ context.Context, _, _ string) (*out.AppointmentView, error) {
	return nil, errors.New("not implemented")
}
