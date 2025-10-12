# cloudscan-api-gateway

> API Gateway for CloudScan - routes HTTP requests, handles authentication, rate limiting

## Features
- JWT authentication
- Request routing to backend services
- Rate limiting (per user/org)
- Request validation
- OpenAPI/Swagger documentation
- CORS handling

## Endpoints
```
POST   /api/v1/auth/login
POST   /api/v1/auth/signup
GET    /api/v1/scans
POST   /api/v1/scans
GET    /api/v1/scans/:id
DELETE /api/v1/scans/:id
```

## Tech Stack
- Go + Echo framework
- JWT for authentication
- Redis for rate limiting
- Prometheus metrics