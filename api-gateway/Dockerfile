FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy the entire project to the container
COPY . .

# Build the API gateway service
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway ./api-gateway

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /api-gateway .

EXPOSE 8000
CMD ["./api-gateway"]
