# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# git para go modules; ca-certificates para o bundle de CAs copiado ao scratch
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Flags comuns para build totalmente estático (sem libc, pronto para scratch)
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Build server
RUN go build \
        -tags 'netgo osusergo static_build' \
        -installsuffix netgo \
        -ldflags="-w -s -extldflags '-static'" \
        -trimpath \
        -o /out/server ./cmd/server

# Build migrate
RUN go build \
        -tags 'netgo osusergo static_build' \
        -installsuffix netgo \
        -ldflags="-w -s -extldflags '-static'" \
        -trimpath \
        -o /out/migrate ./cmd/migrate

# Build seed
RUN go build \
        -tags 'netgo osusergo static_build' \
        -installsuffix netgo \
        -ldflags="-w -s -extldflags '-static'" \
        -trimpath \
        -o /out/seed ./cmd/seed

# Runtime stage (scratch - sem OS, sem shell, sem nada)
FROM scratch

# Bundle de CAs — necessário para o TLS das chamadas HTTPS (Meta Graph API, OpenAI).
# Sem isto, o scratch não tem certificados e o handshake falha com x509.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Binários estáticos
COPY --from=builder /out/server  /server
COPY --from=builder /out/migrate /migrate
COPY --from=builder /out/seed    /seed

# Configs (inclui migrations/ usadas pelo golang-migrate)
COPY --from=builder /app/conf /conf

EXPOSE 8080

ENTRYPOINT ["/server"]
