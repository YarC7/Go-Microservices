# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Common Development Commands

### Project Setup and Development
```powershell
# Clone and setup the project
git clone <repository-url>
cd Go-Microservices

# Start all services with Docker Compose
docker-compose up --build

# Start services in background
docker-compose up -d --build

# Stop all services
docker-compose down

# View service logs
docker-compose logs -f [service-name]
docker-compose logs -f order-service
```

### Testing
```powershell
# Run all tests (unit + integration)
make test

# Run only unit tests  
make test-unit

# Run only integration tests
make test-integration

# Generate test coverage report (creates coverage.html)
make test-coverage

# Clean test cache and coverage files
make clean

# Run tests for specific service
go test -v ./order-service/tests/...

# Run specific test function
go test -v ./order-service/tests -run TestCreateOrder
```

### Development Workflow
```powershell
# Build a specific service
docker build -t service-name ./service-directory/
docker build -t order-service ./order-service/

# Run Go modules commands
go mod download    # Download dependencies
go mod tidy       # Clean up dependencies
go mod vendor     # Vendor dependencies

# Run a single service locally (requires dependencies)
cd order-service
go run main.go

# View running containers
docker-compose ps

# Check service health
curl http://localhost:8000/health    # API Gateway
curl http://localhost:8081/metrics   # Prometheus metrics

# Access service documentation
curl http://localhost:8000/          # API Gateway endpoints list
```

### Database Operations
```powershell
# Connect to databases (when containers are running)
docker exec -it go-microservices-product-db-1 psql -U postgres -d products_db
docker exec -it go-microservices-order-db-1 psql -U postgres -d orders_db
docker exec -it go-microservices-inventory-db-1 psql -U postgres -d inventory_db
docker exec -it go-microservices-notification-db-1 psql -U postgres -d notification_db
docker exec -it go-microservices-payment-db-1 psql -U postgres -d payment_db

# Initialize database schema
docker exec -i go-microservices-product-db-1 psql -U postgres < init.sql
```

### Monitoring and Debugging
```powershell
# Access monitoring dashboards
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3000 (admin/admin)
# RabbitMQ Management: http://localhost:15672 (guest/guest)

# View Redis data
docker exec -it go-microservices-redis-1 redis-cli
redis-cli> keys *
redis-cli> get order:1

# Check RabbitMQ queues
docker exec -it go-microservices-rabbitmq-1 rabbitmqctl list_queues
```

## Architecture Overview

### Microservices Structure
This is a **Domain-Driven Design (DDD) microservices architecture** with **event-driven communication**:

- **API Gateway (8000)**: Single entry point, request routing, CORS handling, static file serving
- **Product Service (8080)**: Product catalog management
- **Order Service (8081)**: Order processing with caching, messaging, batch operations, and payment integration
- **Inventory Service (8082)**: Stock management and availability checks  
- **Notification Service (8083)**: Asynchronous notification handling
- **Payment Service (8084)**: Payment processing with Stripe sandbox integration
- **Web UI (Client)**: React-based frontend with Vite for user interaction

### Key Architectural Patterns

#### Communication Patterns
- **Synchronous**: HTTP/REST via API Gateway with reverse proxy
- **Asynchronous**: RabbitMQ with topic exchange for event publishing
- **Caching**: Redis for order caching (30-min TTL, cache-aside pattern)
- **Resilience**: Circuit breaker pattern for external service calls

#### Data Architecture
- **Database per Service**: Each microservice has its own PostgreSQL database
- **Event Sourcing**: Order events published to RabbitMQ for decoupled processing
- **Caching Strategy**: Redis with automatic TTL and cache invalidation

#### Service Internal Structure (Order Service Example)
```
order-service/
├── cache/           # Redis caching layer
├── controller/      # HTTP request handlers
├── db/             # Database connection and schema
├── model/          # Data models and structs
├── queue/          # RabbitMQ message publishing
├── resilience/     # Circuit breaker implementation
├── routes/         # HTTP route definitions
├── service/        # External service clients (inventory, notification)
├── worker/         # Worker pool for batch processing
├── metrics/        # Prometheus metrics
├── docs/           # Swagger documentation
└── tests/          # Unit and integration tests
```

### Advanced Features

#### Order Service Batch Processing
- **Worker Pool**: Configurable concurrent processing (default: 10 workers)
- **Performance**: Handles 1000+ orders/minute with sub-100ms response times
- **Resilience**: Timeout handling (30s), error tracking, and partial failure support

#### Monitoring and Observability
- **Metrics**: Prometheus metrics on `/metrics` endpoints
- **Monitoring**: Grafana dashboards for service performance
- **Health Checks**: `/health` endpoints for service availability
- **Tracing**: Request logging with correlation IDs

### Environment Configuration

#### Key Environment Variables
```bash
# Database connections (per service)
DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME

# Service URLs for inter-service communication
PRODUCT_SERVICE_URL=http://product-service:8080
ORDER_SERVICE_URL=http://order-service:8081
INVENTORY_SERVICE_URL=http://inventory-service:8082
NOTIFICATION_SERVICE_URL=http://notification-service:8083
PAYMENT_SERVICE_URL=http://payment-service:8084

# Infrastructure components
REDIS_HOST=redis:6379
RABBITMQ_HOST=rabbitmq:5672

# Payment processing
STRIPE_SECRET_KEY=sk_test_...   # Stripe test secret key
STRIPE_PUBLISHABLE_KEY=pk_test_... # Stripe test publishable key

# Performance tuning
WORKER_POOL_SIZE=10        # Batch processing workers
BATCH_TIMEOUT=30s          # Batch operation timeout
```

### Service Ports and Access Points
```
API Gateway:         http://localhost:8000
Product Service:     http://localhost:8080
Order Service:       http://localhost:8081
Inventory Service:   http://localhost:8082
Notification Service: http://localhost:8083
Payment Service:     http://localhost:8084
Web UI:              http://localhost:8000 (served via API Gateway)
Prometheus:          http://localhost:9090
Grafana:            http://localhost:3000
RabbitMQ UI:        http://localhost:15672
Redis:              localhost:6379
```

### CI/CD Pipeline
- **GitHub Actions** with comprehensive pipeline:
  - **Test Stage**: Unit/integration tests with PostgreSQL, Redis, RabbitMQ
  - **Security Stage**: Trivy vulnerability scanning + gosec code analysis
  - **Build Stage**: Multi-service Docker image builds
  - **Deploy Stage**: Kubernetes deployment with image updates

- **Pipeline Triggers**: Push to main, Pull Requests
- **Container Registry**: GitHub Container Registry (ghcr.io)
- **Quality Gates**: Tests must pass, security scans must be clean

### API Design Patterns
- **RESTful APIs**: Standard HTTP verbs and status codes
- **Versioned APIs**: `/api/v1/` prefix for backward compatibility  
- **Batch Operations**: `POST /orders/batch` for bulk processing
- **Resource-based URLs**: `/orders/:id`, `/products/:id`, `/payments/:id`
- **PATCH Support**: `PATCH /orders/:id/status` for partial updates
- **Payment Integration**: `POST /orders/with-payment` for order with Stripe payment
- **Static File Serving**: API Gateway serves React UI from `/` endpoint

### Performance Characteristics
- **Throughput**: 1000+ orders/minute processing capacity
- **Response Time**: Sub-100ms average API response times  
- **Availability**: 99.9% uptime SLA with automatic failover
- **Scalability**: Horizontal scaling via Kubernetes deployments
- **Fault Tolerance**: Circuit breakers, retry mechanisms, graceful degradation

### Development Guidelines
- **Go Version**: 1.21+ required
- **Framework**: Gin for HTTP services, direct SQL for database operations
- **Frontend**: React with Vite for build tooling
- **Payment Processing**: Stripe API v76 for sandbox testing
- **Testing**: Table-driven tests, mocks for external dependencies
- **Error Handling**: Structured error responses with proper HTTP status codes
- **Logging**: Structured logging with correlation IDs for request tracing
- **Documentation**: Swagger/OpenAPI docs available at service `/swagger/index.html` endpoints
- **Security**: API keys managed via environment variables, no secrets in code

### Message Queue Architecture
```
RabbitMQ Topic Exchange: "orders"
├── Routing Key: "order.created"    → Notification processing
├── Routing Key: "order.updated"    → Inventory updates  
├── Routing Key: "order.cancelled"  → Cleanup processes
└── Routing Key: "order.completed"  → Analytics/reporting
```

This architecture supports high scalability, fault tolerance, and maintainability through proper separation of concerns, event-driven communication, and comprehensive monitoring.

## Deployment Options

### Quick Start (Local Development)
```powershell
# 1. Set up environment variables
cp .env.example .env
# Edit .env with your Stripe keys and other configs

# 2. Start all services
docker-compose up --build

# 3. Start frontend (in new terminal)
cd client
npm install
cp .env.example .env.local
# Edit .env.local with frontend configs
npm run dev
```

**Access Points:**
- Frontend UI: http://localhost:5173 (Vite dev server)
- API Gateway: http://localhost:8000
- All services: Individual ports (8080-8084)

### Production Deployment

#### Frontend-Only (Vercel/Netlify)
```powershell
# Deploy frontend to Vercel
cd client
npm run build
npx vercel --prod

# Set environment variables in Vercel dashboard:
# VITE_API_BASE_URL=https://your-api-gateway.com
# VITE_STRIPE_PUBLISHABLE_KEY=pk_test_...
```

#### Full-Stack (Railway/Render)
- Deploy each microservice as separate service
- Configure databases (PostgreSQL, Redis, RabbitMQ)
- Set environment variables for service communication
- Update frontend API_BASE_URL to point to deployed gateway

#### Self-Hosted (VPS/Cloud)
```powershell
# Create production docker-compose
cp docker-compose.yml docker-compose.prod.yml
# Edit with production configs

docker-compose -f docker-compose.prod.yml up -d --build
```

### Deployment Checklist
- [ ] Get Stripe API keys (test/live)
- [ ] Configure environment variables
- [ ] Test locally with `docker-compose up`
- [ ] Deploy backend services
- [ ] Deploy frontend with correct API URLs
- [ ] Test payment flow with Stripe test cards

See `DEPLOYMENT.md` for detailed deployment instructions and troubleshooting.
