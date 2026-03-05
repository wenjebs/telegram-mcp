# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
# Build binary
go build -o ./bin/telegram-mcp .

# Run directly (requires TG_APP_ID and TG_API_HASH env vars or flags)
go run ./. --app-id <id> --api-hash <hash>

# Dry-run smoke test (exercises all tools against real Telegram session)
go run ./. --app-id <id> --api-hash <hash> --dry

# Lint
golangci-lint run --fix ./...
```

No test suite exists yet. Verification is done via the `--dry` flag.

After rebuilding, **restart the MCP client** (e.g. Claude Desktop) to reload the binary — it does not hot-reload.

## Release

Releases are automated via GitHub Actions (`.github/workflows/release.yml`). To publish a new release:

```bash
git tag v0.x.0 && git push origin v0.x.0
```

Pushing a tag triggers goreleaser on CI, which builds cross-platform binaries, creates a GitHub release, and updates the Homebrew tap. No manual `goreleaser` invocation needed.

## Architecture

This is a Go MCP server that bridges the Telegram MTProto API to AI assistants via stdio transport.

**Entry points:**
- `main.go` — CLI setup (`urfave/cli`), two subcommands: `auth` and the default `serve`
- `auth.go` — interactive auth flow, writes session to `~/.telegram-mcp/session.json`
- `serve.go` — creates the MCP server, instantiates `tg.Client`, registers all tools, starts stdio transport

**All tool logic lives in `internal/tg/`:**

| File | Responsibility |
|------|---------------|
| `client.go` | `Client` struct; `T()` creates a fresh `telegram.Client` per call (stateless, `NoUpdates: true`) |
| `dialogs.go` | `GetDialogs` tool — fetches dialog list, filters by `OnlyUnread` and/or `Type` (`user/bot/chat/channel`) |
| `dialogs_offset.go` | Pagination cursor serialized as `"type-id-msgid-date"` string |
| `users.go` | `GetUsers` tool — users-only subset of dialogs with similarity search (exact > prefix > contains > token > fuzzy scoring); fetches all dialogs when a search query is provided |
| `history.go` | `GetHistory` tool — fetches message history; resolves peer names via `getInputPeerFromName` |
| `me.go` | `GetMe` tool |
| `read.go` | `ReadHistory` tool — marks dialog as read |
| `draft.go` | `SendDraft` tool — sends a message |
| `helpers.go` | `getTitle`, `getUsername`, `cleanJSON` utilities |

**Key patterns:**

- Every tool method on `Client` follows `func (c *Client) FooBar(args FooBarArguments) (*mcp.ToolResponse, error)` — the signature MCP-golang requires for tool registration.
- Each tool method calls `c.T().Run(ctx, func(...) error { ... })` to open a fresh Telegram connection. There is no persistent connection.
- Peer name resolution: usernames resolve via `message.NewSender(api).Resolve(name)`; groups use `cht[<id>]`; channels use `chn[<id>:<access_hash>]`. These identifiers come from `getUsername()` in `helpers.go` and are what tools like `tg_dialog`/`tg_send`/`tg_read` accept as the `name` field.
- `DialogType` constants (`user`, `bot`, `chat`, `channel`) are defined in `dialogs.go` and used for filtering in both `GetDialogs` and `GetUsers`.

**Adding a new tool:**
1. Create `internal/tg/<name>.go` with `type <Name>Arguments struct` and `func (c *Client) <Name>(args <Name>Arguments) (*mcp.ToolResponse, error)`
2. Register it in `serve.go` with `server.RegisterTool("tg_<name>", "<description>", client.<Name>)`
