package transcribe

import (
	"sort"
	"strings"
)

type Word struct {
	Text     string
	StartSec float64
	EndSec   float64
}

type Transcribed struct {
	Chunk
	Words []Word
}

func Stitch(chunks []Transcribed) string {
	sort.Slice(chunks, func(i, j int) bool { return chunks[i].StartSec < chunks[j].StartSec })
	var b strings.Builder
	var cursor float64
	for _, c := range chunks {
		for _, w := range c.Words {
			if w.StartSec < cursor {
				continue
			}
			b.WriteString(w.Text)
			b.WriteByte(' ')
			cursor = w.EndSec
		}
	}
	return strings.TrimSpace(b.String())
}
