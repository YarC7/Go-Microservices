# 🚀 Deployment Guide

This guide covers different deployment options for the Go Microservices E-commerce application.

## 📋 Prerequisites

1. **Stripe Account**: Get your API keys from [Stripe Dashboard](https://dashboard.stripe.com/test/apikeys)
2. **Environment Variables**: Configure required environment variables
3. **Domain/URL**: Where your API services will be accessible

## 🎯 Deployment Options

### Option 1: Local Development (Recommended for Testing)

```bash
# 1. Set up environment variables
cp .env.example .env
# Edit .env with your actual values

# 2. Start all services
docker-compose up --build

# 3. Build and start frontend
cd client
npm install
cp .env.example .env.local
# Edit .env.local with your values
npm run dev
```

**Access:**
- Frontend: http://localhost:5173
- API Gateway: http://localhost:8000
- Services: Individual ports (8080-8084)

### Option 2: Frontend-Only Deployment (Vercel/Netlify)

**Perfect for:** Demo purposes, frontend development

#### Vercel Deployment

```bash
# 1. Install Vercel CLI
npm i -g vercel

# 2. Navigate to client directory
cd client

# 3. Build the project
npm run build

# 4. Deploy to Vercel
vercel --prod
```

**Environment Variables on Vercel:**
- `VITE_API_BASE_URL`: Your API Gateway URL
- `VITE_STRIPE_PUBLISHABLE_KEY`: Your Stripe publishable key

#### Netlify Deployment

```bash
# 1. Build the project
cd client
npm run build

# 2. Deploy via Netlify CLI or drag & drop dist/ folder
npm i -g netlify-cli
netlify deploy --prod --dir=dist
```

### Option 3: Full-Stack Deployment (Railway/Render)

**Perfect for:** Production deployment, complete application

#### Railway Deployment

1. **Push to GitHub**:
   ```bash
   git add .
   git commit -m "Ready for deployment"
   git push origin main
   ```

2. **Deploy on Railway**:
   - Go to [Railway.app](https://railway.app)
   - Connect your GitHub repository
   - Deploy each service separately:
     - API Gateway
     - Payment Service
     - Order Service
     - Product Service
     - Inventory Service
     - Notification Service
     - PostgreSQL databases
     - Redis
     - RabbitMQ

3. **Set Environment Variables**:
   ```env
   # For each service
   DB_HOST=your-postgres-host
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=your-password
   DB_NAME=your-database
   
   # For payment service
   STRIPE_SECRET_KEY=sk_test_...
   
   # Service URLs
   PRODUCT_SERVICE_URL=https://your-product-service.railway.app
   ORDER_SERVICE_URL=https://your-order-service.railway.app
   # ... etc
   ```

### Option 4: Docker Compose Production

**Perfect for:** VPS deployment, self-hosting

```bash
# 1. Create production docker-compose
cp docker-compose.yml docker-compose.prod.yml

# 2. Update with production configs
# Edit docker-compose.prod.yml with:
# - Production database passwords
# - SSL certificates
# - Domain names
# - Health checks

# 3. Deploy
docker-compose -f docker-compose.prod.yml up -d --build
```

## 🔧 Required Environment Variables

### Backend Services

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-secure-password
DB_NAME=your-database

# Service URLs
PRODUCT_SERVICE_URL=http://product-service:8080
ORDER_SERVICE_URL=http://order-service:8081
INVENTORY_SERVICE_URL=http://inventory-service:8082
NOTIFICATION_SERVICE_URL=http://notification-service:8083
PAYMENT_SERVICE_URL=http://payment-service:8084

# Infrastructure
REDIS_HOST=redis:6379
RABBITMQ_HOST=rabbitmq:5672

# Payment Processing
STRIPE_SECRET_KEY=sk_test_your_stripe_secret_key
```

### Frontend

```env
# API Configuration
VITE_API_BASE_URL=https://your-api-gateway-url.com

# Stripe Configuration
VITE_STRIPE_PUBLISHABLE_KEY=pk_test_your_stripe_publishable_key

# Optional Features
VITE_ENABLE_ANALYTICS=false
VITE_ENABLE_DEBUG=false
```

## 🌐 Deployment Checklist

### Pre-Deployment
- [ ] Set up Stripe account and get API keys
- [ ] Configure environment variables
- [ ] Test all services locally
- [ ] Run `make test` to ensure all tests pass
- [ ] Update API URLs in frontend configuration

### Post-Deployment
- [ ] Verify all services are running
- [ ] Test API endpoints
- [ ] Test frontend functionality
- [ ] Verify payment processing works
- [ ] Check logs for errors
- [ ] Set up monitoring and alerts

## 🔍 Testing Your Deployment

### API Testing
```bash
# Test API Gateway health
curl https://your-api-gateway-url.com/health

# Test each service
curl https://your-api-gateway-url.com/api/v1/products
curl https://your-api-gateway-url.com/api/v1/orders
curl https://your-api-gateway-url.com/api/v1/payments
```

### Frontend Testing
- [ ] Load the frontend URL
- [ ] Navigate between pages
- [ ] Create a product
- [ ] Create an order
- [ ] Process a test payment

### Payment Testing
Use Stripe test cards:
- **Success**: `4242 4242 4242 4242`
- **Decline**: `4000 0000 0000 0002`
- **Any future expiry date and any 3-digit CVC**

## 🚨 Common Issues & Solutions

### CORS Issues
```javascript
// In API Gateway, ensure CORS is properly configured
app.use(cors({
  origin: [
    'https://your-frontend-domain.com',
    'https://your-frontend-domain.vercel.app'
  ],
  credentials: true
}));
```

### Environment Variables Not Loading
- Ensure variable names start with `VITE_` for frontend
- Restart services after changing environment variables
- Check variable names match exactly (case-sensitive)

### Database Connection Issues
- Verify database credentials
- Check network connectivity
- Ensure database accepts connections from your deployment IP

### Payment Processing Issues
- Verify Stripe keys are correctly set
- Check if using test vs live keys appropriately
- Ensure webhook endpoints are configured if using them

## 📊 Monitoring & Maintenance

### Health Checks
- Set up health check endpoints: `/health`
- Monitor service status regularly
- Set up alerts for service downtime

### Logs
- Monitor application logs for errors
- Set up log aggregation (ELK stack, Datadog, etc.)
- Regular log rotation

### Backups
- Database backups (daily/weekly)
- Configuration backups
- Code repository backups

## 💡 Recommended Deployment Path

1. **Start Local**: Test everything with `docker-compose up`
2. **Frontend First**: Deploy frontend to Vercel/Netlify
3. **Backend Services**: Deploy services to Railway/Render
4. **Production Setup**: Move to dedicated VPS when ready

## 🔗 Useful Links

- [Vercel Documentation](https://vercel.com/docs)
- [Railway Documentation](https://docs.railway.app/)
- [Stripe Testing Guide](https://stripe.com/docs/testing)
- [Docker Compose Production](https://docs.docker.com/compose/production/)

---

**Need Help?** Check the logs, verify environment variables, and ensure all services can communicate with each other!