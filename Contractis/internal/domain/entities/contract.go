package entities

import "time"

// Contract representa un contrato en el sistema
type Contract struct {
	ID          string
	FileName    string
	Content     string
	Size        int64
	UploadedAt  time.Time
	ProcessedAt time.Time
}

// NewContract crea una nueva instancia de Contract
func NewContract(fileName string, content string, size int64) *Contract {
	return &Contract{
		FileName:   fileName,
		Content:    content,
		Size:       size,
		UploadedAt: time.Now(),
	}
}

// Validate valida los datos del contrato
func (c *Contract) Validate() error {
	if c.FileName == "" {
		return ErrInvalidFileName
	}
	if c.Content == "" {
		return ErrEmptyContent
	}
	if c.Size <= 0 {
		return ErrInvalidSize
	}
	if c.Size > MaxFileSize {
		return ErrFileTooLarge
	}
	return nil
}
