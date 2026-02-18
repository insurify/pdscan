# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

pdscan is a Go CLI tool that scans data stores (databases, files, S3) for unencrypted personally identifiable information (PII). It uses data sampling and naming conventions to detect sensitive data with minimal database load.

## Build & Development Commands

```bash
make install    # Install binary to $GOPATH/bin
make test       # Run all tests (go test ./... -v)
make lint       # Run staticcheck (excludes ST1005)
make format     # Run go fmt
```

Run a single test:
```bash
go test ./internal -v -run TestName
go test ./cmd -v -run TestName
```

Note: Integration tests in `cmd/root_test.go` require running database services (MySQL, PostgreSQL, SQL Server, MongoDB, Redis, Elasticsearch). Unit tests in `internal/main_test.go` run without external dependencies.

## Architecture

**Entry flow:** `main.go` → `cmd/root.go` (Cobra CLI) → `internal.Main()` (orchestration)

**Two adapter interfaces for scanning:**

- `DataStoreAdapter` — structured data stores (SQL databases, MongoDB, Redis). Fetches tables/collections, then samples rows. Implementations: `SqlAdapter` (Postgres, MySQL, MariaDB, SQLite, SQL Server), `MongodbAdapter`, `RedisAdapter`.
- `FileAdapter` — file-based sources. Fetches file list, then scans contents. Implementations: `LocalFileAdapter`, `S3Adapter`.
- `ElasticsearchAdapter` implements the top-level `Adapter` interface directly (used for both Elasticsearch and OpenSearch).

**Detection rules** (`internal/rules.go`): Four rule types — `regexRules` (pattern matching), `nameRules` (column name matching), `multiNameRules` (multi-column matching like lat+lon), `tokenRules` (value matching against known lists like surnames).

**MatchFinder** (`internal/match_finder.go`): Orchestrates pattern matching across rule types and computes confidence levels (high/medium/low).

**Output formatters** (`internal/format.go`): `TextFormatter` (colored terminal output) and `JSONFormatter` (ndjson).

**Concurrency:** Data store tables are scanned concurrently via `errgroup`. File scanning uses a worker pool capped at 20.

## Key Conventions

- Adapter selection is URL-scheme-based in `internal/main.go` (`scanDataStore` vs `scanFiles`)
- File processing (`internal/files.go`) handles plain text, ZIP, gzip, and tar.gz with automatic format detection
- The `--only`/`--except` flags filter rules by name (e.g., `email`, `phone`, `location`)
- SQLite support requires CGo (`go-sqlite3`) and is not in prebuilt binaries
