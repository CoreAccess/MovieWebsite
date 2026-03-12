FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum for dependency resolution
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/webserver ./cmd/web/

# Smaller deployment image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/webserver .
# Copy UI files
COPY --from=builder /app/ui ./ui

# Expose port (Cloud Run passes standard PORT env)
EXPOSE 8080

CMD ["./webserver"]
