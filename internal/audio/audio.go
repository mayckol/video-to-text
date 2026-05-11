package audio

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func HasAudio(ctx context.Context, path string) error {
	out, err := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_type",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	).Output()
	if err != nil {
		return fmt.Errorf("probe %s: %w", path, err)
	}
	if strings.TrimSpace(string(out)) != "audio" {
		return fmt.Errorf("no audio stream in %s", path)
	}
	return nil
}

func Extract(ctx context.Context, src, dst string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", src,
		"-vn",
		"-ac", "1",
		"-ar", "16000",
		"-c:a", "pcm_s16le",
		"-y", dst,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg extract %s: %w: %s", src, err, string(out))
	}
	return nil
}

func Duration(ctx context.Context, path string) (time.Duration, error) {
	sec, err := DurationSec(ctx, path)
	if err != nil {
		return 0, err
	}
	return time.Duration(sec) * time.Second, nil
}

func DurationSec(ctx context.Context, path string) (int, error) {
	out, err := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	).Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe duration %s: %w", path, err)
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %w", string(out), err)
	}
	return int(f), nil
}
