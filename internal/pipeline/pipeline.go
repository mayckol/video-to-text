package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mayckol/meeting-summarizer/internal/audio"
	"github.com/mayckol/meeting-summarizer/internal/summarize"
	"github.com/mayckol/meeting-summarizer/internal/transcribe"
)

type Config struct {
	Input          string
	Summarize      bool
	Kind           summarize.Kind
	Lang           summarize.Lang
	Participants   []string
	Out            string
	KeepTranscript string
	WhisperModel   string
	SummaryModel   string
	OpenAIKey      string
}

func Run(ctx context.Context, cfg Config) error {
	if err := audio.HasAudio(ctx, cfg.Input); err != nil {
		return fmt.Errorf("validate audio stream: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "meeting-summarizer-*")
	if err != nil {
		return fmt.Errorf("tempdir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	wavPath := filepath.Join(tmpDir, "input.wav")
	slog.Info("extracting audio", "src", cfg.Input, "dst", wavPath)
	if err := audio.Extract(ctx, cfg.Input, wavPath); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	slog.Info("splitting audio")
	chunks, err := transcribe.Split(ctx, wavPath, tmpDir)
	if err != nil {
		return fmt.Errorf("split: %w", err)
	}
	slog.Info("split complete", "chunks", len(chunks))

	whisper := transcribe.NewClient(cfg.OpenAIKey, cfg.WhisperModel)
	slog.Info("transcribing chunks")
	transcribed, err := whisper.TranscribeAll(ctx, chunks)
	if err != nil {
		return fmt.Errorf("transcribe: %w", err)
	}

	transcript := transcribe.Stitch(transcribed)
	slog.Info("stitched transcript", "chars", len(transcript))

	if !cfg.Summarize {
		if err := os.WriteFile(cfg.Out, []byte(transcript), 0o644); err != nil {
			return fmt.Errorf("write transcript: %w", err)
		}
		slog.Info("done", "out", cfg.Out, "mode", "transcript-only")
		return nil
	}

	if cfg.KeepTranscript != "" {
		if err := os.WriteFile(cfg.KeepTranscript, []byte(transcript), 0o644); err != nil {
			return fmt.Errorf("write transcript: %w", err)
		}
	}

	prompt, err := summarize.Build(summarize.BuildInput{
		Kind:         cfg.Kind,
		Lang:         cfg.Lang,
		Participants: cfg.Participants,
		Transcript:   transcript,
	})
	if err != nil {
		return fmt.Errorf("build prompt: %w", err)
	}

	sum := summarize.NewClient(cfg.OpenAIKey, cfg.SummaryModel)
	slog.Info("calling summary model", "model", cfg.SummaryModel)
	summary, err := sum.Summarize(ctx, prompt)
	if err != nil {
		return fmt.Errorf("summarize: %w", err)
	}

	if err := os.WriteFile(cfg.Out, []byte(summary), 0o644); err != nil {
		return fmt.Errorf("write summary: %w", err)
	}
	slog.Info("done", "out", cfg.Out, "mode", "summary")
	return nil
}
