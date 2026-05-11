package summarize

import (
	"strings"
	"testing"
)

func TestBuild_AllKindsRender(t *testing.T) {
	kinds := []Kind{Refinement, Planning, Retro, OneOnOne, Architecture, TechInterview, Generic}
	for _, k := range kinds {
		t.Run(string(k), func(t *testing.T) {
			out, err := Build(BuildInput{
				Kind:         k,
				Lang:         EnUS,
				Participants: []string{"Alice", "Bob"},
				Transcript:   "hello world",
			})
			if err != nil {
				t.Fatalf("build %s: %v", k, err)
			}
			if strings.Contains(out, "{{") {
				t.Fatalf("unresolved template directive in %s output", k)
			}
			if !strings.Contains(out, "hello world") {
				t.Fatalf("transcript missing from %s output", k)
			}
			if !strings.Contains(out, "Alice, Bob") {
				t.Fatalf("participants missing from %s output", k)
			}
		})
	}
}

func TestBuild_LangInstructionSubstituted(t *testing.T) {
	tests := []struct {
		lang Lang
		want string
	}{
		{PtPT, "Português Europeu"},
		{PtBR, "Português Brasileiro"},
		{EnUS, "Respond in US English."},
	}
	for _, tt := range tests {
		t.Run(string(tt.lang), func(t *testing.T) {
			out, err := Build(BuildInput{
				Kind:       Generic,
				Lang:       tt.lang,
				Transcript: "x",
			})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out, tt.want) {
				t.Fatalf("lang %s: %q not in output", tt.lang, tt.want)
			}
		})
	}
}

func TestBuild_UnknownKindErrors(t *testing.T) {
	_, err := Build(BuildInput{
		Kind:       Kind("nonsense"),
		Lang:       EnUS,
		Transcript: "x",
	})
	if err == nil {
		t.Fatal("expected error for unknown kind")
	}
}

func TestBuild_UnsupportedLangErrors(t *testing.T) {
	_, err := Build(BuildInput{
		Kind:       Generic,
		Lang:       Lang("fr-FR"),
		Transcript: "x",
	})
	if err == nil {
		t.Fatal("expected error for unsupported lang")
	}
}
