# @delvop/cli

**Engineering Command Center for Terminal Coding Agents**

Manage a team of AI coding agents (Claude Code, Codex CLI, Gemini CLI) from a single terminal dashboard. One view for all your agents, one place to approve permissions, one budget to track.

## Install

```bash
npm install -g @delvop/cli
```

### Prerequisites

- **tmux** -- `brew install tmux` (macOS) or `apt install tmux` (Linux)
- **Node.js 16+**

## Usage

```bash
delvop
```

| Key | Action |
|-----|--------|
| `n` | New agent |
| `enter` | Focus agent / drop into terminal |
| `y` / `N` | Approve / deny permissions |
| `m` | Message agent |
| `?` | Help |
| `q` | Quit (agents keep running) |

## Features

- **Unified dashboard** -- All agents, one view. State, cost, tokens, files changed.
- **Agent-agnostic** -- Claude Code, Codex CLI, Gemini CLI, and more.
- **Permission queue** -- Approve/deny from one place, no terminal switching.
- **Native notifications** -- Desktop alerts with sound when agents need you.
- **Templates** -- Deploy pre-configured agent teams from TOML files.
- **Cost tracking** -- Per-session cost and token usage with SQLite persistence.

## How It Works

This npm package downloads the platform-specific `delvop` binary from [GitHub Releases](https://github.com/delvop-dev/command-center/releases) on install. The binary is a single Go executable that uses tmux for session isolation.

No daemon. No Electron. No background processes.

## Configuration

Optional `~/.delvop/config.toml`:

```toml
[general]
default_provider = "claude"
default_model = "opus"

[notify]
channels = ["native", "sound"]
```

## Links

- [GitHub](https://github.com/delvop-dev/command-center)
- [Full documentation](https://github.com/delvop-dev/command-center#readme)
- [Issues](https://github.com/delvop-dev/command-center/issues)

## License

MIT
