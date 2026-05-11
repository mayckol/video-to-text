package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mayckol/meeting-summarizer/internal/pipeline"
	"github.com/mayckol/meeting-summarizer/internal/summarize"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

type errUser struct{ err error }

func (e errUser) Error() string { return e.err.Error() }
func (e errUser) Unwrap() error { return e.err }

func userErr(format string, args ...any) error {
	return errUser{err: fmt.Errorf(format, args...)}
}

var (
	validKinds = map[string]summarize.Kind{
		"refinement":     summarize.Refinement,
		"planning":       summarize.Planning,
		"retro":          summarize.Retro,
		"1on1":           summarize.OneOnOne,
		"architecture":   summarize.Architecture,
		"tech-interview": summarize.TechInterview,
		"generic":        summarize.Generic,
	}
	validLangs = map[string]summarize.Lang{
		"pt-PT": summarize.PtPT,
		"pt-BR": summarize.PtBR,
		"en-US": summarize.EnUS,
	}
)

type flags struct {
	input          string
	summarize      bool
	kind           string
	lang           string
	participants   string
	out            string
	keepTranscript string
	whisperModel   string
	summaryModel   string
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	// Load .env from CWD if present. Real environment variables win over file values.
	_ = godotenv.Load()
	var f flags

	root := &cobra.Command{
		Use:           "summarize",
		Short:         "Transcribe a recording (default) and optionally summarize it to Markdown",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), f)
		},
	}

	root.Flags().StringVar(&f.input, "input", "", "path to video/audio file (required)")
	root.Flags().StringVar(&f.out, "out", "", "path to write output (required)")
	root.Flags().BoolVar(&f.summarize, "summarize", false, "also call OpenAI chat to produce a Markdown summary (default: transcript only)")
	root.Flags().StringVar(&f.kind, "kind", "", "meeting kind (required with --summarize): refinement|planning|retro|1on1|architecture|tech-interview|generic")
	root.Flags().StringVar(&f.lang, "lang", "", "output language (required with --summarize): pt-PT|pt-BR|en-US")
	root.Flags().StringVar(&f.participants, "participants", "", "comma-separated participant names (used with --summarize)")
	root.Flags().StringVar(&f.keepTranscript, "keep-transcript", "", "with --summarize, also write the raw transcript to this path")
	root.Flags().StringVar(&f.whisperModel, "whisper-model", "whisper-1", "OpenAI transcription model id")
	root.Flags().StringVar(&f.summaryModel, "summary-model", "gpt-4o-mini", "OpenAI chat model id for summarization")

	_ = root.MarkFlagRequired("input")
	_ = root.MarkFlagRequired("out")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		if _, ok := errors.AsType[errUser](err); ok {
			os.Exit(1)
		}
		os.Exit(2)
	}
}

func run(ctx context.Context, f flags) error {
	if _, err := os.Stat(f.input); err != nil {
		return userErr("input file: %v", err)
	}
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		return userErr("OPENAI_API_KEY env var is required")
	}

	var kind summarize.Kind
	var lang summarize.Lang
	if f.summarize {
		var ok bool
		if f.kind == "" {
			return userErr("--summarize requires --kind")
		}
		if kind, ok = validKinds[f.kind]; !ok {
			return userErr("invalid --kind %q (want refinement|planning|retro|1on1|architecture|tech-interview|generic)", f.kind)
		}
		if f.lang == "" {
			return userErr("--summarize requires --lang")
		}
		if lang, ok = validLangs[f.lang]; !ok {
			return userErr("invalid --lang %q (want pt-PT|pt-BR|en-US)", f.lang)
		}
	} else if f.kind != "" || f.lang != "" || f.participants != "" || f.keepTranscript != "" {
		slog.Warn("--kind/--lang/--participants/--keep-transcript ignored without --summarize")
	}

	var participants []string
	if f.participants != "" {
		for p := range strings.SplitSeq(f.participants, ",") {
			if p = strings.TrimSpace(p); p != "" {
				participants = append(participants, p)
			}
		}
	}

	return pipeline.Run(ctx, pipeline.Config{
		Input:          f.input,
		Summarize:      f.summarize,
		Kind:           kind,
		Lang:           lang,
		Participants:   participants,
		Out:            f.out,
		KeepTranscript: f.keepTranscript,
		WhisperModel:   f.whisperModel,
		SummaryModel:   f.summaryModel,
		OpenAIKey:      openAIKey,
	})
}
