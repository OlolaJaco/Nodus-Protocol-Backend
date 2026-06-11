# Nodus Protocol Backend

[![CI](https://github.com/Nodus-protocol/Nodus-Protocol-Backend/actions/workflows/ci.yml/badge.svg)](https://github.com/Nodus-protocol/Nodus-Protocol-Backend/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-violet.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](go.mod)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

REST API for the Nodus Protocol AMM DEX — user auth, pool data, payments, and admin endpoints.

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

## API Reference

### Auth
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/auth/register` | Public | Create account |
| POST | `/api/v1/auth/login` | Public | Login, receive token pair |
| POST | `/api/v1/auth/refresh` | Public | Rotate access + refresh tokens |
| POST | `/api/v1/auth/logout` | Bearer | Blacklist current token |
| GET  | `/api/v1/auth/verify-email` | Public | Confirm email via link |
| POST | `/api/v1/auth/resend-verification` | Public | Re-send confirmation email |
| POST | `/api/v1/auth/forgot-password` | Public | Send password reset link |
| POST | `/api/v1/auth/reset-password` | Public | Set new password |

### Users
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET    | `/api/v1/users/me` | Bearer | Get own profile |
| PUT    | `/api/v1/users/me` | Bearer | Update name / avatar |
| DELETE | `/api/v1/users/me` | Bearer | Soft-delete account |
| PUT    | `/api/v1/users/me/password` | Bearer | Change password |
| POST   | `/api/v1/users/me/wallet` | Bearer | Link Stellar address |
| DELETE | `/api/v1/users/me/wallet` | Bearer | Unlink Stellar address |
| GET    | `/api/v1/users/me/lp-position` | Bearer | Live LP balance + redeemable tokens |
| GET    | `/api/v1/users/me/transactions` | Bearer | Paginated transaction history |
| GET    | `/api/v1/users/me/transactions/:id` | Bearer | Single transaction |
| GET    | `/api/v1/users/me/export` | Bearer | GDPR data export |

### Pool (public)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/pool/reserves` | Current XLM/USDC reserves |
| GET | `/api/v1/pool/quote` | Exact-input price quote (optional `?slippage_bps=`) |
| GET | `/api/v1/pool/reverse-quote` | Exact-output price quote |
| GET | `/api/v1/pool/lp-balance` | LP balance for a wallet address |
| GET | `/api/v1/pool/stats` | Prices, k-invariant, fee bps |
| GET | `/api/v1/pool/tvl` | Current pool TVL |
| GET | `/api/v1/pool/overview` | Combined stats + last snapshot timestamp |
| GET | `/api/v1/pool/price-history` | Historical reserve snapshots |
| GET | `/api/v1/pool/snapshots` | Recent raw snapshots (50) |
| GET | `/api/v1/pool/simulate/add-liquidity` | Preview LP tokens minted |
| GET | `/api/v1/pool/simulate/remove-liquidity` | Preview tokens redeemed for an address |

### Pool (authenticated)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/pool/build/swap` | Unsigned swap tx parameters |
| POST | `/api/v1/pool/build/add-liquidity` | Unsigned add-liquidity tx parameters |
| POST | `/api/v1/pool/build/remove-liquidity` | Unsigned remove-liquidity tx parameters |

### Payments
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/payments` | Bearer | Initiate payment via Core Engine |
| GET  | `/api/v1/payments` | Bearer | List own payments |
| GET  | `/api/v1/payments/:id` | Bearer | Get single payment |
| GET  | `/api/v1/payments/:id/receipt` | Bearer | Payment receipt |

### Leaderboard & Admin
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/leaderboard/traders` | Public | Top traders by confirmed volume |
| GET | `/api/v1/admin/users` | Admin | List all users |
| GET | `/api/v1/admin/users/:id` | Admin | Get user by ID |
| PUT | `/api/v1/admin/users/:id/role` | Admin | Change user role |
| DELETE | `/api/v1/admin/users/:id` | Admin | Hard-delete user |
| GET | `/api/v1/admin/stats` | Admin | Protocol-wide aggregates |
| GET | `/api/v1/admin/transactions` | Admin | All transactions |

### Health
```
GET /health   →  { status, database, redis, core_engine }
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