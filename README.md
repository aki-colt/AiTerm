# AiTerm

Say hello to `AiTerm`, your new best friend for chatting with AI right from the terminal! This nifty command-line tool lets you toss natural language commands like "list files" into a slick text-based interface, powered by `tview`. Watch as the AI whips up the perfect terminal command (like `ls`) and runs it in a `tmux` split pane, all while keeping your conversation history scrollable and snappy. Whether you're a terminal ninja or just love a good chat, `AiTerm` makes AI-powered workflows a breeze!

![show](./docs/show.gif)

## Features

- **Natural Language Interaction**: Enter commands in natural language (e.g., "list files"), and the AI generates corresponding terminal commands (e.g., `ls`) and explanations.
- **tmux Integration**: Commands are executed in a dynamically created `tmux` pane, displayed alongside the chat interface.
- **Scrollable Chat History**: View past conversations in a scrollable `TextView`, with support for manual scrolling to review history.
- **Dynamic "Generating" Indicator**: A visual animation in the input field ("generating.", "generating..", "generating...") shows when the AI is processing.
- **Cancel Generation**: Press `Ctrl+C` during AI processing to cancel and resume input.
- **Configuration Management**: Stores AI service details (URL, Token, Model) in `~/.aitermrc`, with interactive terminal-based setup on first run.
- **Version Information**: Use `aiterm -v` to display the tool's version.
- **Error Handling**: Validates AI configuration and handles `tmux` session requirements gracefully.

## Installation

Install `AiTerm` using Go:

```bash
go install github.com/aki-colt/aiterm@latest
```

Ensure `$GOPATH/bin` is in your `PATH`:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Requirements

- **Go**: 1.22 or later
- **tmux**: Required for pane splitting and command execution
- **Terminal**: A terminal emulator supporting ANSI colors and keyboard input

## Quick Start

1. **Start a tmux session**:
   ```bash
   tmux
   ```

2. **Run the tool**:
   ```bash
   aiterm
   ```

3. **Configure AI (first run only)**:
   If `~/.aitermrc` is missing or incomplete, you'll be prompted to enter:
   ```
   Enter AI URL: https://api.example.com/v1
   Enter API Token: your-token
   Enter Model: gpt-4
   ```
   This creates `~/.aitermrc`:
   ```
   [ai]
   url=https://api.example.com/v1
   token=your-token
   model=gpt-4
   ```

4. **Interact**:
   - Type natural language commands (e.g., "list files") in the input field.
   - Press `Enter` to send the command to the AI.
   - View AI responses and command outputs in the split `tmux` pane.
   - Use `Ctrl+C` to cancel AI processing.
   - Press `Ctrl+Q` to exit.

5. **Check version**:
   ```bash
   aiterm -v
   ```
   Output:
   ```
   aiterm version v0.1.0
   ```

## Configuration

The tool stores AI configuration in `~/.aitermrc` locolly. Edit this file directly to update settings, or delete it to trigger the configuration prompt on next run.

Example `~/.aitermrc`:
```
[ai]
url=https://api.example.com/v1
token=your-token
model=gpt-4
```

## For Contributors

### Project Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/yourusername/aiterm.git
   cd aiterm
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Build and test**:
   ```bash
   go build -o aiterm
   ./aiterm
   ```

### Project Structure

```
aiterm/
├── README.md
├── ai
│   ├── ai.go # request to llm
│   ├── prompt.go # prompt
│   └── tools.go # tools to check and execute commands
├── cmd
│   └── main.go # entry point, handles flags, config, and UI setup
└── terminal
│   ├── command.go # cotroller of tmux
│   ├── dialog.go # dialog ui
├── go.mod
├── go.sum
```

### Implementation Overview

`AiTerm` is built with the following components:

- **Command-Line Interface**:
  - Uses `flag` package to handle `-v` for version display.
  - Checks `~/.aitermrc` for AI configuration (URL, Token, Model).
  - Prompts for configuration via terminal input if missing, validated using `openai-go` SDK.

- **Configuration**:
  - Stores settings in `~/.aitermrc` using `ini` format.
  - Validates AI config by attempting to list models with `openai-go`.

- **UI**:
  - Built with `tview`, featuring a `TextView` for chat history and an `InputField` for user input.
  - Supports scrolling (`SetScrollable(true)`), dynamic "generating" animation, and `Ctrl+C` cancellation.
  - Updates are queued via `app.QueueUpdateDraw` for thread safety.

- **tmux Integration**:
  - Managed by `tmux.go`, which splits the terminal using `tmux split-window -P`.
  - Sends commands to the correct pane using `send-keys`, tracked via pane ID.

- **AI Processing**:
  - Uses `openai-go` SDK to process input.
  - Runs in a goroutine with `context` for cancellation support.
  - Displays results in `TextView` and executes commands in `tmux`.

### Contribution Ideas

- Add support for multiple AI providers (e.g., Anthropic, Grok).
- Implement command-line flags for configuration overrides.
- Enhance UI with color-coded messages or a status bar.
- Add unit tests for configuration and tmux logic.
- Support configuration via environment variables.

## License

MIT License
