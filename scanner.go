package filter

import (
	"bufio"
	"io"
	"strings"
)

// newScanner allocates a scanner from an io.Reader
func newScanner(r io.Reader) *scanner {
	return &scanner{r: bufio.NewReader(r)}
}

// scanner tokenizes an input stream
// TODO(b5): set position properly for errors
type scanner struct {
	r *bufio.Reader

	// scanning state
	tok               token
	text              strings.Builder
	line, col, offset int
	readNewline       bool
	err               error
}

// Scan reads one token from the input stream
func (s *scanner) Scan() token {
	inText := false
	inLiteral := false
	s.text.Reset()

	for {
		ch := s.read()

		switch ch {
		case eof:
			if inText || inLiteral {
				return s.newTok(tText)
			}
			return s.newTok(tEOF)
		// ignore line feeds
		case '\r':
			continue
		case '|':
			if !inText {
				return s.newTok(tPipe)
			}
			s.text.WriteRune(ch)
		case '.':
			if inLiteral {
				if err := s.unread(); err != nil {
					panic(err)
				}
				return s.newTextTok()
			}
			if !inText {
				return s.newTok(tDot)
			}
			s.text.WriteRune(ch)
		case '"':
			if inText {
				return s.newTextTok()
			}
			inText = true
		case ' ':
			if inLiteral {
				return s.newTextTok()
			}
		default:
			s.text.WriteRune(ch)
			if !inText {
				inLiteral = true
			}
		}
	}
}

// read reads the next rune from the buffered reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *scanner) unread() error {
	return s.r.UnreadRune()
}

// newTok creates a new token from current scanner state
func (s *scanner) newTok(t tokenType) token {
	return token{
		Type: t,
		Text: strings.TrimSpace(s.text.String()),
		Pos:  position{Line: s.line, Col: s.col, Offset: s.offset},
	}
}

func (s *scanner) newTextTok() token {
	// TODO (b5) - handle numeric literals
	return token{
		Type: tText,
		Text: strings.TrimSpace(s.text.String()),
		Pos:  position{Line: s.line, Col: s.col, Offset: s.offset},
	}
}

// eof represents a marker rune for the end of the reader.
var eof = rune(0)
