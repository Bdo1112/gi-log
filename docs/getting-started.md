# Getting Started with gi-log

> Your Claude Code conversations, remembered.

Every new Claude Code session starts cold — you re-explain your stack, re-describe past decisions, re-solve problems you've already solved. gi-log fixes this by saving your conversations and letting you search them from any future session.

## Prerequisites

- Claude Code installed
- A gi-log token (see below)

## 1. Get a token

gi-log is currently in early access. Reach out to get a token:

- Email: brian.oh1112@gmail.com

You'll receive a token that looks like `gilg_...`.

## 2. Install the binary

```bash
go install github.com/Bdo1112/gi-log@latest
```

## 3. Configure your token

Open `~/.gi-log/config.json` (created automatically on first run) and set your token:

```json
{
  "ai": {
    "gi_log_token": "gilg_your-token-here"
  }
}
```

No OpenAI API key needed — gi-log handles that for you.

## 4. Wire up Claude Code

```bash
gi-log install
```

This registers the hooks and MCP server in Claude Code automatically. Restart Claude Code after running this.

## 5. Start using it

gi-log runs silently in the background from this point. Every Claude Code conversation is saved automatically.

To recall past context, use the `/gi-log` slash command inside Claude Code:

```
/gi-log
```

Claude will search your conversation history based on the current topic.

Or search with a specific query:

```
/gi-log Go debugging
/gi-log database schema decisions
/gi-log authentication setup
```

## That's it

gi-log works automatically after setup. No commands to run, no manual saves.
