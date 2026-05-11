package transcribe

import (
	"strings"
	"testing"
)

func mkWords(start float64, texts []string) []Word {
	words := make([]Word, len(texts))
	for i, t := range texts {
		words[i] = Word{
			Text:     t,
			StartSec: start + float64(i),
			EndSec:   start + float64(i) + 1,
		}
	}
	return words
}

func TestStitch_OverlappingChunksDedupe(t *testing.T) {
	a := Transcribed{
		Chunk: Chunk{StartSec: 0, EndSec: 5},
		Words: mkWords(0, []string{"one", "two", "three", "four", "five"}),
	}
	b := Transcribed{
		Chunk: Chunk{StartSec: 3, EndSec: 8},
		Words: mkWords(3, []string{"four", "five", "six", "seven", "eight"}),
	}
	got := Stitch([]Transcribed{a, b})
	want := "one two three four five six seven eight"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestStitch_NonOverlappingChunks(t *testing.T) {
	a := Transcribed{
		Chunk: Chunk{StartSec: 0, EndSec: 3},
		Words: mkWords(0, []string{"a", "b", "c"}),
	}
	b := Transcribed{
		Chunk: Chunk{StartSec: 10, EndSec: 13},
		Words: mkWords(10, []string{"d", "e", "f"}),
	}
	got := Stitch([]Transcribed{a, b})
	want := "a b c d e f"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestStitch_UnsortedInputSorted(t *testing.T) {
	a := Transcribed{
		Chunk: Chunk{StartSec: 0, EndSec: 2},
		Words: mkWords(0, []string{"x", "y"}),
	}
	b := Transcribed{
		Chunk: Chunk{StartSec: 5, EndSec: 7},
		Words: mkWords(5, []string{"z", "w"}),
	}
	got := Stitch([]Transcribed{b, a})
	want := "x y z w"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestStitch_Empty(t *testing.T) {
	if got := Stitch(nil); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestStitch_NoDoubleSpaces(t *testing.T) {
	a := Transcribed{
		Chunk: Chunk{StartSec: 0, EndSec: 3},
		Words: mkWords(0, []string{"alpha", "beta"}),
	}
	got := Stitch([]Transcribed{a})
	if strings.Contains(got, "  ") {
		t.Fatalf("double spaces in %q", got)
	}
}
