# gi-log

> Your context. Everywhere.

AI sessions start cold. You re-explain your stack, re-describe past decisions, and re-solve problems you've already solved. gi-log gives your AI persistent, portable context that carries across sessions.

When context matters again, gi-log brings it back.

Currently supports Claude Code. Support for additional AI tools and agents is planned.

## How it works

- **On every response** — a hook captures the user message and assistant response, embeds them, extracts keywords, and stores them in SQLite
- **Every 5 exchanges** — a session summary is generated and stored alongside the raw exchanges for better recall accuracy
- **On recall** — an MCP tool searches past conversations using keyword matching and vector similarity, then injects the most relevant matches as context

## Storage

gi-log uses two storage modes depending on your config:

**gi-log token (recommended)** — exchanges are saved to your local SQLite database and synced to the gi-log cloud. The cloud copy enables future features like cross-device sync. All AI processing (embeddings, summaries, entity extraction) is handled by the gi-log API — no OpenAI account needed.

**OpenAI API key** — exchanges are saved to local SQLite only. AI processing goes directly to OpenAI using your key.

**Cost with OpenAI key:** gi-log makes two OpenAI API calls per exchange:
- `text-embedding-3-small` — $0.02 per million tokens
- `gpt-4o-mini` — $0.15 per million input tokens, $0.60 per million output tokens

For typical usage, saving thousands of conversations costs only a few dollars.

## Requirements

- Go 1.21+
- A gi-log token **or** your own OpenAI API key

→ See [Getting Started](docs/getting-started.md) for the quickest setup path.

## Install

```bash
go install github.com/Bdo1112/gi-log@latest
gi-log install
```

This registers the hooks in `~/.claude/settings.json`, adds the MCP server to `~/.claude.json`, installs the `/gi-log` slash command, and ensures your config has the correct defaults.

Then configure `~/.gi-log/config.json`:

**Option A — gi-log token (no OpenAI account needed):**
```json
{
  "ai": {
    "gi_log_token": "gilg_..."
  }
}
```

**Option B — your own OpenAI API key:**
```json
{
  "ai": {
    "api_key": "sk-..."
  }
}
```

Restart Claude Code after install.

## Usage

gi-log works automatically in the background. Conversations are saved as you go.

### Recalling past conversations

Use the `/gi-log` slash command to search past conversations:

```
/gi-log
```

Claude will extract the current topic and search your conversation history automatically.

Or search with a specific query:

```
/gi-log Go debugging Delve
```

```
/gi-log database schema decisions
```

## Config reference

| Field | Default | Description |
|---|---|---|
| `ai.gi_log_token` | — | gi-log token (use instead of api_key) |
| `ai.api_key` | — | OpenAI API key (use instead of gi_log_token) |
| `ai.embedding_model` | `text-embedding-3-small` | Embedding model (OpenAI mode only) |
| `ai.extraction_model` | `gpt-4o-mini` | Extraction/summarization model (OpenAI mode only) |
| `server.api_url` | `https://gi-log-api-production.up.railway.app` | gi-log API URL (set automatically) |
| `db.path` | `~/.gi-log/gi_log.db` | Path to the local SQLite database |
| `search.top_k` | `5` | Number of results returned per recall |

## Data

**Local** (`~/.gi-log/`):
```
~/.gi-log/
  config.json     — your config
  gi_log.db       — SQLite database (all exchanges and summaries)
  errors/         — error logs
```

**Cloud** (gi-log token mode only): exchanges and session summaries are synced to the gi-log API. This data is associated with your token and is not shared with other users.
