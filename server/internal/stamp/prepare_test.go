package stamp_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Polqt/ocealis/internal/stamp"
)

func TestPrepareRequiresSealOrNote(t *testing.T) {
	_, err := stamp.Prepare("  ", "  ")
	if !errors.Is(err, stamp.ErrStampRequired) {
		t.Fatalf("want ErrStampRequired, got %v", err)
	}
}

func TestPrepareRejectsNoteOver80Characters(t *testing.T) {
	_, err := stamp.Prepare("anchor", strings.Repeat("潮", 81))
	if !errors.Is(err, stamp.ErrNoteTooLong) {
		t.Fatalf("want ErrNoteTooLong, got %v", err)
	}
}

func TestPrepareAcceptsSealOrNoteAndSanitizesNote(t *testing.T) {
	tests := []struct {
		name string
		seal string
		note string
	}{
		{name: "seal only", seal: "anchor"},
		{name: "note only", note: "<b>safe passage</b><script>bad</script>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stamp.Prepare(tt.seal, tt.note)
			if err != nil {
				t.Fatal(err)
			}
			if got.SealIcon == "" && got.Note == "" {
				t.Fatal("Stamp must keep a seal or note")
			}
			if strings.Contains(got.Note, "<") || strings.Contains(got.Note, "script") {
				t.Fatalf("note still contains HTML: %q", got.Note)
			}
		})
	}
}
