// Package parser implements a PHP parser.
package parser

import (
	"github.com/alexisbouchez/phpgo/ast"
	"github.com/alexisbouchez/phpgo/lexer"
	"github.com/alexisbouchez/phpgo/token"
)

// Operator precedence levels
const (
	_ int = iota
	LOWEST
	ASSIGN      // = += -= etc
	TERNARY     // ?:
	COALESCE    // ??
	OR          // || or
	XOR         // xor
	AND         // && and
	BITOR       // |
	BITXOR      // ^
	BITAND      // &
	EQUALITY    // == != === !==
	COMPARISON  // < <= > >= <=>
	PIPE        // |>
	CONCAT      // .
	SHIFT       // << >>
	SUM         // + -
	PRODUCT     // * / %
	INSTANCEOF  // instanceof
	PREFIX      // ! ~ - + ++ -- @ (type)
	POW         // **
	CALL        // ()
	INDEX       // []
	PROPERTY    // -> ?-> ::
)

var precedences = map[token.Token]int{
	// Assignment (right-associative, handled specially)
	token.EQUALS:           ASSIGN,
	token.T_PLUS_EQUAL:     ASSIGN,
	token.T_MINUS_EQUAL:    ASSIGN,
	token.T_MUL_EQUAL:      ASSIGN,
	token.T_DIV_EQUAL:      ASSIGN,
	token.T_MOD_EQUAL:      ASSIGN,
	token.T_POW_EQUAL:      ASSIGN,
	token.T_CONCAT_EQUAL:   ASSIGN,
	token.T_AND_EQUAL:      ASSIGN,
	token.T_OR_EQUAL:       ASSIGN,
	token.T_XOR_EQUAL:      ASSIGN,
	token.T_SL_EQUAL:       ASSIGN,
	token.T_SR_EQUAL:       ASSIGN,
	token.T_COALESCE_EQUAL: ASSIGN,

	// Ternary
	token.QUESTION: TERNARY,

	// Null coalesce
	token.T_COALESCE: COALESCE,

	// Logical
	token.T_BOOLEAN_OR:  OR,
	token.T_LOGICAL_OR:  OR,
	token.T_LOGICAL_XOR: XOR,
	token.T_BOOLEAN_AND: AND,
	token.T_LOGICAL_AND: AND,

	// Bitwise
	token.PIPE:      BITOR,
	token.CARET:     BITXOR,
	token.AMPERSAND: BITAND,

	// Comparison
	token.T_IS_EQUAL:         EQUALITY,
	token.T_IS_NOT_EQUAL:     EQUALITY,
	token.T_IS_IDENTICAL:     EQUALITY,
	token.T_IS_NOT_IDENTICAL: EQUALITY,
	token.LESS:               COMPARISON,
	token.GREATER:            COMPARISON,
	token.T_IS_SMALLER_OR_EQUAL:  COMPARISON,
	token.T_IS_GREATER_OR_EQUAL:  COMPARISON,
	token.T_SPACESHIP:        COMPARISON,

	// Pipe
	token.T_PIPE: PIPE,

	// Concat
	token.DOT: CONCAT,

	// Shift
	token.T_SL: SHIFT,
	token.T_SR: SHIFT,

	// Arithmetic
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.ASTERISK: PRODUCT,
	token.SLASH:    PRODUCT,
	token.PERCENT:  PRODUCT,

	// Instanceof
	token.T_INSTANCEOF: INSTANCEOF,

	// Power
	token.T_POW: POW,

	// Call and access
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
	token.T_OBJECT_OPERATOR:          PROPERTY,
	token.T_NULLSAFE_OBJECT_OPERATOR: PROPERTY,
	token.T_PAAMAYIM_NEKUDOTAYIM:     PROPERTY,

	// Increment/decrement (postfix)
	token.T_INC: CALL,
	token.T_DEC: CALL,
}

// Parser parses PHP source code into an AST.
type Parser struct {
	l         *lexer.Lexer
	curToken  lexer.TokenInfo
	peekToken lexer.TokenInfo
	errors    []string
}

// New creates a new Parser.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()
	return p
}

// ParseString parses a PHP source string and returns the AST.
func ParseString(input string) *ast.File {
	l := lexer.New(input)
	p := New(l)
	return p.ParseFile()
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.Token) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.Token) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.Token) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	return false
}

func (p *Parser) curPos() ast.Position {
	return ast.Position{
		Offset: p.curToken.Pos.Offset,
		Line:   p.curToken.Pos.Line,
		Column: p.curToken.Pos.Column,
	}
}

func (p *Parser) skipWhitespace() {
	for p.curTokenIs(token.WHITESPACE) || p.curTokenIs(token.T_COMMENT) || p.curTokenIs(token.T_DOC_COMMENT) {
		p.nextToken()
	}
}

func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}
	return LOWEST
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken.Type]; ok {
		return prec
	}
	return LOWEST
}

// ParseFile parses a complete PHP file.
func (p *Parser) ParseFile() *ast.File {
	file := &ast.File{Stmts: []ast.Stmt{}}

	// Skip to PHP open tag
	for !p.curTokenIs(token.T_OPEN_TAG) && !p.curTokenIs(token.T_OPEN_TAG_WITH_ECHO) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.T_INLINE_HTML) {
			file.Stmts = append(file.Stmts, &ast.InlineHTMLStmt{
				Start: p.curPos(),
				Value: p.curToken.Literal,
			})
		}
		p.nextToken()
	}

	if p.curTokenIs(token.T_OPEN_TAG) || p.curTokenIs(token.T_OPEN_TAG_WITH_ECHO) {
		file.OpenTag = p.curPos()
		isEcho := p.curTokenIs(token.T_OPEN_TAG_WITH_ECHO)
		p.nextToken()

		if isEcho {
			// <?= is equivalent to <?php echo
			p.skipWhitespace()
			expr := p.parseExpression(LOWEST)
			file.Stmts = append(file.Stmts, &ast.EchoStmt{
				EchoPos: file.OpenTag,
				Exprs:   []ast.Expr{expr},
			})
		}
	}

	for !p.curTokenIs(token.EOF) {
		p.skipWhitespace()
		if p.curTokenIs(token.EOF) {
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			file.Stmts = append(file.Stmts, stmt)
		}
	}

	return file
}

func (p *Parser) parseStatement() ast.Stmt {
	p.skipWhitespace()

	switch p.curToken.Type {
	case token.T_NAMESPACE:
		return p.parseNamespaceDecl()
	case token.T_USE:
		return p.parseUseDecl()
	case token.T_CONST:
		return p.parseConstDecl()
	case token.T_FUNCTION:
		return p.parseFunctionDecl()
	case token.T_CLASS:
		return p.parseClassDecl(nil)
	case token.T_ABSTRACT:
		return p.parseAbstractClass()
	case token.T_FINAL:
		return p.parseFinalClass()
	case token.T_READONLY:
		return p.parseReadonlyClass()
	case token.T_INTERFACE:
		return p.parseInterfaceDecl()
	case token.T_TRAIT:
		return p.parseTraitDecl()
	case token.T_ENUM:
		return p.parseEnumDecl()
	case token.T_IF:
		return p.parseIfStmt()
	case token.T_WHILE:
		return p.parseWhileStmt()
	case token.T_DO:
		return p.parseDoWhileStmt()
	case token.T_FOR:
		return p.parseForStmt()
	case token.T_FOREACH:
		return p.parseForeachStmt()
	case token.T_SWITCH:
		return p.parseSwitchStmt()
	case token.T_TRY:
		return p.parseTryStmt()
	case token.T_THROW:
		return p.parseThrowStmt()
	case token.T_RETURN:
		return p.parseReturnStmt()
	case token.T_BREAK:
		return p.parseBreakStmt()
	case token.T_CONTINUE:
		return p.parseContinueStmt()
	case token.T_GOTO:
		return p.parseGotoStmt()
	case token.T_ECHO:
		return p.parseEchoStmt()
	case token.T_GLOBAL:
		return p.parseGlobalStmt()
	case token.T_STATIC:
		return p.parseStaticVarStmt()
	case token.T_UNSET:
		return p.parseUnsetStmt()
	case token.T_DECLARE:
		return p.parseDeclareStmt()
	case token.LBRACE:
		return p.parseBlockStmt()
	case token.SEMICOLON:
		pos := p.curPos()
		p.nextToken()
		return &ast.EmptyStmt{Semicolon: pos}
	case token.T_INLINE_HTML:
		stmt := &ast.InlineHTMLStmt{
			Start: p.curPos(),
			Value: p.curToken.Literal,
		}
		p.nextToken()
		return stmt
	case token.T_CLOSE_TAG:
		p.nextToken()
		return nil
	case token.T_ATTRIBUTE:
		attrs := p.parseAttributeGroups()
		return p.parseStatementWithAttributes(attrs)
	default:
		return p.parseExpressionStmt()
	}
}

func (p *Parser) parseStatementWithAttributes(attrs []*ast.AttributeGroup) ast.Stmt {
	p.skipWhitespace()
	switch p.curToken.Type {
	case token.T_FUNCTION:
		fn := p.parseFunctionDecl()
		fn.Attrs = attrs
		return fn
	case token.T_CLASS:
		class := p.parseClassDecl(nil)
		class.Attrs = attrs
		return class
	default:
		return p.parseExpressionStmt()
	}
}

func (p *Parser) parseBlockStmt() *ast.BlockStmt {
	block := &ast.BlockStmt{Lbrace: p.curPos()}
	p.nextToken() // skip {

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		p.skipWhitespace()
		if p.curTokenIs(token.RBRACE) {
			break
		}
		stmt := p.parseStatement()
		if stmt != nil {
			block.Stmts = append(block.Stmts, stmt)
		}
	}

	block.Rbrace = p.curPos()
	p.nextToken() // skip }
	return block
}

func (p *Parser) parseExpressionStmt() ast.Stmt {
	expr := p.parseExpression(LOWEST)
	if expr == nil {
		p.nextToken()
		return nil
	}

	stmt := &ast.ExprStmt{Expr: expr}

	p.skipWhitespace()
	if p.curTokenIs(token.SEMICOLON) {
		stmt.Semicolon = p.curPos()
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expr {
	p.skipWhitespace()

	left := p.parsePrefixExpr()
	if left == nil {
		return nil
	}

	for {
		p.skipWhitespace()
		peekPrec := p.curPrecedence()
		if peekPrec <= precedence {
			break
		}

		left = p.parseInfixExpr(left, peekPrec)
		if left == nil {
			return nil
		}
	}

	return left
}

func (p *Parser) parsePrefixExpr() ast.Expr {
	switch p.curToken.Type {
	case token.T_VARIABLE:
		return p.parseVariable()
	case token.T_LNUMBER:
		return p.parseLiteral(token.T_LNUMBER)
	case token.T_DNUMBER:
		return p.parseLiteral(token.T_DNUMBER)
	case token.T_CONSTANT_ENCAPSED_STRING:
		return p.parseLiteral(token.T_CONSTANT_ENCAPSED_STRING)
	case token.T_STRING:
		return p.parseIdentOrCall()
	case token.T_NAME_QUALIFIED, token.T_NAME_FULLY_QUALIFIED, token.T_NAME_RELATIVE:
		return p.parseNameOrCall()
	case token.LBRACKET:
		return p.parseArrayLiteral(true)
	case token.T_ARRAY:
		return p.parseArrayLiteral(false)
	case token.LPAREN:
		return p.parseParenExpr()
	case token.MINUS, token.PLUS:
		return p.parseUnaryExpr()
	case token.EXCLAMATION:
		return p.parseUnaryExpr()
	case token.TILDE:
		return p.parseUnaryExpr()
	case token.T_INC, token.T_DEC:
		return p.parsePrefixIncDec()
	case token.AT:
		return p.parseErrorSuppress()
	case token.T_INT_CAST, token.T_DOUBLE_CAST, token.T_STRING_CAST,
		token.T_ARRAY_CAST, token.T_OBJECT_CAST, token.T_BOOL_CAST, token.T_UNSET_CAST:
		return p.parseCastExpr()
	case token.T_NEW:
		return p.parseNewExpr()
	case token.T_CLONE:
		return p.parseCloneExpr()
	case token.T_FUNCTION:
		return p.parseClosureExpr()
	case token.T_FN:
		return p.parseArrowFunc()
	case token.T_STATIC:
		return p.parseStaticExpr()
	case token.T_MATCH:
		return p.parseMatchExpr()
	case token.T_YIELD:
		return p.parseYieldExpr()
	case token.T_YIELD_FROM:
		return p.parseYieldFromExpr()
	case token.T_THROW:
		return p.parseThrowExpr()
	case token.T_PRINT:
		return p.parsePrintExpr()
	case token.T_INCLUDE, token.T_INCLUDE_ONCE, token.T_REQUIRE, token.T_REQUIRE_ONCE:
		return p.parseIncludeExpr()
	case token.T_ISSET:
		return p.parseIssetExpr()
	case token.T_EMPTY:
		return p.parseEmptyExpr()
	case token.T_EVAL:
		return p.parseEvalExpr()
	case token.T_EXIT:
		return p.parseExitExpr()
	case token.T_LIST:
		return p.parseListExpr()
	case token.T_LINE, token.T_FILE, token.T_DIR, token.T_CLASS_C,
		token.T_TRAIT_C, token.T_METHOD_C, token.T_FUNC_C, token.T_NS_C:
		return p.parseMagicConst()
	case token.DOUBLE_QUOTE:
		return p.parseEncapsedString()
	case token.BACKTICK:
		return p.parseShellExec()
	case token.T_START_HEREDOC:
		return p.parseHeredoc()
	case token.AMPERSAND:
		// Could be reference - skip for now
		p.nextToken()
		return p.parseExpression(LOWEST)
	default:
		return nil
	}
}

func (p *Parser) parseInfixExpr(left ast.Expr, precedence int) ast.Expr {
	switch p.curToken.Type {
	case token.PLUS, token.MINUS, token.ASTERISK, token.SLASH, token.PERCENT,
		token.T_POW, token.DOT, token.T_SL, token.T_SR,
		token.PIPE, token.AMPERSAND, token.CARET,
		token.T_BOOLEAN_AND, token.T_BOOLEAN_OR,
		token.T_LOGICAL_AND, token.T_LOGICAL_OR, token.T_LOGICAL_XOR,
		token.T_IS_EQUAL, token.T_IS_NOT_EQUAL, token.T_IS_IDENTICAL, token.T_IS_NOT_IDENTICAL,
		token.LESS, token.GREATER, token.T_IS_SMALLER_OR_EQUAL, token.T_IS_GREATER_OR_EQUAL,
		token.T_SPACESHIP, token.T_PIPE:
		return p.parseBinaryExpr(left, precedence)
	case token.EQUALS, token.T_PLUS_EQUAL, token.T_MINUS_EQUAL, token.T_MUL_EQUAL,
		token.T_DIV_EQUAL, token.T_MOD_EQUAL, token.T_POW_EQUAL, token.T_CONCAT_EQUAL,
		token.T_AND_EQUAL, token.T_OR_EQUAL, token.T_XOR_EQUAL,
		token.T_SL_EQUAL, token.T_SR_EQUAL, token.T_COALESCE_EQUAL:
		return p.parseAssignExpr(left)
	case token.T_COALESCE:
		return p.parseCoalesceExpr(left)
	case token.QUESTION:
		return p.parseTernaryExpr(left)
	case token.T_INSTANCEOF:
		return p.parseInstanceofExpr(left)
	case token.LPAREN:
		return p.parseCallExpr(left)
	case token.LBRACKET:
		return p.parseArrayAccessExpr(left)
	case token.T_OBJECT_OPERATOR, token.T_NULLSAFE_OBJECT_OPERATOR:
		return p.parsePropertyOrMethodExpr(left)
	case token.T_PAAMAYIM_NEKUDOTAYIM:
		return p.parseStaticAccessExpr(left)
	case token.T_INC, token.T_DEC:
		return p.parsePostfixExpr(left)
	default:
		return left
	}
}

func (p *Parser) parseVariable() ast.Expr {
	v := &ast.Variable{
		DollarPos: p.curPos(),
		Name: &ast.Ident{
			NamePos: p.curPos(),
			Name:    p.curToken.Literal[1:], // Remove $
		},
	}
	p.nextToken()
	return v
}

func (p *Parser) parseLiteral(kind token.Token) ast.Expr {
	lit := &ast.Literal{
		ValuePos: p.curPos(),
		Kind:     kind,
		Value:    p.curToken.Literal,
	}
	p.nextToken()
	return lit
}

func (p *Parser) parseIdentOrCall() ast.Expr {
	ident := &ast.Ident{
		NamePos: p.curPos(),
		Name:    p.curToken.Literal,
	}
	p.nextToken()
	return ident
}

func (p *Parser) parseNameOrCall() ast.Expr {
	ident := &ast.Ident{
		NamePos: p.curPos(),
		Name:    p.curToken.Literal,
	}
	p.nextToken()
	return ident
}

func (p *Parser) parseArrayLiteral(isShort bool) ast.Expr {
	arr := &ast.ArrayExpr{
		Lbrack:  p.curPos(),
		IsShort: isShort,
	}
	p.nextToken() // skip [ or 'array'

	if !isShort {
		p.skipWhitespace()
		if !p.curTokenIs(token.LPAREN) {
			return arr
		}
		p.nextToken() // skip (
	}

	endToken := token.RBRACKET
	if !isShort {
		endToken = token.RPAREN
	}

	for !p.curTokenIs(endToken) && !p.curTokenIs(token.EOF) {
		p.skipWhitespace()
		if p.curTokenIs(endToken) {
			break
		}

		item := &ast.ArrayItem{}

		// Check for spread
		if p.curTokenIs(token.T_ELLIPSIS) {
			item.Unpack = true
			p.nextToken()
			p.skipWhitespace()
		}

		// Parse key or value
		expr := p.parseExpression(LOWEST)
		item.Value = expr

		p.skipWhitespace()
		if p.curTokenIs(token.T_DOUBLE_ARROW) {
			// This was a key
			item.Key = expr
			p.nextToken()
			p.skipWhitespace()

			// Check for reference
			if p.curTokenIs(token.AMPERSAND) {
				item.ByRef = true
				p.nextToken()
				p.skipWhitespace()
			}

			item.Value = p.parseExpression(LOWEST)
		}

		arr.Items = append(arr.Items, item)

		p.skipWhitespace()
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	arr.Rbrack = p.curPos()
	p.nextToken()
	return arr
}

func (p *Parser) parseParenExpr() ast.Expr {
	lparen := p.curPos()
	p.nextToken() // skip (
	p.skipWhitespace()

	expr := p.parseExpression(LOWEST)

	p.skipWhitespace()
	rparen := p.curPos()
	if p.curTokenIs(token.RPAREN) {
		p.nextToken()
	}

	return &ast.ParenExpr{
		Lparen: lparen,
		X:      expr,
		Rparen: rparen,
	}
}

func (p *Parser) parseUnaryExpr() ast.Expr {
	un := &ast.UnaryExpr{
		OpPos: p.curPos(),
		Op:    p.curToken.Type,
	}
	p.nextToken()
	p.skipWhitespace()
	un.X = p.parseExpression(PREFIX)
	return un
}

func (p *Parser) parsePrefixIncDec() ast.Expr {
	un := &ast.UnaryExpr{
		OpPos: p.curPos(),
		Op:    p.curToken.Type,
	}
	p.nextToken()
	p.skipWhitespace()
	un.X = p.parseExpression(PREFIX)
	return un
}

func (p *Parser) parseErrorSuppress() ast.Expr {
	expr := &ast.ErrorSuppressExpr{AtPos: p.curPos()}
	p.nextToken()
	p.skipWhitespace()
	expr.Expr = p.parseExpression(PREFIX)
	return expr
}

func (p *Parser) parseCastExpr() ast.Expr {
	cast := &ast.CastExpr{
		CastPos: p.curPos(),
		Type:    p.curToken.Type,
	}
	p.nextToken()
	p.skipWhitespace()
	cast.X = p.parseExpression(PREFIX)
	return cast
}

func (p *Parser) parseNewExpr() ast.Expr {
	new_ := &ast.NewExpr{NewPos: p.curPos()}
	p.nextToken()
	p.skipWhitespace()

	new_.Class = p.parseExpression(CALL)

	p.skipWhitespace()
	if p.curTokenIs(token.LPAREN) {
		new_.Args = p.parseArgumentList()
	}

	return new_
}

func (p *Parser) parseCloneExpr() ast.Expr {
	clone := &ast.CloneExpr{ClonePos: p.curPos()}
	p.nextToken()
	p.skipWhitespace()
	clone.Expr = p.parseExpression(PREFIX)
	return clone
}

func (p *Parser) parseClosureExpr() ast.Expr {
	closure := &ast.ClosureExpr{FuncPos: p.curPos()}
	p.nextToken() // skip function
	p.skipWhitespace()

	// Check for reference
	if p.curTokenIs(token.AMPERSAND) {
		closure.ByRef = true
		p.nextToken()
		p.skipWhitespace()
	}

	// Parameters
	if p.curTokenIs(token.LPAREN) {
		closure.Params = p.parseParameterList()
	}

	p.skipWhitespace()

	// Use clause
	if p.curTokenIs(token.T_USE) {
		p.nextToken()
		p.skipWhitespace()
		if p.curTokenIs(token.LPAREN) {
			p.nextToken()
			for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
				p.skipWhitespace()

				use := &ast.ClosureUse{}
				if p.curTokenIs(token.AMPERSAND) {
					use.ByRef = true
					p.nextToken()
					p.skipWhitespace()
				}

				if p.curTokenIs(token.T_VARIABLE) {
					use.Var = p.parseVariable().(*ast.Variable)
				}
				closure.Uses = append(closure.Uses, use)

				p.skipWhitespace()
				if p.curTokenIs(token.COMMA) {
					p.nextToken()
				}
			}
			p.nextToken() // skip )
		}
	}

	p.skipWhitespace()

	// Return type
	if p.curTokenIs(token.COLON) {
		p.nextToken()
		p.skipWhitespace()
		closure.ReturnType = p.parseTypeExpr()
	}

	p.skipWhitespace()

	// Body
	if p.curTokenIs(token.LBRACE) {
		closure.Body = p.parseBlockStmt()
	}

	return closure
}

func (p *Parser) parseArrowFunc() ast.Expr {
	arrow := &ast.ArrowFuncExpr{FnPos: p.curPos()}
	p.nextToken() // skip fn
	p.skipWhitespace()

	// Check for reference
	if p.curTokenIs(token.AMPERSAND) {
		arrow.ByRef = true
		p.nextToken()
		p.skipWhitespace()
	}

	// Parameters
	if p.curTokenIs(token.LPAREN) {
		arrow.Params = p.parseParameterList()
	}

	p.skipWhitespace()

	// Return type
	if p.curTokenIs(token.COLON) {
		p.nextToken()
		p.skipWhitespace()
		arrow.ReturnType = p.parseTypeExpr()
	}

	p.skipWhitespace()

	// Arrow and body
	if p.curTokenIs(token.T_DOUBLE_ARROW) {
		arrow.Arrow = p.curPos()
		p.nextToken()
		p.skipWhitespace()
		arrow.Body = p.parseExpression(LOWEST)
	}

	return arrow
}

func (p *Parser) parseStaticExpr() ast.Expr {
	// Could be static::, static function, or static fn
	pos := p.curPos()
	p.nextToken()
	p.skipWhitespace()

	if p.curTokenIs(token.T_PAAMAYIM_NEKUDOTAYIM) {
		// static::
		ident := &ast.Ident{NamePos: pos, Name: "static"}
		return p.parseStaticAccessExpr(ident)
	}

	if p.curTokenIs(token.T_FUNCTION) {
		closure := p.parseClosureExpr().(*ast.ClosureExpr)
		closure.Static = true
		return closure
	}

	if p.curTokenIs(token.T_FN) {
		arrow := p.parseArrowFunc().(*ast.ArrowFuncExpr)
		arrow.Static = true
		return arrow
	}

	return &ast.Ident{NamePos: pos, Name: "static"}
}

func (p *Parser) parseMatchExpr() ast.Expr {
	match := &ast.MatchExpr{MatchPos: p.curPos()}
	p.nextToken() // skip match
	p.skipWhitespace()

	// Condition
	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		p.skipWhitespace()
		match.Cond = p.parseExpression(LOWEST)
		p.skipWhitespace()
		if p.curTokenIs(token.RPAREN) {
			p.nextToken()
		}
	}

	p.skipWhitespace()

	// Arms
	if p.curTokenIs(token.LBRACE) {
		match.Lbrace = p.curPos()
		p.nextToken()

		for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
			p.skipWhitespace()
			if p.curTokenIs(token.RBRACE) {
				break
			}

			arm := &ast.MatchArm{}

			// Check for default
			if p.curTokenIs(token.T_DEFAULT) {
				p.nextToken()
			} else {
				// Parse conditions
				for {
					p.skipWhitespace()
					arm.Conds = append(arm.Conds, p.parseExpression(LOWEST))
					p.skipWhitespace()
					if p.curTokenIs(token.COMMA) {
						p.nextToken()
						p.skipWhitespace()
						if p.curTokenIs(token.T_DOUBLE_ARROW) {
							break
						}
					} else {
						break
					}
				}
			}

			p.skipWhitespace()
			if p.curTokenIs(token.T_DOUBLE_ARROW) {
				arm.Arrow = p.curPos()
				p.nextToken()
				p.skipWhitespace()
				arm.Body = p.parseExpression(LOWEST)
			}

			match.Arms = append(match.Arms, arm)

			p.skipWhitespace()
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
			}
		}

		match.Rbrace = p.curPos()
		p.nextToken()
	}

	return match
}

func (p *Parser) parseYieldExpr() ast.Expr {
	yield := &ast.YieldExpr{YieldPos: p.curPos()}
	p.nextToken()
	p.skipWhitespace()

	if p.curTokenIs(token.SEMICOLON) || p.curTokenIs(token.RPAREN) {
		return yield
	}

	expr := p.parseExpression(LOWEST)
	yield.Value = expr

	p.skipWhitespace()
	if p.curTokenIs(token.T_DOUBLE_ARROW) {
		yield.Key = expr
		p.nextToken()
		p.skipWhitespace()
		yield.Value = p.parseExpression(LOWEST)
	}

	return yield
}

func (p *Parser) parseYieldFromExpr() ast.Expr {
	yield := &ast.YieldFromExpr{YieldPos: p.curPos()}
	p.nextToken()
	p.skipWhitespace()
	yield.Expr = p.parseExpression(LOWEST)
	return yield
}

func (p *Parser) parseThrowExpr() ast.Expr {
	throw := &ast.ThrowExpr{ThrowPos: p.curPos()}
	p.nextToken()
	p.skipWhitespace()
	throw.Expr = p.parseExpression(LOWEST)
	return throw
}

func (p *Parser) parsePrintExpr() ast.Expr {
	print := &ast.PrintExpr{PrintPos: p.curPos()}
	p.nextToken()
	p.skipWhitespace()
	print.Expr = p.parseExpression(LOWEST)
	return print
}

func (p *Parser) parseIncludeExpr() ast.Expr {
	include := &ast.IncludeExpr{
		IncludePos: p.curPos(),
		Type:       p.curToken.Type,
	}
	p.nextToken()
	p.skipWhitespace()
	include.Expr = p.parseExpression(LOWEST)
	return include
}

func (p *Parser) parseIssetExpr() ast.Expr {
	isset := &ast.IssetExpr{IssetPos: p.curPos()}
	p.nextToken()
	p.skipWhitespace()

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
			p.skipWhitespace()
			isset.Vars = append(isset.Vars, p.parseExpression(LOWEST))
			p.skipWhitespace()
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
			}
		}
		isset.Rparen = p.curPos()
		p.nextToken()
	}

	return isset
}

func (p *Parser) parseEmptyExpr() ast.Expr {
	empty := &ast.EmptyExpr{EmptyPos: p.curPos()}
	p.nextToken()
	p.skipWhitespace()

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		p.skipWhitespace()
		empty.Expr = p.parseExpression(LOWEST)
		p.skipWhitespace()
		if p.curTokenIs(token.RPAREN) {
			empty.Rparen = p.curPos()
			p.nextToken()
		}
	}

	return empty
}

func (p *Parser) parseEvalExpr() ast.Expr {
	eval := &ast.EvalExpr{EvalPos: p.curPos()}
	p.nextToken()
	p.skipWhitespace()

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		p.skipWhitespace()
		eval.Expr = p.parseExpression(LOWEST)
		p.skipWhitespace()
		if p.curTokenIs(token.RPAREN) {
			eval.Rparen = p.curPos()
			p.nextToken()
		}
	}

	return eval
}

func (p *Parser) parseExitExpr() ast.Expr {
	exit := &ast.ExitExpr{ExitPos: p.curPos()}
	p.nextToken()
	p.skipWhitespace()

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		p.skipWhitespace()
		if !p.curTokenIs(token.RPAREN) {
			exit.Expr = p.parseExpression(LOWEST)
		}
		p.skipWhitespace()
		if p.curTokenIs(token.RPAREN) {
			p.nextToken()
		}
	}

	return exit
}

func (p *Parser) parseListExpr() ast.Expr {
	list := &ast.ListExpr{ListPos: p.curPos()}
	p.nextToken() // skip list
	p.skipWhitespace()

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
			p.skipWhitespace()
			if p.curTokenIs(token.COMMA) {
				list.Items = append(list.Items, nil)
				p.nextToken()
				continue
			}

			item := &ast.ArrayItem{}
			expr := p.parseExpression(LOWEST)
			item.Value = expr

			p.skipWhitespace()
			if p.curTokenIs(token.T_DOUBLE_ARROW) {
				item.Key = expr
				p.nextToken()
				p.skipWhitespace()
				item.Value = p.parseExpression(LOWEST)
			}

			list.Items = append(list.Items, item)
			p.skipWhitespace()
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
			}
		}
		list.Rparen = p.curPos()
		p.nextToken()
	}

	return list
}

func (p *Parser) parseMagicConst() ast.Expr {
	magic := &ast.MagicConstExpr{
		ConstPos: p.curPos(),
		Kind:     p.curToken.Type,
	}
	p.nextToken()
	return magic
}

func (p *Parser) parseEncapsedString() ast.Expr {
	enc := &ast.EncapsedStringExpr{OpenQuote: p.curPos()}
	p.nextToken() // skip opening "

	for !p.curTokenIs(token.DOUBLE_QUOTE) && !p.curTokenIs(token.EOF) {
		switch p.curToken.Type {
		case token.T_ENCAPSED_AND_WHITESPACE:
			enc.Parts = append(enc.Parts, &ast.Literal{
				ValuePos: p.curPos(),
				Kind:     token.T_ENCAPSED_AND_WHITESPACE,
				Value:    p.curToken.Literal,
			})
			p.nextToken()
		case token.T_VARIABLE:
			enc.Parts = append(enc.Parts, p.parseVariable())
		case token.T_CURLY_OPEN:
			p.nextToken()
			enc.Parts = append(enc.Parts, p.parseExpression(LOWEST))
			p.skipWhitespace()
			if p.curTokenIs(token.RBRACE) {
				p.nextToken()
			}
		case token.T_DOLLAR_OPEN_CURLY_BRACES:
			p.nextToken()
			enc.Parts = append(enc.Parts, p.parseExpression(LOWEST))
			p.skipWhitespace()
			if p.curTokenIs(token.RBRACE) {
				p.nextToken()
			}
		default:
			p.nextToken()
		}
	}

	enc.CloseQuote = p.curPos()
	if p.curTokenIs(token.DOUBLE_QUOTE) {
		p.nextToken()
	}

	return enc
}

func (p *Parser) parseShellExec() ast.Expr {
	shell := &ast.ShellExecExpr{OpenTick: p.curPos()}
	p.nextToken() // skip opening `

	for !p.curTokenIs(token.BACKTICK) && !p.curTokenIs(token.EOF) {
		switch p.curToken.Type {
		case token.T_ENCAPSED_AND_WHITESPACE:
			shell.Parts = append(shell.Parts, &ast.Literal{
				ValuePos: p.curPos(),
				Kind:     token.T_ENCAPSED_AND_WHITESPACE,
				Value:    p.curToken.Literal,
			})
			p.nextToken()
		case token.T_VARIABLE:
			shell.Parts = append(shell.Parts, p.parseVariable())
		default:
			p.nextToken()
		}
	}

	shell.CloseTick = p.curPos()
	if p.curTokenIs(token.BACKTICK) {
		p.nextToken()
	}

	return shell
}

func (p *Parser) parseHeredoc() ast.Expr {
	heredoc := &ast.HeredocExpr{StartPos: p.curPos()}
	p.nextToken()

	for !p.curTokenIs(token.T_END_HEREDOC) && !p.curTokenIs(token.EOF) {
		switch p.curToken.Type {
		case token.T_ENCAPSED_AND_WHITESPACE:
			heredoc.Parts = append(heredoc.Parts, &ast.Literal{
				ValuePos: p.curPos(),
				Kind:     token.T_ENCAPSED_AND_WHITESPACE,
				Value:    p.curToken.Literal,
			})
			p.nextToken()
		case token.T_VARIABLE:
			heredoc.Parts = append(heredoc.Parts, p.parseVariable())
		default:
			p.nextToken()
		}
	}

	heredoc.EndPos = p.curPos()
	if p.curTokenIs(token.T_END_HEREDOC) {
		p.nextToken()
	}

	return heredoc
}

func (p *Parser) parseBinaryExpr(left ast.Expr, precedence int) ast.Expr {
	bin := &ast.BinaryExpr{
		Left:  left,
		OpPos: p.curPos(),
		Op:    p.curToken.Type,
	}
	p.nextToken()
	p.skipWhitespace()
	bin.Right = p.parseExpression(precedence)
	return bin
}

func (p *Parser) parseAssignExpr(left ast.Expr) ast.Expr {
	assign := &ast.AssignExpr{
		Var:   left,
		OpPos: p.curPos(),
		Op:    p.curToken.Type,
	}
	p.nextToken()
	p.skipWhitespace()

	// Check for reference assignment
	if p.curTokenIs(token.AMPERSAND) {
		ampPos := p.curPos()
		p.nextToken()
		p.skipWhitespace()
		return &ast.AssignRefExpr{
			Var:    left,
			Equals: assign.OpPos,
			AmpPos: ampPos,
			Value:  p.parseExpression(ASSIGN - 1),
		}
	}

	assign.Value = p.parseExpression(ASSIGN - 1) // Right-associative
	return assign
}

func (p *Parser) parseCoalesceExpr(left ast.Expr) ast.Expr {
	coal := &ast.CoalesceExpr{
		Left:  left,
		OpPos: p.curPos(),
	}
	p.nextToken()
	p.skipWhitespace()
	coal.Right = p.parseExpression(COALESCE - 1) // Right-associative
	return coal
}

func (p *Parser) parseTernaryExpr(left ast.Expr) ast.Expr {
	tern := &ast.TernaryExpr{
		Cond:     left,
		Question: p.curPos(),
	}
	p.nextToken()
	p.skipWhitespace()

	// Check for Elvis operator (?:)
	if p.curTokenIs(token.COLON) {
		tern.Colon = p.curPos()
		p.nextToken()
		p.skipWhitespace()
		tern.Else = p.parseExpression(TERNARY - 1)
		return tern
	}

	tern.Then = p.parseExpression(LOWEST)
	p.skipWhitespace()

	if p.curTokenIs(token.COLON) {
		tern.Colon = p.curPos()
		p.nextToken()
		p.skipWhitespace()
		tern.Else = p.parseExpression(TERNARY - 1)
	}

	return tern
}

func (p *Parser) parseInstanceofExpr(left ast.Expr) ast.Expr {
	inst := &ast.InstanceofExpr{
		Expr:  left,
		OpPos: p.curPos(),
	}
	p.nextToken()
	p.skipWhitespace()
	inst.Class = p.parseExpression(INSTANCEOF)
	return inst
}

func (p *Parser) parseCallExpr(left ast.Expr) ast.Expr {
	call := &ast.CallExpr{
		Func: left,
		Args: p.parseArgumentList(),
	}
	return call
}

func (p *Parser) parseArgumentList() *ast.ArgumentList {
	args := &ast.ArgumentList{Lparen: p.curPos()}
	p.nextToken() // skip (

	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		p.skipWhitespace()
		if p.curTokenIs(token.RPAREN) {
			break
		}

		arg := &ast.Argument{}

		// Check for spread
		if p.curTokenIs(token.T_ELLIPSIS) {
			arg.Unpack = true
			p.nextToken()
			p.skipWhitespace()
		}

		// Check for named argument
		if p.curTokenIs(token.T_STRING) && p.peekTokenIs(token.COLON) {
			arg.Name = &ast.Ident{
				NamePos: p.curPos(),
				Name:    p.curToken.Literal,
			}
			p.nextToken() // skip name
			p.nextToken() // skip :
			p.skipWhitespace()
		}

		arg.Value = p.parseExpression(LOWEST)
		args.Args = append(args.Args, arg)

		p.skipWhitespace()
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	args.Rparen = p.curPos()
	if p.curTokenIs(token.RPAREN) {
		p.nextToken()
	}

	return args
}

func (p *Parser) parseArrayAccessExpr(left ast.Expr) ast.Expr {
	access := &ast.ArrayAccessExpr{
		Array:  left,
		Lbrack: p.curPos(),
	}
	p.nextToken() // skip [
	p.skipWhitespace()

	if !p.curTokenIs(token.RBRACKET) {
		access.Index = p.parseExpression(LOWEST)
	}

	p.skipWhitespace()
	access.Rbrack = p.curPos()
	if p.curTokenIs(token.RBRACKET) {
		p.nextToken()
	}

	return access
}

func (p *Parser) parsePropertyOrMethodExpr(left ast.Expr) ast.Expr {
	nullSafe := p.curTokenIs(token.T_NULLSAFE_OBJECT_OPERATOR)
	arrowPos := p.curPos()
	p.nextToken()
	p.skipWhitespace()

	var property ast.Expr
	if p.curTokenIs(token.T_STRING) {
		property = &ast.Ident{
			NamePos: p.curPos(),
			Name:    p.curToken.Literal,
		}
		p.nextToken()
	} else if p.curTokenIs(token.T_VARIABLE) {
		property = p.parseVariable()
	} else if p.curTokenIs(token.LBRACE) {
		p.nextToken()
		property = p.parseExpression(LOWEST)
		p.skipWhitespace()
		if p.curTokenIs(token.RBRACE) {
			p.nextToken()
		}
	}

	p.skipWhitespace()

	// Check if it's a method call
	if p.curTokenIs(token.LPAREN) {
		return &ast.MethodCallExpr{
			Object:   left,
			Arrow:    arrowPos,
			NullSafe: nullSafe,
			Method:   property,
			Args:     p.parseArgumentList(),
		}
	}

	return &ast.PropertyFetchExpr{
		Object:   left,
		Arrow:    arrowPos,
		NullSafe: nullSafe,
		Property: property,
	}
}

func (p *Parser) parseStaticAccessExpr(left ast.Expr) ast.Expr {
	doubleColonPos := p.curPos()
	p.nextToken() // skip ::
	p.skipWhitespace()

	// Check if it's a variable (static property)
	if p.curTokenIs(token.T_VARIABLE) {
		return &ast.StaticPropertyFetchExpr{
			Class:       left,
			DoubleColon: doubleColonPos,
			Property:    p.parseVariable(),
		}
	}

	// Constant or method
	var member ast.Expr
	if p.curTokenIs(token.T_STRING) || p.curTokenIs(token.T_CLASS) {
		member = &ast.Ident{
			NamePos: p.curPos(),
			Name:    p.curToken.Literal,
		}
		p.nextToken()
	}

	p.skipWhitespace()

	// Check if it's a static method call
	if p.curTokenIs(token.LPAREN) {
		return &ast.StaticCallExpr{
			Class:       left,
			DoubleColon: doubleColonPos,
			Method:      member,
			Args:        p.parseArgumentList(),
		}
	}

	// Class constant
	return &ast.ClassConstFetchExpr{
		Class:       left,
		DoubleColon: doubleColonPos,
		Const:       member.(*ast.Ident),
	}
}

func (p *Parser) parsePostfixExpr(left ast.Expr) ast.Expr {
	post := &ast.PostfixExpr{
		X:     left,
		OpPos: p.curPos(),
		Op:    p.curToken.Type,
	}
	p.nextToken()
	return post
}

func (p *Parser) parseAttributeGroups() []*ast.AttributeGroup {
	var groups []*ast.AttributeGroup

	for p.curTokenIs(token.T_ATTRIBUTE) {
		group := &ast.AttributeGroup{HashBracket: p.curPos()}
		p.nextToken() // skip #[

		for !p.curTokenIs(token.RBRACKET) && !p.curTokenIs(token.EOF) {
			p.skipWhitespace()
			if p.curTokenIs(token.RBRACKET) {
				break
			}

			attr := &ast.Attribute{}
			if p.curTokenIs(token.T_STRING) || p.curTokenIs(token.T_NAME_QUALIFIED) {
				attr.Name = &ast.Ident{
					NamePos: p.curPos(),
					Name:    p.curToken.Literal,
				}
				p.nextToken()
			}

			p.skipWhitespace()
			if p.curTokenIs(token.LPAREN) {
				attr.Args = p.parseArgumentList()
			}

			group.Attrs = append(group.Attrs, attr)
			p.skipWhitespace()
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
			}
		}

		group.Rbrack = p.curPos()
		if p.curTokenIs(token.RBRACKET) {
			p.nextToken()
		}

		groups = append(groups, group)
		p.skipWhitespace()
	}

	return groups
}
