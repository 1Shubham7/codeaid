# codeaid

A TUI based AI coding agent built in Go. codeaid runs an autonomous agentic loop powered by Claude, with tools for file I/O, code execution, directory traversal etc. (with safety guardrails) - iterating until the task is complete and streams output tokens. Built with BubbleTea's Elm architecture.

![CI](https://github.com/1shubham7/codeaid/actions/workflows/ci.yml/badge.svg)

---

## Stack

- **[Anthropic Go SDK](https://github.com/anthropics/anthropic-sdk-go)** - streaming Claude API
- **[BubbleTea](https://github.com/charmbracelet/bubbletea)** - Elm-architecture TUI framework
- **[Bubbles](https://github.com/charmbracelet/bubbles)** - spinner, text input components
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - terminal styling
- **[Cobra](https://github.com/spf13/cobra)** - CLI framework
- **[godotenv](https://github.com/joho/godotenv)** - `.env` file loading
- **[GoReleaser](https://goreleaser.com)** - cross-platform binary builds and GitHub releases on tag push

**Key features:**

- **Autonomous agentic loop** - Claude runs in an agentic loop, calling tools and feeding results back into the conversation until it hits `end_turn`. It lists the directory, reads the relevant files, writes the fix, runs it, and verifies the output on its own.
- **Channel-based streaming** - `agent.CallAPI` runs in a goroutine and pushes text chunks to a `chan string` and per-iteration stats to a `chan IterationMsg`. The BubbleTea TUI re-schedules a `tea.Cmd` on every channel receive, keeping the Elm event loop non-blocking while the response streams in live.
- **Manual stream accumulation** - The Anthropic SDK emits raw SSE events (`ContentBlockStartEvent`, `ContentBlockDeltaEvent`, `MessageDeltaEvent`) with no built-in accumulator. Blocks are collected in a `map[int64]*blockAccum` keyed by stream index, sorted by index before processing to guarantee correct ordering across concurrent text and tool-use blocks.
- **Safety guardrails** - `execute_code` extracts the first token of every command via `strings.Fields` and checks it against a configurable blocklist stored in `~/.codeaid/config.json`. File reads are similarly gated by a configurable size limit - both defaults are seeded into config on first run and are fully user-editable.
- **Persistent conversation history** - Messages are serialized as `[]anthropic.MessageParam` to `~/.codeaid/history.json` after each response. On the next session, history is reloaded and prepended to the context window, giving Claude memory across restarts.
- **Tool-use architecture** - Tools are defined as Anthropic `ToolParam` schemas and registered in a central `Dispatch` router. Claude decides which tool to call, the router executes it, and the result is injected back as a `ToolResultBlock` for the next API call.
- **Structured logging** - `log/slog` with `slog.NewJSONHandler` writes JSON logs to `~/.codeaid/logs/codeaid.log`, opened in `O_CREATE|O_APPEND|O_WRONLY` mode at startup before any other initialization runs. Every tool call, API request, token count, and error is captured with structured fields.

---

## Installation

### Download a binary (recommended)

Grab the latest release for your platform from the [releases page](https://github.com/1shubham7/codeaid/releases):

```bash
# macOS (Apple Silicon)
curl -L https://github.com/1shubham7/codeaid/releases/latest/download/codeaid-<version>-darwin-arm64.tar.gz | tar xz
sudo mv codeaid /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/1shubham7/codeaid/releases/latest/download/codeaid-<version>-darwin-amd64.tar.gz | tar xz
sudo mv codeaid /usr/local/bin/

# Linux
curl -L https://github.com/1shubham7/codeaid/releases/latest/download/codeaid-<version>-linux-amd64.tar.gz | tar xz
sudo mv codeaid /usr/local/bin/
```

### Build from source

Requires Go 1.21+.

```bash
git clone https://github.com/1shubham7/codeaid.git
cd codeaid
go build -o codeaid .
sudo mv codeaid /usr/local/bin/
```

---

## Configuration

codeaid stores all configuration under `~/.codeaid/`.

```
~/.codeaid/
├── config.json       # model, restricted commands, file size limit
├── history.json      # conversation history (auto-managed)
└── logs/
    └── codeaid.log   # structured JSON logs
```

### API key

Set your Anthropic API key before running:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

Or pass it directly:

```bash
codeaid --api-key sk-ant-...
```

You can also put it in a `.env` file in the directory where you run codeaid:

```
ANTHROPIC_API_KEY=sk-ant-...
```

### config.json

Created automatically on first run at `~/.codeaid/config.json` with defaults:

```json
{
  "model": "claude-haiku-4-5-20251001",
  "restricted_commands": [
    "rm", "rmdir", "curl", "wget", "ssh", "sudo",
    "chmod", "chown", "dd", "mkfs", "shutdown", "reboot",
    "kill", "pkill", "iptables", "nc", "ncat", "netcat",
    "passwd", "fdisk", "crontab", "at", "nohup"
  ],
  "max_file_size_kb": 100
}
```

| Field | Description |
|---|---|
| `model` | Claude model to use. Overridden by `--model` flag. |
| `restricted_commands` | Commands Claude is never allowed to execute. Matched against the first token of any shell command. |
| `max_file_size_kb` | Maximum file size (in KB) that `read_file` will read. Files larger than this are blocked. |

**To allow a restricted command** - for example, if you want Claude to be able to use `curl`:

```json
{
  "restricted_commands": ["rm", "rmdir", "ssh", "sudo", ...]
}
```

Just remove it from the list and save the file.

---

## Usage

```bash
codeaid
```

### Keyboard shortcuts

| Key | Action |
|---|---|
| `↑` / `↓` | Navigate menu |
| `1`–`4` | Menu shortcuts |
| `Enter` | Confirm |
| `Esc` | Back to menu |
| `Ctrl+C` | Quit |

### In-session commands

| Command | Action |
|---|---|
| `clear` | Clear the screen |
| `clear history` | Delete conversation history, start a new session |
| `exit` | Return to menu |

## Demo

1. run the binary locally

<img width="905" height="505" alt="image" src="https://github.com/user-attachments/assets/c4885674-93a4-4dee-acb8-fd5e371ed694" />

2. select the model you want to use

<img width="905" height="505" alt="image" src="https://github.com/user-attachments/assets/3aec1d48-511f-4aa4-a034-fc5d4a9b82e6" />

3. that's it. you can now go to code and enter what you want to do

<img width="1544" height="835" alt="image" src="https://github.com/user-attachments/assets/21e72ad3-e710-4f4b-ba05-e69c3da6feba" />

4. and you will see, codeaid created a hello-world.py and executed it as well. this is just an example, you can use a capable model and perform your daily coding tasks with codeaid as well.

<img width="1544" height="835" alt="image" src="https://github.com/user-attachments/assets/9132d0c8-254e-4d4e-b7fb-1a4f02cff2bf" />

## Tools available to Claude

| Tool | Description |
|---|---|
| `read_file` | Reads a file - blocked if it exceeds the configured size limit |
| `write_file` | Writes a file, creating any missing parent directories |
| `list_directory` | Lists directories and files separately at a given path |
| `execute_code` | Runs a shell command via `sh -c` with a 30s timeout, captures stdout/stderr |
| `get_current_time` | Returns current time in any IANA timezone |

---

### CLI flags

```bash
codeaid --api-key sk-ant-...        # set API key inline
codeaid --model claude-sonnet-4-6   # override model
```

### Switching models

From the main menu, select **Model** to switch between:

- `claude-haiku-4-5` - fast, low cost
- `claude-sonnet-4-6` - balanced
- `claude-opus-4-8` - most capable

The selected model is saved to `config.json` and used on the next session.

---

## Development

```bash
# run tests
go test ./...

# run with live reload (requires air)
air

# dry-run a release locally (requires goreleaser)
goreleaser release --snapshot --clean
```

### Logs

All tool calls, API requests, and errors are logged as JSON to `~/.codeaid/logs/codeaid.log`:

```bash
tail -f ~/.codeaid/logs/codeaid.log | jq
```
