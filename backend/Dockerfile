# ---- Stage 1: Build ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod + go.sum first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

# ---- Stage 2: Run ----
FROM alpine:3.21

# Add ca-certificates for HTTPS calls and a non-root user for security
RUN apk --no-cache add ca-certificates && \
    adduser -D -u 1000 appuser

WORKDIR /app
COPY --from=builder /app/server .

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/health || exit 1

CMD ["./server"]
