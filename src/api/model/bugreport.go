package model

type CreateBugReportRequestDTO struct {
	Subject string  `json:"subject" validate:"required,max=160"`
	Message *string `json:"message" validate:"omitempty,max=160"`
}
