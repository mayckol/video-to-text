package transcribe

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	openai "github.com/sashabaranov/go-openai"
)

const concurrency = 4

type Client struct {
	api   *openai.Client
	model string
}

func NewClient(apiKey, model string) *Client {
	if model == "" {
		model = openai.Whisper1
	}
	return &Client{api: openai.NewClient(apiKey), model: model}
}

func (c *Client) Transcribe(ctx context.Context, ch Chunk) (Transcribed, error) {
	resp, err := c.api.CreateTranscription(ctx, openai.AudioRequest{
		Model:    c.model,
		FilePath: ch.Path,
		Format:   openai.AudioResponseFormatVerboseJSON,
		TimestampGranularities: []openai.TranscriptionTimestampGranularity{
			openai.TranscriptionTimestampGranularityWord,
		},
	})
	if err != nil {
		return Transcribed{}, fmt.Errorf("whisper transcribe %s: %w", ch.Path, err)
	}
	offset := float64(ch.StartSec)
	words := make([]Word, 0, len(resp.Words))
	for _, w := range resp.Words {
		words = append(words, Word{
			Text:     w.Word,
			StartSec: w.Start + offset,
			EndSec:   w.End + offset,
		})
	}
	slog.Info("whisper chunk transcribed",
		"start_sec", ch.StartSec,
		"end_sec", ch.EndSec,
		"detected_lang", resp.Language,
		"words", len(words),
		"text_chars", len(resp.Text),
	)
	if len(words) == 0 {
		slog.Warn("whisper returned zero words for chunk", "start_sec", ch.StartSec)
	}
	return Transcribed{Chunk: ch, Words: words}, nil
}

func (c *Client) TranscribeAll(ctx context.Context, chunks []Chunk) ([]Transcribed, error) {
	sem := make(chan struct{}, concurrency)
	out := make([]Transcribed, len(chunks))
	errCh := make(chan error, len(chunks))
	var wg sync.WaitGroup

	for i, ch := range chunks {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, ch Chunk) {
			defer wg.Done()
			defer func() { <-sem }()
			t, err := c.Transcribe(ctx, ch)
			if err != nil {
				errCh <- fmt.Errorf("chunk %d: %w", i, err)
				return
			}
			out[i] = t
		}(i, ch)
	}
	wg.Wait()
	close(errCh)
	for e := range errCh {
		if e != nil {
			return nil, e
		}
	}
	return out, nil
}
