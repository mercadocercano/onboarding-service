# ==============================================
# Onboarding Service - Multi-stage Dockerfile
# ==============================================

# ==============================================
# Stage 1: Dependencies
# ==============================================
FROM golang:1.24-alpine AS deps
WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

# Configure private Go modules
ARG GITHUB_TOKEN
ENV GOPRIVATE=github.com/mercadocercano/*
RUN if [ -n "$GITHUB_TOKEN" ]; then git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; fi

COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ==============================================
# Stage 2: Build
# ==============================================
FROM deps AS builder

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -trimpath \
    -o onboarding-service ./src/main.go

# ==============================================
# Stage 3: Development (with Air hot reload)
# ==============================================
FROM golang:1.24-alpine AS development

RUN addgroup -g 1001 -S appgroup && \
    adduser -S -D -h /app -s /bin/sh -G appgroup -u 1001 appuser

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    postgresql-client \
    git \
    && cp /usr/share/zoneinfo/UTC /etc/localtime \
    && echo "UTC" > /etc/timezone \
    && apk del tzdata

RUN go install github.com/cosmtrek/air@v1.49.0

ARG GITHUB_TOKEN
ENV GOPRIVATE=github.com/mercadocercano/*
RUN if [ -n "$GITHUB_TOKEN" ]; then git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; fi

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN mkdir -p tmp scripts migrations logs /go/pkg/mod && \
    chmod -R 777 /go/pkg && \
    chown -R appuser:appgroup /app tmp scripts migrations logs

COPY --chown=appuser:appgroup . .

USER appuser

HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8110/health || exit 1

EXPOSE 8110

CMD sh -c 'if [ -n "$GITHUB_TOKEN" ]; then git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"; fi && air -c .air.toml'

# ==============================================
# Stage 4: Production (Distroless)
# ==============================================
FROM gcr.io/distroless/static-debian12:nonroot AS production

LABEL org.opencontainers.image.title="Onboarding Service" \
      org.opencontainers.image.description="Multi-tenant Onboarding service" \
      org.opencontainers.image.vendor="SaaS MT Team"

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /app/onboarding-service ./
COPY --from=builder --chown=nonroot:nonroot /app/migrations ./migrations/

USER nonroot

EXPOSE 8110

ENTRYPOINT ["./onboarding-service"]

# ==============================================
# Default stage: Development
# ==============================================
FROM development
