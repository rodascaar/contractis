# Contractis

Sistema de análisis de contratos legales con IA utilizando RAG (Retrieval-Augmented Generation).

## 🏗️ Arquitectura

El proyecto sigue los principios de **Clean Architecture**, organizando el código en capas bien definidas:

```
Contractis/
├── cmd/                          # Punto de entrada de la aplicación
│   └── main.go                   # Dependency Injection y configuración inicial
├── internal/
│   ├── domain/                   # Capa de dominio (reglas de negocio)
│   │   ├── entities/             # Entidades del dominio
│   │   │   ├── contract.go       # Entidad de contrato
│   │   │   ├── llm_config.go     # Configuración del LLM
│   │   │   ├── analysis.go       # Resultados de análisis
│   │   │   ├── constants.go      # Constantes del dominio
│   │   │   └── errors.go         # Errores del dominio
│   │   ├── repositories/         # Interfaces de repositorios
│   │   │   ├── pdf_repository.go
│   │   │   └── llm_repository.go
│   │   └── services/             # Interfaces de servicios
│   │       ├── text_processor.go
│   │       └── analysis_service.go
│   ├── usecases/                 # Casos de uso (lógica de aplicación)
│   │   ├── analyze_contract.go   # Análisis de contratos
│   │   ├── estimate_tokens.go    # Estimación de tokens
│   │   └── test_llm_connection.go
│   ├── infrastructure/           # Implementaciones técnicas
│   │   ├── pdf/                  # Extractor de PDF
│   │   ├── llm/                  # Cliente LLM
│   │   └── text/                 # Procesador de texto
│   └── adapters/                 # Adaptadores (HTTP, etc.)
│       └── http/
│           ├── handlers/         # Manejadores HTTP
│           ├── middleware/       # Middlewares
│           ├── router/           # Configuración de rutas
│           ├── dto/              # DTOs para HTTP
│           └── http_server.go    # Servidor HTTP
├── go.mod
└── go.sum
```

## 📋 Capas de la Arquitectura

### 1. **Domain** (Dominio)
- **Entidades**: Modelos de negocio puros sin dependencias externas
- **Repositorios**: Interfaces que definen cómo acceder a datos
- **Servicios**: Interfaces para servicios del dominio
- **Sin dependencias**: Esta capa NO depende de ninguna otra

### 2. **Use Cases** (Casos de Uso)
- Implementan la lógica de negocio específica de la aplicación
- Orquestan el flujo entre repositorios y servicios
- Dependen solo de las interfaces del dominio

### 3. **Infrastructure** (Infraestructura)
- Implementaciones concretas de los repositorios
- Detalles técnicos (PDF, LLM, procesamiento de texto)
- Adaptadores para servicios externos

### 4. **Adapters** (Adaptadores)
- HTTP handlers, DTOs, middlewares
- Traducen entre el mundo externo y los casos de uso
- Entrada/salida del sistema

## 🚀 Características

- ✅ **Análisis de contratos PDF** con IA
- ✅ **Procesamiento por fragmentos** para documentos grandes
- ✅ **Consolidación jerárquica** de análisis
- ✅ **Estimación de tokens** antes del análisis
- ✅ **Soporte para LLM local y online**
- ✅ **Reintentos automáticos** en caso de fallo
- ✅ **Interfaz web moderna**

## 🛠️ Tecnologías

- **Go 1.24+**: Lenguaje de programación
- **Clean Architecture**: Patrón arquitectónico
- **github.com/ledongthuc/pdf**: Extracción de texto de PDF
- **HTTP estándar de Go**: Servidor web
- **Vanilla JS**: Frontend sin frameworks

## 📦 Instalación

1. Clonar el repositorio
```bash
git clone https://github.com/rodascaar/contractis.git
cd Contractis
```

2. Instalar dependencias
```bash
go mod download
```

3. Compilar
```bash
go build -o contractis ./cmd
```

## 🏃 Ejecución

### Desarrollo
```bash
go run ./cmd/main.go
```

### Producción
```bash
./contractis
```

El servidor se iniciará en `http://localhost:8080`

## 🔧 Configuración

### LLM Local
1. Instalar LM Studio o similar
2. Iniciar servidor local en `http://localhost:1234`
3. Configurar en la interfaz web

### LLM Online (OpenAI, etc.)
1. Obtener API key
2. Configurar en la interfaz web:
   - URL de API
   - API Key
   - Nombre del modelo

## 🧪 Casos de Uso

### Analizar un contrato
```go
// El caso de uso orquesta todo el flujo
analyzeUseCase := usecases.NewAnalyzeContractUseCase(
    pdfExtractor,
    llmClient,
    textProcessor,
)

result, err := analyzeUseCase.Execute(ctx, pdfPath, llmConfig)
```

### Estimar tokens
```go
estimateUseCase := usecases.NewEstimateTokensUseCase(
    pdfExtractor,
    textProcessor,
)

estimation, err := estimateUseCase.Execute(pdfPath, maxTokens)
```

## 📐 Principios de Clean Architecture Aplicados

### 1. **Dependency Rule** (Regla de Dependencia)
- Las dependencias apuntan hacia adentro
- El dominio no conoce la infraestructura
- Los casos de uso dependen solo de interfaces

### 2. **Separation of Concerns** (Separación de Responsabilidades)
- Cada capa tiene una responsabilidad clara
- Fácil de testear cada capa independientemente

### 3. **Dependency Inversion** (Inversión de Dependencias)
- Se usan interfaces en lugar de implementaciones concretas
- La inyección de dependencias se hace en `main.go`

## 🧩 Extensibilidad

### Agregar un nuevo repositorio
1. Definir interfaz en `internal/domain/repositories/`
2. Implementar en `internal/infrastructure/`
3. Inyectar en `cmd/main.go`

### Agregar un nuevo caso de uso
1. Crear en `internal/usecases/`
2. Usar interfaces del dominio
3. Crear handler en `internal/adapters/http/handlers/`
4. Configurar ruta en router

### Agregar un nuevo endpoint HTTP
1. Crear handler en `internal/adapters/http/handlers/`
2. Agregar DTO en `internal/adapters/http/dto/`
3. Configurar en `internal/adapters/http/router/`

## 🧪 Testing

La arquitectura facilita el testing:

```go
// Mock del repositorio
type MockPDFRepository struct{}
func (m *MockPDFRepository) ExtractText(path string) (string, error) {
    return "mock text", nil
}

// Test del caso de uso
func TestAnalyzeContract(t *testing.T) {
    mockPDF := &MockPDFRepository{}
    mockLLM := &MockLLMRepository{}
    processor := text.NewProcessor()
    
    useCase := usecases.NewAnalyzeContractUseCase(mockPDF, mockLLM, processor)
    // ... test logic
}
```

## 📝 Notas Importantes

- El archivo `main_old.go` contiene el código monolítico original
- Todos los imports usan el módulo `github.com/rodascaar/contractis`

## 🤝 Contribuir

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto está bajo la licencia MIT.

## 👥 Autores

- Carlos Barrios- Trabajo inicial

## 🙏 Agradecimientos

- Clean Architecture por Robert C. Martin
- Comunidad de Go
