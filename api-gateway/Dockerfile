FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod & go.sum từ root
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy toàn bộ source từ root
COPY . .

# Build service api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway ./api-gateway

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy binary từ builder
COPY --from=builder /api-gateway .

EXPOSE 8000
CMD ["./api-gateway"]
