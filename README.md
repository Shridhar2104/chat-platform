# 🚀 Real-Time Chat Authentication Service

> Lightning-fast authentication microservice built with Go - **2ms response times, 415 req/sec**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Performance](https://img.shields.io/badge/Performance-⚡%20Sub--10ms-brightgreen.svg)](PERFORMANCE.md)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](Dockerfile)

## ✨ Why This Rocks

- ⚡ **Sub-10ms response times** - Faster than most database queries
- 🔥 **415 requests/sec** - Handles serious traffic
- 🛡️ **Enterprise security** - JWT + bcrypt + rate limiting
- 🐳 **Production ready** - Docker + Kubernetes deployment
- 📊 **Zero failures** - 100% success rate under load

## 🚀 Quick Start

### Run It Locally

```bash
# 1. Clone and setup
git clone https://github.com/yourusername/chat-platform.git
cd chat-platform

# 2. Start services
make dev-up

# 3. Configure
cp .env.example .env

# 4. Run auth service
cd services/auth-service
go run cmd/server/main.go
Test It Works
bash# Health check
curl http://localhost:8080/health

# Register a user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","display_name":"Test User"}'
📖 API Endpoints
MethodEndpointWhat It DoesPOST/api/v1/auth/registerCreate new userPOST/api/v1/auth/loginUser loginPOST/api/v1/auth/refreshRefresh tokenGET/api/v1/auth/meGet user infoGET/healthService health
Example Usage
Register:
bashcurl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"securepass123","display_name":"John"}'
Response:
json{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "display_name": "John"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": 1748778512
}
🏗️ What's Inside
Tech Stack:

Go + Gin - Fast HTTP framework
PostgreSQL - User data storage
Redis - Session management
JWT - Stateless authentication
Docker - Easy deployment

Security Features:

bcrypt password hashing
JWT token authentication
Rate limiting (60 req/min)
CORS protection
Input validation

⚡ Performance
See PERFORMANCE.md for detailed benchmarks.
Quick stats:

Health checks: 2ms avg, 415 req/sec
User registration: 6ms avg, 166 req/sec
User login: 3ms avg, 300 req/sec
Concurrent users: 25+ supported

🐳 Deployment
Development:
bashmake dev-up
Production:
bashkubectl apply -f k8s/
Environment Variables:
bashPORT=8080
DATABASE_URL=postgres://user:pass@localhost:5432/db
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key
RATE_LIMIT_RPM=60
🧪 Testing
bash# Run tests
make test

# Load testing
./tests/benchmark.sh
📚 More Info

Performance Benchmarks - Detailed speed tests
Contributing Guide - How to contribute
Issues - Report bugs

🤝 Contributing
Contributions welcome! See CONTRIBUTING.md for guidelines.

Built for speed, security, and scale 🚀