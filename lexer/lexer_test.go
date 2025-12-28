package lexer

import (
	"testing"

	"github.com/alexisbouchez/phpgo/token"
)

// TokenResult holds expected token results for testing
type TokenResult struct {
	Type    token.Token
	Literal string
}

func TestOpenTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenResult
	}{
		{
			name:  "standard open tag",
			input: "<?php",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php"},
				{token.EOF, ""},
			},
		},
		{
			name:  "open tag with newline",
			input: "<?php\n",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php\n"},
				{token.EOF, ""},
			},
		},
		{
			name:  "open tag with space",
			input: "<?php ",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.EOF, ""},
			},
		},
		{
			name:  "short echo tag",
			input: "<?=",
			expected: []TokenResult{
				{token.T_OPEN_TAG_WITH_ECHO, "<?="},
				{token.EOF, ""},
			},
		},
		{
			name:  "inline HTML before tag",
			input: "<html><?php",
			expected: []TokenResult{
				{token.T_INLINE_HTML, "<html>"},
				{token.T_OPEN_TAG, "<?php"},
				{token.EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.Type {
					t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
						i, exp.Type, tok.Type)
				}
				if tok.Literal != exp.Literal {
					t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
						i, exp.Literal, tok.Literal)
				}
			}
		})
	}
}

func TestCloseTag(t *testing.T) {
	input := "<?php echo 1; ?>"
	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php "},
		{token.T_ECHO, "echo"},
		{token.WHITESPACE, " "},
		{token.T_LNUMBER, "1"},
		{token.SEMICOLON, ";"},
		{token.WHITESPACE, " "},
		{token.T_CLOSE_TAG, "?>"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestKeywords(t *testing.T) {
	input := `<?php
if else elseif endif
while endwhile do for endfor foreach endforeach
switch endswitch case default match
break continue goto return
try catch finally throw
function fn class trait interface enum
extends implements new clone instanceof
public protected private readonly static abstract final var
global use unset isset empty eval
include include_once require require_once
echo print list array callable declare as insteadof
namespace const exit die
or and xor
`

	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php\n"},
		{token.T_IF, "if"},
		{token.WHITESPACE, " "},
		{token.T_ELSE, "else"},
		{token.WHITESPACE, " "},
		{token.T_ELSEIF, "elseif"},
		{token.WHITESPACE, " "},
		{token.T_ENDIF, "endif"},
		{token.WHITESPACE, "\n"},
		{token.T_WHILE, "while"},
		{token.WHITESPACE, " "},
		{token.T_ENDWHILE, "endwhile"},
		{token.WHITESPACE, " "},
		{token.T_DO, "do"},
		{token.WHITESPACE, " "},
		{token.T_FOR, "for"},
		{token.WHITESPACE, " "},
		{token.T_ENDFOR, "endfor"},
		{token.WHITESPACE, " "},
		{token.T_FOREACH, "foreach"},
		{token.WHITESPACE, " "},
		{token.T_ENDFOREACH, "endforeach"},
		{token.WHITESPACE, "\n"},
		{token.T_SWITCH, "switch"},
		{token.WHITESPACE, " "},
		{token.T_ENDSWITCH, "endswitch"},
		{token.WHITESPACE, " "},
		{token.T_CASE, "case"},
		{token.WHITESPACE, " "},
		{token.T_DEFAULT, "default"},
		{token.WHITESPACE, " "},
		{token.T_MATCH, "match"},
		{token.WHITESPACE, "\n"},
		{token.T_BREAK, "break"},
		{token.WHITESPACE, " "},
		{token.T_CONTINUE, "continue"},
		{token.WHITESPACE, " "},
		{token.T_GOTO, "goto"},
		{token.WHITESPACE, " "},
		{token.T_RETURN, "return"},
		{token.WHITESPACE, "\n"},
		{token.T_TRY, "try"},
		{token.WHITESPACE, " "},
		{token.T_CATCH, "catch"},
		{token.WHITESPACE, " "},
		{token.T_FINALLY, "finally"},
		{token.WHITESPACE, " "},
		{token.T_THROW, "throw"},
		{token.WHITESPACE, "\n"},
		{token.T_FUNCTION, "function"},
		{token.WHITESPACE, " "},
		{token.T_FN, "fn"},
		{token.WHITESPACE, " "},
		{token.T_CLASS, "class"},
		{token.WHITESPACE, " "},
		{token.T_TRAIT, "trait"},
		{token.WHITESPACE, " "},
		{token.T_INTERFACE, "interface"},
		{token.WHITESPACE, " "},
		{token.T_ENUM, "enum"},
		{token.WHITESPACE, "\n"},
		{token.T_EXTENDS, "extends"},
		{token.WHITESPACE, " "},
		{token.T_IMPLEMENTS, "implements"},
		{token.WHITESPACE, " "},
		{token.T_NEW, "new"},
		{token.WHITESPACE, " "},
		{token.T_CLONE, "clone"},
		{token.WHITESPACE, " "},
		{token.T_INSTANCEOF, "instanceof"},
		{token.WHITESPACE, "\n"},
		{token.T_PUBLIC, "public"},
		{token.WHITESPACE, " "},
		{token.T_PROTECTED, "protected"},
		{token.WHITESPACE, " "},
		{token.T_PRIVATE, "private"},
		{token.WHITESPACE, " "},
		{token.T_READONLY, "readonly"},
		{token.WHITESPACE, " "},
		{token.T_STATIC, "static"},
		{token.WHITESPACE, " "},
		{token.T_ABSTRACT, "abstract"},
		{token.WHITESPACE, " "},
		{token.T_FINAL, "final"},
		{token.WHITESPACE, " "},
		{token.T_VAR, "var"},
		{token.WHITESPACE, "\n"},
		{token.T_GLOBAL, "global"},
		{token.WHITESPACE, " "},
		{token.T_USE, "use"},
		{token.WHITESPACE, " "},
		{token.T_UNSET, "unset"},
		{token.WHITESPACE, " "},
		{token.T_ISSET, "isset"},
		{token.WHITESPACE, " "},
		{token.T_EMPTY, "empty"},
		{token.WHITESPACE, " "},
		{token.T_EVAL, "eval"},
		{token.WHITESPACE, "\n"},
		{token.T_INCLUDE, "include"},
		{token.WHITESPACE, " "},
		{token.T_INCLUDE_ONCE, "include_once"},
		{token.WHITESPACE, " "},
		{token.T_REQUIRE, "require"},
		{token.WHITESPACE, " "},
		{token.T_REQUIRE_ONCE, "require_once"},
		{token.WHITESPACE, "\n"},
		{token.T_ECHO, "echo"},
		{token.WHITESPACE, " "},
		{token.T_PRINT, "print"},
		{token.WHITESPACE, " "},
		{token.T_LIST, "list"},
		{token.WHITESPACE, " "},
		{token.T_ARRAY, "array"},
		{token.WHITESPACE, " "},
		{token.T_CALLABLE, "callable"},
		{token.WHITESPACE, " "},
		{token.T_DECLARE, "declare"},
		{token.WHITESPACE, " "},
		{token.T_AS, "as"},
		{token.WHITESPACE, " "},
		{token.T_INSTEADOF, "insteadof"},
		{token.WHITESPACE, "\n"},
		{token.T_NAMESPACE, "namespace"},
		{token.WHITESPACE, " "},
		{token.T_CONST, "const"},
		{token.WHITESPACE, " "},
		{token.T_EXIT, "exit"},
		{token.WHITESPACE, " "},
		{token.T_EXIT, "die"},
		{token.WHITESPACE, "\n"},
		{token.T_LOGICAL_OR, "or"},
		{token.WHITESPACE, " "},
		{token.T_LOGICAL_AND, "and"},
		{token.WHITESPACE, " "},
		{token.T_LOGICAL_XOR, "xor"},
		{token.WHITESPACE, "\n"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestMagicConstants(t *testing.T) {
	input := `<?php __LINE__ __FILE__ __DIR__ __CLASS__ __TRAIT__ __METHOD__ __FUNCTION__ __NAMESPACE__`

	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php "},
		{token.T_LINE, "__LINE__"},
		{token.WHITESPACE, " "},
		{token.T_FILE, "__FILE__"},
		{token.WHITESPACE, " "},
		{token.T_DIR, "__DIR__"},
		{token.WHITESPACE, " "},
		{token.T_CLASS_C, "__CLASS__"},
		{token.WHITESPACE, " "},
		{token.T_TRAIT_C, "__TRAIT__"},
		{token.WHITESPACE, " "},
		{token.T_METHOD_C, "__METHOD__"},
		{token.WHITESPACE, " "},
		{token.T_FUNC_C, "__FUNCTION__"},
		{token.WHITESPACE, " "},
		{token.T_NS_C, "__NAMESPACE__"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestVariables(t *testing.T) {
	input := `<?php $var $_var $var123 $Var $camelCase $snake_case`

	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php "},
		{token.T_VARIABLE, "$var"},
		{token.WHITESPACE, " "},
		{token.T_VARIABLE, "$_var"},
		{token.WHITESPACE, " "},
		{token.T_VARIABLE, "$var123"},
		{token.WHITESPACE, " "},
		{token.T_VARIABLE, "$Var"},
		{token.WHITESPACE, " "},
		{token.T_VARIABLE, "$camelCase"},
		{token.WHITESPACE, " "},
		{token.T_VARIABLE, "$snake_case"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestIntegerLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenResult
	}{
		{
			name:  "decimal integer",
			input: "<?php 123",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_LNUMBER, "123"},
				{token.EOF, ""},
			},
		},
		{
			name:  "zero",
			input: "<?php 0",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_LNUMBER, "0"},
				{token.EOF, ""},
			},
		},
		{
			name:  "hexadecimal",
			input: "<?php 0xFF 0x1A2B",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_LNUMBER, "0xFF"},
				{token.WHITESPACE, " "},
				{token.T_LNUMBER, "0x1A2B"},
				{token.EOF, ""},
			},
		},
		{
			name:  "octal with prefix",
			input: "<?php 0o755 0O644",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_LNUMBER, "0o755"},
				{token.WHITESPACE, " "},
				{token.T_LNUMBER, "0O644"},
				{token.EOF, ""},
			},
		},
		{
			name:  "binary",
			input: "<?php 0b1010 0B1111",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_LNUMBER, "0b1010"},
				{token.WHITESPACE, " "},
				{token.T_LNUMBER, "0B1111"},
				{token.EOF, ""},
			},
		},
		{
			name:  "underscore separators",
			input: "<?php 1_000_000 0xFF_FF 0b1010_1010",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_LNUMBER, "1_000_000"},
				{token.WHITESPACE, " "},
				{token.T_LNUMBER, "0xFF_FF"},
				{token.WHITESPACE, " "},
				{token.T_LNUMBER, "0b1010_1010"},
				{token.EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.Type {
					t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
						i, exp.Type, tok.Type)
				}
				if tok.Literal != exp.Literal {
					t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
						i, exp.Literal, tok.Literal)
				}
			}
		})
	}
}

func TestFloatLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenResult
	}{
		{
			name:  "simple float",
			input: "<?php 1.5",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_DNUMBER, "1.5"},
				{token.EOF, ""},
			},
		},
		{
			name:  "float without leading digit",
			input: "<?php .5",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_DNUMBER, ".5"},
				{token.EOF, ""},
			},
		},
		{
			name:  "float without trailing digit",
			input: "<?php 1.",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_DNUMBER, "1."},
				{token.EOF, ""},
			},
		},
		{
			name:  "scientific notation",
			input: "<?php 1e5 1E5 1e+5 1e-5",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_DNUMBER, "1e5"},
				{token.WHITESPACE, " "},
				{token.T_DNUMBER, "1E5"},
				{token.WHITESPACE, " "},
				{token.T_DNUMBER, "1e+5"},
				{token.WHITESPACE, " "},
				{token.T_DNUMBER, "1e-5"},
				{token.EOF, ""},
			},
		},
		{
			name:  "float with scientific notation",
			input: "<?php 1.5e10 .5e-3",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_DNUMBER, "1.5e10"},
				{token.WHITESPACE, " "},
				{token.T_DNUMBER, ".5e-3"},
				{token.EOF, ""},
			},
		},
		{
			name:  "float with underscore",
			input: "<?php 1_000.5_5",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_DNUMBER, "1_000.5_5"},
				{token.EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.Type {
					t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
						i, exp.Type, tok.Type)
				}
				if tok.Literal != exp.Literal {
					t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
						i, exp.Literal, tok.Literal)
				}
			}
		})
	}
}

func TestSingleQuotedStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenResult
	}{
		{
			name:  "simple string",
			input: `<?php 'hello'`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_CONSTANT_ENCAPSED_STRING, "'hello'"},
				{token.EOF, ""},
			},
		},
		{
			name:  "empty string",
			input: `<?php ''`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_CONSTANT_ENCAPSED_STRING, "''"},
				{token.EOF, ""},
			},
		},
		{
			name:  "escaped quote",
			input: `<?php 'it\'s'`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_CONSTANT_ENCAPSED_STRING, `'it\'s'`},
				{token.EOF, ""},
			},
		},
		{
			name:  "escaped backslash",
			input: `<?php 'path\\to'`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_CONSTANT_ENCAPSED_STRING, `'path\\to'`},
				{token.EOF, ""},
			},
		},
		{
			name:  "variable not interpolated",
			input: `<?php '$var'`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_CONSTANT_ENCAPSED_STRING, "'$var'"},
				{token.EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.Type {
					t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
						i, exp.Type, tok.Type)
				}
				if tok.Literal != exp.Literal {
					t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
						i, exp.Literal, tok.Literal)
				}
			}
		})
	}
}

func TestDoubleQuotedStringsSimple(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenResult
	}{
		{
			name:  "simple string without variables",
			input: `<?php "hello"`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_CONSTANT_ENCAPSED_STRING, `"hello"`},
				{token.EOF, ""},
			},
		},
		{
			name:  "empty string",
			input: `<?php ""`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_CONSTANT_ENCAPSED_STRING, `""`},
				{token.EOF, ""},
			},
		},
		{
			name:  "escape sequences",
			input: `<?php "hello\nworld\t!"`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_CONSTANT_ENCAPSED_STRING, `"hello\nworld\t!"`},
				{token.EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.Type {
					t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
						i, exp.Type, tok.Type)
				}
				if tok.Literal != exp.Literal {
					t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
						i, exp.Literal, tok.Literal)
				}
			}
		})
	}
}

func TestDoubleQuotedStringsWithVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenResult
	}{
		{
			name:  "simple variable interpolation",
			input: `<?php "Hello $name"`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.DOUBLE_QUOTE, `"`},
				{token.T_ENCAPSED_AND_WHITESPACE, "Hello "},
				{token.T_VARIABLE, "$name"},
				{token.DOUBLE_QUOTE, `"`},
				{token.EOF, ""},
			},
		},
		{
			name:  "curly brace syntax",
			input: `<?php "Hello {$name}"`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.DOUBLE_QUOTE, `"`},
				{token.T_ENCAPSED_AND_WHITESPACE, "Hello "},
				{token.T_CURLY_OPEN, "{"},
				{token.T_VARIABLE, "$name"},
				{token.RBRACE, "}"},
				{token.DOUBLE_QUOTE, `"`},
				{token.EOF, ""},
			},
		},
		{
			name:  "dollar curly syntax",
			input: `<?php "Hello ${name}"`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.DOUBLE_QUOTE, `"`},
				{token.T_ENCAPSED_AND_WHITESPACE, "Hello "},
				{token.T_DOLLAR_OPEN_CURLY_BRACES, "${"},
				{token.T_STRING_VARNAME, "name"},
				{token.RBRACE, "}"},
				{token.DOUBLE_QUOTE, `"`},
				{token.EOF, ""},
			},
		},
		{
			name:  "variable at start",
			input: `<?php "$name is here"`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.DOUBLE_QUOTE, `"`},
				{token.T_VARIABLE, "$name"},
				{token.T_ENCAPSED_AND_WHITESPACE, " is here"},
				{token.DOUBLE_QUOTE, `"`},
				{token.EOF, ""},
			},
		},
		{
			name:  "variable at end",
			input: `<?php "Hello $name"`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.DOUBLE_QUOTE, `"`},
				{token.T_ENCAPSED_AND_WHITESPACE, "Hello "},
				{token.T_VARIABLE, "$name"},
				{token.DOUBLE_QUOTE, `"`},
				{token.EOF, ""},
			},
		},
		{
			name:  "array access in string",
			input: `<?php "Value: $arr[0]"`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.DOUBLE_QUOTE, `"`},
				{token.T_ENCAPSED_AND_WHITESPACE, "Value: "},
				{token.T_VARIABLE, "$arr"},
				{token.LBRACKET, "["},
				{token.T_NUM_STRING, "0"},
				{token.RBRACKET, "]"},
				{token.DOUBLE_QUOTE, `"`},
				{token.EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.Type {
					t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
						i, exp.Type, tok.Type)
				}
				if tok.Literal != exp.Literal {
					t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
						i, exp.Literal, tok.Literal)
				}
			}
		})
	}
}

func TestHeredoc(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenResult
	}{
		{
			name: "simple heredoc",
			input: `<?php <<<EOT
Hello World
EOT;`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_START_HEREDOC, "<<<EOT\n"},
				{token.T_ENCAPSED_AND_WHITESPACE, "Hello World\n"},
				{token.T_END_HEREDOC, "EOT"},
				{token.SEMICOLON, ";"},
				{token.EOF, ""},
			},
		},
		{
			name: "heredoc with variable",
			input: `<?php <<<EOT
Hello $name
EOT;`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_START_HEREDOC, "<<<EOT\n"},
				{token.T_ENCAPSED_AND_WHITESPACE, "Hello "},
				{token.T_VARIABLE, "$name"},
				{token.T_ENCAPSED_AND_WHITESPACE, "\n"},
				{token.T_END_HEREDOC, "EOT"},
				{token.SEMICOLON, ";"},
				{token.EOF, ""},
			},
		},
		{
			name: "quoted heredoc",
			input: `<?php <<<"EOT"
Hello World
EOT;`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_START_HEREDOC, "<<<\"EOT\"\n"},
				{token.T_ENCAPSED_AND_WHITESPACE, "Hello World\n"},
				{token.T_END_HEREDOC, "EOT"},
				{token.SEMICOLON, ";"},
				{token.EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.Type {
					t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
						i, exp.Type, tok.Type)
				}
				if tok.Literal != exp.Literal {
					t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
						i, exp.Literal, tok.Literal)
				}
			}
		})
	}
}

func TestNowdoc(t *testing.T) {
	input := `<?php <<<'EOT'
Hello $name
EOT;`
	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php "},
		{token.T_START_HEREDOC, "<<<'EOT'\n"},
		{token.T_ENCAPSED_AND_WHITESPACE, "Hello $name\n"},
		{token.T_END_HEREDOC, "EOT"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenResult
	}{
		{
			name:  "single line //",
			input: "<?php // comment\n$x",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_COMMENT, "// comment\n"},
				{token.T_VARIABLE, "$x"},
				{token.EOF, ""},
			},
		},
		{
			name:  "single line #",
			input: "<?php # comment\n$x",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_COMMENT, "# comment\n"},
				{token.T_VARIABLE, "$x"},
				{token.EOF, ""},
			},
		},
		{
			name:  "multi-line comment",
			input: "<?php /* comment */ $x",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_COMMENT, "/* comment */"},
				{token.WHITESPACE, " "},
				{token.T_VARIABLE, "$x"},
				{token.EOF, ""},
			},
		},
		{
			name:  "doc comment",
			input: "<?php /** doc */ $x",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_DOC_COMMENT, "/** doc */"},
				{token.WHITESPACE, " "},
				{token.T_VARIABLE, "$x"},
				{token.EOF, ""},
			},
		},
		{
			name: "multi-line doc comment",
			input: `<?php /**
 * Documentation
 * @param string $x
 */
$x`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_DOC_COMMENT, "/**\n * Documentation\n * @param string $x\n */"},
				{token.WHITESPACE, "\n"},
				{token.T_VARIABLE, "$x"},
				{token.EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.Type {
					t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
						i, exp.Type, tok.Type)
				}
				if tok.Literal != exp.Literal {
					t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
						i, exp.Literal, tok.Literal)
				}
			}
		})
	}
}

func TestOperators(t *testing.T) {
	input := `<?php
+= -= *= /= %= **= &= |= ^= <<= >>= .= ??=
++ -- ** ?? <=> |>
== != <> === !== <= >= < >
|| && or and xor
<< >> & | ^
+ - * / % .
= => -> ?-> :: ...
`
	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php\n"},
		{token.T_PLUS_EQUAL, "+="},
		{token.WHITESPACE, " "},
		{token.T_MINUS_EQUAL, "-="},
		{token.WHITESPACE, " "},
		{token.T_MUL_EQUAL, "*="},
		{token.WHITESPACE, " "},
		{token.T_DIV_EQUAL, "/="},
		{token.WHITESPACE, " "},
		{token.T_MOD_EQUAL, "%="},
		{token.WHITESPACE, " "},
		{token.T_POW_EQUAL, "**="},
		{token.WHITESPACE, " "},
		{token.T_AND_EQUAL, "&="},
		{token.WHITESPACE, " "},
		{token.T_OR_EQUAL, "|="},
		{token.WHITESPACE, " "},
		{token.T_XOR_EQUAL, "^="},
		{token.WHITESPACE, " "},
		{token.T_SL_EQUAL, "<<="},
		{token.WHITESPACE, " "},
		{token.T_SR_EQUAL, ">>="},
		{token.WHITESPACE, " "},
		{token.T_CONCAT_EQUAL, ".="},
		{token.WHITESPACE, " "},
		{token.T_COALESCE_EQUAL, "??="},
		{token.WHITESPACE, "\n"},
		{token.T_INC, "++"},
		{token.WHITESPACE, " "},
		{token.T_DEC, "--"},
		{token.WHITESPACE, " "},
		{token.T_POW, "**"},
		{token.WHITESPACE, " "},
		{token.T_COALESCE, "??"},
		{token.WHITESPACE, " "},
		{token.T_SPACESHIP, "<=>"},
		{token.WHITESPACE, " "},
		{token.T_PIPE, "|>"},
		{token.WHITESPACE, "\n"},
		{token.T_IS_EQUAL, "=="},
		{token.WHITESPACE, " "},
		{token.T_IS_NOT_EQUAL, "!="},
		{token.WHITESPACE, " "},
		{token.T_IS_NOT_EQUAL, "<>"},
		{token.WHITESPACE, " "},
		{token.T_IS_IDENTICAL, "==="},
		{token.WHITESPACE, " "},
		{token.T_IS_NOT_IDENTICAL, "!=="},
		{token.WHITESPACE, " "},
		{token.T_IS_SMALLER_OR_EQUAL, "<="},
		{token.WHITESPACE, " "},
		{token.T_IS_GREATER_OR_EQUAL, ">="},
		{token.WHITESPACE, " "},
		{token.LESS, "<"},
		{token.WHITESPACE, " "},
		{token.GREATER, ">"},
		{token.WHITESPACE, "\n"},
		{token.T_BOOLEAN_OR, "||"},
		{token.WHITESPACE, " "},
		{token.T_BOOLEAN_AND, "&&"},
		{token.WHITESPACE, " "},
		{token.T_LOGICAL_OR, "or"},
		{token.WHITESPACE, " "},
		{token.T_LOGICAL_AND, "and"},
		{token.WHITESPACE, " "},
		{token.T_LOGICAL_XOR, "xor"},
		{token.WHITESPACE, "\n"},
		{token.T_SL, "<<"},
		{token.WHITESPACE, " "},
		{token.T_SR, ">>"},
		{token.WHITESPACE, " "},
		{token.AMPERSAND, "&"},
		{token.WHITESPACE, " "},
		{token.PIPE, "|"},
		{token.WHITESPACE, " "},
		{token.CARET, "^"},
		{token.WHITESPACE, "\n"},
		{token.PLUS, "+"},
		{token.WHITESPACE, " "},
		{token.MINUS, "-"},
		{token.WHITESPACE, " "},
		{token.ASTERISK, "*"},
		{token.WHITESPACE, " "},
		{token.SLASH, "/"},
		{token.WHITESPACE, " "},
		{token.PERCENT, "%"},
		{token.WHITESPACE, " "},
		{token.DOT, "."},
		{token.WHITESPACE, "\n"},
		{token.EQUALS, "="},
		{token.WHITESPACE, " "},
		{token.T_DOUBLE_ARROW, "=>"},
		{token.WHITESPACE, " "},
		{token.T_OBJECT_OPERATOR, "->"},
		{token.WHITESPACE, " "},
		{token.T_NULLSAFE_OBJECT_OPERATOR, "?->"},
		{token.WHITESPACE, " "},
		{token.T_PAAMAYIM_NEKUDOTAYIM, "::"},
		{token.WHITESPACE, " "},
		{token.T_ELLIPSIS, "..."},
		{token.WHITESPACE, "\n"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestTypeCasts(t *testing.T) {
	input := `<?php (int) (integer) (float) (double) (real) (string) (binary) (array) (object) (bool) (boolean) (unset)`
	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php "},
		{token.T_INT_CAST, "(int)"},
		{token.WHITESPACE, " "},
		{token.T_INT_CAST, "(integer)"},
		{token.WHITESPACE, " "},
		{token.T_DOUBLE_CAST, "(float)"},
		{token.WHITESPACE, " "},
		{token.T_DOUBLE_CAST, "(double)"},
		{token.WHITESPACE, " "},
		{token.T_DOUBLE_CAST, "(real)"},
		{token.WHITESPACE, " "},
		{token.T_STRING_CAST, "(string)"},
		{token.WHITESPACE, " "},
		{token.T_STRING_CAST, "(binary)"},
		{token.WHITESPACE, " "},
		{token.T_ARRAY_CAST, "(array)"},
		{token.WHITESPACE, " "},
		{token.T_OBJECT_CAST, "(object)"},
		{token.WHITESPACE, " "},
		{token.T_BOOL_CAST, "(bool)"},
		{token.WHITESPACE, " "},
		{token.T_BOOL_CAST, "(boolean)"},
		{token.WHITESPACE, " "},
		{token.T_UNSET_CAST, "(unset)"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestPunctuation(t *testing.T) {
	input := `<?php ; : , [ ] ( ) { } ? @ $ ~ !`
	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php "},
		{token.SEMICOLON, ";"},
		{token.WHITESPACE, " "},
		{token.COLON, ":"},
		{token.WHITESPACE, " "},
		{token.COMMA, ","},
		{token.WHITESPACE, " "},
		{token.LBRACKET, "["},
		{token.WHITESPACE, " "},
		{token.RBRACKET, "]"},
		{token.WHITESPACE, " "},
		{token.LPAREN, "("},
		{token.WHITESPACE, " "},
		{token.RPAREN, ")"},
		{token.WHITESPACE, " "},
		{token.LBRACE, "{"},
		{token.WHITESPACE, " "},
		{token.RBRACE, "}"},
		{token.WHITESPACE, " "},
		{token.QUESTION, "?"},
		{token.WHITESPACE, " "},
		{token.AT, "@"},
		{token.WHITESPACE, " "},
		{token.DOLLAR, "$"},
		{token.WHITESPACE, " "},
		{token.TILDE, "~"},
		{token.WHITESPACE, " "},
		{token.EXCLAMATION, "!"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestNamespaces(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenResult
	}{
		{
			name:  "namespace separator",
			input: `<?php \Foo`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_NAME_FULLY_QUALIFIED, "\\Foo"},
				{token.EOF, ""},
			},
		},
		{
			name:  "qualified name",
			input: `<?php Foo\Bar`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_NAME_QUALIFIED, "Foo\\Bar"},
				{token.EOF, ""},
			},
		},
		{
			name:  "fully qualified name",
			input: `<?php \Foo\Bar\Baz`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_NAME_FULLY_QUALIFIED, "\\Foo\\Bar\\Baz"},
				{token.EOF, ""},
			},
		},
		{
			name:  "relative name",
			input: `<?php namespace\Foo`,
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.T_NAME_RELATIVE, "namespace\\Foo"},
				{token.EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.Type {
					t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
						i, exp.Type, tok.Type)
				}
				if tok.Literal != exp.Literal {
					t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
						i, exp.Literal, tok.Literal)
				}
			}
		})
	}
}

func TestAttributes(t *testing.T) {
	input := `<?php #[Attribute] #[Route("/api")]`
	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php "},
		{token.T_ATTRIBUTE, "#["},
		{token.T_STRING, "Attribute"},
		{token.RBRACKET, "]"},
		{token.WHITESPACE, " "},
		{token.T_ATTRIBUTE, "#["},
		{token.T_STRING, "Route"},
		{token.LPAREN, "("},
		{token.T_CONSTANT_ENCAPSED_STRING, `"/api"`},
		{token.RPAREN, ")"},
		{token.RBRACKET, "]"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestCompleteProgram(t *testing.T) {
	input := `<?php
namespace App\Controllers;

use App\Models\User;

#[Route("/users")]
class UserController extends Controller
{
    public function __construct(
        private readonly UserRepository $repo
    ) {}

    public function index(): array
    {
        return $this->repo->findAll();
    }

    public function show(int $id): ?User
    {
        return $this->repo?->find($id);
    }
}
`

	l := New(input)
	tokens := []TokenInfo{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.ILLEGAL {
			t.Fatalf("Unexpected ILLEGAL token: %q at position %d", tok.Literal, tok.Pos.Offset)
		}
	}

	// Verify some key tokens are present
	foundTokens := map[token.Token]bool{}
	for _, tok := range tokens {
		foundTokens[tok.Type] = true
	}

	expectedTokens := []token.Token{
		token.T_NAMESPACE,
		token.T_USE,
		token.T_CLASS,
		token.T_EXTENDS,
		token.T_PUBLIC,
		token.T_FUNCTION,
		token.T_PRIVATE,
		token.T_READONLY,
		token.T_RETURN,
		token.T_ATTRIBUTE,
		token.T_VARIABLE,
		token.T_STRING,
		token.T_OBJECT_OPERATOR,
		token.T_NULLSAFE_OBJECT_OPERATOR,
	}

	for _, exp := range expectedTokens {
		if !foundTokens[exp] {
			t.Errorf("Expected token %s not found in program", exp)
		}
	}
}

func TestBacktickStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenResult
	}{
		{
			name:  "simple backtick",
			input: "<?php `ls -la`",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.BACKTICK, "`"},
				{token.T_ENCAPSED_AND_WHITESPACE, "ls -la"},
				{token.BACKTICK, "`"},
				{token.EOF, ""},
			},
		},
		{
			name:  "backtick with variable",
			input: "<?php `ls $dir`",
			expected: []TokenResult{
				{token.T_OPEN_TAG, "<?php "},
				{token.BACKTICK, "`"},
				{token.T_ENCAPSED_AND_WHITESPACE, "ls "},
				{token.T_VARIABLE, "$dir"},
				{token.BACKTICK, "`"},
				{token.EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.Type {
					t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
						i, exp.Type, tok.Type)
				}
				if tok.Literal != exp.Literal {
					t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
						i, exp.Literal, tok.Literal)
				}
			}
		})
	}
}

func TestPositionTracking(t *testing.T) {
	input := `<?php
$x = 1;
`
	l := New(input)

	// <?php\n
	tok := l.NextToken()
	if tok.Type != token.T_OPEN_TAG {
		t.Errorf("expected T_OPEN_TAG, got %s", tok.Type)
	}
	if tok.Pos.Line != 1 {
		t.Errorf("expected line 1, got %d", tok.Pos.Line)
	}

	// $x
	tok = l.NextToken()
	if tok.Type != token.T_VARIABLE {
		t.Errorf("expected T_VARIABLE, got %s", tok.Type)
	}
	if tok.Pos.Line != 2 {
		t.Errorf("expected line 2, got %d", tok.Pos.Line)
	}
	if tok.Pos.Column != 1 {
		t.Errorf("expected column 1, got %d", tok.Pos.Column)
	}

	// whitespace
	tok = l.NextToken()
	if tok.Type != token.WHITESPACE {
		t.Errorf("expected WHITESPACE, got %s", tok.Type)
	}

	// =
	tok = l.NextToken()
	if tok.Type != token.EQUALS {
		t.Errorf("expected EQUALS, got %s", tok.Type)
	}
	if tok.Pos.Column != 4 {
		t.Errorf("expected column 4, got %d", tok.Pos.Column)
	}
}

func TestYieldFrom(t *testing.T) {
	input := `<?php yield from $gen`
	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php "},
		{token.T_YIELD_FROM, "yield from"},
		{token.WHITESPACE, " "},
		{token.T_VARIABLE, "$gen"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestPropertyAccessInStrings(t *testing.T) {
	input := `<?php "Hello $obj->name"`
	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php "},
		{token.DOUBLE_QUOTE, `"`},
		{token.T_ENCAPSED_AND_WHITESPACE, "Hello "},
		{token.T_VARIABLE, "$obj"},
		{token.T_OBJECT_OPERATOR, "->"},
		{token.T_STRING, "name"},
		{token.DOUBLE_QUOTE, `"`},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestUnicodeIdentifiers(t *testing.T) {
	input := `<?php $日本語 $émoji $über`
	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php "},
		{token.T_VARIABLE, "$日本語"},
		{token.WHITESPACE, " "},
		{token.T_VARIABLE, "$émoji"},
		{token.WHITESPACE, " "},
		{token.T_VARIABLE, "$über"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestLegacyOctalNumbers(t *testing.T) {
	// Legacy octal notation (leading 0)
	input := `<?php 0755 0123`
	l := New(input)

	tok := l.NextToken() // <?php
	if tok.Type != token.T_OPEN_TAG {
		t.Fatalf("expected T_OPEN_TAG, got %s", tok.Type)
	}

	tok = l.NextToken() // 0755
	if tok.Type != token.T_LNUMBER {
		t.Errorf("expected T_LNUMBER, got %s", tok.Type)
	}
	if tok.Literal != "0755" {
		t.Errorf("expected '0755', got %s", tok.Literal)
	}
}

func TestNestedCurlyBracesInStrings(t *testing.T) {
	input := `<?php "Hello {$arr['key']}"`
	l := New(input)

	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php "},
		{token.DOUBLE_QUOTE, `"`},
		{token.T_ENCAPSED_AND_WHITESPACE, "Hello "},
		{token.T_CURLY_OPEN, "{"},
		{token.T_VARIABLE, "$arr"},
		{token.LBRACKET, "["},
		{token.T_CONSTANT_ENCAPSED_STRING, "'key'"},
		{token.RBRACKET, "]"},
		{token.RBRACE, "}"},
		{token.DOUBLE_QUOTE, `"`},
		{token.EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestComplexArrowFunction(t *testing.T) {
	input := `<?php $fn = fn($x, ...$rest) => $x + array_sum($rest);`
	l := New(input)

	tokens := []TokenInfo{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.ILLEGAL {
			t.Fatalf("Unexpected ILLEGAL token: %q at position %d", tok.Literal, tok.Pos.Offset)
		}
	}

	// Verify key tokens are present
	foundTokens := map[token.Token]bool{}
	for _, tok := range tokens {
		foundTokens[tok.Type] = true
	}

	if !foundTokens[token.T_FN] {
		t.Error("Expected T_FN token not found")
	}
	if !foundTokens[token.T_ELLIPSIS] {
		t.Error("Expected T_ELLIPSIS token not found")
	}
	if !foundTokens[token.T_DOUBLE_ARROW] {
		t.Error("Expected T_DOUBLE_ARROW token not found")
	}
}

func TestMatchExpression(t *testing.T) {
	input := `<?php
$result = match($x) {
    1, 2 => 'small',
    3, 4 => 'medium',
    default => 'large',
};`

	l := New(input)
	tokens := []TokenInfo{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.ILLEGAL {
			t.Fatalf("Unexpected ILLEGAL token: %q", tok.Literal)
		}
	}

	foundTokens := map[token.Token]bool{}
	for _, tok := range tokens {
		foundTokens[tok.Type] = true
	}

	if !foundTokens[token.T_MATCH] {
		t.Error("Expected T_MATCH token not found")
	}
	if !foundTokens[token.T_DEFAULT] {
		t.Error("Expected T_DEFAULT token not found")
	}
	if !foundTokens[token.T_DOUBLE_ARROW] {
		t.Error("Expected T_DOUBLE_ARROW token not found")
	}
}

func TestEnumDeclaration(t *testing.T) {
	input := `<?php
enum Status: string {
    case Active = 'active';
    case Inactive = 'inactive';
}`

	l := New(input)
	tokens := []TokenInfo{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.ILLEGAL {
			t.Fatalf("Unexpected ILLEGAL token: %q", tok.Literal)
		}
	}

	foundTokens := map[token.Token]bool{}
	for _, tok := range tokens {
		foundTokens[tok.Type] = true
	}

	if !foundTokens[token.T_ENUM] {
		t.Error("Expected T_ENUM token not found")
	}
	if !foundTokens[token.T_CASE] {
		t.Error("Expected T_CASE token not found")
	}
}

func TestMultipleStringInterpolations(t *testing.T) {
	input := `<?php "$a $b $c"`
	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php "},
		{token.DOUBLE_QUOTE, `"`},
		{token.T_VARIABLE, "$a"},
		{token.T_ENCAPSED_AND_WHITESPACE, " "},
		{token.T_VARIABLE, "$b"},
		{token.T_ENCAPSED_AND_WHITESPACE, " "},
		{token.T_VARIABLE, "$c"},
		{token.DOUBLE_QUOTE, `"`},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestNullCoalescing(t *testing.T) {
	input := `<?php $a ?? $b ??= $c`
	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php "},
		{token.T_VARIABLE, "$a"},
		{token.WHITESPACE, " "},
		{token.T_COALESCE, "??"},
		{token.WHITESPACE, " "},
		{token.T_VARIABLE, "$b"},
		{token.WHITESPACE, " "},
		{token.T_COALESCE_EQUAL, "??="},
		{token.WHITESPACE, " "},
		{token.T_VARIABLE, "$c"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestSpaceshipOperator(t *testing.T) {
	input := `<?php $a <=> $b`
	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php "},
		{token.T_VARIABLE, "$a"},
		{token.WHITESPACE, " "},
		{token.T_SPACESHIP, "<=>"},
		{token.WHITESPACE, " "},
		{token.T_VARIABLE, "$b"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestTraitUsage(t *testing.T) {
	input := `<?php
trait MyTrait {
    use AnotherTrait;
    public function foo() {}
}

class MyClass {
    use MyTrait, OtherTrait {
        MyTrait::foo insteadof OtherTrait;
    }
}`

	l := New(input)
	tokens := []TokenInfo{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.ILLEGAL {
			t.Fatalf("Unexpected ILLEGAL token: %q", tok.Literal)
		}
	}

	foundTokens := map[token.Token]bool{}
	for _, tok := range tokens {
		foundTokens[tok.Type] = true
	}

	if !foundTokens[token.T_TRAIT] {
		t.Error("Expected T_TRAIT token not found")
	}
	if !foundTokens[token.T_USE] {
		t.Error("Expected T_USE token not found")
	}
	if !foundTokens[token.T_INSTEADOF] {
		t.Error("Expected T_INSTEADOF token not found")
	}
	if !foundTokens[token.T_PAAMAYIM_NEKUDOTAYIM] {
		t.Error("Expected T_PAAMAYIM_NEKUDOTAYIM token not found")
	}
}

func TestGeneratorYield(t *testing.T) {
	input := `<?php
function gen() {
    yield 1;
    yield $key => $value;
    yield from [1, 2, 3];
}`

	l := New(input)
	tokens := []TokenInfo{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.ILLEGAL {
			t.Fatalf("Unexpected ILLEGAL token: %q", tok.Literal)
		}
	}

	foundYield := 0
	foundYieldFrom := 0
	for _, tok := range tokens {
		if tok.Type == token.T_YIELD {
			foundYield++
		}
		if tok.Type == token.T_YIELD_FROM {
			foundYieldFrom++
		}
	}

	if foundYield < 2 {
		t.Errorf("Expected at least 2 T_YIELD tokens, found %d", foundYield)
	}
	if foundYieldFrom < 1 {
		t.Errorf("Expected at least 1 T_YIELD_FROM token, found %d", foundYieldFrom)
	}
}

func TestIndentedHeredoc(t *testing.T) {
	input := `<?php
$text = <<<EOT
    Hello World
    EOT;`
	expected := []TokenResult{
		{token.T_OPEN_TAG, "<?php\n"},
		{token.T_VARIABLE, "$text"},
		{token.WHITESPACE, " "},
		{token.EQUALS, "="},
		{token.WHITESPACE, " "},
		{token.T_START_HEREDOC, "<<<EOT\n"},
		{token.T_ENCAPSED_AND_WHITESPACE, "    Hello World\n    "},
		{token.T_END_HEREDOC, "EOT"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.Type {
			t.Errorf("test[%d] - token type wrong. expected=%q, got=%q",
				i, exp.Type, tok.Type)
		}
		if tok.Literal != exp.Literal {
			t.Errorf("test[%d] - literal wrong. expected=%q, got=%q",
				i, exp.Literal, tok.Literal)
		}
	}
}

func TestExceptionHandling(t *testing.T) {
	input := `<?php
try {
    throw new Exception("Error");
} catch (Exception $e) {
    echo $e->getMessage();
} finally {
    cleanup();
}`

	l := New(input)
	tokens := []TokenInfo{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.ILLEGAL {
			t.Fatalf("Unexpected ILLEGAL token: %q", tok.Literal)
		}
	}

	foundTokens := map[token.Token]bool{}
	for _, tok := range tokens {
		foundTokens[tok.Type] = true
	}

	if !foundTokens[token.T_TRY] {
		t.Error("Expected T_TRY token not found")
	}
	if !foundTokens[token.T_THROW] {
		t.Error("Expected T_THROW token not found")
	}
	if !foundTokens[token.T_CATCH] {
		t.Error("Expected T_CATCH token not found")
	}
	if !foundTokens[token.T_FINALLY] {
		t.Error("Expected T_FINALLY token not found")
	}
	if !foundTokens[token.T_NEW] {
		t.Error("Expected T_NEW token not found")
	}
}

func TestNamespaceDeclarations(t *testing.T) {
	input := `<?php
namespace App\Controllers;

use App\Models\User;
use Illuminate\Http\Request as HttpRequest;

class UserController {}`

	l := New(input)
	tokens := []TokenInfo{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.ILLEGAL {
			t.Fatalf("Unexpected ILLEGAL token: %q", tok.Literal)
		}
	}

	foundTokens := map[token.Token]bool{}
	for _, tok := range tokens {
		foundTokens[tok.Type] = true
	}

	if !foundTokens[token.T_NAMESPACE] {
		t.Error("Expected T_NAMESPACE token not found")
	}
	if !foundTokens[token.T_USE] {
		t.Error("Expected T_USE token not found")
	}
	if !foundTokens[token.T_AS] {
		t.Error("Expected T_AS token not found")
	}
	if !foundTokens[token.T_NAME_QUALIFIED] {
		t.Error("Expected T_NAME_QUALIFIED token not found")
	}
}

func TestClosureWithUse(t *testing.T) {
	input := `<?php $fn = function($x) use ($y, &$z) { return $x + $y; };`

	l := New(input)
	tokens := []TokenInfo{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.ILLEGAL {
			t.Fatalf("Unexpected ILLEGAL token: %q", tok.Literal)
		}
	}

	foundTokens := map[token.Token]bool{}
	for _, tok := range tokens {
		foundTokens[tok.Type] = true
	}

	if !foundTokens[token.T_FUNCTION] {
		t.Error("Expected T_FUNCTION token not found")
	}
	if !foundTokens[token.T_USE] {
		t.Error("Expected T_USE token not found")
	}
	if !foundTokens[token.AMPERSAND] {
		t.Error("Expected AMPERSAND token not found")
	}
}

func TestEmptyScript(t *testing.T) {
	input := `<?php`
	l := New(input)

	tok := l.NextToken()
	if tok.Type != token.T_OPEN_TAG {
		t.Errorf("expected T_OPEN_TAG, got %s", tok.Type)
	}

	tok = l.NextToken()
	if tok.Type != token.EOF {
		t.Errorf("expected EOF, got %s", tok.Type)
	}
}

func TestMultiplePHPBlocks(t *testing.T) {
	input := `<html>
<?php echo "Hello"; ?>
<body>
<?= $content ?>
</body>
</html>`

	l := New(input)
	tokens := []TokenInfo{}
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}

	// Count different token types
	openTags := 0
	closeTags := 0
	inlineHTML := 0

	for _, tok := range tokens {
		switch tok.Type {
		case token.T_OPEN_TAG:
			openTags++
		case token.T_OPEN_TAG_WITH_ECHO:
			openTags++
		case token.T_CLOSE_TAG:
			closeTags++
		case token.T_INLINE_HTML:
			inlineHTML++
		}
	}

	if openTags != 2 {
		t.Errorf("expected 2 open tags, got %d", openTags)
	}
	if closeTags != 2 {
		t.Errorf("expected 2 close tags, got %d", closeTags)
	}
	if inlineHTML < 2 {
		t.Errorf("expected at least 2 inline HTML blocks, got %d", inlineHTML)
	}
}
