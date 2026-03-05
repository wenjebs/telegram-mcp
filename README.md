[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# Telegram MCP server

A fork of [chaindead/telegram-mcp](https://github.com/chaindead/telegram-mcp) with additional tools for browsing contacts and groups.

Connects Telegram to AI assistants via the [Model Context Protocol](https://modelcontextprotocol.io), using MTProto directly (no bot token needed — this runs as your user account).

> [!IMPORTANT]
> Read the [Telegram API Terms of Service](https://core.telegram.org/api/terms) before using this. Misuse can get your account suspended.

## Tools

| Tool | Description |
|------|-------------|
| `tg_me` | Get your account info |
| `tg_dialogs` | List dialogs, filterable by type (`user`, `bot`, `chat`, `channel`) and unread status |
| `tg_users` | List users you can message, with fuzzy name/username search |
| `tg_groups` | List group chats (not channels), with fuzzy name search |
| `tg_dialog` | Get message history for a dialog |
| `tg_send` | Send a message |
| `tg_read` | Mark a dialog as read |

### tg_users / tg_groups search

Both tools support a `search` parameter that does case-insensitive matching against display name and username. Results are ranked: exact > prefix > substring > fuzzy (Levenshtein distance up to 3). The tools auto-paginate through your full dialog history, so contacts you haven't messaged recently still show up.

## Setup

### 1. Get Telegram API credentials

Go to [my.telegram.org/auth](https://my.telegram.org/auth) and create an app to get your `api_id` and `api_hash`.

### 2. Authenticate

```bash
telegram-mcp auth --app-id <your-api-id> --api-hash <your-api-hash> --phone <your-phone-number>
```

Add `--password <2fa_password>` if you have 2FA enabled. Add `--new` to replace an existing session.

This writes a session file to `~/.telegram-mcp/session.json`.

### 3. Install via Homebrew

```bash
brew install wenjebs/tap/telegram-mcp
```

### 4. Or build from source

```bash
git clone https://github.com/wenjebs/telegram-mcp
cd telegram-mcp
go build -o ./telegram-mcp .
```

Requires Go 1.24+.

### 5. Configure your MCP client

For Claude Desktop (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "telegram": {
      "command": "telegram-mcp",
      "env": {
        "TG_APP_ID": "<your-app-id>",
        "TG_API_HASH": "<your-api-hash>",
        "HOME": "<your-home-directory>"
      }
    }
  }
}
```

For Cursor:

```json
{
  "mcpServers": {
    "telegram-mcp": {
      "command": "telegram-mcp",
      "env": {
        "TG_APP_ID": "<your-app-id>",
        "TG_API_HASH": "<your-api-hash>"
      }
    }
  }
}
```

### JSON schema version

VS Code doesn't support JSON Schema Draft 2020-12. Use the `--schema-version` flag or `TG_SCHEMA_VERSION` env var to override:

```
TG_SCHEMA_VERSION=https://json-schema.org/draft-07/schema#
```
