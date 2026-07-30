package stamp

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/Polqt/ocealis/util"
)

const MaxNoteRunes = 80

var (
	ErrEmpty       = errors.New("seal or note required")
	ErrNoteTooLong = errors.New("stamp note must be ≤80 characters")
)

type Details struct {
	SealIcon string
	Note     string
}

// Prepare validates and sanitizes one passport-style Stamp.
func Prepare(sealIcon, note string) (Details, error) {
	sealIcon = strings.TrimSpace(util.SanitizeMessage(sealIcon))
	note = strings.TrimSpace(util.SanitizeMessage(note))

	if utf8.RuneCountInString(note) > MaxNoteRunes {
		return Details{}, ErrNoteTooLong
	}
	if sealIcon == "" && note == "" {
		return Details{}, ErrEmpty
	}

	return Details{SealIcon: sealIcon, Note: note}, nil
}
