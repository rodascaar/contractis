package entities

import "errors"

// Domain errors
var (
	// Contract errors
	ErrInvalidFileName = errors.New("invalid file name")
	ErrEmptyContent    = errors.New("empty content")
	ErrInvalidSize     = errors.New("invalid file size")
	ErrFileTooLarge    = errors.New("file too large (maximum 10MB)")
	ErrInvalidFileType = errors.New("only PDF files are allowed")

	// LLM errors
	ErrLLMConnectionFailed = errors.New("failed to connect to LLM")
	ErrLLMTimeout          = errors.New("LLM request timeout")
	ErrInvalidLLMResponse  = errors.New("invalid LLM response")

	// Processing errors
	ErrProcessingFailed = errors.New("processing failed")
	ErrExtractionFailed = errors.New("text extraction failed")
)
