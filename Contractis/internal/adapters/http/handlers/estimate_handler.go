package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/rodascaar/contractis/internal/adapters/http/dto"
	"github.com/rodascaar/contractis/internal/domain/entities"
	"github.com/rodascaar/contractis/internal/usecases"
)

// EstimateHandler maneja las solicitudes de estimación de tokens
type EstimateHandler struct {
	estimateUseCase *usecases.EstimateTokensUseCase
}

// NewEstimateHandler crea una nueva instancia de EstimateHandler
func NewEstimateHandler(estimateUseCase *usecases.EstimateTokensUseCase) *EstimateHandler {
	return &EstimateHandler{
		estimateUseCase: estimateUseCase,
	}
}

// Handle procesa la solicitud de estimación
func (h *EstimateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		h.sendError(w, "Method not allowed")
		return
	}

	// Obtener archivo
	file, header, err := r.FormFile("file")
	if err != nil {
		h.sendError(w, "Error al obtener el archivo")
		return
	}
	defer file.Close()

	// Validar tamaño
	if header.Size > entities.MaxFileSize {
		h.sendError(w, "Archivo demasiado grande (máximo 10MB)")
		return
	}

	// Validar tipo
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		h.sendError(w, "Solo se permiten archivos PDF")
		return
	}

	// Obtener maxTokens
	maxTokensStr := r.FormValue("maxTokens")
	maxTokens := 800 // valor por defecto
	if maxTokensStr != "" {
		if parsed, err := json.Number(maxTokensStr).Int64(); err == nil {
			maxTokens = int(parsed)
		}
	}

	// Guardar archivo temporal
	tempFile, err := os.CreateTemp("", "estimate-*.pdf")
	if err != nil {
		h.sendError(w, "Error al crear archivo temporal")
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		h.sendError(w, "Error al guardar el archivo")
		return
	}

	// Ejecutar estimación
	estimation, err := h.estimateUseCase.Execute(tempFile.Name(), maxTokens)
	if err != nil {
		log.Printf("Error en estimación: %v", err)
		h.sendError(w, "Error al extraer texto del PDF")
		return
	}

	log.Printf("Estimación: %d tokens totales, %d chunks, maxTokens recomendado: %d",
		estimation.TotalTokens, estimation.Chunks, estimation.RecommendedMaxTokens)

	// Convertir a DTO de respuesta
	response := dto.TokenEstimationResponse{
		Success:              estimation.Success,
		CharacterCount:       estimation.CharacterCount,
		EstimatedTokens:      estimation.EstimatedTokens,
		Chunks:               estimation.Chunks,
		SystemPromptTokens:   estimation.SystemPromptTokens,
		Phase1Tokens:         estimation.Phase1Tokens,
		Phase2InputTokens:    estimation.Phase2InputTokens,
		Phase2OutputTokens:   estimation.Phase2OutputTokens,
		TotalTokens:          estimation.TotalTokens,
		RecommendedMaxTokens: estimation.RecommendedMaxTokens,
		Warning:              estimation.Warning,
		Error:                estimation.Error,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *EstimateHandler) sendError(w http.ResponseWriter, message string) {
	json.NewEncoder(w).Encode(dto.TokenEstimationResponse{
		Success: false,
		Error:   message,
	})
}
