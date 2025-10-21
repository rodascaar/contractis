# 🚀 Guía Rápida - Contractis

## Inicio Rápido

### 1. Compilar el Proyecto
```bash
cd Contractis
go build -o contractis ./cmd
```

### 2. Ejecutar el Servidor
```bash
./contractis
```

O directamente sin compilar:
```bash
go run ./cmd/main.go
```

El servidor estará disponible en: **http://localhost:8080**

### 3. Configurar el LLM

#### Opción A: LLM Local (LM Studio)
1. Descargar e instalar [LM Studio](https://lmstudio.ai/)
2. Descargar el modelo Qwen 3 4B o similar
3. Iniciar servidor local (puerto 1234 por defecto)
4. En la interfaz web de Contractis:
   - Tipo: Local
   - URL: `http://localhost:1234/v1/chat/completions`

#### Opción B: LLM Online (OpenAI, etc.)
1. Obtener API Key de tu proveedor
2. En la interfaz web de Contractis:
   - Tipo: Online
   - URL: `https://api.openai.com/v1/chat/completions`
   - API Key: tu-api-key
   - Modelo: `gpt-3.5-turbo` o `gpt-4`

### 4. Analizar un Contrato
1. Abrir http://localhost:8080
2. Click en "Seleccionar Archivo"
3. Subir un PDF (máximo 10MB)
4. Click en "📊 Estimar Tokens" (opcional pero recomendado)
5. Click en "Analizar Contrato"
6. Esperar el resultado

## 📁 Estructura del Proyecto

```
Contractis/
├── cmd/
│   └── main.go                    # Punto de entrada
├── internal/
│   ├── domain/                    # Reglas de negocio
│   │   ├── entities/              # Entidades del dominio
│   │   ├── repositories/          # Interfaces de repositorios
│   │   └── services/              # Interfaces de servicios
│   ├── usecases/                  # Casos de uso
│   │   ├── analyze_contract.go
│   │   ├── estimate_tokens.go
│   │   └── test_llm_connection.go
│   ├── infrastructure/            # Implementaciones técnicas
│   │   ├── pdf/                   # Extractor PDF
│   │   ├── llm/                   # Cliente LLM
│   │   └── text/                  # Procesador de texto
│   └── adapters/                  # Adaptadores HTTP
│       └── http/
├── main_old.go                    # Código monolítico original
├── go.mod
├── README.md
├── ARCHITECTURE.md
└── MIGRATION_GUIDE.md
```

## 🔧 Comandos Útiles

### Desarrollo
```bash
# Ejecutar sin compilar
go run ./cmd/main.go

# Compilar
go build -o contractis ./cmd

# Limpiar dependencias
go mod tidy

# Ver dependencias
go list -m all
```

### Testing (cuando se agreguen tests)
```bash
# Ejecutar todos los tests
go test ./...

# Tests con cobertura
go test -cover ./...

# Tests verbose
go test -v ./...
```

### Producción
```bash
# Compilar para producción (optimizado)
go build -ldflags="-s -w" -o contractis ./cmd

# Cross-compile para Linux
GOOS=linux GOARCH=amd64 go build -o contractis-linux ./cmd

# Cross-compile para Windows
GOOS=windows GOARCH=amd64 go build -o contractis.exe ./cmd
```

## 🎯 Casos de Uso Principales

### 1. Analizar Contrato
**Endpoint**: `POST /upload`

**Parámetros**:
- `file`: PDF del contrato
- `llmConfig`: Configuración del LLM (JSON)

**Respuesta**:
```json
{
  "success": true,
  "data": "Análisis del contrato..."
}
```

### 2. Estimar Tokens
**Endpoint**: `POST /estimate`

**Parámetros**:
- `file`: PDF del contrato
- `maxTokens`: Máximo de tokens

**Respuesta**:
```json
{
  "success": true,
  "characterCount": 15000,
  "estimatedTokens": 5000,
  "chunks": 5,
  "totalTokens": 12000,
  "recommendedMaxTokens": 1000,
  "warning": ""
}
```

## ⚙️ Configuración

### Variables de Entorno (futuro)
Actualmente no se usan, pero puedes agregar:

```bash
# .env
PORT=8080
STATIC_PATH=../static
MAX_FILE_SIZE=10485760  # 10MB
```

### Personalización

#### Cambiar Puerto
Editar `cmd/main.go`:
```go
server := httpAdapter.NewServer("3000", appRouter)  // Puerto 3000
```

#### Cambiar Ruta Estática
Editar `cmd/main.go`:
```go
appRouter := router.NewRouter(
    uploadHandler,
    estimateHandler,
    "/ruta/a/static",  // Nueva ruta
)
```

## 🐛 Troubleshooting

### Error: "No se pudo conectar al LLM"
- Verificar que el servidor LLM esté corriendo
- Verificar la URL del LLM
- Verificar la API Key (si es online)

### Error: "Archivo demasiado grande"
- Máximo: 10MB
- Dividir el PDF en partes más pequeñas

### Error: "Solo se permiten archivos PDF"
- Verificar que el archivo tenga extensión `.pdf`
- Convertir documentos Word/Excel a PDF primero

### El servidor no inicia
```bash
# Verificar que el puerto 8080 esté libre
lsof -i :8080

# O usar otro puerto
# Editar cmd/main.go y cambiar "8080" por "3000"
```

## 📊 Rendimiento

### Tiempos Estimados de Análisis

| Tamaño del Documento | Fragmentos | Tiempo Estimado |
|---------------------|------------|-----------------|
| < 5 páginas | 1-2 | 30-60 segundos |
| 5-10 páginas | 3-5 | 1-3 minutos |
| 10-20 páginas | 6-10 | 3-8 minutos |
| 20-50 páginas | 11-20 | 8-20 minutos |
| > 50 páginas | 20+ | 20+ minutos |

**Nota**: Los tiempos dependen del modelo LLM usado y su velocidad.

## 🔐 Seguridad

### Buenas Prácticas
1. **No compartir API Keys**: Nunca subir al repositorio
2. **Validar archivos**: El sistema valida tipo y tamaño
3. **Límite de tamaño**: 10MB máximo
4. **Archivos temporales**: Se eliminan automáticamente

### Para Producción
- Agregar autenticación
- Usar HTTPS
- Agregar rate limiting
- Validar origen de requests (CORS más restrictivo)

## 📚 Documentación Adicional

- [README.md](./README.md) - Documentación completa
- [ARCHITECTURE.md](./ARCHITECTURE.md) - Detalles de arquitectura
- [MIGRATION_GUIDE.md](./MIGRATION_GUIDE.md) - Guía de migración

## ❓ FAQ

**P: ¿Puedo usar con otros modelos LLM?**  
R: Sí, cualquier modelo compatible con API de OpenAI.

**P: ¿Funciona offline?**  
R: Sí, con un LLM local como LM Studio.

**P: ¿Puedo analizar contratos en inglés?**  
R: Sí, pero está optimizado para español.

**P: ¿Guarda los contratos?**  
R: No, los archivos temporales se eliminan después del análisis.

**P: ¿Cuánto cuesta usar con API online?**  
R: Depende del proveedor y tamaño del documento. Usa la estimación de tokens para calcular.

## 🤝 Soporte

Para reportar bugs o solicitar features, crear un issue en el repositorio.
