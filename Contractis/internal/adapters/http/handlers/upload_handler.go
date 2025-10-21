package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rodascaar/contractis/internal/adapters/http/dto"
	"github.com/rodascaar/contractis/internal/domain/entities"
	"github.com/rodascaar/contractis/internal/infrastructure/utils"
	"github.com/rodascaar/contractis/internal/usecases"
)

// UploadHandler maneja las solicitudes de carga y análisis de contratos
type UploadHandler struct {
	analyzeUseCase  *usecases.AnalyzeContractUseCase
	processingMutex sync.Mutex
}

// NewUploadHandler crea una nueva instancia de UploadHandler
func NewUploadHandler(analyzeUseCase *usecases.AnalyzeContractUseCase) *UploadHandler {
	return &UploadHandler{
		analyzeUseCase: analyzeUseCase,
	}
}

// Handle procesa la solicitud de carga
func (h *UploadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Configurar CORS con restricciones de seguridad
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight for 24 hours
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

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

	// Parsear configuración LLM
	llmConfigStr := r.FormValue("llmConfig")
	var llmConfigReq dto.LLMConfigRequest
	if err := json.Unmarshal([]byte(llmConfigStr), &llmConfigReq); err != nil {
		h.sendError(w, "Error al parsear configuración LLM")
		return
	}

	// Evitar procesamiento concurrente para modelos locales (pueden ser lentos)
	if llmConfigReq.Type == "local" {
		h.processingMutex.Lock()
		defer h.processingMutex.Unlock()
		log.Printf("🔒 Procesamiento secuencial activado para modelo local")
	}

	// Convertir a entidad de dominio
	llmConfig := entities.NewLLMConfig(
		llmConfigReq.Type,
		llmConfigReq.LocalUrl,
		llmConfigReq.ApiUrl,
		llmConfigReq.ApiKey,
		llmConfigReq.ModelName,
		llmConfigReq.MaxTokens,
	)

	if err := llmConfig.Validate(); err != nil {
		log.Printf("❌ Configuración LLM inválida: %v", err)
		h.sendError(w, fmt.Sprintf("Configuración LLM inválida: %v", err))
		return
	}

	log.Printf("🔧 Configuración LLM recibida - Tipo: %s, URL: %s", llmConfig.Type, llmConfig.GetEndpointURL())

	// Guardar archivo temporal
	tempFile, err := os.CreateTemp("", "contrato-*.pdf")
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

	// Calcular hash del archivo para caché
	fileHash, err := utils.CalculateFileHash(tempFile.Name())
	if err != nil {
		log.Printf("⚠️  Error calculando hash: %v", err)
		fileHash = "" // Continuar sin hash
	}

	// Crear contexto con timeout más largo para modelos locales
	timeout := 30 * time.Minute
	if llmConfig.Type == "local" {
		timeout = 45 * time.Minute // Más tiempo para modelos locales
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Ejecutar análisis
	result, err := h.analyzeUseCase.Execute(ctx, tempFile.Name(), header.Filename, fileHash, header.Size, llmConfig)
	if err != nil {
		log.Printf("Error en análisis: %v", err)
		h.sendError(w, fmt.Sprintf("Error al analizar: %v", err))
		return
	}

	// Enviar respuesta exitosa
	response := dto.AnalysisResponse{
		Success: result.Success,
		Data:    result.Content,
	}

	log.Printf("✅ Análisis completado exitosamente, enviando respuesta (longitud: %d caracteres)", len(result.Content))
	json.NewEncoder(w).Encode(response)
}

func (h *UploadHandler) sendError(w http.ResponseWriter, message string) {
	log.Printf("❌ Error en upload handler: %s", message)
	json.NewEncoder(w).Encode(dto.AnalysisResponse{
		Success: false,
		Error:   message,
	})
}
