// Package lexer implements a PHP lexer that produces tokens compatible
// with PHP's official tokenizer.
package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/alexisbouchez/phpgo/token"
)

// Position represents a position in the source code.
type Position struct {
	Offset int // byte offset
	Line   int // line number (1-based)
	Column int // column number (1-based)
}

// TokenInfo represents a token with its metadata.
type TokenInfo struct {
	Type    token.Token
	Literal string
	Pos     Position
}

// LexerState represents the current state of the lexer.
type LexerState int

const (
	StateInitial LexerState = iota
	StateInScripting
	StateDoubleQuotes
	StateBacktick
	StateHeredoc
	StateNowdoc
	StateLookingForProperty
	StateLookingForVarname
	StateVarOffset
)

// Lexer tokenizes PHP source code.
type Lexer struct {
	input        string
	pos          int  // current position in input
	readPos      int  // reading position (after current char)
	ch           byte // current char under examination
	line         int
	column       int
	state        LexerState
	stateStack   []LexerState
	heredocLabel string
	isNowdoc     bool
}

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	l := &Lexer{
		input:      input,
		line:       1,
		column:     1,
		state:      StateInitial,
		stateStack: make([]LexerState, 0),
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	if l.ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) peekCharN(n int) byte {
	pos := l.readPos + n - 1
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

func (l *Lexer) pushState(state LexerState) {
	l.stateStack = append(l.stateStack, l.state)
	l.state = state
}

func (l *Lexer) popState() {
	if len(l.stateStack) > 0 {
		l.state = l.stateStack[len(l.stateStack)-1]
		l.stateStack = l.stateStack[:len(l.stateStack)-1]
	}
}

func (l *Lexer) currentPos() Position {
	return Position{
		Offset: l.pos,
		Line:   l.line,
		Column: l.column - 1,
	}
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() TokenInfo {
	switch l.state {
	case StateInitial:
		return l.scanInitial()
	case StateInScripting:
		return l.scanInScripting()
	case StateDoubleQuotes:
		return l.scanDoubleQuotes()
	case StateBacktick:
		return l.scanBacktick()
	case StateHeredoc:
		return l.scanHeredoc()
	case StateNowdoc:
		return l.scanNowdoc()
	case StateLookingForVarname:
		return l.scanLookingForVarname()
	case StateVarOffset:
		return l.scanVarOffset()
	case StateLookingForProperty:
		return l.scanLookingForProperty()
	default:
		return l.scanInScripting()
	}
}

func (l *Lexer) scanInitial() TokenInfo {
	if l.ch == 0 {
		return TokenInfo{Type: token.EOF, Literal: "", Pos: l.currentPos()}
	}

	// Look for PHP open tag
	startPos := l.currentPos()
	startIdx := l.pos

	for l.ch != 0 {
		if l.ch == '<' && l.peekChar() == '?' {
			if startIdx < l.pos {
				// Return inline HTML before the tag
				return TokenInfo{
					Type:    token.T_INLINE_HTML,
					Literal: l.input[startIdx:l.pos],
					Pos:     startPos,
				}
			}

			// Check for <?php or <?=
			pos := l.currentPos()
			l.readChar() // <
			l.readChar() // ?

			if l.ch == '=' {
				l.readChar()
				l.state = StateInScripting
				return TokenInfo{Type: token.T_OPEN_TAG_WITH_ECHO, Literal: "<?=", Pos: pos}
			}

			// Check for "php" followed by whitespace
			if (l.ch == 'p' || l.ch == 'P') &&
				(l.peekChar() == 'h' || l.peekChar() == 'H') &&
				(l.peekCharN(2) == 'p' || l.peekCharN(2) == 'P') {
				l.readChar() // p
				l.readChar() // h
				l.readChar() // p
				// Read trailing whitespace (at least one space/newline required)
				if isWhitespace(l.ch) {
					for isWhitespace(l.ch) {
						if l.ch == '\n' {
							l.readChar()
							break
						}
						l.readChar()
					}
					l.state = StateInScripting
					return TokenInfo{
						Type:    token.T_OPEN_TAG,
						Literal: l.input[pos.Offset:l.pos],
						Pos:     pos,
					}
				} else if l.ch == 0 {
					l.state = StateInScripting
					return TokenInfo{
						Type:    token.T_OPEN_TAG,
						Literal: l.input[pos.Offset:l.pos],
						Pos:     pos,
					}
				}
				// Not followed by whitespace, continue looking
				continue
			}
			continue
		}
		l.readChar()
	}

	if startIdx < l.pos {
		return TokenInfo{
			Type:    token.T_INLINE_HTML,
			Literal: l.input[startIdx:l.pos],
			Pos:     startPos,
		}
	}

	return TokenInfo{Type: token.EOF, Literal: "", Pos: l.currentPos()}
}

func (l *Lexer) scanInScripting() TokenInfo {
	pos := l.currentPos()

	if l.ch == 0 {
		return TokenInfo{Type: token.EOF, Literal: "", Pos: pos}
	}

	// Whitespace
	if isWhitespace(l.ch) {
		return l.scanWhitespace()
	}

	// Comments
	if l.ch == '/' && l.peekChar() == '/' {
		return l.scanSingleLineComment()
	}
	if l.ch == '#' {
		if l.peekChar() == '[' {
			// Attribute
			l.readChar() // #
			l.readChar() // [
			return TokenInfo{Type: token.T_ATTRIBUTE, Literal: "#[", Pos: pos}
		}
		return l.scanSingleLineComment()
	}
	if l.ch == '/' && l.peekChar() == '*' {
		return l.scanMultiLineComment()
	}

	// String literals
	if l.ch == '\'' {
		return l.scanSingleQuotedString()
	}
	if l.ch == '"' {
		return l.scanDoubleQuotedStringStart()
	}
	if l.ch == '`' {
		l.readChar()
		l.pushState(StateBacktick)
		return TokenInfo{Type: token.BACKTICK, Literal: "`", Pos: pos}
	}

	// Heredoc/Nowdoc
	if l.ch == '<' && l.peekChar() == '<' && l.peekCharN(2) == '<' {
		return l.scanHeredocStart()
	}

	// Close tag
	if l.ch == '?' && l.peekChar() == '>' {
		l.readChar()
		l.readChar()
		l.state = StateInitial
		return TokenInfo{Type: token.T_CLOSE_TAG, Literal: "?>", Pos: pos}
	}

	// Variable
	if l.ch == '$' {
		if isIdentStart(l.peekChar()) {
			return l.scanVariable()
		}
		l.readChar()
		return TokenInfo{Type: token.DOLLAR, Literal: "$", Pos: pos}
	}

	// Numbers
	if isDigit(l.ch) || (l.ch == '.' && isDigit(l.peekChar())) {
		return l.scanNumber()
	}

	// Identifiers and keywords
	if isIdentStart(l.ch) {
		return l.scanIdentifier()
	}

	// Namespace separator (fully qualified name)
	if l.ch == '\\' {
		return l.scanNamespaceName()
	}

	// Operators and punctuation
	return l.scanOperator()
}

func (l *Lexer) scanWhitespace() TokenInfo {
	pos := l.currentPos()
	startIdx := l.pos
	for isWhitespace(l.ch) {
		l.readChar()
	}
	return TokenInfo{
		Type:    token.WHITESPACE,
		Literal: l.input[startIdx:l.pos],
		Pos:     pos,
	}
}

func (l *Lexer) scanSingleLineComment() TokenInfo {
	pos := l.currentPos()
	startIdx := l.pos
	for l.ch != 0 && l.ch != '\n' {
		if l.ch == '?' && l.peekChar() == '>' {
			break
		}
		l.readChar()
	}
	if l.ch == '\n' {
		l.readChar()
	}
	return TokenInfo{
		Type:    token.T_COMMENT,
		Literal: l.input[startIdx:l.pos],
		Pos:     pos,
	}
}

func (l *Lexer) scanMultiLineComment() TokenInfo {
	pos := l.currentPos()
	startIdx := l.pos
	l.readChar() // /
	l.readChar() // *

	isDoc := l.ch == '*' && l.peekChar() != '/'

	for l.ch != 0 {
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // *
			l.readChar() // /
			break
		}
		l.readChar()
	}

	tokType := token.T_COMMENT
	if isDoc {
		tokType = token.T_DOC_COMMENT
	}

	return TokenInfo{
		Type:    tokType,
		Literal: l.input[startIdx:l.pos],
		Pos:     pos,
	}
}

func (l *Lexer) scanSingleQuotedString() TokenInfo {
	pos := l.currentPos()
	startIdx := l.pos
	l.readChar() // opening quote

	for l.ch != 0 {
		if l.ch == '\\' && (l.peekChar() == '\'' || l.peekChar() == '\\') {
			l.readChar()
			l.readChar()
			continue
		}
		if l.ch == '\'' {
			l.readChar()
			break
		}
		l.readChar()
	}

	return TokenInfo{
		Type:    token.T_CONSTANT_ENCAPSED_STRING,
		Literal: l.input[startIdx:l.pos],
		Pos:     pos,
	}
}

func (l *Lexer) scanDoubleQuotedStringStart() TokenInfo {
	pos := l.currentPos()
	startIdx := l.pos
	l.readChar() // opening quote

	// Check if the string contains variables
	hasVar := false
	checkPos := l.pos
	for checkPos < len(l.input) {
		ch := l.input[checkPos]
		if ch == '\\' {
			checkPos += 2
			continue
		}
		if ch == '"' {
			break
		}
		if ch == '$' {
			if checkPos+1 < len(l.input) {
				next := l.input[checkPos+1]
				if isIdentStart(next) || next == '{' {
					hasVar = true
					break
				}
			}
		}
		if ch == '{' && checkPos+1 < len(l.input) && l.input[checkPos+1] == '$' {
			hasVar = true
			break
		}
		checkPos++
	}

	if !hasVar {
		// Simple string without variables
		for l.ch != 0 && l.ch != '"' {
			if l.ch == '\\' {
				l.readChar()
				if l.ch != 0 {
					l.readChar()
				}
				continue
			}
			l.readChar()
		}
		if l.ch == '"' {
			l.readChar()
		}
		return TokenInfo{
			Type:    token.T_CONSTANT_ENCAPSED_STRING,
			Literal: l.input[startIdx:l.pos],
			Pos:     pos,
		}
	}

	// String with variables - enter double quotes state
	l.pushState(StateDoubleQuotes)
	return TokenInfo{Type: token.DOUBLE_QUOTE, Literal: `"`, Pos: pos}
}

func (l *Lexer) scanDoubleQuotes() TokenInfo {
	pos := l.currentPos()

	// End of string
	if l.ch == '"' {
		l.readChar()
		l.popState()
		return TokenInfo{Type: token.DOUBLE_QUOTE, Literal: `"`, Pos: pos}
	}

	// Variable interpolation: {$var}
	if l.ch == '{' && l.peekChar() == '$' {
		l.readChar() // {
		l.pushState(StateInScripting)
		return TokenInfo{Type: token.T_CURLY_OPEN, Literal: "{", Pos: pos}
	}

	// Variable interpolation: ${var}
	if l.ch == '$' && l.peekChar() == '{' {
		l.readChar() // $
		l.readChar() // {
		l.pushState(StateLookingForVarname)
		return TokenInfo{Type: token.T_DOLLAR_OPEN_CURLY_BRACES, Literal: "${", Pos: pos}
	}

	// Simple variable
	if l.ch == '$' && isIdentStart(l.peekChar()) {
		return l.scanVariableInString()
	}

	// Regular string content
	return l.scanEncapsedString('"')
}

func (l *Lexer) scanEncapsedString(endChar byte) TokenInfo {
	pos := l.currentPos()
	startIdx := l.pos

	for l.ch != 0 && l.ch != endChar {
		// Check for variable interpolation
		if l.ch == '$' {
			if isIdentStart(l.peekChar()) || l.peekChar() == '{' {
				break
			}
		}
		if l.ch == '{' && l.peekChar() == '$' {
			break
		}

		// Handle escape sequences
		if l.ch == '\\' {
			l.readChar()
			if l.ch != 0 {
				l.readChar()
			}
			continue
		}

		l.readChar()
	}

	if startIdx == l.pos {
		// No content scanned, this shouldn't happen
		l.readChar()
		return TokenInfo{Type: token.ILLEGAL, Literal: string(l.ch), Pos: pos}
	}

	return TokenInfo{
		Type:    token.T_ENCAPSED_AND_WHITESPACE,
		Literal: l.input[startIdx:l.pos],
		Pos:     pos,
	}
}

func (l *Lexer) scanVariableInString() TokenInfo {
	pos := l.currentPos()
	startIdx := l.pos
	l.readChar() // $

	for isIdentChar(l.ch) {
		l.readChar()
	}

	literal := l.input[startIdx:l.pos]

	// Check for array access or property access
	if l.ch == '[' {
		// Will be handled by caller with state
		l.pushState(StateVarOffset)
	} else if l.ch == '-' && l.peekChar() == '>' {
		// Property access in string - push state to handle it
		l.pushState(StateLookingForProperty)
	}

	return TokenInfo{
		Type:    token.T_VARIABLE,
		Literal: literal,
		Pos:     pos,
	}
}

func (l *Lexer) scanLookingForProperty() TokenInfo {
	pos := l.currentPos()

	// Handle ->
	if l.ch == '-' && l.peekChar() == '>' {
		l.readChar() // -
		l.readChar() // >
		return TokenInfo{Type: token.T_OBJECT_OPERATOR, Literal: "->", Pos: pos}
	}

	// Read property name
	if isIdentStart(l.ch) {
		startIdx := l.pos
		for isIdentChar(l.ch) {
			l.readChar()
		}
		l.popState()
		return TokenInfo{
			Type:    token.T_STRING,
			Literal: l.input[startIdx:l.pos],
			Pos:     pos,
		}
	}

	l.popState()
	return l.scanDoubleQuotes()
}

func (l *Lexer) scanVarOffset() TokenInfo {
	pos := l.currentPos()

	if l.ch == '[' {
		l.readChar()
		return TokenInfo{Type: token.LBRACKET, Literal: "[", Pos: pos}
	}

	if l.ch == ']' {
		l.readChar()
		l.popState()
		return TokenInfo{Type: token.RBRACKET, Literal: "]", Pos: pos}
	}

	// Numeric index
	if isDigit(l.ch) {
		startIdx := l.pos
		for isDigit(l.ch) {
			l.readChar()
		}
		return TokenInfo{
			Type:    token.T_NUM_STRING,
			Literal: l.input[startIdx:l.pos],
			Pos:     pos,
		}
	}

	// String index
	if isIdentStart(l.ch) {
		startIdx := l.pos
		for isIdentChar(l.ch) {
			l.readChar()
		}
		return TokenInfo{
			Type:    token.T_STRING,
			Literal: l.input[startIdx:l.pos],
			Pos:     pos,
		}
	}

	// Variable index
	if l.ch == '$' && isIdentStart(l.peekChar()) {
		return l.scanVariable()
	}

	l.popState()
	return l.scanDoubleQuotes()
}

func (l *Lexer) scanLookingForVarname() TokenInfo {
	pos := l.currentPos()

	if l.ch == '}' {
		l.readChar()
		l.popState()
		return TokenInfo{Type: token.RBRACE, Literal: "}", Pos: pos}
	}

	if isIdentStart(l.ch) {
		startIdx := l.pos
		for isIdentChar(l.ch) {
			l.readChar()
		}

		if l.ch == '[' || l.ch == '}' {
			return TokenInfo{
				Type:    token.T_STRING_VARNAME,
				Literal: l.input[startIdx:l.pos],
				Pos:     pos,
			}
		}

		// Complex expression - switch to scripting
		l.state = StateInScripting
		return TokenInfo{
			Type:    token.T_STRING_VARNAME,
			Literal: l.input[startIdx:l.pos],
			Pos:     pos,
		}
	}

	l.popState()
	return l.scanDoubleQuotes()
}

func (l *Lexer) scanBacktick() TokenInfo {
	pos := l.currentPos()

	if l.ch == '`' {
		l.readChar()
		l.popState()
		return TokenInfo{Type: token.BACKTICK, Literal: "`", Pos: pos}
	}

	if l.ch == '$' && isIdentStart(l.peekChar()) {
		return l.scanVariableInString()
	}

	if l.ch == '{' && l.peekChar() == '$' {
		l.readChar()
		l.pushState(StateInScripting)
		return TokenInfo{Type: token.T_CURLY_OPEN, Literal: "{", Pos: pos}
	}

	// Regular content
	startIdx := l.pos
	for l.ch != 0 && l.ch != '`' {
		if l.ch == '$' && isIdentStart(l.peekChar()) {
			break
		}
		if l.ch == '{' && l.peekChar() == '$' {
			break
		}
		if l.ch == '\\' {
			l.readChar()
			if l.ch != 0 {
				l.readChar()
			}
			continue
		}
		l.readChar()
	}

	return TokenInfo{
		Type:    token.T_ENCAPSED_AND_WHITESPACE,
		Literal: l.input[startIdx:l.pos],
		Pos:     pos,
	}
}

func (l *Lexer) scanHeredocStart() TokenInfo {
	pos := l.currentPos()
	startIdx := l.pos
	l.readChar() // <
	l.readChar() // <
	l.readChar() // <

	// Check for quotes
	l.isNowdoc = false
	quoted := byte(0)
	if l.ch == '\'' {
		l.isNowdoc = true
		quoted = '\''
		l.readChar()
	} else if l.ch == '"' {
		quoted = '"'
		l.readChar()
	}

	// Read label
	labelStart := l.pos
	for isIdentChar(l.ch) {
		l.readChar()
	}
	l.heredocLabel = l.input[labelStart:l.pos]

	if quoted != 0 {
		l.readChar() // closing quote
	}

	// Skip to end of line
	for l.ch != 0 && l.ch != '\n' {
		l.readChar()
	}
	if l.ch == '\n' {
		l.readChar()
	}

	if l.isNowdoc {
		l.state = StateNowdoc
	} else {
		l.state = StateHeredoc
	}

	return TokenInfo{
		Type:    token.T_START_HEREDOC,
		Literal: l.input[startIdx:l.pos],
		Pos:     pos,
	}
}

func (l *Lexer) scanHeredoc() TokenInfo {
	pos := l.currentPos()

	// Check for end label
	if l.checkHeredocEnd() {
		startIdx := l.pos
		for isIdentChar(l.ch) {
			l.readChar()
		}
		l.state = StateInScripting
		return TokenInfo{
			Type:    token.T_END_HEREDOC,
			Literal: l.input[startIdx:l.pos],
			Pos:     pos,
		}
	}

	// Variable interpolation
	if l.ch == '$' && isIdentStart(l.peekChar()) {
		return l.scanVariableInString()
	}

	if l.ch == '{' && l.peekChar() == '$' {
		l.readChar()
		l.pushState(StateInScripting)
		return TokenInfo{Type: token.T_CURLY_OPEN, Literal: "{", Pos: pos}
	}

	if l.ch == '$' && l.peekChar() == '{' {
		l.readChar()
		l.readChar()
		l.pushState(StateLookingForVarname)
		return TokenInfo{Type: token.T_DOLLAR_OPEN_CURLY_BRACES, Literal: "${", Pos: pos}
	}

	// Regular content
	startIdx := l.pos
	for l.ch != 0 {
		if l.checkHeredocEnd() {
			break
		}
		if l.ch == '$' && isIdentStart(l.peekChar()) {
			break
		}
		if l.ch == '{' && l.peekChar() == '$' {
			break
		}
		if l.ch == '$' && l.peekChar() == '{' {
			break
		}
		l.readChar()
	}

	return TokenInfo{
		Type:    token.T_ENCAPSED_AND_WHITESPACE,
		Literal: l.input[startIdx:l.pos],
		Pos:     pos,
	}
}

func (l *Lexer) scanNowdoc() TokenInfo {
	pos := l.currentPos()

	// Check for end label
	if l.checkHeredocEnd() {
		startIdx := l.pos
		for isIdentChar(l.ch) {
			l.readChar()
		}
		l.state = StateInScripting
		return TokenInfo{
			Type:    token.T_END_HEREDOC,
			Literal: l.input[startIdx:l.pos],
			Pos:     pos,
		}
	}

	// Read content until end label
	startIdx := l.pos
	for l.ch != 0 {
		if l.checkHeredocEnd() {
			break
		}
		l.readChar()
	}

	return TokenInfo{
		Type:    token.T_ENCAPSED_AND_WHITESPACE,
		Literal: l.input[startIdx:l.pos],
		Pos:     pos,
	}
}

func (l *Lexer) checkHeredocEnd() bool {
	if l.pos+len(l.heredocLabel) > len(l.input) {
		return false
	}

	// Check if we're at the start of a line or the content matches
	label := l.heredocLabel
	remaining := l.input[l.pos:]

	if strings.HasPrefix(remaining, label) {
		// Check what follows the label
		afterLabel := l.pos + len(label)
		if afterLabel >= len(l.input) {
			return true
		}
		next := l.input[afterLabel]
		// Label must be followed by ; or newline or end of file
		if next == ';' || next == '\n' || next == '\r' {
			return true
		}
	}

	return false
}

func (l *Lexer) scanVariable() TokenInfo {
	pos := l.currentPos()
	startIdx := l.pos
	l.readChar() // $

	for isIdentChar(l.ch) {
		l.readChar()
	}

	return TokenInfo{
		Type:    token.T_VARIABLE,
		Literal: l.input[startIdx:l.pos],
		Pos:     pos,
	}
}

func (l *Lexer) scanNumber() TokenInfo {
	pos := l.currentPos()
	startIdx := l.pos

	if l.ch == '0' {
		l.readChar()
		// Hexadecimal
		if l.ch == 'x' || l.ch == 'X' {
			l.readChar()
			for isHexDigit(l.ch) || l.ch == '_' {
				l.readChar()
			}
			return TokenInfo{
				Type:    token.T_LNUMBER,
				Literal: l.input[startIdx:l.pos],
				Pos:     pos,
			}
		}
		// Binary
		if l.ch == 'b' || l.ch == 'B' {
			l.readChar()
			for l.ch == '0' || l.ch == '1' || l.ch == '_' {
				l.readChar()
			}
			return TokenInfo{
				Type:    token.T_LNUMBER,
				Literal: l.input[startIdx:l.pos],
				Pos:     pos,
			}
		}
		// Octal with explicit prefix
		if l.ch == 'o' || l.ch == 'O' {
			l.readChar()
			for isOctalDigit(l.ch) || l.ch == '_' {
				l.readChar()
			}
			return TokenInfo{
				Type:    token.T_LNUMBER,
				Literal: l.input[startIdx:l.pos],
				Pos:     pos,
			}
		}
	}

	// Read integer part
	isFloat := false
	if l.ch != '.' {
		for isDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
	}

	// Check for decimal point
	if l.ch == '.' {
		if isDigit(l.peekChar()) || l.pos == startIdx {
			isFloat = true
			l.readChar() // .
			for isDigit(l.ch) || l.ch == '_' {
				l.readChar()
			}
		} else if !isDigit(l.peekChar()) && l.peekChar() != '.' {
			// Could be method call like 1.toString()
			// Check if followed by identifier
			if isIdentStart(l.peekChar()) {
				// This is just an integer
				return TokenInfo{
					Type:    token.T_LNUMBER,
					Literal: l.input[startIdx:l.pos],
					Pos:     pos,
				}
			}
			// 1. is a valid float
			isFloat = true
			l.readChar()
		}
	}

	// Check for exponent
	if l.ch == 'e' || l.ch == 'E' {
		isFloat = true
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) || l.ch == '_' {
			l.readChar()
		}
	}

	tokType := token.T_LNUMBER
	if isFloat {
		tokType = token.T_DNUMBER
	}

	return TokenInfo{
		Type:    tokType,
		Literal: l.input[startIdx:l.pos],
		Pos:     pos,
	}
}

func (l *Lexer) scanIdentifier() TokenInfo {
	pos := l.currentPos()
	startIdx := l.pos

	// Handle multi-byte characters
	for {
		if l.ch == 0 {
			break
		}
		if l.ch < 0x80 {
			if !isIdentChar(l.ch) {
				break
			}
			l.readChar()
		} else {
			// Multi-byte UTF-8 character
			r, size := utf8.DecodeRuneInString(l.input[l.pos:])
			if r == utf8.RuneError || !isIdentRuneContinue(r) {
				break
			}
			for i := 0; i < size; i++ {
				l.readChar()
			}
		}
	}

	literal := l.input[startIdx:l.pos]

	// Check for yield from
	if strings.ToLower(literal) == "yield" {
		// Look ahead for "from"
		savedPos := l.pos
		savedReadPos := l.readPos
		savedCh := l.ch
		savedLine := l.line
		savedColumn := l.column

		// Skip whitespace
		for isWhitespace(l.ch) {
			l.readChar()
		}

		// Check for "from"
		fromStart := l.pos
		for isIdentChar(l.ch) {
			l.readChar()
		}
		if strings.ToLower(l.input[fromStart:l.pos]) == "from" {
			return TokenInfo{
				Type:    token.T_YIELD_FROM,
				Literal: l.input[startIdx:l.pos],
				Pos:     pos,
			}
		}

		// Restore position
		l.pos = savedPos
		l.readPos = savedReadPos
		l.ch = savedCh
		l.line = savedLine
		l.column = savedColumn
	}

	// Check for namespace\Name (relative name)
	if strings.ToLower(literal) == "namespace" && l.ch == '\\' {
		for l.ch == '\\' || isIdentChar(l.ch) {
			if l.ch == '\\' {
				l.readChar()
			}
			for isIdentChar(l.ch) {
				l.readChar()
			}
		}
		return TokenInfo{
			Type:    token.T_NAME_RELATIVE,
			Literal: l.input[startIdx:l.pos],
			Pos:     pos,
		}
	}

	// Check for qualified name (Name\SubName)
	if l.ch == '\\' {
		for l.ch == '\\' {
			l.readChar()
			for isIdentChar(l.ch) {
				l.readChar()
			}
		}
		return TokenInfo{
			Type:    token.T_NAME_QUALIFIED,
			Literal: l.input[startIdx:l.pos],
			Pos:     pos,
		}
	}

	// Look up keyword
	tok := token.LookupIdent(literal)

	return TokenInfo{
		Type:    tok,
		Literal: literal,
		Pos:     pos,
	}
}

func (l *Lexer) scanNamespaceName() TokenInfo {
	pos := l.currentPos()
	startIdx := l.pos
	l.readChar() // \

	// Read the name parts
	for {
		for isIdentChar(l.ch) {
			l.readChar()
		}
		if l.ch == '\\' {
			l.readChar()
		} else {
			break
		}
	}

	return TokenInfo{
		Type:    token.T_NAME_FULLY_QUALIFIED,
		Literal: l.input[startIdx:l.pos],
		Pos:     pos,
	}
}

func (l *Lexer) scanOperator() TokenInfo {
	pos := l.currentPos()

	switch l.ch {
	case '+':
		if l.peekChar() == '+' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_INC, Literal: "++", Pos: pos}
		}
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_PLUS_EQUAL, Literal: "+=", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.PLUS, Literal: "+", Pos: pos}

	case '-':
		if l.peekChar() == '-' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_DEC, Literal: "--", Pos: pos}
		}
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_MINUS_EQUAL, Literal: "-=", Pos: pos}
		}
		if l.peekChar() == '>' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_OBJECT_OPERATOR, Literal: "->", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.MINUS, Literal: "-", Pos: pos}

	case '*':
		if l.peekChar() == '*' {
			l.readChar()
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return TokenInfo{Type: token.T_POW_EQUAL, Literal: "**=", Pos: pos}
			}
			return TokenInfo{Type: token.T_POW, Literal: "**", Pos: pos}
		}
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_MUL_EQUAL, Literal: "*=", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.ASTERISK, Literal: "*", Pos: pos}

	case '/':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_DIV_EQUAL, Literal: "/=", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.SLASH, Literal: "/", Pos: pos}

	case '%':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_MOD_EQUAL, Literal: "%=", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.PERCENT, Literal: "%", Pos: pos}

	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_BOOLEAN_AND, Literal: "&&", Pos: pos}
		}
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_AND_EQUAL, Literal: "&=", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.AMPERSAND, Literal: "&", Pos: pos}

	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_BOOLEAN_OR, Literal: "||", Pos: pos}
		}
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_OR_EQUAL, Literal: "|=", Pos: pos}
		}
		if l.peekChar() == '>' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_PIPE, Literal: "|>", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.PIPE, Literal: "|", Pos: pos}

	case '^':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_XOR_EQUAL, Literal: "^=", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.CARET, Literal: "^", Pos: pos}

	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return TokenInfo{Type: token.T_IS_IDENTICAL, Literal: "===", Pos: pos}
			}
			return TokenInfo{Type: token.T_IS_EQUAL, Literal: "==", Pos: pos}
		}
		if l.peekChar() == '>' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_DOUBLE_ARROW, Literal: "=>", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.EQUALS, Literal: "=", Pos: pos}

	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return TokenInfo{Type: token.T_IS_NOT_IDENTICAL, Literal: "!==", Pos: pos}
			}
			return TokenInfo{Type: token.T_IS_NOT_EQUAL, Literal: "!=", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.EXCLAMATION, Literal: "!", Pos: pos}

	case '<':
		if l.peekChar() == '<' {
			l.readChar()
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return TokenInfo{Type: token.T_SL_EQUAL, Literal: "<<=", Pos: pos}
			}
			return TokenInfo{Type: token.T_SL, Literal: "<<", Pos: pos}
		}
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			if l.ch == '>' {
				l.readChar()
				return TokenInfo{Type: token.T_SPACESHIP, Literal: "<=>", Pos: pos}
			}
			return TokenInfo{Type: token.T_IS_SMALLER_OR_EQUAL, Literal: "<=", Pos: pos}
		}
		if l.peekChar() == '>' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_IS_NOT_EQUAL, Literal: "<>", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.LESS, Literal: "<", Pos: pos}

	case '>':
		if l.peekChar() == '>' {
			l.readChar()
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return TokenInfo{Type: token.T_SR_EQUAL, Literal: ">>=", Pos: pos}
			}
			return TokenInfo{Type: token.T_SR, Literal: ">>", Pos: pos}
		}
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_IS_GREATER_OR_EQUAL, Literal: ">=", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.GREATER, Literal: ">", Pos: pos}

	case '?':
		if l.peekChar() == '?' {
			l.readChar()
			l.readChar()
			if l.ch == '=' {
				l.readChar()
				return TokenInfo{Type: token.T_COALESCE_EQUAL, Literal: "??=", Pos: pos}
			}
			return TokenInfo{Type: token.T_COALESCE, Literal: "??", Pos: pos}
		}
		if l.peekChar() == '-' && l.peekCharN(2) == '>' {
			l.readChar()
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_NULLSAFE_OBJECT_OPERATOR, Literal: "?->", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.QUESTION, Literal: "?", Pos: pos}

	case '.':
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_CONCAT_EQUAL, Literal: ".=", Pos: pos}
		}
		if l.peekChar() == '.' && l.peekCharN(2) == '.' {
			l.readChar()
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_ELLIPSIS, Literal: "...", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.DOT, Literal: ".", Pos: pos}

	case ':':
		if l.peekChar() == ':' {
			l.readChar()
			l.readChar()
			return TokenInfo{Type: token.T_PAAMAYIM_NEKUDOTAYIM, Literal: "::", Pos: pos}
		}
		l.readChar()
		return TokenInfo{Type: token.COLON, Literal: ":", Pos: pos}

	case '(':
		// Check for type casts
		if cast := l.tryScanCast(); cast.Type != token.ILLEGAL {
			return cast
		}
		l.readChar()
		return TokenInfo{Type: token.LPAREN, Literal: "(", Pos: pos}

	case ')':
		l.readChar()
		return TokenInfo{Type: token.RPAREN, Literal: ")", Pos: pos}

	case '[':
		l.readChar()
		return TokenInfo{Type: token.LBRACKET, Literal: "[", Pos: pos}

	case ']':
		l.readChar()
		return TokenInfo{Type: token.RBRACKET, Literal: "]", Pos: pos}

	case '{':
		l.readChar()
		return TokenInfo{Type: token.LBRACE, Literal: "{", Pos: pos}

	case '}':
		// Check if we're closing a string interpolation
		if len(l.stateStack) > 0 {
			l.popState()
		}
		l.readChar()
		return TokenInfo{Type: token.RBRACE, Literal: "}", Pos: pos}

	case ';':
		l.readChar()
		return TokenInfo{Type: token.SEMICOLON, Literal: ";", Pos: pos}

	case ',':
		l.readChar()
		return TokenInfo{Type: token.COMMA, Literal: ",", Pos: pos}

	case '@':
		l.readChar()
		return TokenInfo{Type: token.AT, Literal: "@", Pos: pos}

	case '~':
		l.readChar()
		return TokenInfo{Type: token.TILDE, Literal: "~", Pos: pos}

	case '\\':
		return l.scanNamespaceName()
	}

	// Unknown character
	ch := l.ch
	l.readChar()
	return TokenInfo{Type: token.ILLEGAL, Literal: string(ch), Pos: pos}
}

func (l *Lexer) tryScanCast() TokenInfo {
	pos := l.currentPos()
	savedPos := l.pos
	savedReadPos := l.readPos
	savedCh := l.ch
	savedLine := l.line
	savedColumn := l.column

	l.readChar() // (

	// Skip whitespace
	for isWhitespace(l.ch) {
		l.readChar()
	}

	// Read type name
	startIdx := l.pos
	for isIdentChar(l.ch) {
		l.readChar()
	}
	typeName := strings.ToLower(l.input[startIdx:l.pos])

	// Skip whitespace
	for isWhitespace(l.ch) {
		l.readChar()
	}

	if l.ch != ')' {
		// Not a cast, restore position
		l.pos = savedPos
		l.readPos = savedReadPos
		l.ch = savedCh
		l.line = savedLine
		l.column = savedColumn
		return TokenInfo{Type: token.ILLEGAL}
	}

	l.readChar() // )

	literal := l.input[pos.Offset:l.pos]

	switch typeName {
	case "int", "integer":
		return TokenInfo{Type: token.T_INT_CAST, Literal: literal, Pos: pos}
	case "float", "double", "real":
		return TokenInfo{Type: token.T_DOUBLE_CAST, Literal: literal, Pos: pos}
	case "string", "binary":
		return TokenInfo{Type: token.T_STRING_CAST, Literal: literal, Pos: pos}
	case "array":
		return TokenInfo{Type: token.T_ARRAY_CAST, Literal: literal, Pos: pos}
	case "object":
		return TokenInfo{Type: token.T_OBJECT_CAST, Literal: literal, Pos: pos}
	case "bool", "boolean":
		return TokenInfo{Type: token.T_BOOL_CAST, Literal: literal, Pos: pos}
	case "unset":
		return TokenInfo{Type: token.T_UNSET_CAST, Literal: literal, Pos: pos}
	}

	// Not a cast, restore position
	l.pos = savedPos
	l.readPos = savedReadPos
	l.ch = savedCh
	l.line = savedLine
	l.column = savedColumn
	return TokenInfo{Type: token.ILLEGAL}
}

// Helper functions

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isOctalDigit(ch byte) bool {
	return ch >= '0' && ch <= '7'
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch >= 0x80
}

func isIdentChar(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch)
}

func isIdentRuneContinue(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
