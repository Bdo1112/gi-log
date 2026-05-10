# gi-log

> Your context. Everywhere.

## Why gi-log exists

AI tools are powerful, but their context is fragmented.

Your workflows, decisions, and conversations become trapped inside isolated sessions, apps, and providers. Switching tools often means starting over.

Static notes and `memory.md` files help, but they still require manual organization and retrieval.

gi-log makes context portable by automatically capturing conversations, indexing them semantically, and bringing relevant context back when it matters again.

All conversations and search indexes are stored locally on your machine.

Currently supports Claude Code. Support for additional AI tools and agents is planned.

## Use cases

gi-log works especially well for:

- engineering projects that span weeks or months
- debugging sessions that require historical context
- architecture discussions and technical decision tracking
- AI-assisted workflows across multiple tools and sessions
- developers who switch frequently between AI providers

## How it works

- **On every response** — a hook captures the user message and assistant response, embeds them via OpenAI, and stores them in a local SQLite database
- **On recall** — an MCP tool searches past conversations using cosine similarity and injects the most relevant matches as context


## Requirements

- Go 1.21+
- OpenAI API key

## Install

```bash
go install github.com/Bdo1112/gi-log@latest
gi-log install
```

This registers the hooks in `~/.claude/settings.json`, adds the MCP server to `~/.claude.json`, and installs the `/gi-log` slash command.

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


**Cost:** gi-log makes two OpenAI API calls per conversation exchange:
- `text-embedding-3-small` for embeddings — $0.02 per million tokens
- `gpt-4o-mini` for entity extraction — $0.15 per million input tokens, $0.60 per million output tokens

For typical usage, saving thousands of conversations costs only a few dollars.

## Data

All data is stored locally at `~/.gi-log/`:

```
~/.gi-log/
  config.json     — your config
  gi_log.db       — SQLite database
  errors/         — error logs
  tmp/            — temporary files during hook execution
```
