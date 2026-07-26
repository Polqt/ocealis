package stamp

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/Polqt/ocealis/util"
)

const MaxNoteRunes = 80

var (
	ErrStampRequired = errors.New("seal or note required")
	ErrNoteTooLong   = errors.New("stamp note must be ≤80 characters")
)

type Prepared struct {
	SealIcon string
	Note     string
}

// Prepare validates and sanitizes one passport-style Stamp.
func Prepare(sealIcon, note string) (Prepared, error) {
	sealIcon = strings.TrimSpace(util.SanitizeMessage(sealIcon))
	note = strings.TrimSpace(util.SanitizeMessage(note))

	if sealIcon == "" && note == "" {
		return Prepared{}, ErrStampRequired
	}
	if utf8.RuneCountInString(note) > MaxNoteRunes {
		return Prepared{}, ErrNoteTooLong
	}

	return Prepared{SealIcon: sealIcon, Note: note}, nil
}
