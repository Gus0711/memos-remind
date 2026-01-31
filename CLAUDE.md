# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Memos is a self-hosted note-taking and knowledge management platform with:
- **Backend:** Go 1.25, Echo HTTP server, Connect RPC + gRPC-Gateway
- **Frontend:** React 18.3, TypeScript, Vite 7, Tailwind CSS v4
- **Databases:** SQLite (default), MySQL, PostgreSQL

## Development Commands

### Backend
```bash
go run ./cmd/memos --port 8081          # Start dev server
go test ./...                            # Run all tests
go test ./store/...                      # Run specific package tests
golangci-lint run                        # Lint Go code
```

### Frontend
```bash
cd web && pnpm install                   # Install dependencies
pnpm dev                                 # Dev server (proxies to localhost:8081)
pnpm lint                                # Type checking via Biome
pnpm lint:fix                            # Auto-fix lint issues
pnpm build                               # Production build
pnpm release                             # Build and copy to backend
```

### Protocol Buffers
```bash
cd proto && buf generate                 # Regenerate Go and TypeScript
cd proto && buf lint                     # Lint proto files
```

## Architecture

### Backend Structure
- `cmd/memos/` - Entry point, Cobra CLI setup
- `server/router/api/v1/` - gRPC service implementations
- `server/auth/` - JWT and PAT authentication
- `store/` - Data layer with driver interface pattern
- `store/db/{sqlite,mysql,postgres}/` - Database driver implementations
- `plugin/` - Pluggable components (scheduler, email, webhook, markdown, storage/s3)
- `proto/` - Protocol Buffer definitions

### Frontend Structure
- `web/src/components/` - React components
- `web/src/contexts/` - React Context for client state (Auth, View, MemoFilter)
- `web/src/hooks/` - React Query hooks for server state
- `web/src/pages/` - Page components
- `web/src/types/proto/` - Generated TypeScript from .proto files

### Key Patterns

**Dual Protocol API:** Connect RPC for browsers (type-safe), gRPC-Gateway for REST (`/api/v1/*`)

**Driver Interface:** All database operations through `store.Driver` interface with SQLite/MySQL/PostgreSQL implementations

**State Management:** React Query v5 for server state, React Context for client state

**Authentication:** JWT tokens (15-min expiration) and Personal Access Tokens (long-lived)

## Adding New Features

### New API Endpoint
1. Edit `proto/api/v1/*_service.proto`
2. Run `cd proto && buf generate`
3. Implement in `server/router/api/v1/*_service.go`
4. If public, add to `server/router/api/v1/acl_config.go`
5. Create frontend hook in `web/src/hooks/use*Queries.ts`

### Database Schema Changes
1. Create migration files in `store/migration/{sqlite,mysql,postgres}/{version}/`
2. Update `store/migration/{driver}/LATEST.sql`
3. Add methods to `store/driver.go` if new table
4. Implement in `store/db/{driver}/*.go`

## Linting

**Go:** golangci-lint with revive, govet, staticcheck, gocritic. Forbidden: `fmt.Errorf`, `ioutil.ReadDir`

**TypeScript:** Biome (replaces ESLint+Prettier). Line width: 140 chars, semicolons required

## Testing

Backend tests use testify and support all three database drivers. Set `DRIVER` env var to test against MySQL/PostgreSQL:
```bash
DRIVER=mysql DSN="user:pass@tcp(localhost:3306)/memos" go test ./...
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MEMOS_PORT` | `8081` | HTTP port |
| `MEMOS_DATA` | `~/.memos` | Data directory |
| `MEMOS_DRIVER` | `sqlite` | Database: sqlite, mysql, postgres |
| `MEMOS_DSN` | | Database connection string |
| `MEMOS_DEMO` | `false` | Enable demo mode |
