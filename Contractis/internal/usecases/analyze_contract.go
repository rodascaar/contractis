package usecases

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/rodascaar/contractis/internal/domain/entities"
	"github.com/rodascaar/contractis/internal/domain/repositories"
	"github.com/rodascaar/contractis/internal/domain/services"
)

// AnalyzeContractUseCase maneja el caso de uso de análisis de contratos
type AnalyzeContractUseCase struct {
	pdfRepo       repositories.PDFRepository
	llmRepo       repositories.LLMRepository
	contractRepo  repositories.ContractRepository
	textProcessor services.TextProcessor
}

// NewAnalyzeContractUseCase crea una nueva instancia del caso de uso
func NewAnalyzeContractUseCase(
	pdfRepo repositories.PDFRepository,
	llmRepo repositories.LLMRepository,
	contractRepo repositories.ContractRepository,
	textProcessor services.TextProcessor,
) *AnalyzeContractUseCase {
	return &AnalyzeContractUseCase{
		pdfRepo:       pdfRepo,
		llmRepo:       llmRepo,
		contractRepo:  contractRepo,
		textProcessor: textProcessor,
	}
}

// Execute ejecuta el análisis del contrato
func (uc *AnalyzeContractUseCase) Execute(
	ctx context.Context,
	pdfPath string,
	filename string,
	fileHash string,
	fileSize int64,
	config *entities.LLMConfig,
) (*entities.AnalysisResult, error) {
	startTime := time.Now()

	// Validar configuración
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid LLM config: %w", err)
	}

	// Verificar si ya existe registro para este archivo
	existingRecord, err := uc.contractRepo.GetByHash(ctx, fileHash)
	if err != nil {
		log.Printf("⚠️  Error verificando registro existente: %v", err)
		// Continuar con el análisis
	}

	var record *entities.ContractRecord
	var recordID int64

	if existingRecord != nil {
		// Usar registro existente
		record = existingRecord
		recordID = existingRecord.ID
		log.Printf("📝 Usando registro existente en BD (ID: %d, Status: %s)", recordID, existingRecord.Status)
	} else {
		// Crear nuevo registro en BD
		record = entities.NewContractRecord(
			filename,
			fileHash,
			fileSize,
			config.Type,
			config.ModelName,
			config.MaxTokens,
		)

		recordID, err = uc.contractRepo.Create(ctx, record)
		if err != nil {
			log.Printf("⚠️  Error creando registro en BD: %v", err)
			// Continuar con el análisis aunque falle la BD
		} else {
			record.ID = recordID
			log.Printf("💾 Registro creado en BD (ID: %d)", recordID)
		}
	}

	// Probar conexión con LLM
	if err := uc.llmRepo.TestConnection(ctx, config); err != nil {
		if record.ID > 0 {
			record.MarkFailed(err.Error())
			uc.contractRepo.Update(ctx, record)
		}
		return nil, fmt.Errorf("LLM connection test failed: %w", err)
	}

	// Marcar como analizando
	if record.ID > 0 {
		record.MarkAnalyzing()
		uc.contractRepo.Update(ctx, record)
	}

	// Extraer texto del PDF
	content, err := uc.pdfRepo.ExtractText(pdfPath)
	if err != nil {
		if record.ID > 0 {
			record.MarkFailed(err.Error())
			uc.contractRepo.Update(ctx, record)
		}
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}

	log.Printf("Iniciando análisis de documento (%d caracteres)", len(content))

	// Generar respuesta con RAG
	result, err := uc.generateResponseWithRAG(ctx, content, config)
	if err != nil {
		if record.ID > 0 {
			record.MarkFailed(err.Error())
			uc.contractRepo.Update(ctx, record)
		}
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	duration := time.Since(startTime)
	log.Printf("Análisis completado en %.2f segundos", duration.Seconds())

	// Calcular chunks para metadata
	maxChunkSize := uc.calculateMaxChunkSize("")
	chunks := uc.textProcessor.SplitText(content, maxChunkSize)

	// Validar resultado antes de marcar como exitoso
	if result == "" {
		log.Printf("⚠️ Advertencia: El resultado del análisis está vacío, usando mensaje por defecto")
		result = "No se pudo generar un análisis válido del contrato. Es posible que el documento esté vacío, corrupto o el modelo de lenguaje no haya podido procesarlo correctamente."
	}

	// Guardar resultado en BD
	if record.ID > 0 {
		record.MarkCompleted(result, len(content), len(content)/entities.CharsPerToken, len(chunks), duration.Seconds())
		if err := uc.contractRepo.Update(ctx, record); err != nil {
			log.Printf("⚠️  Error actualizando registro en BD: %v", err)
		}
	}

	analysisResult := entities.NewAnalysisResult("")
	analysisResult.MarkSuccess(result, duration, len(chunks))

	return analysisResult, nil
}

func (uc *AnalyzeContractUseCase) generateResponseWithRAG(
	ctx context.Context,
	documentContent string,
	llmConfig *entities.LLMConfig,
) (string, error) {
	systemPrompt := `Analiza contratos legales en español. Identifica: terminación unilateral, penalizaciones, jurisdicción, riesgos. Respuesta completa en español, sin emojis ni formato markdown.`

	// Determinar si procesar en una sola petición o por chunks
	totalChars := len(documentContent)
	totalTokens := totalChars / entities.CharsPerToken

	// Para modelos online con mucho contexto, procesar en una sola petición si cabe
	if llmConfig.IsOnline() && totalTokens < (entities.OnlineContextWindow-entities.SafetyMargin-entities.MaxOutputTokens) {
		log.Printf("📄 Documento pequeño (%d tokens), procesando en una sola petición", totalTokens)
		return uc.processSingleRequest(ctx, documentContent, systemPrompt, llmConfig)
	}

	// Procesamiento por chunks para documentos grandes o modelos locales
	log.Printf("📄 Documento grande (%d tokens), procesando por chunks", totalTokens)

	// Calcular tamaño de chunk
	maxChunkSize := uc.calculateMaxChunkSize(systemPrompt)
	log.Printf("Tamaño de chunk calculado: %d caracteres (~%d tokens)",
		maxChunkSize, maxChunkSize/entities.CharsPerToken)

	chunks := uc.textProcessor.SplitText(documentContent, maxChunkSize)
	var analysisFragments []string

	// FASE 1: Análisis por fragmento
	userQuery := `Analiza este fragmento del contrato. Identifica terminación unilateral, penalizaciones (con montos), jurisdicción/arbitraje y riesgos principales. Respuesta en español, sin emojis.`

	for i, chunk := range chunks {
		fmt.Printf("📄 Procesando parte %d/%d (%d caracteres)...\n", i+1, len(chunks), len(chunk))

		prompt := fmt.Sprintf(
			"Parte %d/%d del contrato:\n%s\n\nInstrucción: %s",
			i+1, len(chunks), chunk, userQuery,
		)

		messages := []repositories.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		}

		responseText, err := uc.llmRepo.SendChatRequest(ctx, llmConfig, messages, entities.Phase1MaxTokens)
		if err != nil {
			return "", fmt.Errorf("error processing part %d/%d: %w", i+1, len(chunks), err)
		}

		log.Printf("📄 Respuesta parte %d/%d: %d caracteres", i+1, len(chunks), len(responseText))

		responseText = uc.textProcessor.CleanFragment(responseText)
		analysisFragments = append(analysisFragments,
			fmt.Sprintf("PARTE %d/%d:\n%s", i+1, len(chunks), responseText))
	}

	// FASE 2: Consolidación final
	return uc.consolidateFragments(ctx, analysisFragments, systemPrompt, llmConfig)
}

func (uc *AnalyzeContractUseCase) calculateMaxChunkSize(systemPrompt string) int {
	systemPromptTokens := len(systemPrompt) / entities.CharsPerToken
	userInstructionTokens := 50
	availableTokensForChunk := entities.MaxInputTokens - systemPromptTokens - userInstructionTokens
	maxChunkSize := availableTokensForChunk * entities.CharsPerToken

	if maxChunkSize > entities.DefaultChunkSize {
		maxChunkSize = entities.DefaultChunkSize
	}
	if maxChunkSize < entities.MinChunkSize {
		maxChunkSize = entities.MinChunkSize
	}

	return maxChunkSize
}

func (uc *AnalyzeContractUseCase) processSingleRequest(
	ctx context.Context,
	documentContent string,
	systemPrompt string,
	llmConfig *entities.LLMConfig,
) (string, error) {
	userQuery := `Analiza este contrato completo. Identifica terminación unilateral, penalizaciones (con montos), jurisdicción/arbitraje y riesgos principales. Respuesta completa en español, sin emojis, sin formato markdown.`

	messages := []repositories.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userQuery + "\n\n" + documentContent},
	}

	// Calcular tokens disponibles para respuesta
	totalTokens := len(documentContent)/entities.CharsPerToken + len(systemPrompt)/entities.CharsPerToken + len(userQuery)/entities.CharsPerToken
	availableTokens := entities.OnlineContextWindow - totalTokens - entities.SafetyMargin

	if availableTokens > entities.MaxOutputTokens {
		availableTokens = entities.MaxOutputTokens
	}
	if availableTokens < 800 {
		availableTokens = 800
	}

	log.Printf("Procesando documento completo en una petición - Tokens disponibles: %d", availableTokens)

	responseText, err := uc.llmRepo.SendChatRequest(ctx, llmConfig, messages, availableTokens)
	if err != nil {
		return "", fmt.Errorf("error processing single request: %w", err)
	}

	log.Printf("📄 Respuesta documento completo: %d caracteres", len(responseText))
	return responseText, nil
}

func (uc *AnalyzeContractUseCase) consolidateFragments(
	ctx context.Context,
	analysisFragments []string,
	systemPrompt string,
	llmConfig *entities.LLMConfig,
) (string, error) {
	consolidationPrompt := `Consolida estos fragmentos en un reporte final completo en español sobre: terminación unilateral, penalizaciones, jurisdicción y riesgos. Incluye todos los detalles importantes sin omitir información. Respuesta en español, sin emojis, sin formato markdown.`

	// Limitar tamaño de fragmentos
	maxCharsPerFragment := 2500
	for i, fragment := range analysisFragments {
		fragment = uc.textProcessor.CleanFragment(fragment)
		if len(fragment) > maxCharsPerFragment {
			fragment = fragment[:maxCharsPerFragment] + "\n[Continuación omitida]"
		}
		analysisFragments[i] = fragment
	}

	// Consolidación jerárquica si hay muchos fragmentos
	if len(analysisFragments) > entities.MaxFragments {
		analysisFragments = uc.hierarchicalConsolidation(analysisFragments, maxCharsPerFragment)
	}

	// Construir prompt final
	finalPrompt := uc.buildFinalPrompt(consolidationPrompt, analysisFragments)

	// Calcular tokens disponibles
	estimatedInputTokens := len(finalPrompt) / entities.CharsPerToken

	// Usar ventana de contexto apropiada
	contextWindow := entities.QwenContextWindow
	if llmConfig.IsOnline() {
		contextWindow = entities.OnlineContextWindow
	}

	availableTokens := contextWindow - estimatedInputTokens - entities.SafetyMargin

	minConsolidationTokens := 1800
	if availableTokens < minConsolidationTokens {
		availableTokens = minConsolidationTokens
	}
	if availableTokens > entities.MaxOutputTokens {
		availableTokens = entities.MaxOutputTokens
	}

	log.Printf("Tokens para consolidación - Input: ~%d, Output: %d", estimatedInputTokens, availableTokens)

	messages := []repositories.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: finalPrompt},
	}

	finalResult, err := uc.llmRepo.SendChatRequest(ctx, llmConfig, messages, availableTokens)
	if err != nil {
		return "", fmt.Errorf("error in consolidation: %w", err)
	}

	log.Printf("📄 Resultado consolidación: %d caracteres", len(finalResult))
	return finalResult, nil
}

func (uc *AnalyzeContractUseCase) hierarchicalConsolidation(fragments []string, maxCharsPerFragment int) []string {
	log.Printf("Aplicando consolidación jerárquica: %d fragmentos -> grupos", len(fragments))

	groupSize := 3
	var consolidatedGroups []string

	for i := 0; i < len(fragments); i += groupSize {
		end := i + groupSize
		if end > len(fragments) {
			end = len(fragments)
		}

		group := fragments[i:end]
		groupText := strings.Join(group, "\n\n")
		groupText = uc.textProcessor.CleanFragment(groupText)

		if len(groupText) > maxCharsPerFragment*2 {
			groupText = groupText[:maxCharsPerFragment*2] + "\n[Continuación omitida]"
		}

		consolidatedGroups = append(consolidatedGroups,
			fmt.Sprintf("GRUPO %d-%d:\n%s", i+1, end, groupText))
	}

	log.Printf("Fragmentos después de consolidación jerárquica: %d", len(consolidatedGroups))
	return consolidatedGroups
}

func (uc *AnalyzeContractUseCase) buildFinalPrompt(consolidationPrompt string, fragments []string) string {
	var combinedFragments strings.Builder
	totalChars := len(consolidationPrompt)
	maxTotalChars := entities.MaxInputTokens * entities.CharsPerToken

	for i, fragment := range fragments {
		if totalChars+len(fragment)+4 > maxTotalChars {
			log.Printf("⚠️ Omitiendo fragmentos %d-%d por límite de tokens", i+1, len(fragments))
			combinedFragments.WriteString("\n\n[Fragmentos adicionales omitidos por límite de tokens]")
			break
		}

		combinedFragments.WriteString(fragment)
		combinedFragments.WriteString("\n\n")
		totalChars += len(fragment) + 2
	}

	finalPrompt := fmt.Sprintf("%s\n\n%s", consolidationPrompt, combinedFragments.String())

	// Verificación final de tamaño
	estimatedPromptTokens := len(finalPrompt) / entities.CharsPerToken
	if estimatedPromptTokens > entities.MaxInputTokens {
		log.Printf("⚠️ Advertencia: Prompt muy grande (%d tokens), truncando...", estimatedPromptTokens)
		maxPromptChars := entities.MaxInputTokens * entities.CharsPerToken
		if len(finalPrompt) > maxPromptChars {
			finalPrompt = finalPrompt[:maxPromptChars] + "\n\n[Contenido truncado por límite de tokens]"
		}
	} else {
		log.Printf("✅ Prompt de consolidación: %d tokens (dentro del límite)", estimatedPromptTokens)
	}

	return finalPrompt
}
