package lexer

import (
	"github.com/alexisbouchez/phpgo/token"
)

// TokenizeAll tokenizes the entire input and returns all tokens including EOF.
func TokenizeAll(input string) []TokenInfo {
	l := New(input)
	tokens := []TokenInfo{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}

// TokenizeFiltered tokenizes the input and returns only non-whitespace tokens.
func TokenizeFiltered(input string) []TokenInfo {
	l := New(input)
	tokens := []TokenInfo{}
	for {
		tok := l.NextToken()
		if tok.Type != token.WHITESPACE {
			tokens = append(tokens, tok)
		}
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}

// TokenizeCode tokenizes PHP code that doesn't include the opening tag.
// It wraps the input in <?php ... for convenience.
func TokenizeCode(input string) []TokenInfo {
	return TokenizeAll("<?php " + input)
}

// CountTokens returns the number of tokens of each type in the input.
func CountTokens(input string) map[token.Token]int {
	l := New(input)
	counts := make(map[token.Token]int)
	for {
		tok := l.NextToken()
		counts[tok.Type]++
		if tok.Type == token.EOF {
			break
		}
	}
	return counts
}

// FindTokens returns all tokens of the specified types.
func FindTokens(input string, types ...token.Token) []TokenInfo {
	typeSet := make(map[token.Token]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	l := New(input)
	tokens := []TokenInfo{}
	for {
		tok := l.NextToken()
		if typeSet[tok.Type] {
			tokens = append(tokens, tok)
		}
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}

// HasToken checks if the input contains a token of the specified type.
func HasToken(input string, t token.Token) bool {
	l := New(input)
	for {
		tok := l.NextToken()
		if tok.Type == t {
			return true
		}
		if tok.Type == token.EOF {
			break
		}
	}
	return false
}

// ValidateSyntax performs basic lexical validation.
// Returns nil if no illegal tokens are found, otherwise returns the first illegal token.
func ValidateSyntax(input string) *TokenInfo {
	l := New(input)
	for {
		tok := l.NextToken()
		if tok.Type == token.ILLEGAL || tok.Type == token.T_ERROR || tok.Type == token.T_BAD_CHARACTER {
			return &tok
		}
		if tok.Type == token.EOF {
			break
		}
	}
	return nil
}

// ExtractVariables returns all variable names found in the input.
func ExtractVariables(input string) []string {
	tokens := FindTokens(input, token.T_VARIABLE)
	vars := make([]string, len(tokens))
	for i, tok := range tokens {
		vars[i] = tok.Literal
	}
	return vars
}

// ExtractStrings returns all string literals found in the input.
func ExtractStrings(input string) []string {
	tokens := FindTokens(input, token.T_CONSTANT_ENCAPSED_STRING)
	strs := make([]string, len(tokens))
	for i, tok := range tokens {
		strs[i] = tok.Literal
	}
	return strs
}

// ExtractFunctionCalls returns identifiers that appear before opening parentheses,
// which are likely function calls.
func ExtractFunctionCalls(input string) []string {
	all := TokenizeFiltered(input)
	calls := []string{}

	for i := 0; i < len(all)-1; i++ {
		if all[i].Type == token.T_STRING && all[i+1].Type == token.LPAREN {
			calls = append(calls, all[i].Literal)
		}
	}

	return calls
}

// IsKeyword checks if the given identifier is a PHP keyword.
func IsKeyword(ident string) bool {
	tok := token.LookupIdent(ident)
	return tok != token.T_STRING
}
