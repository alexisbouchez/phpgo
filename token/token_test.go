package token

import "testing"

func TestTokenString(t *testing.T) {
	tests := []struct {
		tok      Token
		expected string
	}{
		{T_IF, "T_IF"},
		{T_ELSE, "T_ELSE"},
		{T_FUNCTION, "T_FUNCTION"},
		{T_CLASS, "T_CLASS"},
		{T_LNUMBER, "T_LNUMBER"},
		{T_DNUMBER, "T_DNUMBER"},
		{T_VARIABLE, "T_VARIABLE"},
		{T_STRING, "T_STRING"},
		{T_OPEN_TAG, "T_OPEN_TAG"},
		{T_CLOSE_TAG, "T_CLOSE_TAG"},
		{SEMICOLON, ";"},
		{LPAREN, "("},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{RBRACE, "}"},
		{EOF, "EOF"},
	}

	for _, tt := range tests {
		if got := tt.tok.String(); got != tt.expected {
			t.Errorf("Token(%d).String() = %q, want %q", tt.tok, got, tt.expected)
		}
	}
}

func TestLookupIdent(t *testing.T) {
	tests := []struct {
		ident    string
		expected Token
	}{
		{"if", T_IF},
		{"else", T_ELSE},
		{"elseif", T_ELSEIF},
		{"while", T_WHILE},
		{"for", T_FOR},
		{"foreach", T_FOREACH},
		{"function", T_FUNCTION},
		{"class", T_CLASS},
		{"trait", T_TRAIT},
		{"interface", T_INTERFACE},
		{"enum", T_ENUM},
		{"return", T_RETURN},
		{"echo", T_ECHO},
		{"print", T_PRINT},
		{"public", T_PUBLIC},
		{"private", T_PRIVATE},
		{"protected", T_PROTECTED},
		{"static", T_STATIC},
		{"final", T_FINAL},
		{"abstract", T_ABSTRACT},
		{"namespace", T_NAMESPACE},
		{"use", T_USE},
		{"const", T_CONST},
		{"exit", T_EXIT},
		{"die", T_EXIT},
		{"or", T_LOGICAL_OR},
		{"and", T_LOGICAL_AND},
		{"xor", T_LOGICAL_XOR},
		{"__LINE__", T_LINE},
		{"__FILE__", T_FILE},
		{"__DIR__", T_DIR},
		{"__CLASS__", T_CLASS_C},
		{"__FUNCTION__", T_FUNC_C},
		{"__METHOD__", T_METHOD_C},
		{"__NAMESPACE__", T_NS_C},
		// Non-keywords should return T_STRING
		{"foo", T_STRING},
		{"myFunction", T_STRING},
		{"MyClass", T_STRING},
		{"variable123", T_STRING},
	}

	for _, tt := range tests {
		if got := LookupIdent(tt.ident); got != tt.expected {
			t.Errorf("LookupIdent(%q) = %s, want %s", tt.ident, got, tt.expected)
		}
	}
}

func TestIsKeyword(t *testing.T) {
	keywords := []Token{
		T_IF, T_ELSE, T_WHILE, T_FOR, T_FUNCTION, T_CLASS,
		T_RETURN, T_ECHO, T_PRINT, T_PUBLIC, T_PRIVATE,
	}

	for _, tok := range keywords {
		if !tok.IsKeyword() {
			t.Errorf("Expected %s to be a keyword", tok)
		}
	}

	nonKeywords := []Token{
		T_VARIABLE, T_LNUMBER, T_DNUMBER, T_STRING,
		T_CONSTANT_ENCAPSED_STRING, SEMICOLON, LPAREN,
	}

	for _, tok := range nonKeywords {
		if tok.IsKeyword() {
			t.Errorf("Expected %s to NOT be a keyword", tok)
		}
	}
}

func TestIsLiteral(t *testing.T) {
	literals := []Token{
		T_LNUMBER, T_DNUMBER, T_STRING, T_VARIABLE,
		T_CONSTANT_ENCAPSED_STRING, T_ENCAPSED_AND_WHITESPACE,
	}

	for _, tok := range literals {
		if !tok.IsLiteral() {
			t.Errorf("Expected %s to be a literal", tok)
		}
	}

	nonLiterals := []Token{
		T_IF, T_ELSE, T_FUNCTION, SEMICOLON, LPAREN,
		T_PLUS_EQUAL, T_OBJECT_OPERATOR,
	}

	for _, tok := range nonLiterals {
		if tok.IsLiteral() {
			t.Errorf("Expected %s to NOT be a literal", tok)
		}
	}
}

func TestIsOperator(t *testing.T) {
	operators := []Token{
		T_PLUS_EQUAL, T_MINUS_EQUAL, T_MUL_EQUAL,
		T_IS_EQUAL, T_IS_IDENTICAL, T_BOOLEAN_AND,
		T_DOUBLE_ARROW, T_OBJECT_OPERATOR, T_ELLIPSIS,
	}

	for _, tok := range operators {
		if !tok.IsOperator() {
			t.Errorf("Expected %s to be an operator", tok)
		}
	}

	nonOperators := []Token{
		T_IF, T_ELSE, T_VARIABLE, T_LNUMBER, SEMICOLON,
	}

	for _, tok := range nonOperators {
		if tok.IsOperator() {
			t.Errorf("Expected %s to NOT be an operator", tok)
		}
	}
}
