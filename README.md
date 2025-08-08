### osheet2xlsx (CLI)

Command‑line tool for converting Osheet (.osheet) files into Microsoft Excel (.xlsx). Designed for both developers and non‑technical users: easy installation, clear commands, and helpful errors.

## What it is

Osheet is a ZIP container with spreadsheet data. This tool:
- opens .osheet (ZIP),
- reads `document.json` or `sheets/*.json`,
- converts values, formulas and some sheet parameters,
- writes the result to `.xlsx`.

## Features

- Single‑file and batch conversion
- Parallel processing, safe overwrite, dry‑run
- Formulas and basic date/time styling in Excel output
- Flexible number/date parsing with locale awareness
- Structured JSON logs (`--json`) and exit codes for automation
- Live progress: percent, speed and ETA
- Configuration via file and environment variables

## Who needs an Osheet converter and why?

- Users migrating away from Synology Office or other platforms to different tools.
- Backup and migration of personal or enterprise datasets.
- Integration into automated workflows, data exchange, and analytics pipelines.

## Requirements

- Go 1.22+ (for building from source)
- macOS / Linux / Windows (x86_64/arm64)

## Installation

### Option 1: build from source

```bash
git clone https://github.com/romanitalian/osheet2xlsx.git
cd osheet2xlsx
go build -o osheet2xlsx .
./osheet2xlsx version
```

On Windows replace `./osheet2xlsx` with `osheet2xlsx.exe`.

### Option 2: local development

```bash
# Clone and install locally
git clone https://github.com/romanitalian/osheet2xlsx.git
cd osheet2xlsx
go install .
```

### Option 3: go install

```bash
# This will work after first release is created
go install github.com/romanitalian/osheet2xlsx/v2@latest
# binary will be in $GOBIN or $GOPATH/bin/osheet2xlsx
```

**Note:** Currently there are no releases, so `@latest` won't work. Use build from source or local development options.

## Quick start

```bash
# Convert a single file
./osheet2xlsx convert path/to/file.osheet --out file.xlsx

# Inspect sheet names
./osheet2xlsx inspect path/to/file.osheet --json

# Validate structure (exit codes: 0 ok, 4 structural issue)
./osheet2xlsx validate path/to/file.osheet
```

## Commands

### convert

Convert .osheet to .xlsx — single input or batch.

Args and flags:
- `convert [path]` — path to file or directory. Default: current directory.
- `--out string` — output `.xlsx` path (single input)
- `--out-dir string` — output directory (batch)
- `--pattern string` — input file glob within a directory (default `*.osheet`)
- `--recursive` — scan subdirectories
- `--overwrite` — overwrite outputs if exist
- `--parallel int` — worker count (0=auto→1)
- `--dry-run` — do not write files, only report
- `--progress` — show progress (TTY)
- `--fail-fast` — stop the batch on first error

Examples:

```bash
# Single file → explicit output
./osheet2xlsx convert ./sample.osheet --out ./sample.xlsx

# Batch (non‑recursive)
./osheet2xlsx convert ./data --pattern "*.osheet" --out-dir out

# Recursive, with progress, overwrite
./osheet2xlsx convert ./data --pattern "*.osheet" --recursive --out-dir out --overwrite --progress

# Parallel (4 workers) and fail fast
./osheet2xlsx convert ./data --pattern "*.osheet" --parallel 4 --fail-fast

# Dry‑run (no writes)
./osheet2xlsx convert ./data --dry-run
```

### inspect

Show a brief summary (sheet names, etc.).

```bash
./osheet2xlsx inspect file.osheet
./osheet2xlsx inspect file.osheet --json
```

### validate

Check .osheet structure. In JSON mode prints machine‑readable `code` and `issue`.

```bash
./osheet2xlsx validate file.osheet
./osheet2xlsx validate file.osheet --json
```

### version

Print tool version.

### completion

Generate shell completions (bash/zsh/fish/powershell). Bash example:

```bash
./osheet2xlsx completion bash > /etc/bash_completion.d/osheet2xlsx
```

## Logging and output format

Base flags:
- `--log-level error|warn|info|debug`
- `--json` — emit JSON events (stdout)
- `--quiet` — suppress non‑error output
- `--no-color` — disable colors

Sample JSON events during conversion:

```text
{"event":"convert_start","input":"in.osheet","output":"out.xlsx"}
{"event":"convert_ok","input":"in.osheet","output":"out.xlsx"}
{"event":"convert_error","input":"bad.osheet","error":"parse failed"}
```

## Configuration (env/file)

You can configure the tool via a file or environment variables. CLI flags take precedence.

Search order (first found wins):
- `./osheet2xlsx.json`
- `$XDG_CONFIG_HOME/osheet2xlsx/config.json`
- `$HOME/.osheet2xlsx.json`
- exact path via `$OS2X_CONFIG`

Example `osheet2xlsx.json`:

```json
{
  "logLevel": "info",
  "json": false,
  "convert": {
    "pattern": "*.osheet",
    "recursive": true,
    "outDir": "out",
    "overwrite": false,
    "parallel": 0,
    "dryRun": false,
    "progress": true,
    "failFast": false
  }
}
```

Environment variables (override file):
- `OS2X_LOG_LEVEL`, `OS2X_JSON`, `OS2X_QUIET`, `OS2X_NO_COLOR`
- `OS2X_CONVERT_PATTERN`, `OS2X_CONVERT_RECURSIVE`, `OS2X_CONVERT_OUT_DIR`,
  `OS2X_CONVERT_OVERWRITE`, `OS2X_CONVERT_PARALLEL`, `OS2X_CONVERT_DRY_RUN`,
  `OS2X_CONVERT_PROGRESS`, `OS2X_CONVERT_FAIL_FAST`

## Exit codes

- 0 — success
- 2 — invalid arguments/usage
- 3 — I/O errors
- 4 — parse/validation (structural) errors
- 5 — partial success (some errors in batch)

## Input format (Osheet)

- Expected to be a ZIP file
- Inside: `document.json` (preferred) or `sheets/*.json`
- Supported `document.json` variants:
  - `{"sheets":[{"name":"S","rows":[["a","b"],...]},...]}`
  - `{"sheets":[{"name":"S","cells":[[{"t":"n","v":1},{"f":"SUM(A1:B1)"}],...]}]}`
- If `document.json` is missing, the tool attempts `sheets/*.json`.
- If nothing matches, text entries under `sheets/` are embedded or a simple archive listing is produced.

## Progress

With `--progress` (and an attached TTY) the tool shows: percent complete, processing speed (items/sec), elapsed time and ETA.

## Limitations

- Minimal styling in `.xlsx` (dates/time use a basic style)
- Formulas are written as‑is (no validation)
- Ambiguous number/date formats are parsed with best‑effort heuristics

## Security

- With `--out-dir`, path traversal is prevented: outputs are created only inside the provided directory.

## Troubleshooting

- Use `--log-level debug` for more details
- In `--json` mode events are suitable for parsing in scripts/CI
- `validate` returns exit code 4 for structural issues and prints `code`/`issue` in JSON mode

## FAQ

— I don’t have a real .osheet. How can I test?

Create a test container: it’s a ZIP with `document.json`:

```bash
mkdir tmp && printf '%s' '{"sheets":[{"name":"S","rows":[["1","2"],["3","4"]]}]}' > tmp/document.json
(cd tmp && zip -rq ../sample.osheet .)
./osheet2xlsx convert sample.osheet --out sample.xlsx
```

— How do I enable shell completion?

Generate a script for your shell (bash/zsh/fish/pwsh) using `completion` and source it.

## License

License is not specified. For commercial/open‑source use, please agree on terms with the author.


