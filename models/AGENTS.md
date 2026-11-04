# Models Package

**Purpose**: Database layer for TeleNews using GORM

## OVERVIEW

Manages three entities: `User`, `Source`, `Item`. Uses PostgreSQL with GORM ORM. Supports soft deletes, associations, and custom error handling.

## STRUCTURE

```
models/
├── models.go       # DB initialization and global vars
├── user.go         # User entity (Telegram chat users)
├── source.go       # Source entity (RSS/Twitter/VK subscriptions)
├── item.go         # Item entity (individual news items)
└── *_test.go       # Unit tests
```

## WHERE TO LOOK

| Task | File | Function |
|------|------|----------|
| Add user field | `user.go` | Modify `User` struct |
| Add source type | `source.go` | Add to `SourceType` const block |
| Query items by source | `item.go` | `item.Select(source)` |
| Cascading delete | `source.go`, `user.go` | `Delete()` methods |
| DB connection | `models.go` | `Init()` function |
| Add custom error | `models.go` | Add to `var (...)` block |

## CONVENTIONS

**GORM Patterns**:
- `Source` and `Item` embed `gorm.Model`; `User` has explicit ID/timestamps and no soft deletion.
- Use tags for constraints: `` `gorm:"unique;not null"` ``
- Associations via slice fields: `Items []Item`

**Method Patterns**:
```go
// Query methods - value receiver
func (m Model) Select(...) (results, error)

// Modify methods - pointer receiver
func (m *Model) Insert(...) (count, error)
func (m *Model) Save() error
func (m *Model) Delete() error
```

**Error Handling**:
- Return `models.ErrRecordNotFound` for "not found" errors
- Return `models.ErrAlreadyExists` for duplicates
- Always check `err != nil` before using results

**Cascading Deletes**:
- Manual implementation in `Delete()` methods
- Load associations first, then delete children
- Finally delete parent

## ANTI-PATTERNS

- **DO NOT** use `db.Create()` directly (use model methods)
- **DO NOT** ignore GORM errors (always check return)
- **AVOID** raw SQL (use GORM query builder)
- **NEVER** delete without checking associations

## NOTES

- Auto-migration runs on `models.Init()`
- Database logging controlled by `db.LogMode(false)` in models.go
- The current connection builder disables PostgreSQL TLS; production TLS support is a separate configuration change.
- Sources/items soft-delete through `deleted_at`; deleting a user physically removes its row.
- Integration tests require an explicitly configured disposable PostgreSQL database; see the root AGENTS.md.
