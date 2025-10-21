package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/rodascaar/contractis/internal/domain/repositories"
)

// HistoryHandler maneja las solicitudes de historial de contratos
type HistoryHandler struct {
	contractRepo repositories.ContractRepository
}

// NewHistoryHandler crea una nueva instancia de HistoryHandler
func NewHistoryHandler(contractRepo repositories.ContractRepository) *HistoryHandler {
	return &HistoryHandler{
		contractRepo: contractRepo,
	}
}

// HandleList lista todos los contratos con paginación
func (h *HistoryHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parsear parámetros de paginación
	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Obtener contratos
	contracts, err := h.contractRepo.List(r.Context(), limit, offset)
	if err != nil {
		log.Printf("Error listing contracts: %v", err)
		http.Error(w, "Error al obtener historial", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    contracts,
		"limit":   limit,
		"offset":  offset,
	})
}

// HandleSearch busca contratos por nombre
func (h *HistoryHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	contracts, err := h.contractRepo.Search(r.Context(), query, limit, offset)
	if err != nil {
		log.Printf("Error searching contracts: %v", err)
		http.Error(w, "Error al buscar contratos", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    contracts,
		"query":   query,
		"limit":   limit,
		"offset":  offset,
	})
}

// HandleGetByID obtiene un contrato por ID
func (h *HistoryHandler) HandleGetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	contract, err := h.contractRepo.GetByID(r.Context(), id)
	if err != nil {
		log.Printf("Error getting contract: %v", err)
		http.Error(w, "Contrato no encontrado", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    contract,
	})
}

// HandleGetRecent obtiene los contratos más recientes
func (h *HistoryHandler) HandleGetRecent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	contracts, err := h.contractRepo.GetRecent(r.Context(), limit)
	if err != nil {
		log.Printf("Error getting recent contracts: %v", err)
		http.Error(w, "Error al obtener contratos recientes", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    contracts,
	})
}

// HandleGetStats obtiene estadísticas de contratos
func (h *HistoryHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.contractRepo.GetStats(r.Context())
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		http.Error(w, "Error al obtener estadísticas", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// HandleDelete elimina un contrato
func (h *HistoryHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "DELETE" && r.Method != "POST" {
		log.Printf("Delete handler: Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		log.Printf("Delete handler: Missing ID parameter")
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("Delete handler: Invalid ID format: %s", idStr)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	log.Printf("Attempting to delete contract with ID: %d", id)

	if err := h.contractRepo.Delete(r.Context(), id); err != nil {
		log.Printf("Error deleting contract ID %d: %v", id, err)
		http.Error(w, "Error al eliminar contrato", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted contract ID: %d", id)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Contrato eliminado exitosamente",
	})
}
