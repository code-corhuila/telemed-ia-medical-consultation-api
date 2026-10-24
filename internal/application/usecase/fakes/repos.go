package fakes

import (
	"context"
	"fmt"
	"sync"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

// ConsultationRepo is an in-memory fake for tests.
type ConsultationRepo struct {
	mu   sync.Mutex
	data map[string]*model.Consultation
	next int64
}

func NewConsultationRepo() *ConsultationRepo {
	return &ConsultationRepo{data: map[string]*model.Consultation{}, next: 1}
}

func (r *ConsultationRepo) Save(_ context.Context, c *model.Consultation) (*model.Consultation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c.ID == "" {
		c.ID = fmt.Sprintf("00000000-0000-0000-0000-%012d", r.next)
		r.next++
	}
	cp := *c
	r.data[c.ID] = &cp
	return &cp, nil
}

func (r *ConsultationRepo) FindByID(_ context.Context, id string) (*model.Consultation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.data[id]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, nil
}

func (r *ConsultationRepo) FindByAppointmentID(_ context.Context, appointmentID string) (*model.Consultation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.data {
		if c.AppointmentID == appointmentID {
			cp := *c
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *ConsultationRepo) Complete(_ context.Context, id string) (*model.Consultation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.data[id]
	if !ok {
		return nil, nil
	}
	_ = c.Complete()
	cp := *c
	return &cp, nil
}

// AttentionRepo is an in-memory fake for tests.
type AttentionRepo struct {
	mu   sync.Mutex
	data map[string]*model.AttentionSummary
	next int64
}

func NewAttentionRepo() *AttentionRepo {
	return &AttentionRepo{data: map[string]*model.AttentionSummary{}, next: 1}
}

func (r *AttentionRepo) Save(_ context.Context, s *model.AttentionSummary) (*model.AttentionSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.ID == "" {
		s.ID = fmt.Sprintf("00000000-0000-0000-0000-%012d", r.next)
		r.next++
	}
	cp := *s
	r.data[s.ID] = &cp
	return &cp, nil
}

func (r *AttentionRepo) SaveAndComplete(_ context.Context, s *model.AttentionSummary) (*model.AttentionSummary, error) {
	return r.Save(context.Background(), s)
}

func (r *AttentionRepo) FindByID(_ context.Context, id string) (*model.AttentionSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.data[id]; ok {
		cp := *s
		return &cp, nil
	}
	return nil, nil
}

func (r *AttentionRepo) FindByConsultationID(_ context.Context, consultationID string) (*model.AttentionSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.data {
		if s.ConsultationID == consultationID {
			cp := *s
			return &cp, nil
		}
	}
	return nil, nil
}

// PostSummaryRepo is an in-memory fake for tests.
type PostSummaryRepo struct {
	mu   sync.Mutex
	data map[string]*model.PostSummary
	next int64
}

func NewPostSummaryRepo() *PostSummaryRepo {
	return &PostSummaryRepo{data: map[string]*model.PostSummary{}, next: 1}
}

func (r *PostSummaryRepo) SaveWithOutbox(_ context.Context, s *model.PostSummary, _ *model.OutboxEvent) (*model.PostSummary, error) {
	return r.Save(context.Background(), s)
}

func (r *PostSummaryRepo) Save(_ context.Context, s *model.PostSummary) (*model.PostSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.ID == "" {
		s.ID = fmt.Sprintf("00000000-0000-0000-0000-%012d", r.next)
		r.next++
	}
	cp := *s
	r.data[s.ID] = &cp
	return &cp, nil
}

func (r *PostSummaryRepo) FindByID(_ context.Context, id string) (*model.PostSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.data[id]; ok {
		cp := *s
		return &cp, nil
	}
	return nil, nil
}

func (r *PostSummaryRepo) FindByConsultationID(_ context.Context, consultationID string) (*model.PostSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.data {
		if s.ConsultationID == consultationID {
			cp := *s
			return &cp, nil
		}
	}
	return nil, nil
}
