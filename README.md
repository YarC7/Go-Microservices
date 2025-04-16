# Go Microservices Project

A modern microservices project built with Go, implementing industry best practices and scalable architecture.

## Architecture

The project consists of the following microservices:

- **API Gateway** (Port: 8000): Single entry point for all client requests
- **Product Service** (Port: 8080): Product management
- **Order Service** (Port: 8081): Order processing with caching and message queue
- **Inventory Service** (Port: 8082): Inventory management
- **Notification Service** (Port: 8083): Notification handling

### Technologies Used

- **Go**: Primary programming language
- **Gin**: Web framework
- **PostgreSQL**: Primary database
- **Redis**: Caching layer
- **RabbitMQ**: Message queue
- **Docker & Docker Compose**: Containerization and orchestration
- **Prometheus & Grafana**: Monitoring and metrics
- **Circuit Breaker**: Fault tolerance handling
- **Swagger/OpenAPI**: API Documentation
- **Postman**: API Testing

## Key Features

### API Gateway
- Single entry point for all client requests
- Intelligent request routing
- CORS support
- Automatic API documentation
- Health check endpoints

### Order Service
- **Redis Caching**:
  - Order caching with 30-minute TTL
  - Automatic cache invalidation
  - Cache-aside pattern implementation

- **RabbitMQ Message Queue**:
  - Event publishing for new orders
  - Topic exchange for order events
  - Asynchronous notification processing

- **Batch Processing**:
  - Parallel processing of multiple orders
  - Configurable worker pool
  - Timeout handling
  - Success/failure tracking
  - Performance optimization for bulk operations

- **Resilience**:
  - Circuit breaker for service calls
  - Retry mechanism
  - Async notification handling
  - Error handling and logging

### Database
- PostgreSQL for each service
- Separate databases for isolation
- Optimized queries and indexing

### Monitoring
- Prometheus metrics
- Grafana dashboards
- Service health monitoring
- Performance metrics

## Installation and Running

1. Clone repository:
\`\`\`bash
git clone <repository-url>
cd go-microservices
\`\`\`

2. Run services with Docker Compose:
\`\`\`bash
docker-compose up --build
\`\`\`

3. Check services:
- API Gateway: http://localhost:8000
- Product Service: http://localhost:8080
- Order Service: http://localhost:8081
- Inventory Service: http://localhost:8082
- Notification Service: http://localhost:8083
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000

## API Endpoints

### API Gateway (http://localhost:8000)
- `/api/v1/products/*`: Product service endpoints
- `/api/v1/orders/*`: Order service endpoints
- `/api/v1/inventory/*`: Inventory service endpoints
- `/api/v1/notifications/*`: Notification service endpoints
- `/health`: Health check endpoint
- `/docs`: API documentation

### Order Service (http://localhost:8081)
- `POST /orders`: Create new order
  - Inventory check
  - Cache result
  - Publish event to RabbitMQ
  - Async notification
- `POST /orders/batch`: Process multiple orders in parallel
  - Concurrent processing using worker pool
  - Configurable number of workers
  - Timeout handling
  - Detailed success/failure tracking
- `GET /orders/:id`: Get order details (with Redis cache)
- `GET /orders`: List all orders
- `PUT /orders/:id`: Update order
- `DELETE /orders/:id`: Delete order
- `PATCH /orders/:id/status`: Update order status

## Batch Processing

### Features
- Parallel processing of large order volumes
- Configurable worker pool size (default: 10 workers)
- Timeout handling (default: 30 seconds)
- Detailed success/failure tracking
- Performance optimization

### Example Request
\`\`\`bash
curl -X POST http://localhost:8081/orders/batch \
  -H "Content-Type: application/json" \
  -d '[
    {
      "product_id": 1,
      "customer_id": 1,
      "quantity": 2
    },
    {
      "product_id": 2,
      "customer_id": 1,
      "quantity": 1
    }
    // ... more orders ...
  ]'
\`\`\`

### Example Response
\`\`\`json
{
  "total_orders": 1000,
  "successful": 990,
  "failed": 10,
  "failed_orders": [
    {
      "order_id": 5,
      "error": "Product not available"
    }
  ],
  "processing_time": "30s"
}
\`\`\`

### Performance
- Processing capacity: Up to 1000 orders/minute
- Average processing time: ~100ms per order
- Concurrent processing: 10 orders at a time
- Automatic timeout after 30 seconds

## Monitoring

### Prometheus Metrics
- Order processing time
- Cache hit/miss ratio
- Message queue performance
- Batch processing metrics
- Service health metrics

### Grafana Dashboards
- Service performance monitoring
- Error rate tracking
- Resource utilization
- Business metrics

## Environment Variables

### Order Service
- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `REDIS_HOST`: Redis host
- `RABBITMQ_HOST`: RabbitMQ host
- `INVENTORY_SERVICE_URL`: Inventory service URL
- `NOTIFICATION_SERVICE_URL`: Notification service URL
- `WORKER_POOL_SIZE`: Number of workers for batch processing
- `BATCH_TIMEOUT`: Timeout for batch processing

## Contributing

1. Fork repository
2. Create feature branch
3. Commit changes
4. Push to branch
5. Create Pull Request

## License

MIT License

## Testing

The project implements comprehensive testing strategies across different levels:

### Unit Tests (`/order-service/tests`)
- **Controller Tests**
  - Mock external services (Inventory, Notification)
  - Test business logic
  - Test error handling
  - Test request validation
  - Example test cases:
    ```go
    func TestCreateOrder(t *testing.T)
    func TestGetOrder(t *testing.T)
    func TestCreateBatchOrders(t *testing.T)
    ```

### Integration Tests (`/order-service/tests/integration`)
- **End-to-End Flow Tests**
  - Test complete order creation flow
  - Test batch processing
  - Real Redis integration
  - Real RabbitMQ integration
  - Example test cases:
    ```go
    func TestOrderFlowIntegration(t *testing.T)
    func TestCacheIntegration(t *testing.T)
    func TestMessageQueueIntegration(t *testing.T)
    ```

### Test Coverage
- Coverage reports in HTML format
- Track code coverage metrics
- Identify untested code paths

### Running Tests

Use the provided Makefile commands:

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Generate coverage report
make test-coverage

# Clean test cache and coverage files
make clean
```

### Test Environment
- Separate test database
- Isolated Redis instance (DB 1)
- Test-specific RabbitMQ queues
- Mock external services
- Cleanup after tests

### Test Features
- Table-driven tests
- Mock implementations
- Parallel test execution
- Timeout handling
- Cleanup routines
- Detailed assertions

## API Documentation

### Swagger/OpenAPI Documentation

Each service provides Swagger documentation for its API endpoints. Access the documentation at:

- API Gateway: http://localhost:8000/swagger/index.html
- Order Service: http://localhost:8081/swagger/index.html
- Product Service: http://localhost:8080/swagger/index.html
- Inventory Service: http://localhost:8082/swagger/index.html
- Notification Service: http://localhost:8083/swagger/index.html

The Swagger documentation includes:
- Detailed endpoint descriptions
- Request/response schemas
- Authentication requirements
- Example requests
- Response codes and examples

### Postman Collection

A comprehensive Postman collection is available for testing the APIs:

1. Import the collection from `order-service/docs/Order_Service.postman_collection.json`
2. Set up environment variables:
   - `base_url`: Base URL for the service (e.g., http://localhost:8081)
3. Use the collection to test:
   - Create Order
   - Get Order
   - Create Batch Orders
   - Update Order
   - Delete Order
   - Update Order Status

## CI/CD Pipeline

### Overview
Project sử dụng GitHub Actions để tự động hóa quy trình CI/CD, bao gồm testing, security scanning và deployment.

### Pipeline Stages

#### 1. Test Stage
- Chạy unit tests và integration tests
- Tạo báo cáo test coverage
- Upload kết quả test lên Codecov
- Môi trường test bao gồm:
  - PostgreSQL 13
  - Redis 6
  - RabbitMQ 3

#### 2. Security Scan
- Trivy: Quét lỗ hổng bảo mật trong dependencies và container images
- gosec: Phân tích mã nguồn Go để tìm các vấn đề bảo mật
- Chặn pipeline nếu phát hiện lỗ hổng nghiêm trọng

#### 3. Build Stage
- Build Docker images cho tất cả services
- Push images lên GitHub Container Registry (ghcr.io)
- Tag images với commit SHA

#### 4. Deploy Stage
- Tự động deploy khi merge vào nhánh main
- Deploy lên Kubernetes cluster
- Verify deployment status

### Trigger Events
Pipeline được kích hoạt khi:
- Push code vào nhánh main
- Tạo Pull Request vào nhánh main

### Setup Requirements

1. GitHub Repository Configuration:
   ```bash
   # Add required secrets
   KUBE_CONFIG: Base64 encoded kubeconfig file
   ```

2. Enable GitHub Container Registry:
   - Go to Settings > Packages
   - Enable GitHub Container Registry

3. Kubernetes Configuration:
   - Cluster đã được setup
   - Deployments cho tất cả services
   - Correct RBAC permissions

### Monitoring Pipeline

1. View Pipeline Status:
   - Go to repository's Actions tab
   - Select workflow run to view details

2. Test Results:
   - Test reports available as artifacts
   - Coverage reports on Codecov

3. Security Scan Results:
   - Trivy scan results in workflow logs
   - gosec analysis results in workflow logs

### Best Practices

1. Commit Guidelines:
   - Viết commit message rõ ràng
   - Mỗi commit chỉ nên chứa một thay đổi logic
   - Tham khảo [Conventional Commits](https://www.conventionalcommits.org/)

2. Branch Strategy:
   - Develop trên feature branches
   - Tạo Pull Request để merge vào main
   - Đảm bảo CI pass trước khi merge

3. Security:
   - Không commit secrets vào repository
   - Regular dependency updates
   - Review security scan results

### Troubleshooting

Common Issues:
1. Test Failures:
   - Check test logs in Actions tab
   - Verify test environment configuration
   - Check service dependencies

2. Build Failures:
   - Verify Dockerfile configurations
   - Check resource limits
   - Validate image tags

3. Deploy Failures:
   - Verify Kubernetes configuration
   - Check cluster access
   - Validate deployment manifests

### Continuous Improvement

1. Metrics to Monitor:
   - Build time
   - Test coverage
   - Deployment frequency
   - Failure rate
   - Mean time to recovery

2. Regular Maintenance:
   - Update dependencies
   - Review and optimize pipeline
   - Update documentation
   - Security patches 