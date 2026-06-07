# Multi-stage build
# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git for private modules (if any)
RUN apk add --no-cache git

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o english-tutor ./cmd/main.go

# Stage 2: Runtime
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# Copy binary
COPY --from=builder /app/english-tutor .

# Copy Swagger docs
COPY --from=builder /app/docs ./docs

# Create audio storage directory
RUN mkdir -p /app/audio

EXPOSE 8080

CMD ["./english-tutor"]
