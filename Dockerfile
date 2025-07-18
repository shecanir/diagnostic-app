# Build stage
FROM golang:1.23.4-alpine AS builder
WORKDIR /app

# Copy go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 go build -o shecan-diagnostic

# Runtime stage
FROM alpine:latest
WORKDIR /app

# Needed tools for pinger/nslookup
RUN apk add --no-cache bind-tools iputils

COPY --from=builder /app/shecan-diagnostic /usr/local/bin/shecan-diagnostic

# Optional: copy .env if available
# COPY .env /app/

CMD ["shecan-diagnostic"]
