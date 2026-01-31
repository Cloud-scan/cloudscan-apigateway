# CloudScan API Gateway

> Unified API Gateway for CloudScan - handles authentication, routing, rate limiting, and request validation

## 🎯 Overview

The API Gateway serves as the single entry point for all client requests to CloudScan services. It provides:

- 🔐 **Authentication & Authorization** (JWT-based)
- 🚦 **Rate Limiting** (per user/organization)
- 🔀 **Request Routing** to backend services
- ✅ **Request Validation**
- 📝 **OpenAPI/Swagger Documentation**
- 🌐 **CORS Handling**
- 📊 **Metrics & Logging**

## 🏗️ Architecture

```
┌──────────────────────────────────────────┐
│          React UI / External Clients     │
└────────────────┬─────────────────────────┘
                 │
                 │ HTTP/REST
                 ▼
┌──────────────────────────────────────────────────────┐
│          API Gateway (This Service)                  │
│          Port: 8080 (HTTP/REST)                      │
│                                                      │
│  ┌────────────────────────────────────────────┐    │
│  │  Middleware Chain                          │    │
│  │  ├─ CORS                                   │    │
│  │  ├─ Auth (JWT validation)                  │    │
│  │  ├─ Rate Limiting (Redis-backed)           │    │
│  │  ├─ Request Logging                        │    │
│  │  └─ Request Validation                     │    │
│  └────────────┬───────────────────────────────┘    │
│               │                                      │
│               ▼                                      │
│  ┌────────────────────────────────────────────┐    │
│  │  Route Handler                             │    │
│  │  - /api/v1/scans/*    → Orchestrator       │    │
│  │  - /api/v1/storage/*  → Storage Service    │    │
│  │  - /api/v1/auth/*     → User Service       │    │
│  └────────────┬───────────────────────────────┘    │
└───────────────┼──────────────────────────────────┘
                │
       ┌────────┼────────┐
       │        │        │
       ▼        ▼        ▼
┌─────────┐ ┌──────────────┐ ┌──────────────┐
│  User   │ │ Orchestrator │ │   Storage    │
│ Service │ │   Service    │ │   Service    │
│ (gRPC)  │ │   (gRPC)     │ │   (gRPC)     │
└─────────┘ └──────────────┘ └──────────────┘
```

## 📡 API Endpoints

### Authentication
```
POST   /api/v1/auth/login       # User login
POST   /api/v1/auth/signup      # User registration
POST   /api/v1/auth/refresh     # Refresh JWT token
POST   /api/v1/auth/logout      # Invalidate token
```

### Scans
```
POST   /api/v1/scans            # Create new scan
GET    /api/v1/scans            # List scans (with filters)
GET    /api/v1/scans/:id        # Get scan details
DELETE /api/v1/scans/:id        # Cancel/delete scan
GET    /api/v1/scans/:id/findings  # Get scan findings
```

### Storage
```
POST   /api/v1/storage/artifacts      # Create artifact (get upload URL)
GET    /api/v1/storage/artifacts/:id  # Get artifact (get download URL)
DELETE /api/v1/storage/artifacts/:id  # Delete artifact
```

### Projects & Organizations
```
GET    /api/v1/organizations    # List user's organizations
POST   /api/v1/organizations    # Create organization
GET    /api/v1/projects         # List projects
POST   /api/v1/projects         # Create project
```

## 🔐 Authentication Flow

```
1. Client → POST /api/v1/auth/login
   {email, password}

2. API Gateway → User Service (gRPC)
   ValidateCredentials(email, password_hash)

3. User Service → Response
   {user_id, org_id, roles}

4. API Gateway → Generate JWT
   Token includes: {user_id, org_id, roles, exp}

5. API Gateway → Client
   {token: "eyJhbGc...", expires_at: "..."}

6. Client → Subsequent requests
   Headers: {Authorization: "Bearer eyJhbGc..."}

7. API Gateway → Validate JWT on each request
   - Verify signature
   - Check expiration
   - Extract user context
```

## ⚙️ Configuration

```bash
# Server
export SERVER_PORT=8080
export SERVER_HOST=0.0.0.0

# JWT
export JWT_SECRET=your-secret-key-here
export JWT_EXPIRATION=24h

# Backend Services
export ORCHESTRATOR_URL=cloudscan-orchestrator:9999
export STORAGE_URL=cloudscan-storage:8082
export USER_SERVICE_URL=cloudscan-users:8083

# Redis (for rate limiting)
export REDIS_URL=redis:6379
export REDIS_PASSWORD=

# Rate Limiting
export RATE_LIMIT_REQUESTS=100  # requests per window
export RATE_LIMIT_WINDOW=1m     # time window
```

## 🚀 Running

```bash
# Development
go run cmd/main.go

# Production
./cloudscan-apigateway-amd64
```

## 📊 Metrics

Prometheus metrics exposed on `/metrics`:

- `http_requests_total{method, path, status}` - Total HTTP requests
- `http_request_duration_seconds{method, path}` - Request latency
- `rate_limit_rejected_total{org_id}` - Rate limit rejections
- `jwt_validation_errors_total` - Failed JWT validations

## 🧪 Tech Stack

- **Framework:** Go + Echo
- **Authentication:** JWT (golang-jwt)
- **Rate Limiting:** Redis + go-redis
- **Service Communication:** gRPC (to backend services)
- **Metrics:** Prometheus
- **API Docs:** Swagger/OpenAPI

## 📄 License

Apache 2.0