# meeting-summarizer

Go CLI that turns a meeting recording (video or audio) into a transcript or a structured Markdown summary. Transcription via OpenAI Whisper. Optional summarization via an OpenAI chat model (default `gpt-4o-mini`).

**Default mode: transcript only.** No chat call, no summary token cost. Pass `--summarize` to also produce a kind-specific Markdown summary.

## Requirements

- Go 1.22+
- `ffmpeg` and `ffprobe` on `PATH`
- `OPENAI_API_KEY` env var

## Install

```bash
go build -o summarize ./cmd/summarize
```

## Usage

### Transcript only (default)

```bash
export OPENAI_API_KEY=...

./summarize \
  --input meeting.mp4 \
  --out ./transcript.txt
```

Writes the stitched transcript directly to `--out`. Only Whisper is called.

### Transcript + Markdown summary

```bash
./summarize \
  --input meeting.mp4 \
  --summarize \
  --kind refinement \
  --lang pt-PT \
  --participants "Mayckol,João,Ana" \
  --out ./summary.md \
  [--keep-transcript ./transcript.txt] \
  [--whisper-model whisper-1] \
  [--summary-model gpt-4o-mini]
```

With `--summarize`, the chat model is called and the Markdown summary is written to `--out`. The raw transcript can be saved alongside via `--keep-transcript`.

### Flags

| Flag | Required | Description |
|---|---|---|
| `--input` | yes | Path to a video or audio file. Anything ffmpeg can decode. |
| `--out` | yes | Path to write output (transcript `.txt` by default, or `.md` with `--summarize`). |
| `--summarize` | no | Also call OpenAI chat to produce a Markdown summary (default: transcript only). |
| `--kind` | with `--summarize` | `refinement` \| `planning` \| `retro` \| `1on1` \| `architecture` \| `tech-interview` \| `generic` |
| `--lang` | with `--summarize` | `pt-PT` \| `pt-BR` \| `en-US` |
| `--participants` | no | Comma-separated participant names (used with `--summarize`). |
| `--keep-transcript` | no | With `--summarize`, also write the raw transcript to this path. |
| `--whisper-model` | no | OpenAI transcription model id (default `whisper-1`). |
| `--summary-model` | no | OpenAI chat model id for summarization (default `gpt-4o-mini`). Override with `gpt-4o` or `gpt-4.1` for higher quality. |
| `--version` | no | Print version and exit. |

### Exit codes

| Code | Meaning |
|---|---|
| `0` | success |
| `1` | user error (bad flags, missing input, missing env var) |
| `2` | external failure (ffmpeg, OpenAI API) |

## Pipeline

1. Probe input for an audio stream (`ffprobe`); fast-fail if none.
2. Extract to mono 16 kHz PCM WAV (`ffmpeg`).
3. Probe duration; split into 10 min chunks with 5 s overlap.
4. Transcribe chunks in parallel (semaphore size 4) using Whisper `verbose_json` with word timestamps.
5. Shift each chunk's word timestamps onto the absolute timeline.
6. Stitch chunks, dropping overlap duplicates via an end-cursor.
7. **Without `--summarize`:** write the transcript to `--out` and stop.
8. **With `--summarize`:** optionally write the transcript to `--keep-transcript`, render the kind-specific prompt, call the OpenAI chat model (4096 max tokens), and write the Markdown summary to `--out`.

Tempdir is cleaned up on every exit path.

## Layout

```
cmd/summarize/main.go
internal/
  audio/       ffmpeg extract, ffprobe duration, audio-stream probe
  transcribe/  chunker (planning + ffmpeg cut), whisper adapter, stitcher
  summarize/   prompt templates per meeting kind, OpenAI chat adapter
  pipeline/    end-to-end orchestration
```

## Tests

```bash
go test ./...
```

Covers chunker math, stitcher overlap dedupe, and prompt rendering. Network adapters (Whisper, chat) and ffmpeg wrappers are deliberately untested.
