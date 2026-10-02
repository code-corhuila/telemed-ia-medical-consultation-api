package model

import (
	"errors"
	"strings"
)

// MedicationInstruction is a value object representing one medication
// recorded by the healthcare professional during the consultation.
//
// It has no identity of its own: two instructions with the same fields are
// equivalent. The aggregate owns its lifecycle.
type MedicationInstruction struct {
	Name        string  `json:"name"`
	Dosage      *string `json:"dosage,omitempty"`
	Frequency   *string `json:"frequency,omitempty"`
	Duration    *string `json:"duration,omitempty"`
	Instructions *string `json:"instructions,omitempty"`
}

// NewMedicationInstruction creates a valid medication instruction.
// Only Name is required by the contract; the other fields are optional.
func NewMedicationInstruction(
	name string,
	dosage, frequency, duration, instructions *string,
) (MedicationInstruction, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return MedicationInstruction{}, errors.New("medication name is required")
	}
	return MedicationInstruction{
		Name:         trimmed,
		Dosage:       clean(dosage),
		Frequency:    clean(frequency),
		Duration:     clean(duration),
		Instructions: clean(instructions),
	}, nil
}

func clean(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}
