# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

# Configure private Go modules
ARG GITHUB_TOKEN
ENV GOPRIVATE=github.com/mercadocercano/*
RUN if [ -n "$GITHUB_TOKEN" ]; then git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; fi

# Copiar archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./src/main.go

# Final stage
FROM alpine:3.18

# Variables de entorno configurables
ARG APP_VERSION=1.0
ARG APP_PORT=8110
ARG APP_ENV=development

# Añadir etiquetas para mejor gestión desde Terraform
LABEL maintainer="SaaS Team" \
      application="onboarding" \
      version="${APP_VERSION}" \
      description="Multi-tenant Onboarding service" \
      environment="${APP_ENV}"

# Definir variables de entorno
ENV PORT=${APP_PORT} \
    APP_ENV=${APP_ENV}

# Instalar dependencias necesarias en una sola capa
RUN apk add --no-cache postgresql-client dos2unix ca-certificates tzdata wget && \
    cp /usr/share/zoneinfo/UTC /etc/localtime && \
    echo "UTC" > /etc/timezone && \
    adduser -D appuser

WORKDIR /app

# Copiar el binario compilado desde el stage anterior
COPY --from=builder /app/main .

# Copiar las migraciones
COPY --from=builder /app/migrations ./migrations

# Asignar permisos adecuados
RUN chmod +x /app/main && \
    chown -R appuser:appuser /app

# Cambiar al usuario no-root para seguridad
USER appuser

# Exponer el puerto (configurable a través de ARG)
EXPOSE ${PORT}

# Configurar healthcheck para que Terraform sepa si el servicio está listo
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:${PORT}/health || exit 1

# Comando para ejecutar la aplicación
CMD ["./main"] 
