package transcribe

import (
	"reflect"
	"testing"
)

func TestPlanChunks_ShortFileNoSplit(t *testing.T) {
	got := planChunks(30, 600, 5)
	want := []chunkSpec{{0, 30}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestPlanChunks_25MinuteFile(t *testing.T) {
	got := planChunks(1500, 600, 5)
	want := []chunkSpec{
		{0, 600},
		{595, 1195},
		{1190, 1500},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestPlanChunks_LastChunkBounded(t *testing.T) {
	got := planChunks(700, 600, 5)
	want := []chunkSpec{
		{0, 600},
		{595, 700},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for _, c := range got {
		if c.EndSec > 700 {
			t.Fatalf("end %d exceeds total", c.EndSec)
		}
	}
}

func TestPlanChunks_ExactBoundary(t *testing.T) {
	got := planChunks(600, 600, 5)
	want := []chunkSpec{{0, 600}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestPlanChunks_ZeroOrNegative(t *testing.T) {
	if planChunks(0, 600, 5) != nil {
		t.Fatal("expected nil for zero total")
	}
	if planChunks(-1, 600, 5) != nil {
		t.Fatal("expected nil for negative total")
	}
}

func TestPlanChunks_OverlapEqualOrGreaterIgnored(t *testing.T) {
	got := planChunks(1200, 600, 600)
	want := []chunkSpec{{0, 600}, {600, 1200}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
