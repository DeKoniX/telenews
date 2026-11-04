# Parse Package

**Purpose**: News parsers for multiple source types (Twitter, VK, RSS)

## OVERVIEW

Extracts news from external APIs and feeds. Each parser converts source-specific data into standardized `NewsStruct`. Supports retry logic for rate-limited sources.

## STRUCTURE

```
parse/
├── parse.go        # Core types and hash generation
├── twitter.go      # X API v2 recent-search parser
├── vk.go           # VKontakte wall parser
├── rss.go          # RSS/Atom feed parser
└── parse_test.go   # Unit tests
```

## WHERE TO LOOK

| Task | File | Function |
|------|------|----------|
| Add new source type | `parse.go` | Extend `ParseNewsStruct` |
| Fix Twitter parsing | `twitter.go` | `ParseTwitter()` |
| Fix VK parsing | `vk.go` | `ParseVKWall()` |
| Fix RSS parsing | `rss.go` | `ParseRSS()` |
| Change hash algorithm | `parse.go` | `GenHash()` |
| Modify retry logic | `vk.go` | `parseVKWall()` and the worker in `../helper.go` |

## CONVENTIONS

**Parser Pattern**:
```go
// Initialization (called once in main)
func (p *ParseNewsStruct) InitXXX(credentials...)

// Parsing (called per source)
func (p ParseNewsStruct) ParseXXX(query string) ([]NewsStruct, error)
```

**NewsStruct Fields**:
- `URL` - Source identifier
- `Title` - News headline
- `MSG` - Content/text
- `Link` - Permalink to original
- `PhotoURL []string` - Attached images
- `Hash` - MD5 deduplication key

**Error Handling**:
- Return `error` for failures (caller logs with `[ERR][SOURCE]`)
- Empty slice + nil error = no new items
- Source-specific errors wrapped with context

**Rate Limiting**:
- VK: API errors return immediately; the polling worker delays retry by 10 minutes.
- Twitter: No built-in retry (relies on 2-minute polling interval)
- RSS: No rate limiting

## ANTI-PATTERNS

- **DO NOT** panic on parse errors (return error)
- **DO NOT** log inside parsers (return to caller)
- **AVOID** hardcoded limits (use constants)
- **NEVER** ignore HTTP response body close (`defer resp.Body.Close()`)

## NOTES

- The worker passes the source ID to `NewsStruct.GenHash`; the parser has no SourceID field.
- X excludes retweets and replies through `-is:retweet -is:reply` in the query.
- VK uses API version 5.199, with the access token in the POST form.
- RSS uses Firefox User-Agent to avoid blocks
- Photo selection: VK picks highest resolution image
