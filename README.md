# OllamaTUI

This repository builds a Go-based CLI that provides a **terminal UI (TUI)** for chatting with an [Ollama](https://ollama.com/) instance. It’s built with [Charm](https://charm.sh/)’s Bubble Tea stack and focuses on fast streaming responses, keyboard-first navigation, and quality-of-life features like in-TUI search and clipboard copying.

## Features

- **Streaming chat** via the official Ollama Go API client (`github.com/ollama/ollama/api`).
- **TUI interface** (Bubble Tea) with Markdown rendering (Glamour).
- **Conversation search** (`Ctrl+F`) with match navigation and highlighting.
- **Clipboard support**
  - `Ctrl+Y`: copy the last assistant message
  - `Ctrl+Shift+C`: copy the full conversation
- **Theme toggle** (`Ctrl+T`) between light/dark.
- **Multi-line input** with input history
  - `Enter`: submit
  - `Alt+Enter` / `Ctrl+J`: newline
  - `↑/↓` or `Ctrl+P/Ctrl+N`: navigate input history
- **Model listing**: `ollama-go list` prints locally available models from your Ollama instance.

## Prerequisites

- **Ollama installed and running** (default: `http://localhost:11434`).
  - Start it with `ollama serve`
  - Pull a model with `ollama pull <model>`
- **Go** (the module declares `go 1.26.1` in `go.mod`).

## Install

### Option A: build/install from source (recommended)

```bash
git clone https://github.com/gagan-devv/OllamaTUI.git
cd OllamaTUI
make build
./ollama-go
```

To install into your `GOPATH/bin`:

```bash
make install
```

### Option B: `go install <module>@latest` (when published)

This project’s module path is defined in `go.mod`. If you publish a tagged release for that module path, you can install directly with:

```bash
go install github.com/gagan-devv/ollama-go@latest
```

## Usage

### Start the TUI chat

```bash
ollama-go
```

### Pick a model / system prompt

```bash
ollama-go --model qwen3.5:4b --system "You are a concise assistant."
```

### List downloaded models

```bash
ollama-go list
```

## Keybindings (TUI)

Global:

- `Ctrl+C`: quit
- `Ctrl+T`: toggle theme
- `Ctrl+F`: search conversation
- `Esc`: clear current input

Clipboard:

- `Ctrl+Y`: copy last assistant message
- `Ctrl+Shift+C`: copy full conversation

Search mode (`Ctrl+F`):

- `Enter` / `Ctrl+F`: search
- `Ctrl+N` / `F3`: next match
- `Ctrl+P` / `Shift+F3`: previous match
- `Ctrl+I`: toggle case sensitivity
- `Esc`: exit search

Multi-line input:

- `Enter`: submit
- `Alt+Enter` / `Ctrl+J`: insert newline
- `↑/↓` (or `Ctrl+P/Ctrl+N`): input history

## Configuration

On first run, `ollama-go` creates a config file (if missing) at:

- `~/.ollama-go/config.yaml`

This file controls UI preferences, output paths, retries/timeouts, and other defaults. The schema is defined in `internal/config/config.go`.

### Default config (generated)

```yaml
model:
  default: qwen3.5:4b
  system_prompt: You are a helpful assistant.
  parameters:
    temperature: 0.7
    top_p: 0.9
    top_k: 40

ui:
  theme: dark
  vim_mode: false
  show_metrics: true
  colors:
    user_message: "#5FAFD7"
    ai_message: "#FFD787"
    background: "#1E1E1E"
    border: "#3E3E3E"
    status_bar: "#2E2E2E"

paths:
  test_folder: test_output
  config_dir: "~/.ollama-go"
  plugin_dir: "~/.ollama-go/plugins"

behavior:
  auto_save: true
  auto_save_interval: 5s
  cache_enabled: true
  cache_size_mb: 100
  confirm_destructive: true

network:
  retry_count: 3
  retry_delay: 2s
  timeout: 30s

keybindings:
  quit: ctrl+c
  submit: enter
  multiline: shift+enter
  search: ctrl+f
  theme_toggle: ctrl+t
  help: f1
  copy: ctrl+y

advanced:
  debug_mode: false
  log_level: info
  max_history: 1000
  memory_limit_mb: 500
```

Notes:

- Duration values use Go-style duration strings (for example: `250ms`, `5s`, `2m`).
- The TUI implements several keybindings directly (see `internal/ui/tui.go`), so changing `keybindings` in the config may not affect all shortcuts yet.

## Data & output files

By default, the app uses `test_output/` as a workspace for local artifacts (this folder is gitignored):

- `test_output/history/`: chat/session data (and input history)
- `test_output/exports/`: exported sessions (utility code exists in `internal/storage/export.go`)
- `test_output/logs/`: log files (see `internal/util/logger.go`)

## Environment variables

The Ollama client is created via `api.ClientFromEnvironment()`, so standard Ollama environment variables apply. Common ones include:

- `OLLAMA_HOST` (example: `http://localhost:11434`)

## Development

```bash
make build
make test
make test-coverage
```

Repository layout:

- `cmd/`: Cobra CLI commands (`ollama-go`, `ollama-go list`)
- `internal/ui/`: Bubble Tea TUI + components (search, clipboard, status bar)
- `internal/config/`: config loading/validation and default generation
- `internal/client/`: streaming chat handler + retry policy utilities
- `internal/session/`, `internal/storage/`, `internal/performance/`: session persistence and performance utilities (foundations for richer session features)

## Troubleshooting

### “Error connecting to Ollama…”

1. Ensure Ollama is running: `ollama serve`
2. Verify the host/port (default is `http://localhost:11434`)
3. If Ollama is remote, set `OLLAMA_HOST` (for example: `export OLLAMA_HOST=http://server:11434`)

### Model not found

- Pull it first: `ollama pull <model>`
- Confirm what’s installed: `ollama-go list`

## License

See `LICENSE`.
