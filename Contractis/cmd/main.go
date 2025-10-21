package main

import (
	"log"

	httpAdapter "github.com/rodascaar/contractis/internal/adapters/http"
	"github.com/rodascaar/contractis/internal/adapters/http/handlers"
	"github.com/rodascaar/contractis/internal/adapters/http/router"
	"github.com/rodascaar/contractis/internal/infrastructure/database"
	"github.com/rodascaar/contractis/internal/infrastructure/llm"
	"github.com/rodascaar/contractis/internal/infrastructure/pdf"
	"github.com/rodascaar/contractis/internal/infrastructure/text"
	"github.com/rodascaar/contractis/internal/usecases"
)

func main() {
	// Database initialization
	db, err := database.NewSQLiteDB("./contractis.db")
	if err != nil {
		log.Fatalf("❌ Error inicializando base de datos: %v", err)
	}
	defer db.Close()
	log.Printf("✅ Base de datos SQLite inicializada: ./contractis.db")

	// Infrastructure layer
	pdfExtractor := pdf.NewExtractor()
	llmClient := llm.NewClient()
	textProcessor := text.NewProcessor()
	contractRepo := database.NewContractRepository(db)

	// Use cases layer
	analyzeUseCase := usecases.NewAnalyzeContractUseCase(
		pdfExtractor,
		llmClient,
		contractRepo,
		textProcessor,
	)

	estimateUseCase := usecases.NewEstimateTokensUseCase(
		pdfExtractor,
		textProcessor,
	)

	// HTTP handlers (adapters layer)
	uploadHandler := handlers.NewUploadHandler(analyzeUseCase)
	estimateHandler := handlers.NewEstimateHandler(estimateUseCase)
	historyHandler := handlers.NewHistoryHandler(contractRepo)

	// Router setup
	appRouter := router.NewRouter(
		uploadHandler,
		estimateHandler,
		historyHandler,
		"./static",
	)

	// HTTP server
	server := httpAdapter.NewServer("8080", appRouter)

	// Start server
	log.Printf("🚀 Contractis servidor iniciado en http://localhost:8080")
	if err := server.Start(); err != nil {
		log.Fatalf("❌ Error iniciando servidor: %v", err)
	}
}
