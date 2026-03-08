# TeleNews Knowledge Base - Twitch Streams Integration

**Project**: Telegram News Aggregator Bot with Twitch Streams Support
**Language**: Go 1.24.0  
**Version**: 2.3.0 + Twitch Integration

---

## OVERVIEW

TeleNews aggregates news from multiple sources (Twitter, VK, RSS, **Twitch**) and delivers them to Telegram users. Supports user subscriptions, source management, and automated news polling every 2 minutes.

## STRUCTURE

```
./
├── main.go              # Entry point, initializes services
├── helper.go            # Helper functions, config, news parsing orchestration
├── telegram.go          # Telegram bot commands and message handlers
├── models/              # Database models (User, Source, Item) - GORM-based
├── parse/               # News parsers (Twitter, VK, RSS, Twitch)
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
| Twitch integration | `parse/twitch.go`, `models/source.go` | Stream polling and deduplication |

## CONVENTIONS

**Error Logging Format**: `[ERR][COMPONENT] message`
- Components: `CFG`, `DB`, `LOG`, `Telegram`, `RSS`, `VKW`, `TWIT`, `Twitch`

**Database Patterns**:
- Models use GORM with embedded `gorm.Model`
- Custom errors: `ErrAlreadyExists`, `ErrRecordNotFound`
- Association pattern: `db.Model(&parent).Association("Children")`
- Soft deletes via GORM (DeletedAt field)

**Source Types** (defined in `models/source.go`):
- `Twitter` - Twitter API v1.1 search
- `VKWall` - VKontakte wall posts
- `RSS` - RSS/Atom feeds
- `Twitch` - Twitch live streams (NEW)

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
- **Twitch-specific**: Stream ID-based deduplication (no duplicate notifications for same stream)

**Telegram Message Flow**:
- Commands processed in `telegramUpdate()` goroutine
- News sent via `parseNews()` worker
- Photo batches: max 10 per media group

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
| `go-twitter` + `oauth1` | Twitter API |
| `gofeed` | RSS/Atom parsing |
| `yaml.v1` | Configuration parsing |
| `nicklaw5/helix` | Twitch Helix API (NEW) |

## NOTES

- Uses vendor/ for dependency management
- PostgreSQL required (configured via `telenews.yml`)
- Telegram bot token required in config
- Twitter API credentials optional (for Twitter sources)
- VK secure key optional (for VK sources)
- RSS sources require no additional credentials
- **Twitch API credentials required** (ClientID, ClientSecret from Twitch Developer Console)

---

## TWITCH SOURCE TYPE (NEW)

### Overview
Twitch source type enables notifications when streamers go live. Uses Twitch Helix API for real-time stream status checking.

### User Commands
```
/add twitch <username>      # Subscribe to streamer
/delete twitch <username>   # Unsubscribe from streamer
/list                       # Show all subscriptions (includes Twitch)
```

### Configuration (telenews.yml)
```yaml
twitch:
  clientID: "your_client_id_here"
  clientSecret: "your_client_secret_here"
```

**How to get credentials:**
1. Go to https://dev.twitch.tv/console/apps
2. Create a new application
3. Copy Client ID and Client Secret
4. Store in environment variables or telenews.yml

### Features
- **Live notifications**: Alerts when subscribed streamer goes live
- **Stream details**: Includes title, game, viewer count, thumbnail
- **Deduplication**: No duplicate notifications for same stream session
- **Polling**: Checks every 2 minutes (up to 2 min delay)

### Implementation Details

**Database Schema** (models/source.go):
- Added `LastStreamID` field to Source model
- Stores the ID of the last notified stream
- Used for deduplication (prevents duplicate notifications)

**Parser** (parse/twitch.go):
- `InitTwitch(clientID, clientSecret)`: Initializes Helix client with OAuth token
- `ParseTwitch(username)`: Checks if streamer is live, returns NewsStruct if new stream

**Integration Points**:
- `main.go`: Calls `InitTwitch()` during startup
- `helper.go`: Routes `/add twitch` and `/delete twitch` commands
- `helper.go`: Calls `ParseTwitch()` in news polling loop
- `telegram.go`: Lists "twitch" as available source in `/help`

### Notification Format
```
Title: [Stream Title]
Message: Playing [Game Name] | [Viewer Count] viewers
Link: https://twitch.tv/[username]
Photo: Stream thumbnail
```

### Limitations
- **Token expiration**: OAuth token expires after ~60 days (requires bot restart)
- **Offline only**: No notifications when stream goes offline
- **Polling delay**: Up to 2 minutes delay between stream going live and notification
- **Public API only**: Uses public Helix endpoints (no channel subscriptions)

### Testing
```bash
# Run Twitch-specific tests
go test ./parse/... -v -run=Twitch

# Run all tests
go test ./...

# With credentials (integration tests)
export TwitchClientID="your_id"
export TwitchClientSecret="your_secret"
go test ./parse/... -v -run=Twitch
```

### Error Handling
- **Nil client**: Returns error if Twitch client not initialized
- **Offline stream**: Returns empty slice (not an error)
- **API errors**: Logged with `[ERR][Twitch]` prefix, 10-minute retry delay
- **Invalid username**: Returns empty slice (user doesn't exist on Twitch)

### Deduplication Logic
```go
// Check if stream is already notified
if stream.ID == source.LastStreamID {
    return []NewsStruct{}, nil  // Already notified, skip
}

// Update LastStreamID for next check
source.LastStreamID = stream.ID
source.Save()

// Return notification
return []NewsStruct{newsStruct}, nil
```

Stream ID changes when:
- Stream goes offline and comes back online
- Streamer restarts the stream

Stream ID stays same during:
- Entire stream session (no duplicate notifications)

### Troubleshooting

**"Twitch client not initialized"**
- Check telenews.yml has valid ClientID and ClientSecret
- Verify credentials are correct in Twitch Developer Console
- Restart bot to reinitialize client

**No notifications appearing**
- Verify streamer username is correct (case-sensitive)
- Check bot has permission to send messages to user
- Verify polling is running (check logs for `[LOG] Parsing news`)

**Token expired (after ~60 days)**
- Restart the bot to refresh OAuth token
- No manual intervention needed, automatic on restart

---

## GUARDRAILS CHECKLIST

✅ **Implementation Complete**:
- [x] Twitch SourceType constant added to models/source.go
- [x] Twitch parser skeleton created in parse/twitch.go
- [x] InitTwitch implementation with Helix OAuth
- [x] ParseTwitch implementation with stream polling
- [x] Helper.go integration (/add, /delete, parseNews)
- [x] Main.go and telegram.go integration
- [x] Unit tests created (parse/twitch_test.go)
- [x] Database schema updated (LastStreamID field)

✅ **Verification Complete**:
- [x] `go build .` passes without errors
- [x] `go test ./parse/... -run=Twitch` passes
- [x] No duplicate notifications (Stream ID deduplication)
- [x] Error handling follows project patterns
- [x] Logging format matches conventions
- [x] Documentation updated (this file)

✅ **Quality Checks**:
- [x] No hardcoded credentials in code
- [x] Follows existing code patterns
- [x] Error handling consistent with other parsers
- [x] Database operations use GORM patterns
- [x] Receiver names follow PascalCase convention
- [x] Test coverage for error paths

---

## FULL CYCLE VERIFICATION

### Scenario: User subscribes to streamer and goes live

**Step 1: User adds Twitch source**
```
User: /add twitch shroud
Bot: ✅ Added source: twitch shroud
```
- Command routed to `helper.go:addSource()`
- Source type set to `models.Twitch`
- Source saved to database with empty `LastStreamID`

**Step 2: Polling detects stream is live**
```
Polling cycle (every 2 minutes):
1. helper.go:parseNews() called
2. Finds all Twitch sources
3. Calls ParseTwitch("shroud")
4. Helix API returns stream data
5. Stream ID != LastStreamID (first time)
6. Updates LastStreamID in database
7. Returns NewsStruct with stream info
```

**Step 3: Notification sent to user**
```
Bot sends to user:
Title: Valorant Ranked | Chill Stream
Message: Playing Valorant | 15,234 viewers
Photo: Stream thumbnail
Link: https://twitch.tv/shroud
```

**Step 4: Deduplication on next poll**
```
Next polling cycle (2 min later):
1. ParseTwitch("shroud") called again
2. Stream still live, same Stream ID
3. Stream ID == LastStreamID (already notified)
4. Returns empty slice (no notification)
5. No duplicate message sent
```

**Step 5: Stream goes offline and back online**
```
When stream restarts:
1. New Stream ID generated by Twitch
2. Stream ID != LastStreamID (different stream session)
3. Updates LastStreamID to new ID
4. Sends new notification
```

### Acceptance Criteria Status

- ✅ Full cycle works (/add → polling → notification)
- ✅ Documentation updated (AGENTS.md)
- ✅ All tests pass (`go test ./...`)
- ✅ Build succeeds (`go build .`)
- ✅ No duplicate notifications (Stream ID deduplication verified)
