package lexer

import (
	"testing"

	"github.com/alexisbouchez/phpgo/token"
)

func TestTokenizeAll(t *testing.T) {
	input := `<?php $x = 1;`
	tokens := TokenizeAll(input)

	if len(tokens) == 0 {
		t.Fatal("Expected tokens")
	}

	// Last token should be EOF
	if tokens[len(tokens)-1].Type != token.EOF {
		t.Errorf("Expected last token to be EOF, got %s", tokens[len(tokens)-1].Type)
	}
}

func TestTokenizeFiltered(t *testing.T) {
	input := `<?php $x = 1;`
	tokens := TokenizeFiltered(input)

	// Should not contain whitespace tokens
	for _, tok := range tokens {
		if tok.Type == token.WHITESPACE {
			t.Error("TokenizeFiltered should not return whitespace tokens")
		}
	}
}

func TestTokenizeCode(t *testing.T) {
	// TokenizeCode adds <?php prefix
	tokens := TokenizeCode("$x = 1;")

	if tokens[0].Type != token.T_OPEN_TAG {
		t.Errorf("Expected first token to be T_OPEN_TAG, got %s", tokens[0].Type)
	}

	foundVar := false
	for _, tok := range tokens {
		if tok.Type == token.T_VARIABLE && tok.Literal == "$x" {
			foundVar = true
		}
	}
	if !foundVar {
		t.Error("Expected to find $x variable")
	}
}

func TestCountTokens(t *testing.T) {
	input := `<?php $a = $b + $c;`
	counts := CountTokens(input)

	if counts[token.T_VARIABLE] != 3 {
		t.Errorf("Expected 3 variables, got %d", counts[token.T_VARIABLE])
	}
}

func TestFindTokens(t *testing.T) {
	input := `<?php $a = 1; $b = 2;`
	vars := FindTokens(input, token.T_VARIABLE)

	if len(vars) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(vars))
	}
}

func TestHasToken(t *testing.T) {
	input := `<?php class Foo {}`

	if !HasToken(input, token.T_CLASS) {
		t.Error("Expected to find T_CLASS")
	}

	if HasToken(input, token.T_INTERFACE) {
		t.Error("Should not find T_INTERFACE")
	}
}

func TestValidateSyntax(t *testing.T) {
	valid := `<?php $x = 1;`
	if tok := ValidateSyntax(valid); tok != nil {
		t.Errorf("Expected valid syntax, got illegal token: %s", tok.Literal)
	}
}

func TestExtractVariables(t *testing.T) {
	input := `<?php $foo = $bar + $baz;`
	vars := ExtractVariables(input)

	if len(vars) != 3 {
		t.Errorf("Expected 3 variables, got %d", len(vars))
	}

	expected := map[string]bool{"$foo": true, "$bar": true, "$baz": true}
	for _, v := range vars {
		if !expected[v] {
			t.Errorf("Unexpected variable: %s", v)
		}
	}
}

func TestExtractStrings(t *testing.T) {
	input := `<?php $a = "hello"; $b = 'world';`
	strs := ExtractStrings(input)

	if len(strs) != 2 {
		t.Errorf("Expected 2 strings, got %d", len(strs))
	}
}

func TestExtractFunctionCalls(t *testing.T) {
	input := `<?php foo(); bar($x); baz(1, 2);`
	calls := ExtractFunctionCalls(input)

	if len(calls) != 3 {
		t.Errorf("Expected 3 function calls, got %d", len(calls))
	}

	expected := map[string]bool{"foo": true, "bar": true, "baz": true}
	for _, c := range calls {
		if !expected[c] {
			t.Errorf("Unexpected function call: %s", c)
		}
	}
}

func TestIsKeyword(t *testing.T) {
	keywords := []string{"if", "else", "class", "function", "return"}
	for _, kw := range keywords {
		if !IsKeyword(kw) {
			t.Errorf("Expected %q to be a keyword", kw)
		}
	}

	nonKeywords := []string{"foo", "bar", "myFunction", "MyClass"}
	for _, nk := range nonKeywords {
		if IsKeyword(nk) {
			t.Errorf("Expected %q to NOT be a keyword", nk)
		}
	}
}
