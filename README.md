# todox

A simple CLI tool for managing TODO lists in Markdown format, with tmux integration.

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/s-k-zaman/todox.git
   cd todox
   ```

2. Build and install:
   ```bash
   make install
   ```

   This installs `todox` to `~/.local/bin`. Make sure `~/.local/bin` is in your PATH.

## Usage

`todox` manages a `TODO.md` file in the root of your git repository (or current directory if not in a git repo).

### Basic Commands

- **Edit TODO.md**: `todox -e`
- **Add a task**: `todox -a "Your task here"` or `todox "Your task here"`
- **Specify custom file**: `todox -f mytasks.md -e`

### Tmux Integration

When inside tmux, `todox -e` opens the TODO file in a new tmux window named "TODOs" (configurable).

>[!TIP] For tmux keyboard shortcut, bind `prefix + M` (change M to your liking):
```tmux
bind -r M run-shell "tmux neww todox -e"
```

## Options

- `-a, --add <task>`: Add a new task to TODO.md
- `-e, --edit`: Open TODO.md in your editor (with tmux support)
- `-f, --file <filename>`: Specify TODO file name (default: TODO.md)
- `-y, --yes`: Bypass confirmation prompts
- `-w, --twin <name>`: Tmux window name (default: TODOs)
- `-n, --tnum <number>`: Tmux window number (default: 9)

## Examples

Add a task:
```bash
todox -a "Implement user authentication"
```

Edit TODO.md:
```bash
todox --edit
```

Use custom file:
```bash
todox -f project-todos.md -a "Fix bug #123"
```
