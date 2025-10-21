package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/rodascaar/contractis/internal/domain/entities"
	"github.com/rodascaar/contractis/internal/domain/repositories"
)

// Client implementa el cliente para interactuar con LLMs
type Client struct {
	// No usamos un httpClient fijo, lo creamos dinámicamente según el tipo de LLM
}

// NewClient crea una nueva instancia de Client
func NewClient() *Client {
	return &Client{}
}

// getHTTPClientWithDynamicTimeout retorna un cliente HTTP con timeout dinámico basado en tokens esperados
func (c *Client) getHTTPClientWithDynamicTimeout(config *entities.LLMConfig, expectedTokens int) *http.Client {
	// Timeouts base
	baseTimeout := entities.HTTPTimeoutLocal
	tokensPerSecond := 10.0 // Modelos locales suelen ser más lentos
	marginMultiplier := 3.0 // Margen de seguridad conservador

	if config.IsOnline() {
		baseTimeout = entities.HTTPTimeoutOnline
		tokensPerSecond = 15.0 // Modelos online: más conservador (antes 30.0)
		marginMultiplier = 4.0 // Margen aún más conservador para online
	}

	// Calcular tiempo estimado basado en tokens esperados
	estimatedSeconds := float64(expectedTokens) / tokensPerSecond
	estimatedTime := time.Duration(estimatedSeconds) * time.Second

	// Agregar margen de seguridad (3-4x) + overhead de red (30s)
	dynamicTimeout := (estimatedTime * time.Duration(marginMultiplier)) + (30 * time.Second)

	// Usar el mayor entre el timeout base y el dinámico
	if dynamicTimeout > baseTimeout {
		log.Printf("⏱️  Timeout dinámico: %.0fs (esperando ~%d tokens @ %.0f tps, margen %.1fx)",
			dynamicTimeout.Seconds(), expectedTokens, tokensPerSecond, marginMultiplier)
		return &http.Client{Timeout: dynamicTimeout}
	}

	log.Printf("⏱️  Usando timeout base: %.0fs", baseTimeout.Seconds())
	return &http.Client{Timeout: baseTimeout}
}

// ChatRequest representa una solicitud de chat para modelos locales
type ChatRequest struct {
	Model     string        `json:"model"`
	Messages  []ChatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
	Stream    bool          `json:"stream"`
}

// OnlineChatRequest representa una solicitud de chat para modelos online
type OnlineChatRequest struct {
	Model               string        `json:"model"`
	Messages            []ChatMessage `json:"messages"`
	MaxCompletionTokens int           `json:"max_completion_tokens,omitempty"`
	Stream              bool          `json:"stream"`
}

// ChatMessage representa un mensaje en el chat
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse representa una respuesta del chat para modelos online (OpenAI-style)
type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

// LocalChatResponse representa una respuesta del chat para modelos locales (Ollama-style)
type LocalChatResponse struct {
	Message ChatMessage `json:"message"`
	Done    bool        `json:"done,omitempty"` // Para respuestas de streaming
}

// StreamingResponse representa una respuesta de streaming
type StreamingResponse struct {
	Message ChatMessage `json:"message"`
	Done    bool        `json:"done"`
}

// SendChatRequest envía una solicitud de chat al LLM
func (c *Client) SendChatRequest(
	ctx context.Context,
	config *entities.LLMConfig,
	messages []repositories.ChatMessage,
	maxTokens int,
) (string, error) {
	return c.SendChatRequestWithStreaming(ctx, config, messages, maxTokens, false)
}

// SendChatRequestWithStreaming envía una solicitud de chat al LLM con opción de streaming
func (c *Client) SendChatRequestWithStreaming(
	ctx context.Context,
	config *entities.LLMConfig,
	messages []repositories.ChatMessage,
	maxTokens int,
	stream bool,
) (string, error) {
	// Convertir mensajes al formato interno
	chatMessages := make([]ChatMessage, len(messages))
	for i, msg := range messages {
		chatMessages[i] = ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	modelName := config.ModelName
	if modelName == "" {
		log.Printf("⚠️  Advertencia: No se especificó un modelo LLM. Se requiere configuración explícita desde el panel de configuración.")
		return "", fmt.Errorf("modelo LLM no configurado: por favor configura un modelo específico desde el panel de configuración")
	}

	var reqBody interface{}
	var paramName string

	if config.IsOnline() {
		// Para modelos online, usar max_completion_tokens
		reqBody = OnlineChatRequest{
			Model:               modelName,
			Messages:            chatMessages,
			MaxCompletionTokens: maxTokens,
			Stream:              stream,
		}
		paramName = "max_completion_tokens"
	} else {
		// Para modelos locales, usar max_tokens
		reqBody = ChatRequest{
			Model:     modelName,
			Messages:  chatMessages,
			MaxTokens: maxTokens,
			Stream:    stream,
		}
		paramName = "max_tokens"
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error al serializar request: %w", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	if authHeader := config.GetAuthorizationHeader(); authHeader != "" {
		headers["Authorization"] = authHeader
	}

	// Usar timeout dinámico basado en maxTokens esperados
	httpClient := c.getHTTPClientWithDynamicTimeout(config, maxTokens)
	streamMode := ""
	if stream {
		streamMode = " (streaming)"
	}
	log.Printf("🔗 Enviando petición a %s (modelo: %s, %s: %d)%s", config.GetEndpointURL(), modelName, paramName, maxTokens, streamMode)

	if stream {
		return c.handleStreamingResponse(ctx, httpClient, config.GetEndpointURL(), jsonData, headers)
	}

	resp, err := c.makeRequestWithRetryCustomClient(ctx, httpClient, config.GetEndpointURL(), jsonData, headers, entities.MaxRetries)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error al leer respuesta: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error del servidor (%d): %s", resp.StatusCode, string(body))
	}

	var content string
	var parseErr error

	// Intentar parsear respuesta local (Ollama-style) primero para modelos locales
	var parsed LocalChatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		parseErr = fmt.Errorf("error al parsear respuesta local: %v", err)
	} else if parsed.Message.Content != "" {
		content = parsed.Message.Content
		log.Printf("✅ Respuesta local parseada correctamente")
	} else {
		parseErr = fmt.Errorf("respuesta local no contiene contenido válido")
	}

	// Si el parsing local falló, intentar formato online (OpenAI-style)
	if parseErr != nil {
		log.Printf("⚠️  %v - intentando formato online", parseErr)

		var parsed ChatResponse
		if err := json.Unmarshal(body, &parsed); err == nil && len(parsed.Choices) > 0 {
			content = parsed.Choices[0].Message.Content
			log.Printf("✅ Fallback exitoso: respuesta parseada como online")
		} else {
			return "", fmt.Errorf("error al parsear respuesta local y fallback falló: %v\nCuerpo: %s", parseErr, string(body))
		}
	}

	// Validación básica de la respuesta antes de procesar
	if strings.TrimSpace(content) == "" {
		log.Printf("⚠️  Respuesta del LLM está vacía")
		return "", fmt.Errorf("el LLM devolvió una respuesta vacía")
	}

	// Log raw content for debugging
	log.Printf("🔍 Raw LLM response: %d caracteres", len(content))
	log.Printf("🔍 Raw content preview: %s", content[:min(200, len(content))])

	// Procesamiento inteligente de contenido según el tipo de modelo
	processedContent := c.processLLMResponse(content)

	// Validación final del contenido procesado
	if strings.TrimSpace(processedContent) == "" {
		log.Printf("⚠️  Content is empty after processing, returning raw content")
		return content, nil // Devolver contenido original si el procesamiento lo dejó vacío
	}

	log.Printf("✅ Respuesta procesada exitosamente: %d caracteres", len(processedContent))
	return processedContent, nil
}

// TestConnection verifica la conexión con el LLM
func (c *Client) TestConnection(ctx context.Context, config *entities.LLMConfig) error {
	log.Printf("🔗 Probando conexión con LLM (%s) - URL: %s", config.Type, config.GetEndpointURL())

	testClient := &http.Client{Timeout: entities.TestConnectionTimeout}

	var reqBody interface{}

	if config.IsOnline() {
		reqBody = OnlineChatRequest{
			Model:               config.ModelName,
			Messages:            []ChatMessage{{Role: "user", Content: "test"}},
			MaxCompletionTokens: 10,
			Stream:              false,
		}
	} else {
		reqBody = ChatRequest{
			Model: config.ModelName,
			Messages: []ChatMessage{
				{Role: "user", Content: "test"},
			},
			MaxTokens: 10,
			Stream:    false,
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error al serializar test request: %w", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	if authHeader := config.GetAuthorizationHeader(); authHeader != "" {
		headers["Authorization"] = authHeader
	}

	resp, err := c.makeRequestWithRetryCustomClient(ctx, testClient, config.GetEndpointURL(), jsonData, headers, 1)
	if err != nil {
		return fmt.Errorf("no se pudo conectar al LLM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("el servidor LLM no está disponible (status: %d)", resp.StatusCode)
	}

	log.Printf("✓ Conexión con LLM exitosa (tipo: %s, URL: %s)", config.Type, config.GetEndpointURL())
	return nil
}

func (c *Client) makeRequestWithRetryCustomClient(
	ctx context.Context,
	client *http.Client,
	url string,
	jsonData []byte,
	headers map[string]string,
	maxRetries int,
) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Reintentando petición (intento %d/%d)...", attempt+1, maxRetries)
			time.Sleep(entities.RetryDelay * time.Duration(attempt))
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("error al crear request: %w", err)
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := client.Do(req)
		if err == nil {
			if resp.StatusCode < 500 {
				return resp, nil
			}
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
			continue
		}

		lastErr = err

		// Check for timeout errors
		if ctx.Err() != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("timeout: el servidor no respondió a tiempo. Verifica que la URL sea correcta y el servidor esté disponible")
			}
			return nil, ctx.Err()
		}
	}

	// Provide more specific error message
	if lastErr != nil {
		errMsg := lastErr.Error()
		if ctx.Err() == context.DeadlineExceeded ||
			strings.Contains(errMsg, "timeout") ||
			strings.Contains(errMsg, "deadline exceeded") {
			return nil, fmt.Errorf("timeout después de %d intentos: el servidor no respondió. Verifica la configuración del LLM", maxRetries)
		}
	}

	return nil, fmt.Errorf("falló después de %d intentos: %w", maxRetries, lastErr)
}

// handleStreamingResponse maneja respuestas de streaming del LLM
func (c *Client) handleStreamingResponse(
	ctx context.Context,
	client *http.Client,
	url string,
	jsonData []byte,
	headers map[string]string,
) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error al crear streaming request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error en streaming request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error del servidor streaming (%d): %s", resp.StatusCode, string(body))
	}

	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var streamResp StreamingResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				log.Printf("⚠️  Error parseando línea de streaming: %v", err)
				continue
			}

			if streamResp.Message.Content != "" {
				fullContent.WriteString(streamResp.Message.Content)
			}

			if streamResp.Done {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error leyendo streaming response: %w", err)
	}

	content := fullContent.String()
	if content == "" {
		return "", fmt.Errorf("no se recibió contenido en streaming")
	}

	log.Printf("✅ Streaming completado: %d caracteres", len(content))
	return content, nil
}

// processLLMResponse procesa la respuesta del LLM según el tipo de modelo
func (c *Client) processLLMResponse(content string) string {
	originalContent := content

	// 1. Modelos razonadores (como DeepSeek-R1) - extraer respuesta final después del thinking
	if strings.Contains(content, "</think>") {
		log.Printf("🔍 Detected reasoning model (thinking tags), extracting final answer")
		parts := strings.Split(content, "</think>")
		if len(parts) > 1 {
			content = strings.TrimSpace(parts[1])
			log.Printf("🔍 Extracted content after </think>: %d caracteres", len(content))
		} else {
			log.Printf("⚠️  </think> tag found but no content after it")
			// Si no hay contenido después del thinking, intentar extraer del thinking mismo
			if strings.Contains(content, "<think>") {
				thinkStart := strings.Index(content, "<think>")
				thinkEnd := strings.Index(content, "</think>")
				if thinkStart >= 0 && thinkEnd > thinkStart {
					thinkingContent := content[thinkStart+7 : thinkEnd]
					log.Printf("🔍 Thinking content length: %d caracteres", len(thinkingContent))
					// Si el thinking es muy largo, probablemente es la respuesta
					if len(thinkingContent) > 100 {
						content = strings.TrimSpace(thinkingContent)
						log.Printf("🔍 Using thinking content as response: %d caracteres", len(content))
					}
				}
			}
		}
	}

	// 2. Modelos que podrían devolver JSON - extraer texto plano si es necesario
	if strings.HasPrefix(strings.TrimSpace(content), "{") && strings.HasSuffix(strings.TrimSpace(content), "}") {
		log.Printf("🔍 Detected potential JSON response, checking if it contains text content")
		// Intentar parsear como JSON y extraer campos de texto comunes
		var jsonResponse map[string]interface{}
		if err := json.Unmarshal([]byte(content), &jsonResponse); err == nil {
			// Buscar campos comunes que podrían contener la respuesta
			if text, ok := jsonResponse["text"].(string); ok && text != "" {
				log.Printf("🔍 Extracted text from JSON response")
				content = text
			} else if response, ok := jsonResponse["response"].(string); ok && response != "" {
				log.Printf("🔍 Extracted response from JSON response")
				content = response
			} else if contentField, ok := jsonResponse["content"].(string); ok && contentField != "" {
				log.Printf("🔍 Extracted content from JSON response")
				content = contentField
			}
		}
	}

	// 3. Limpieza general de formato markdown no deseado
	content = c.cleanMarkdownFormatting(content)

	// 4. Si el contenido quedó vacío pero el original tenía algo, devolver el original
	if strings.TrimSpace(content) == "" && strings.TrimSpace(originalContent) != "" {
		log.Printf("⚠️  Content became empty after processing, returning original content")
		return originalContent
	}

	return content
}

// cleanMarkdownFormatting limpia formato markdown no deseado
func (c *Client) cleanMarkdownFormatting(content string) string {
	// Remover encabezados markdown que no aportan valor
	content = strings.ReplaceAll(content, "# ", "")
	content = strings.ReplaceAll(content, "## ", "")
	content = strings.ReplaceAll(content, "### ", "")

	// Remover negritas y cursivas excesivas si dominan el texto
	asterisks := strings.Count(content, "*")
	if asterisks > len(content)/10 { // Si más del 10% son asteriscos
		log.Printf("🔍 Removing excessive markdown formatting")
		content = strings.ReplaceAll(content, "**", "")
		content = strings.ReplaceAll(content, "*", "")
	}

	return strings.TrimSpace(content)
}

// Helper function for min operation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
