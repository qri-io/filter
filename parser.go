package filter

import (
	"fmt"
	"io"
	"strconv"
)

// parser is a state machine for serializing a documentation struct from a byte stream
type parser struct {
	s *scanner

	buf struct {
		tok          token
		line, indent int
		n            int
	}

	line   int
	indent int // indentation level of current line
}

func (p *parser) scan() (tok token) {
	if p.buf.n > 0 {
		tok = p.buf.tok
		p.buf.n = 0
		return
	}

	defer func() {
		// fmt.Println("read", tok.Type, tok.Text)
		p.buf.tok = tok
	}()

	tok = p.s.Scan()
	return tok
}

func (p *parser) unscan() {
	p.buf.n = 1
}

func (p *parser) read() (fs []filter, err error) {
	for {
		f, err := p.readFilter()
		// fmt.Println("read filter:", f, err)
		if err != nil {
			if err.Error() == "EOF" {
				return fs, nil
			}
			return nil, err
		}
		if f != nil {
			fs = append(fs, f)
		}
	}
}

func (p *parser) readFilter() (f filter, err error) {
	for {
		t := p.scan()

		switch t.Type {
		case tDot:
			p.unscan()
			return p.readSelector()
		case tNumber:
			num, err := strconv.ParseFloat(t.Text, 64)
			if err != nil {
				return nil, err
			}
			return fNumericLiteral(num), nil
		case tText:
			return p.parseTextFilter(t)
		case tPipe:
			return nil, nil
		case tEOF:
			return f, io.EOF
		}
	}
}

func (p *parser) readSelector() (sel fSelector, err error) {
	for {
		t := p.scan()
		switch t.Type {
		case tDot:
			sel = append(sel, fIdentity('.'))
		case tText:
			sel = append(sel, fKeySelector(t.Text))
		case tLeftBracket:
			return nil, fmt.Errorf("range selection not yet supported")
		default:
			p.unscan()
			return sel, nil
		}
	}
}

func (p *parser) parseTextFilter(t token) (f filter, err error) {
	switch t.Text {
	case "length":
		return fLength(0), nil
	default:
		return fStringLiteral(t.Text), nil
	}
}

func (p *parser) errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
