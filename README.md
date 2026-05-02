# gi-log

gi-log gives Claude Code a long-term memory. Every conversation you have is automatically saved and embedded. When you reference something from a past session, Claude recalls it — no re-explaining needed.

## How it works

- **On every response** — a hook captures the user message and assistant response, embeds them via OpenAI, and stores them in a local SQLite database
- **On recall** — an MCP tool searches past conversations using cosine similarity and injects the most relevant matches as context

Everything runs locally. The only external call is to OpenAI for embeddings.

## Requirements

- Go 1.21+
- OpenAI API key

## Install

```bash
git clone <repo>
cd gi-log
make install
```

This builds the binary, moves it to `~/.local/bin/gi-log`, registers the hooks in `~/.claude/settings.json`, and adds the MCP server to `~/.claude.json`.

Then set your OpenAI API key:

```bash
# ~/.gi-log/config.json is created automatically on first run
# open it and fill in your key
```

```json
{
  "embedding": {
    "api_key": "sk-...",
    "model": "text-embedding-3-small"
  },
  "db": {
    "path": "~/.gi-log/gi_log.db"
  },
  "search": {
    "top_k": 5
  }
}
```

Restart Claude Code after install.

## Usage

gi-log works automatically in the background. Conversations are saved as you go.

To trigger recall, just reference a past conversation naturally:

> "I remember we talked about how to handle the database schema"

> "We discussed this before — how did we handle auth?"

Claude will call the `recall` tool, search your past conversations, and use the results as context.

## Config reference

| Field | Default | Description |
|---|---|---|
| `embedding.api_key` | — | Your OpenAI API key (required) |
| `embedding.model` | `text-embedding-3-small` | OpenAI embedding model |
| `db.path` | `~/.gi-log/gi_log.db` | Path to the SQLite database |
| `search.top_k` | `5` | Number of results returned per recall |

## Data

All data is stored locally at `~/.gi-log/`:

```
~/.gi-log/
  config.json     — your config
  gi_log.db       — SQLite database
  errors/         — error logs
  tmp/            — temporary files during hook execution
```
