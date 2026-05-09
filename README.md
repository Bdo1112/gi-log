# gi-log

Every new Claude Code session starts cold. You re-explain your stack, re-describe past decisions, re-solve problems you've already solved. gi-log fixes this by automatically saving every conversation to a local database with semantic search. When you reference something from a past session, Claude recalls it — no re-explaining needed.

**Cost:** gi-log uses OpenAI `text-embedding-3-small` for embeddings — one call per conversation exchange. At $0.02 per million tokens, saving thousands of conversations costs less than $1.

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

This builds the binary, moves it to `~/.local/bin/gi-log`, registers the hooks in `~/.claude/settings.json`, adds the MCP server to `~/.claude.json`, and installs the `/gi-log` slash command.

Then set your OpenAI API key:

```bash
# ~/.gi-log/config.json is created automatically on first run
# open it and fill in your key
```

```json
{
  "ai": {
    "api_key": "sk-...",
    "embedding_model": "text-embedding-3-small",
    "extraction_model": "gpt-4o-mini"
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

### How saving works

On every Claude response, gi-log automatically:
1. Captures your message and Claude's response
2. Generates an embedding vector via OpenAI
3. Extracts key entities from the conversation
4. Stores everything in a local SQLite database

## Config reference

| Field | Default | Description |
|---|---|---|
| `ai.api_key` | — | Your OpenAI API key (required) |
| `ai.embedding_model` | `text-embedding-3-small` | OpenAI embedding model |
| `ai.extraction_model` | `gpt-4o-mini` | OpenAI model for entity extraction |
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
