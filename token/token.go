// Package token defines constants representing the lexical tokens of PHP
// and basic operations on tokens (printing, predicates).
package token

// Token represents a lexical token type.
type Token int

// The list of tokens, matching PHP's official token definitions.
const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	WHITESPACE

	// Literals
	T_LNUMBER                    // Integer literal
	T_DNUMBER                    // Float literal
	T_STRING                     // Identifier/name
	T_VARIABLE                   // Variable ($var)
	T_INLINE_HTML                // HTML outside PHP tags
	T_ENCAPSED_AND_WHITESPACE    // String content with interpolation
	T_CONSTANT_ENCAPSED_STRING   // Single-quoted or simple double-quoted string
	T_STRING_VARNAME             // Variable name in ${varname}
	T_NUM_STRING                 // Numeric array index in string

	// Namespace names
	T_NAME_FULLY_QUALIFIED // \Namespace\Class
	T_NAME_QUALIFIED       // Namespace\Class
	T_NAME_RELATIVE        // namespace\Class

	// Keywords - Control Flow
	T_IF
	T_ELSEIF
	T_ELSE
	T_ENDIF
	T_WHILE
	T_ENDWHILE
	T_DO
	T_FOR
	T_ENDFOR
	T_FOREACH
	T_ENDFOREACH
	T_SWITCH
	T_ENDSWITCH
	T_CASE
	T_DEFAULT
	T_MATCH
	T_BREAK
	T_CONTINUE
	T_GOTO
	T_RETURN
	T_YIELD
	T_YIELD_FROM
	T_TRY
	T_CATCH
	T_FINALLY
	T_THROW

	// Keywords - Functions & Classes
	T_FUNCTION
	T_FN
	T_CLASS
	T_TRAIT
	T_INTERFACE
	T_ENUM
	T_EXTENDS
	T_IMPLEMENTS
	T_NEW
	T_CLONE
	T_INSTANCEOF

	// Keywords - Visibility Modifiers
	T_PUBLIC
	T_PROTECTED
	T_PRIVATE
	T_PUBLIC_SET
	T_PROTECTED_SET
	T_PRIVATE_SET
	T_READONLY
	T_STATIC
	T_ABSTRACT
	T_FINAL
	T_VAR

	// Keywords - Scope & Variables
	T_GLOBAL
	T_USE
	T_UNSET
	T_ISSET
	T_EMPTY
	T_EVAL

	// Keywords - Include
	T_INCLUDE
	T_INCLUDE_ONCE
	T_REQUIRE
	T_REQUIRE_ONCE

	// Keywords - Language Constructs
	T_ECHO
	T_PRINT
	T_LIST
	T_ARRAY
	T_CALLABLE
	T_DECLARE
	T_ENDDECLARE
	T_AS
	T_INSTEADOF
	T_HALT_COMPILER
	T_NAMESPACE
	T_CONST
	T_EXIT // exit/die

	// Operators - Arithmetic Assignment
	T_PLUS_EQUAL  // +=
	T_MINUS_EQUAL // -=
	T_MUL_EQUAL   // *=
	T_DIV_EQUAL   // /=
	T_MOD_EQUAL   // %=
	T_POW         // **
	T_POW_EQUAL   // **=
	T_INC         // ++
	T_DEC         // --

	// Operators - Comparison
	T_IS_EQUAL             // ==
	T_IS_NOT_EQUAL         // != or <>
	T_IS_IDENTICAL         // ===
	T_IS_NOT_IDENTICAL     // !==
	T_IS_SMALLER_OR_EQUAL  // <=
	T_IS_GREATER_OR_EQUAL  // >=
	T_SPACESHIP            // <=>

	// Operators - Logical
	T_LOGICAL_OR  // or
	T_LOGICAL_AND // and
	T_LOGICAL_XOR // xor
	T_BOOLEAN_OR  // ||
	T_BOOLEAN_AND // &&

	// Operators - Bitwise & Special
	T_AND_EQUAL      // &=
	T_OR_EQUAL       // |=
	T_XOR_EQUAL      // ^=
	T_SL             // <<
	T_SR             // >>
	T_SL_EQUAL       // <<=
	T_SR_EQUAL       // >>=
	T_CONCAT_EQUAL   // .=
	T_COALESCE       // ??
	T_COALESCE_EQUAL // ??=
	T_PIPE           // |> (pipe operator)

	// Operators - Ampersand Context
	T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG     // & before variable
	T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG // & as bitwise AND

	// Type Casts
	T_INT_CAST    // (int)
	T_DOUBLE_CAST // (float)/(double)/(real)
	T_STRING_CAST // (string)
	T_ARRAY_CAST  // (array)
	T_OBJECT_CAST // (object)
	T_BOOL_CAST   // (bool)/(boolean)
	T_UNSET_CAST  // (unset)

	// Operators - Array & Object
	T_DOUBLE_ARROW             // =>
	T_OBJECT_OPERATOR          // ->
	T_NULLSAFE_OBJECT_OPERATOR // ?->
	T_PAAMAYIM_NEKUDOTAYIM     // :: (double colon)
	T_ELLIPSIS                 // ...
	T_NS_SEPARATOR             // \ (namespace separator)

	// Magic Constants
	T_LINE      // __LINE__
	T_FILE      // __FILE__
	T_DIR       // __DIR__
	T_CLASS_C   // __CLASS__
	T_TRAIT_C   // __TRAIT__
	T_METHOD_C  // __METHOD__
	T_FUNC_C    // __FUNCTION__
	T_PROPERTY_C // __PROPERTY__
	T_NS_C      // __NAMESPACE__

	// Comments
	T_COMMENT     // // or # style
	T_DOC_COMMENT // /** */

	// PHP Tags
	T_OPEN_TAG           // <?php
	T_OPEN_TAG_WITH_ECHO // <?=
	T_CLOSE_TAG          // ?>

	// String Delimiters
	T_START_HEREDOC           // <<<LABEL
	T_END_HEREDOC             // LABEL
	T_DOLLAR_OPEN_CURLY_BRACES // ${
	T_CURLY_OPEN              // {$

	// Attributes
	T_ATTRIBUTE // #[

	// Error tokens
	T_ERROR
	T_BAD_CHARACTER

	// Single character tokens - these are represented by their ASCII value in PHP
	// but we define them here for clarity
	SEMICOLON      // ;
	COLON          // :
	COMMA          // ,
	DOT            // .
	LBRACKET       // [
	RBRACKET       // ]
	LPAREN         // (
	RPAREN         // )
	LBRACE         // {
	RBRACE         // }
	QUESTION       // ?
	AT             // @
	DOLLAR         // $
	BACKTICK       // `
	TILDE          // ~
	EXCLAMATION    // !
	PLUS           // +
	MINUS          // -
	ASTERISK       // *
	SLASH          // /
	PERCENT        // %
	CARET          // ^
	AMPERSAND      // &
	PIPE           // |
	LESS           // <
	GREATER        // >
	EQUALS         // =
	DOUBLE_QUOTE   // "
	SINGLE_QUOTE   // '
)

var tokens = [...]string{
	ILLEGAL:    "ILLEGAL",
	EOF:        "EOF",
	WHITESPACE: "T_WHITESPACE",

	T_LNUMBER:                  "T_LNUMBER",
	T_DNUMBER:                  "T_DNUMBER",
	T_STRING:                   "T_STRING",
	T_VARIABLE:                 "T_VARIABLE",
	T_INLINE_HTML:              "T_INLINE_HTML",
	T_ENCAPSED_AND_WHITESPACE:  "T_ENCAPSED_AND_WHITESPACE",
	T_CONSTANT_ENCAPSED_STRING: "T_CONSTANT_ENCAPSED_STRING",
	T_STRING_VARNAME:           "T_STRING_VARNAME",
	T_NUM_STRING:               "T_NUM_STRING",

	T_NAME_FULLY_QUALIFIED: "T_NAME_FULLY_QUALIFIED",
	T_NAME_QUALIFIED:       "T_NAME_QUALIFIED",
	T_NAME_RELATIVE:        "T_NAME_RELATIVE",

	T_IF:         "T_IF",
	T_ELSEIF:     "T_ELSEIF",
	T_ELSE:       "T_ELSE",
	T_ENDIF:      "T_ENDIF",
	T_WHILE:      "T_WHILE",
	T_ENDWHILE:   "T_ENDWHILE",
	T_DO:         "T_DO",
	T_FOR:        "T_FOR",
	T_ENDFOR:     "T_ENDFOR",
	T_FOREACH:    "T_FOREACH",
	T_ENDFOREACH: "T_ENDFOREACH",
	T_SWITCH:     "T_SWITCH",
	T_ENDSWITCH:  "T_ENDSWITCH",
	T_CASE:       "T_CASE",
	T_DEFAULT:    "T_DEFAULT",
	T_MATCH:      "T_MATCH",
	T_BREAK:      "T_BREAK",
	T_CONTINUE:   "T_CONTINUE",
	T_GOTO:       "T_GOTO",
	T_RETURN:     "T_RETURN",
	T_YIELD:      "T_YIELD",
	T_YIELD_FROM: "T_YIELD_FROM",
	T_TRY:        "T_TRY",
	T_CATCH:      "T_CATCH",
	T_FINALLY:    "T_FINALLY",
	T_THROW:      "T_THROW",

	T_FUNCTION:   "T_FUNCTION",
	T_FN:         "T_FN",
	T_CLASS:      "T_CLASS",
	T_TRAIT:      "T_TRAIT",
	T_INTERFACE:  "T_INTERFACE",
	T_ENUM:       "T_ENUM",
	T_EXTENDS:    "T_EXTENDS",
	T_IMPLEMENTS: "T_IMPLEMENTS",
	T_NEW:        "T_NEW",
	T_CLONE:      "T_CLONE",
	T_INSTANCEOF: "T_INSTANCEOF",

	T_PUBLIC:        "T_PUBLIC",
	T_PROTECTED:     "T_PROTECTED",
	T_PRIVATE:       "T_PRIVATE",
	T_PUBLIC_SET:    "T_PUBLIC_SET",
	T_PROTECTED_SET: "T_PROTECTED_SET",
	T_PRIVATE_SET:   "T_PRIVATE_SET",
	T_READONLY:      "T_READONLY",
	T_STATIC:        "T_STATIC",
	T_ABSTRACT:      "T_ABSTRACT",
	T_FINAL:         "T_FINAL",
	T_VAR:           "T_VAR",

	T_GLOBAL: "T_GLOBAL",
	T_USE:    "T_USE",
	T_UNSET:  "T_UNSET",
	T_ISSET:  "T_ISSET",
	T_EMPTY:  "T_EMPTY",
	T_EVAL:   "T_EVAL",

	T_INCLUDE:      "T_INCLUDE",
	T_INCLUDE_ONCE: "T_INCLUDE_ONCE",
	T_REQUIRE:      "T_REQUIRE",
	T_REQUIRE_ONCE: "T_REQUIRE_ONCE",

	T_ECHO:          "T_ECHO",
	T_PRINT:         "T_PRINT",
	T_LIST:          "T_LIST",
	T_ARRAY:         "T_ARRAY",
	T_CALLABLE:      "T_CALLABLE",
	T_DECLARE:       "T_DECLARE",
	T_ENDDECLARE:    "T_ENDDECLARE",
	T_AS:            "T_AS",
	T_INSTEADOF:     "T_INSTEADOF",
	T_HALT_COMPILER: "T_HALT_COMPILER",
	T_NAMESPACE:     "T_NAMESPACE",
	T_CONST:         "T_CONST",
	T_EXIT:          "T_EXIT",

	T_PLUS_EQUAL:  "T_PLUS_EQUAL",
	T_MINUS_EQUAL: "T_MINUS_EQUAL",
	T_MUL_EQUAL:   "T_MUL_EQUAL",
	T_DIV_EQUAL:   "T_DIV_EQUAL",
	T_MOD_EQUAL:   "T_MOD_EQUAL",
	T_POW:         "T_POW",
	T_POW_EQUAL:   "T_POW_EQUAL",
	T_INC:         "T_INC",
	T_DEC:         "T_DEC",

	T_IS_EQUAL:            "T_IS_EQUAL",
	T_IS_NOT_EQUAL:        "T_IS_NOT_EQUAL",
	T_IS_IDENTICAL:        "T_IS_IDENTICAL",
	T_IS_NOT_IDENTICAL:    "T_IS_NOT_IDENTICAL",
	T_IS_SMALLER_OR_EQUAL: "T_IS_SMALLER_OR_EQUAL",
	T_IS_GREATER_OR_EQUAL: "T_IS_GREATER_OR_EQUAL",
	T_SPACESHIP:           "T_SPACESHIP",

	T_LOGICAL_OR:  "T_LOGICAL_OR",
	T_LOGICAL_AND: "T_LOGICAL_AND",
	T_LOGICAL_XOR: "T_LOGICAL_XOR",
	T_BOOLEAN_OR:  "T_BOOLEAN_OR",
	T_BOOLEAN_AND: "T_BOOLEAN_AND",

	T_AND_EQUAL:      "T_AND_EQUAL",
	T_OR_EQUAL:       "T_OR_EQUAL",
	T_XOR_EQUAL:      "T_XOR_EQUAL",
	T_SL:             "T_SL",
	T_SR:             "T_SR",
	T_SL_EQUAL:       "T_SL_EQUAL",
	T_SR_EQUAL:       "T_SR_EQUAL",
	T_CONCAT_EQUAL:   "T_CONCAT_EQUAL",
	T_COALESCE:       "T_COALESCE",
	T_COALESCE_EQUAL: "T_COALESCE_EQUAL",
	T_PIPE:           "T_PIPE",

	T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG:     "T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG",
	T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG: "T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG",

	T_INT_CAST:    "T_INT_CAST",
	T_DOUBLE_CAST: "T_DOUBLE_CAST",
	T_STRING_CAST: "T_STRING_CAST",
	T_ARRAY_CAST:  "T_ARRAY_CAST",
	T_OBJECT_CAST: "T_OBJECT_CAST",
	T_BOOL_CAST:   "T_BOOL_CAST",
	T_UNSET_CAST:  "T_UNSET_CAST",

	T_DOUBLE_ARROW:             "T_DOUBLE_ARROW",
	T_OBJECT_OPERATOR:          "T_OBJECT_OPERATOR",
	T_NULLSAFE_OBJECT_OPERATOR: "T_NULLSAFE_OBJECT_OPERATOR",
	T_PAAMAYIM_NEKUDOTAYIM:     "T_PAAMAYIM_NEKUDOTAYIM",
	T_ELLIPSIS:                 "T_ELLIPSIS",
	T_NS_SEPARATOR:             "T_NS_SEPARATOR",

	T_LINE:      "T_LINE",
	T_FILE:      "T_FILE",
	T_DIR:       "T_DIR",
	T_CLASS_C:   "T_CLASS_C",
	T_TRAIT_C:   "T_TRAIT_C",
	T_METHOD_C:  "T_METHOD_C",
	T_FUNC_C:    "T_FUNC_C",
	T_PROPERTY_C: "T_PROPERTY_C",
	T_NS_C:      "T_NS_C",

	T_COMMENT:     "T_COMMENT",
	T_DOC_COMMENT: "T_DOC_COMMENT",

	T_OPEN_TAG:           "T_OPEN_TAG",
	T_OPEN_TAG_WITH_ECHO: "T_OPEN_TAG_WITH_ECHO",
	T_CLOSE_TAG:          "T_CLOSE_TAG",

	T_START_HEREDOC:            "T_START_HEREDOC",
	T_END_HEREDOC:              "T_END_HEREDOC",
	T_DOLLAR_OPEN_CURLY_BRACES: "T_DOLLAR_OPEN_CURLY_BRACES",
	T_CURLY_OPEN:               "T_CURLY_OPEN",

	T_ATTRIBUTE: "T_ATTRIBUTE",

	T_ERROR:         "T_ERROR",
	T_BAD_CHARACTER: "T_BAD_CHARACTER",

	SEMICOLON:    ";",
	COLON:        ":",
	COMMA:        ",",
	DOT:          ".",
	LBRACKET:     "[",
	RBRACKET:     "]",
	LPAREN:       "(",
	RPAREN:       ")",
	LBRACE:       "{",
	RBRACE:       "}",
	QUESTION:     "?",
	AT:           "@",
	DOLLAR:       "$",
	BACKTICK:     "`",
	TILDE:        "~",
	EXCLAMATION:  "!",
	PLUS:         "+",
	MINUS:        "-",
	ASTERISK:     "*",
	SLASH:        "/",
	PERCENT:      "%",
	CARET:        "^",
	AMPERSAND:    "&",
	PIPE:         "|",
	LESS:         "<",
	GREATER:      ">",
	EQUALS:       "=",
	DOUBLE_QUOTE: "\"",
	SINGLE_QUOTE: "'",
}

// String returns the string representation of the token.
func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		return "token(" + string(rune(tok)) + ")"
	}
	return s
}

// keywords maps keyword strings to their token types.
var keywords = map[string]Token{
	"if":              T_IF,
	"elseif":          T_ELSEIF,
	"else":            T_ELSE,
	"endif":           T_ENDIF,
	"while":           T_WHILE,
	"endwhile":        T_ENDWHILE,
	"do":              T_DO,
	"for":             T_FOR,
	"endfor":          T_ENDFOR,
	"foreach":         T_FOREACH,
	"endforeach":      T_ENDFOREACH,
	"switch":          T_SWITCH,
	"endswitch":       T_ENDSWITCH,
	"case":            T_CASE,
	"default":         T_DEFAULT,
	"match":           T_MATCH,
	"break":           T_BREAK,
	"continue":        T_CONTINUE,
	"goto":            T_GOTO,
	"return":          T_RETURN,
	"yield":           T_YIELD,
	"try":             T_TRY,
	"catch":           T_CATCH,
	"finally":         T_FINALLY,
	"throw":           T_THROW,
	"function":        T_FUNCTION,
	"fn":              T_FN,
	"class":           T_CLASS,
	"trait":           T_TRAIT,
	"interface":       T_INTERFACE,
	"enum":            T_ENUM,
	"extends":         T_EXTENDS,
	"implements":      T_IMPLEMENTS,
	"new":             T_NEW,
	"clone":           T_CLONE,
	"instanceof":      T_INSTANCEOF,
	"public":          T_PUBLIC,
	"protected":       T_PROTECTED,
	"private":         T_PRIVATE,
	"readonly":        T_READONLY,
	"static":          T_STATIC,
	"abstract":        T_ABSTRACT,
	"final":           T_FINAL,
	"var":             T_VAR,
	"global":          T_GLOBAL,
	"use":             T_USE,
	"unset":           T_UNSET,
	"isset":           T_ISSET,
	"empty":           T_EMPTY,
	"eval":            T_EVAL,
	"include":         T_INCLUDE,
	"include_once":    T_INCLUDE_ONCE,
	"require":         T_REQUIRE,
	"require_once":    T_REQUIRE_ONCE,
	"echo":            T_ECHO,
	"print":           T_PRINT,
	"list":            T_LIST,
	"array":           T_ARRAY,
	"callable":        T_CALLABLE,
	"declare":         T_DECLARE,
	"enddeclare":      T_ENDDECLARE,
	"as":              T_AS,
	"insteadof":       T_INSTEADOF,
	"__halt_compiler": T_HALT_COMPILER,
	"namespace":       T_NAMESPACE,
	"const":           T_CONST,
	"exit":            T_EXIT,
	"die":             T_EXIT,
	"or":              T_LOGICAL_OR,
	"and":             T_LOGICAL_AND,
	"xor":             T_LOGICAL_XOR,
	"__LINE__":        T_LINE,
	"__FILE__":        T_FILE,
	"__DIR__":         T_DIR,
	"__CLASS__":       T_CLASS_C,
	"__TRAIT__":       T_TRAIT_C,
	"__METHOD__":      T_METHOD_C,
	"__FUNCTION__":    T_FUNC_C,
	"__PROPERTY__":    T_PROPERTY_C,
	"__NAMESPACE__":   T_NS_C,
}

// LookupIdent returns the token type for the given identifier.
// If the identifier is a keyword, returns the keyword token.
// Otherwise returns T_STRING.
func LookupIdent(ident string) Token {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return T_STRING
}

// IsKeyword returns true if the token is a keyword.
func (tok Token) IsKeyword() bool {
	return tok >= T_IF && tok <= T_EXIT
}

// IsLiteral returns true if the token is a literal value.
func (tok Token) IsLiteral() bool {
	return tok >= T_LNUMBER && tok <= T_NUM_STRING
}

// IsOperator returns true if the token is an operator.
func (tok Token) IsOperator() bool {
	return tok >= T_PLUS_EQUAL && tok <= T_ELLIPSIS
}
