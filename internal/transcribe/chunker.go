package transcribe

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/mayckol/meeting-summarizer/internal/audio"
)

const (
	chunkSec   = 600
	overlapSec = 5
)

type Chunk struct {
	Path     string
	StartSec int
	EndSec   int
}

type chunkSpec struct {
	StartSec int
	EndSec   int
}

func planChunks(total, chunk, overlap int) []chunkSpec {
	if total <= 0 || chunk <= 0 {
		return nil
	}
	if overlap >= chunk {
		overlap = 0
	}
	step := chunk - overlap
	var out []chunkSpec
	for start := 0; start < total; start += step {
		end := min(start+chunk, total)
		out = append(out, chunkSpec{StartSec: start, EndSec: end})
		if end == total {
			break
		}
	}
	return out
}

func Split(ctx context.Context, audioPath, outDir string) ([]Chunk, error) {
	total, err := audio.DurationSec(ctx, audioPath)
	if err != nil {
		return nil, fmt.Errorf("probe: %w", err)
	}
	specs := planChunks(total, chunkSec, overlapSec)
	chunks := make([]Chunk, 0, len(specs))
	for _, s := range specs {
		out := filepath.Join(outDir, fmt.Sprintf("chunk_%04d.wav", s.StartSec))
		cmd := exec.CommandContext(ctx, "ffmpeg",
			"-ss", fmt.Sprint(s.StartSec),
			"-t", fmt.Sprint(s.EndSec-s.StartSec),
			"-i", audioPath,
			"-ac", "1", "-ar", "16000", "-c:a", "pcm_s16le",
			"-y", out,
		)
		if combined, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("chunk %ds: %w: %s", s.StartSec, err, string(combined))
		}
		chunks = append(chunks, Chunk{Path: out, StartSec: s.StartSec, EndSec: s.EndSec})
	}
	return chunks, nil
}
