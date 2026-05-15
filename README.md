# Nodus Protocol Backend

Production-ready Go REST API with full user authentication.

## Stack
- **Framework**: Gin
- **ORM**: GORM + PostgreSQL
- **Cache / Sessions**: Redis
- **Auth**: JWT RS256 (asymmetric key pair)
- **Config**: Viper (12-factor)
- **Logging**: Zap (structured)
- **Containerisation**: Docker + docker-compose

## Quick Start

### 1. Prerequisites
- Go 1.22+
- Docker + Docker Compose
- `openssl` (for key generation)

### 2. Generate JWT Keys
```bash
make keys
```

### 3. Configure Environment
```bash
cp .env.example .env
# Edit .env with your values
```

### 4. Start Infrastructure
```bash
make up        # Starts Postgres + Redis in Docker
```

### 5. Run the Server
```bash
make run
```

The API will be available at `http://localhost:8080`.

---

## Auth Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/auth/register` | Public | Create account |
| POST | `/api/v1/auth/login` | Public | Login, get tokens |
| POST | `/api/v1/auth/refresh` | Public | Rotate token pair |
| POST | `/api/v1/auth/logout` | Bearer | Invalidate token |
| GET  | `/api/v1/auth/verify-email` | Public | Verify email via link |
| POST | `/api/v1/auth/resend-verification` | Public | Re-send verification |
| POST | `/api/v1/auth/forgot-password` | Public | Send reset link |
| POST | `/api/v1/auth/reset-password` | Public | Set new password |
| GET  | `/api/v1/users/me` | Bearer | Get own profile |
| PUT  | `/api/v1/users/me` | Bearer | Update profile |
| PUT  | `/api/v1/users/me/password` | Bearer | Change password |
| DELETE | `/api/v1/users/me` | Bearer | Delete account |

## Health Check
```
GET /health
```

## Make Commands
```bash
make help          # List all commands
make up            # Start Docker services
make down          # Stop Docker services
make run           # Run server
make build         # Build binary
make keys          # Generate RSA key pair
make migrate       # Run DB migrations
make migrate-down  # Rollback last migration
make test          # Run tests
make lint          # Run linter
make docs          # Generate Swagger docs
make docker-build  # Build production Docker image
```

## Project Structure
```
.
├── cmd/api/main.go          # Entry point
├── internal/
│   ├── config/              # Viper config
│   ├── database/            # Postgres + Redis connections
│   ├── middleware/          # Auth, CORS, rate-limit, logger, security
│   ├── models/              # GORM models (User, Token)
│   ├── modules/
│   │   ├── auth/            # Register, login, refresh, logout, reset
│   │   └── users/           # Profile CRUD
│   └── utils/               # JWT, crypto, email, response helpers
├── migrations/              # SQL migration files
├── scripts/                 # Key generation
├── certs/                   # RSA key pair (gitignored)
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

## Security Highlights
- **bcrypt** cost 12 for all passwords
- **RS256 JWT** with short-lived access tokens (15min) + rotating refresh tokens (7 days)
- **Redis token blacklist** for immediate logout
- **Rate limiting** per IP (strict on auth endpoints)
- **Security headers**: HSTS, CSP, X-Frame-Options, etc.
- **Soft deletes** throughout
- **Structured audit logging** with Zap
- Input validation on all request bodies