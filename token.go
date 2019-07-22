package filter

// position of a token within the scan stream
type position struct {
	Line, Col, Offset int
}

// token is a recognized token from the outline lexicon
type token struct {
	Type tokenType
	Pos  position
	Text string
}

// String implements the stringer interface for token
func (t token) String() string {
	return t.Text
}

// tokenType enumerates the different types of tokens
type tokenType int

const (
	// IllegalTok is the default for unrecognized tokens
	IllegalTok tokenType = iota

	// tEOF is the end-of-file token
	tEOF

	// literalBegin marks the beginning of literal tokens in the token enumeration
	literalBegin
	// tText is a token for arbitrary text
	tText
	// tNumber is a number
	tNumber
	// tDot is the "." character
	tDot
	// tComma is the "," character
	tComma
	// tColon is the ":" character
	tColon
	// tPipe is the "|" character
	tPipe
	// tLeftBracket is the "[" character
	tLeftBracket
	// tRightBracket is the "]" character
	tRightBracket
	// tLeftBrace is the "{" character
	tLeftBrace
	// tLeftBrace is the "}" character
	tRightBrace
	// tPlus is the "+" character
	tPlus
	// tMinus is the "-" character
	tMinus
	// tForwardSlash is the "/" character
	tForwardSlash
	// literalEnd marks the end of literal tokens in the token enumeration
	literalEnd

	// keywordBegin marks the start of keyword tokens in the token enumeration
	keywordBegin
	// length is the "length" token
	tLength
	// keywordEnd marks the end of keyword tokens in the token enumeration
	keywordEnd
)
