# Getting Started with gi-log

> Your Claude Code conversations, remembered.

Every new Claude Code session starts cold — you re-explain your stack, re-describe past decisions, re-solve problems you've already solved. gi-log fixes this by saving your conversations and letting you search them from any future session.

## Prerequisites

- Go 1.21+
- Claude Code installed
- A gi-log token (see step 1)

## 1. Get a token

gi-log is currently in early access. Reach out to get a token:

- Email: brian.oh1112@gmail.com

You'll receive a token that looks like `gilg_...`. No OpenAI account needed — gi-log handles AI processing for you.

## 2. Install and set up

```bash
go install github.com/Bdo1112/gi-log@latest
gi-log install
```

`gi-log install` registers the hooks and MCP server in Claude Code and creates your config at `~/.gi-log/config.json`.

## 3. Add your token

Open `~/.gi-log/config.json` and set your token:

```json
{
  "ai": {
    "gi_log_token": "gilg_your-token-here"
  }
}
```

## 4. Restart Claude Code

Hooks and the MCP server only take effect after a restart.

## 5. Verify it's working

```bash
gi-log status
```

You should see:
- `mode: gi-log token` — token is configured
- `reachable: YES` — API is reachable
- `UserPromptSubmit: registered` and `Stop: registered` — hooks are wired up

## 6. Start using it

gi-log runs silently in the background. After each Claude response, your conversation is automatically saved and indexed.

To recall past context, use the `/gi-log` slash command inside Claude Code:

```
/gi-log
```

Claude will search your conversation history based on the current topic. Or search with a specific query:

```
/gi-log Go debugging
/gi-log database schema decisions
/gi-log authentication setup
```

## Troubleshooting

**`gi-log status` shows `exchanges: 0` after a few conversations**
Make sure you restarted Claude Code after running `gi-log install`. The hooks only activate after a restart.

**`reachable: NO`**
Check your internet connection. If the issue persists, email brian.oh1112@gmail.com.

**Hook not registered**
Run `gi-log install` again, then restart Claude Code.
