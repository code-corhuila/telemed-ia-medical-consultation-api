package fakes

import (
	"context"
	"sync"

	"github.com/code-corhuila/telemed-ia-medical-consultation-api/internal/domain/model"
)

// ConsultationRepo is an in-memory fake for tests.
type ConsultationRepo struct {
	mu   sync.Mutex
	data map[int64]*model.Consultation
	next int64
}

func NewConsultationRepo() *ConsultationRepo {
	return &ConsultationRepo{data: map[int64]*model.Consultation{}, next: 1}
}

func (r *ConsultationRepo) Save(_ context.Context, c *model.Consultation) (*model.Consultation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c.ID == 0 {
		c.ID = r.next
		r.next++
	}
	cp := *c
	r.data[c.ID] = &cp
	return &cp, nil
}

func (r *ConsultationRepo) FindByID(_ context.Context, id int64) (*model.Consultation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.data[id]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, nil
}

func (r *ConsultationRepo) FindByAppointmentID(_ context.Context, appointmentID int64) (*model.Consultation, error) {
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

func (r *ConsultationRepo) Complete(_ context.Context, id int64) (*model.Consultation, error) {
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
	data map[int64]*model.AttentionSummary
	next int64
}

func NewAttentionRepo() *AttentionRepo {
	return &AttentionRepo{data: map[int64]*model.AttentionSummary{}, next: 1}
}

func (r *AttentionRepo) Save(_ context.Context, s *model.AttentionSummary) (*model.AttentionSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.ID == 0 {
		s.ID = r.next
		r.next++
	}
	cp := *s
	r.data[s.ID] = &cp
	return &cp, nil
}

func (r *AttentionRepo) FindByID(_ context.Context, id int64) (*model.AttentionSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.data[id]; ok {
		cp := *s
		return &cp, nil
	}
	return nil, nil
}

func (r *AttentionRepo) FindByConsultationID(_ context.Context, consultationID int64) (*model.AttentionSummary, error) {
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
	data map[int64]*model.PostSummary
	next int64
}

func NewPostSummaryRepo() *PostSummaryRepo {
	return &PostSummaryRepo{data: map[int64]*model.PostSummary{}, next: 1}
}

func (r *PostSummaryRepo) Save(_ context.Context, s *model.PostSummary) (*model.PostSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.ID == 0 {
		s.ID = r.next
		r.next++
	}
	cp := *s
	r.data[s.ID] = &cp
	return &cp, nil
}

func (r *PostSummaryRepo) FindByID(_ context.Context, id int64) (*model.PostSummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s, ok := r.data[id]; ok {
		cp := *s
		return &cp, nil
	}
	return nil, nil
}

func (r *PostSummaryRepo) FindByConsultationID(_ context.Context, consultationID int64) (*model.PostSummary, error) {
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
