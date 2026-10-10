package dto

// MedicationDTO is the wire representation of one medication.
type MedicationDTO struct {
	Name         string  `json:"name"`
	Dosage       *string `json:"dosage,omitempty"`
	Frequency    *string `json:"frequency,omitempty"`
	Duration     *string `json:"duration,omitempty"`
	Instructions *string `json:"instructions,omitempty"`
}
