# SilentRaven - Remote-ID Enforcement System

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 16 with TimescaleDB
- Redpanda or Kafka
- Docker (for Redpanda)

### Installation

1. Clone and setup:
```bash
cd silentraven
go mod download
```

2. Setup database:
```bash
go run scripts/setup_db.go
```

3. Start services:
```bash
# Terminal 1 - Gateway
go run cmd/gateway/main.go

# Terminal 2 - Ingestion
go run cmd/ingestion/main.go

# Terminal 3 - API
go run cmd/api/main.go
```

## Architecture
```
Edge Gateway → Ingestion Service → Redpanda → TimescaleDB → API → Frontend
```

## Project Structure
```
silentraven/
├── cmd/                    # Main applications
│   ├── gateway/           # Edge gateway service
│   ├── ingestion/         # Data ingestion service
│   └── api/               # REST API service
├── internal/              # Private application code
│   ├── auth/             # Authentication & authorization
│   ├── database/         # Database operations
│   ├── models/           # Data models
│   ├── crypto/           # ECDSA verification
│   └── queue/            # Redpanda integration
├── pkg/                   # Public libraries
│   └── config/           # Configuration management
└── scripts/              # Setup and utility scripts
```

## License
Proprietary - Altivion Technologies Inc.