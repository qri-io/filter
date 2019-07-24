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

func (p *parser) filters() (fs []filter, err error) {
	for {
		f, err := p.readFilter()
		// fmt.Println("read filter:", f, err)
		if err != nil {
			if err.Error() == "EOF" {
				fs = append(fs, f)
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
	var fs fSlice

	for {
		t := p.scan()

		switch t.Type {
		case tDot:
			p.unscan()
			if f, err = p.readSelector(); err != nil {
				return
			}
		case tNumber:
			num, err := strconv.ParseFloat(t.Text, 64)
			if err != nil {
				return nil, err
			}
			f = fNumericLiteral(num)
		case tStar, tPlus, tMinus:
			if f, err = p.parseBinaryOp(f, t); err != nil {
				return f, err
			}
		case tLeftBracket:
			if f, err = p.parseSliceFilter(); err != nil {
				return nil, err
			}
		case tText:
			if f, err = p.parseTextFilter(t); err != nil {
				return nil, err
			}
		case tComma:
			fs = append(fs, f)
		case tPipe:
			if len(fs) > 0 {
				return append(fs, f), nil
			}
			// nil returns won't be added
			return f, nil
		case tEOF:
			if len(fs) > 0 {
				return append(fs, f), io.EOF
			}
			return f, io.EOF
		}
	}
}

func (p *parser) parseBinaryOp(left filter, t token) (f fBinaryOp, err error) {
	f = fBinaryOp{left: left, op: t.Type}
	f.right, err = p.readFilter()
	return f, err
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
			sf, err := p.parseSliceFilter()
			if err != nil {
				return nil, err
			}
			sel = append(sel, sf)
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

func (p *parser) parseSliceFilter() (f selector, err error) {
	r := &fIndexRangeSelector{}
	hasColon := false
	empty := true
	hasDigit := false

	for {
		t := p.scan()
		switch t.Type {
		case tNumber:
			num, err := strconv.ParseInt(t.Text, 10, 64)
			if err != nil {
				return nil, err
			}
			if !hasColon {
				r.start = int(num)
			} else {
				r.stop = int(num)
			}
			hasDigit = true
			empty = false
		case tColon:
			hasColon = true
		case tLeftBracket:
			continue
		case tRightBracket:
			if !hasColon && !empty {
				return fIndexSelector(int(r.start)), nil
			}
			if empty || !hasDigit && hasColon {
				r.all = true
			}
			return r, nil
		default:
			return nil, p.errorf("unexpected token: %#v", t)
		}
	}
}

func (p *parser) errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
