# Etapa de Construcción (Builder Stage)
# Usamos la versión de Go especificada en go.mod (1.24.1) con Alpine.
FROM golang:1.24.1-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
# Descargar dependencias
RUN go mod download
COPY . .

# Compilar el binario: La mejor práctica para Alpine es usar ENLACE ESTÁTICO.
# Esto desactiva CGO y evita el error 'gcc not found' al hacer que el binario sea autocontenido.
# Las etiquetas '-a -tags netgo -ldflags...' fuerzan esta compilación estática.
RUN GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /app/contractis ./cmd/main.go


# Etapa Final (Runtime Stage)
# Usamos Alpine 3.21 para compatibilidad con las dependencias.
FROM alpine:3.21

WORKDIR /app
# Instalar certificados CA y timezone data para compatibilidad
RUN apk add --no-cache ca-certificates tzdata
# Copiar el binario desde la etapa de construcción
COPY --from=builder /app/contractis /app/contractis
# Copiar archivos estáticos desde el directorio padre (fuera del contexto de build)
COPY ../static /app/static
# Crear directorios necesarios
RUN mkdir -p /app/document /app/data
EXPOSE 8080
ENTRYPOINT ["/app/contractis"]
