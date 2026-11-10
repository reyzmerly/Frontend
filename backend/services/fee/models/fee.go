package models

import "time"

type FeeType struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	ProgrammeID   uint      `json:"programme_id"`
	SemesterID    uint      `json:"semester_id"`
	Amount        int64     `json:"amount"`
	EffectiveFrom string    `json:"effective_from"`
	CreatedBy     uint      `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	ProgrammeName string    `json:"programme_name"`
	SemesterName  string    `json:"semester_name"`
}

type CreateFeeTypeRequest struct {
	Name          string `json:"name" validate:"required"`
	ProgrammeID   uint   `json:"programme_id" validate:"required"`
	SemesterID    uint   `json:"semester_id" validate:"required"`
	Amount        int64  `json:"amount" validate:"required,gt=0"`
	EffectiveFrom string `json:"effective_from" validate:"required"`
}

type UpdateFeeTypeRequest struct {
	Name          string `json:"name" validate:"required"`
	Amount        int64  `json:"amount" validate:"required,gt=0"`
	EffectiveFrom string `json:"effective_from" validate:"required"`
}
