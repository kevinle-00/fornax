# Fornax

I came up with Fornax because I was tired of using ad-heavy YouTube-to-MP4
websites and wanted a simple tool that could do the same job locally. It is a
terminal media processing app I built while learning Go and concurrent
programming.

Fornax can download videos, convert local media files, or run both steps
together. Multiple jobs can run at the same time, with their status and
progress shown in a terminal dashboard.

## Features

- Download media from a URL with `yt-dlp`
- Convert local media files with `ffmpeg`
- Download and convert media in one job
- Process several jobs concurrently
- Use either the interactive dashboard or CLI commands
- View job status and progress in a terminal dashboard
- Manually requeue failed jobs

## Requirements

- Go 1.25 or newer
- `yt-dlp`
- `ffmpeg` and `ffprobe`

On macOS, the external tools can be installed with Homebrew:

```bash
brew install yt-dlp ffmpeg
```

## Running Fornax

Build the application:

```bash
go build -o fornax .
```

### TUI

Run Fornax without a command to open the interactive dashboard:

```bash
./fornax
```

Choose an action from the menu and enter the requested values. Output
directories should already exist.

### Controls

| Key | Action |
| --- | --- |
| `up` / `k` | Move up |
| `down` / `j` | Move down |
| `enter` | Select or submit |
| `esc` | Return to the menu from an input screen |
| `m` | Return to the menu from the dashboard |
| `r` | Requeue the selected failed job |
| `q` | Quit from the menu or dashboard |
| `ctrl+c` | Quit from any screen |

### CLI

The same workflows can run without the interactive dashboard:

```bash
./fornax download <url>... --output output/
./fornax encode <file>... --format mp4 --output output/
./fornax process <url>... --format mp4 --output output/
```

Use `--quality` to pass a format selector to `yt-dlp`. Each command prints a
result for every input and returns an error if any job fails. Run
`./fornax <command> --help` for all available flags.

## Architecture

```mermaid
flowchart LR
    User[User] --> TUI[Bubble Tea TUI]
    User --> CLI[CLI Commands]
    TUI --> Queue[Job Queue]
    CLI --> Queue
    Queue --> Workers[Worker Pool]

    Workers --> Download[Download Job]
    Workers --> Encode[Encode Job]
    Workers --> Process[Process Job]

    Download --> YTDLP[yt-dlp]
    Encode --> FFmpeg[ffprobe + ffmpeg]
    Process --> YTDLP
    Process --> Temp[Temporary File]
    Temp --> FFmpeg

    Download --> Output[Output Files]
    Encode --> Output
    FFmpeg --> Output

    Workers -. status and progress .-> TUI
    Workers -. final results .-> CLI
```

The TUI and CLI both add jobs to a queue. A pool of workers reads from that
queue and runs jobs concurrently. Each job stores its own status, progress, and
error so either interface can report what happened.

## Development

```bash
go fmt ./...
go vet ./...
go test ./...
```

The project is split into small packages:

- `cmd`: contains the CLI commands and starts the TUI
- `download`: runs `yt-dlp`
- `encode`: runs `ffprobe` and `ffmpeg`
- `job`: defines download, encode, and combined jobs
- `queue`: stores pending jobs
- `worker`: processes jobs concurrently
- `ui`: contains the Bubble Tea terminal interface
- `validate`: contains input validation helpers
