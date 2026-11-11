// Package appointments reads appointments from appointment-scheduling over HTTP (DEC-006).
package appointments

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"net/url"
	"time"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/application/ports/out"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/correlation"
	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/exception"
)

// maxBody bounds the appointment answer we read.
const maxBody = 64 << 10

// Client implements out.AppointmentReaderPort.
type Client struct {
	baseURL string
	http    *nethttp.Client
}

// NewClient builds a client for baseURL (e.g. http://appointment-scheduling-api:8080).
// timeout bounds the whole call, so a slow appointment service cannot hold the request.
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{baseURL: baseURL, http: &nethttp.Client{Timeout: timeout}}
}

type appointmentBody struct {
	ID             string `json:"id"`
	PatientID      string `json:"patientId"`
	ProfessionalID string `json:"professionalId"`
	Status         string `json:"status"`
}

// FindAppointment calls GET /api/appointments/{id} with the caller's bearer token.
func (c *Client) FindAppointment(ctx context.Context, appointmentID, accessToken string) (*out.AppointmentView, error) {
	req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodGet,
		c.baseURL+"/api/appointments/"+url.PathEscape(appointmentID), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", exception.ErrAppointmentServiceUnavailable, err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	if id := correlation.FromContext(ctx); id != "" {
		req.Header.Set("X-Correlation-Id", id)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		// Timeout or network error. The error text never carries the token.
		return nil, fmt.Errorf("%w: %v", exception.ErrAppointmentServiceUnavailable, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case nethttp.StatusOK:
		var body appointmentBody
		if err := json.NewDecoder(io.LimitReader(resp.Body, maxBody)).Decode(&body); err != nil {
			return nil, fmt.Errorf("%w: unreadable appointment: %v", exception.ErrAppointmentServiceUnavailable, err)
		}
		return &out.AppointmentView{
			ID: body.ID, PatientID: body.PatientID, ProfessionalID: body.ProfessionalID, Status: body.Status,
		}, nil
	case nethttp.StatusNotFound:
		return nil, exception.ErrAppointmentNotFound
	case nethttp.StatusForbidden:
		return nil, exception.ErrForbidden
	default:
		return nil, fmt.Errorf("%w: appointment service answered %d",
			exception.ErrAppointmentServiceUnavailable, resp.StatusCode)
	}
}
