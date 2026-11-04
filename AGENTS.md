# TeleNews Knowledge Base

**Project**: Telegram News Aggregator Bot
**Language**: Go 1.26.0 (minimum in `go.mod`)
**Version**: 2.3.0

---

## OVERVIEW

TeleNews aggregates news from multiple sources (Twitter, VK, RSS) and delivers them to Telegram users. Supports user subscriptions, source management, and automated news polling every 2 minutes.

## STRUCTURE

```
./
├── main.go              # Entry point, initializes services
├── helper.go            # Helper functions, config, news parsing orchestration
├── telegram.go          # Telegram bot commands and message handlers
├── models/              # Database models (User, Source, Item) - GORM-based
├── parse/               # News parsers (Twitter, VK, RSS)
├── telenews.yml         # Application configuration
├── go.mod               # Dependencies
├── Taskfile.yaml        # Task runner commands
└── docker-compose.yml   # Docker development environment
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add new command | `telegram.go` | Extend `telegramUpdate()` switch statement |
| Add new source type | `parse/`, `models/source.go` | Add type constant + parser implementation |
| Modify database schema | `models/*.go` | GORM auto-migration in `models.Init()` |
| Fix parsing error | `parse/*.go` | Check source-specific parser |
| User management | `models/user.go` | ChatID-based user lookup |
| Configuration | `helper.go:configStruct` | YAML unmarshaling |

## CONVENTIONS

**Error Logging Format**: `[ERR][COMPONENT] message`
- Components: `CFG`, `DB`, `LOG`, `Telegram`, `RSS`, `VKW`, `TWIT`

**Database Patterns**:
- Models use GORM with embedded `gorm.Model`
- Custom errors: `ErrAlreadyExists`, `ErrRecordNotFound`
- Association pattern: `db.Model(&parent).Association("Children")`
- Soft deletes via GORM (DeletedAt field)

**Source Types** (defined in `models/source.go`):
- `Twitter` - X API v2 recent search
- `VKWall` - VKontakte wall posts
- `RSS` - RSS/Atom feeds

**Naming**:
- Receiver names: PascalCase struct name (e.g., `ParseNews`, `teleNews`)
- Test files: `*_test.go` suffix
- Config struct: `configStruct`

## ANTI-PATTERNS

- **DO NOT** use `fmt.Println()` for production logging (use `logger.Println()`)
- **DO NOT** modify `vendor/` directly (use `go mod vendor`)
- **AVOID** blocking operations in `telegramUpdate()` loop (use goroutines)
- **NEVER** expose real credentials in `telenews.yml` (use `.env` or secrets)

## UNIQUE STYLES

**News Polling Architecture**:
- Background worker (`workNews()`) triggers every 2 minutes
- First run detection: checks if items exist for source
- Deduplication via MD5 hash of source+title+message+link
- Error backoff: 10-minute retry delay on parse failure

**Telegram Message Flow**:
- Commands processed in `telegramUpdate()` goroutine
- News sent via `parseNews()` worker
- Photo batches: 2–10 per media group; a single photo uses `sendPhoto`.

**Hash Generation**:
```go
func (newsStruct *NewsStruct) GenHash(sourceID uint) string
// MD5 of: sourceID + Title + MSG + Link
```

## COMMANDS

```bash
# Development (Docker)
task debug                    # Start with debugger

# Manual
go build -o telenews .        # Build binary
go test ./...                 # Run tests

# Docker
docker compose up --build     # Start services
docker compose -f docker-compose-debug.yml up --build  # Debug mode
```

## DEPENDENCIES

| Package | Purpose |
|---------|---------|
| `go-telegram-bot-api/v5` | Telegram Bot API |
| `gorm` + `lib/pq` | PostgreSQL ORM |
| Standard `net/http` | X API v2 and VK API clients |
| `gofeed` | RSS/Atom parsing |
| `go.yaml.in/yaml/v3` | Configuration parsing |

## NOTES

- `go.mod` and `go.sum` are authoritative. Local `vendor/` is ignored; regenerate it with `go mod vendor` after dependency changes.
- PostgreSQL required (configured via `telenews.yml`)
- Telegram bot token required in config
- X bearer token is optional (`twitter.bearerToken`); X polling and source advertisement are disabled without it.
- VK access token is configured through the legacy `vk.secureKey` YAML field.
- RSS sources require no additional credentials

## VALIDATION AND API BOUNDARIES

- Run `go test -race -shuffle=on -count=1 ./...`, `go vet ./...`, and `go build ./...`.
- Ordinary parser and Telegram transport tests use local fixtures, not real accounts.
- PostgreSQL tests skip unless `DBADDRESS`, `DBNAME`, and `DBUSER` are set; `DBPASS` is optional. Use only a disposable test database: initialization auto-migrates tables and tests create/delete rows.
- Live parser tests are opt-in via `X_BEARER_TOKEN` and `VkSecureKey`; never print credentials or enable them implicitly.
- X uses `GET https://api.x.com/2/tweets/search/recent`, bearer authentication, author expansion, and excludes retweets/replies. Access depends on the account's API entitlement.
- VK uses form-encoded POST to `https://api.vk.com/method/wall.get`, version 5.199. Keep the token out of URLs.
- Parser HTTP clients have 15-second timeouts. Parser errors reach the polling worker, which delays retry by 10 minutes; there is no in-parser VK sleep/retry loop.
- GORM remains `github.com/jinzhu/gorm` v1. A migration to `gorm.io/gorm` is a separate database compatibility task.
- `Source` and `Item` embed `gorm.Model` and soft-delete; `User` has explicit timestamps and no `DeletedAt`.
- Existing Compose files use PostgreSQL 10 and development-only trust authentication. Do not use them as production deployment templates or upgrade an existing database volume in place.
- Known existing limits: Telegram text is not split at its 4096-character limit; database initialization ignores the AutoMigrate result. Do not describe these paths as validated production behavior.

# >>> agentsmd:project >>>
## Project

- telenews — Go project
- Package manager: `go modules`

## Commands

```bash
go run .  # dev
go build ./...  # build
go test ./...  # test
go vet ./...  # lint
```
# <<< agentsmd:project <<<
# >>> agentsmd:conventions >>>
## Conventions

> Distilled from this project's source by the `agentsmd-analyze` skill.
> Run it to fill this block. Hand-edits belong OUTSIDE the managed block (they are preserved; this block is regenerated).
# <<< agentsmd:conventions <<<
