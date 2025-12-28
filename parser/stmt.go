package parser

import (
	"github.com/alexisbouchez/phpgo/ast"
	"github.com/alexisbouchez/phpgo/token"
)

func (p *Parser) parseIfStmt() *ast.IfStmt {
	ifStmt := &ast.IfStmt{IfPos: p.curPos()}
	p.nextToken() // skip if
	p.skipWhitespace()

	// Condition
	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		p.skipWhitespace()
		ifStmt.Cond = p.parseExpression(LOWEST)
		p.skipWhitespace()
		if p.curTokenIs(token.RPAREN) {
			p.nextToken()
		}
	}

	p.skipWhitespace()

	// Check for alternative syntax
	if p.curTokenIs(token.COLON) {
		ifStmt.UseAltSyntax = true
		p.nextToken()
		ifStmt.Body = p.parseAltSyntaxBody(token.T_ELSEIF, token.T_ELSE, token.T_ENDIF)
	} else {
		ifStmt.Body = p.parseStatementBody()
	}

	p.skipWhitespace()

	// Elseif clauses
	for p.curTokenIs(token.T_ELSEIF) {
		elseif := &ast.ElseIfClause{ElseIfPos: p.curPos()}
		p.nextToken()
		p.skipWhitespace()

		if p.curTokenIs(token.LPAREN) {
			p.nextToken()
			p.skipWhitespace()
			elseif.Cond = p.parseExpression(LOWEST)
			p.skipWhitespace()
			if p.curTokenIs(token.RPAREN) {
				p.nextToken()
			}
		}

		p.skipWhitespace()

		if ifStmt.UseAltSyntax {
			p.nextToken() // skip :
			elseif.Body = p.parseAltSyntaxBody(token.T_ELSEIF, token.T_ELSE, token.T_ENDIF)
		} else {
			elseif.Body = p.parseStatementBody()
		}

		ifStmt.ElseIfs = append(ifStmt.ElseIfs, elseif)
		p.skipWhitespace()
	}

	// Else clause
	if p.curTokenIs(token.T_ELSE) {
		elseClause := &ast.ElseClause{ElsePos: p.curPos()}
		p.nextToken()
		p.skipWhitespace()

		if ifStmt.UseAltSyntax {
			p.nextToken() // skip :
			elseClause.Body = p.parseAltSyntaxBody(token.T_ENDIF)
		} else {
			elseClause.Body = p.parseStatementBody()
		}

		ifStmt.Else = elseClause
	}

	// End if (alternative syntax)
	if ifStmt.UseAltSyntax && p.curTokenIs(token.T_ENDIF) {
		ifStmt.EndIf = p.curPos()
		p.nextToken()
		p.skipWhitespace()
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	return ifStmt
}

func (p *Parser) parseStatementBody() ast.Stmt {
	if p.curTokenIs(token.LBRACE) {
		return p.parseBlockStmt()
	}
	return p.parseStatement()
}

func (p *Parser) parseAltSyntaxBody(endTokens ...token.Token) ast.Stmt {
	block := &ast.BlockStmt{Lbrace: p.curPos()}

	for !p.curTokenIs(token.EOF) {
		for _, end := range endTokens {
			if p.curTokenIs(end) {
				return block
			}
		}
		p.skipWhitespace()
		stmt := p.parseStatement()
		if stmt != nil {
			block.Stmts = append(block.Stmts, stmt)
		}
	}

	return block
}

func (p *Parser) parseWhileStmt() *ast.WhileStmt {
	whileStmt := &ast.WhileStmt{WhilePos: p.curPos()}
	p.nextToken() // skip while
	p.skipWhitespace()

	// Condition
	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		p.skipWhitespace()
		whileStmt.Cond = p.parseExpression(LOWEST)
		p.skipWhitespace()
		if p.curTokenIs(token.RPAREN) {
			p.nextToken()
		}
	}

	p.skipWhitespace()

	// Body
	if p.curTokenIs(token.COLON) {
		whileStmt.UseAltSyntax = true
		p.nextToken()
		whileStmt.Body = p.parseAltSyntaxBody(token.T_ENDWHILE)
		if p.curTokenIs(token.T_ENDWHILE) {
			p.nextToken()
			p.skipWhitespace()
			if p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}
	} else {
		whileStmt.Body = p.parseStatementBody()
	}

	return whileStmt
}

func (p *Parser) parseDoWhileStmt() *ast.DoWhileStmt {
	doStmt := &ast.DoWhileStmt{DoPos: p.curPos()}
	p.nextToken() // skip do
	p.skipWhitespace()

	// Body
	doStmt.Body = p.parseStatementBody()
	p.skipWhitespace()

	// While
	if p.curTokenIs(token.T_WHILE) {
		doStmt.WhilePos = p.curPos()
		p.nextToken()
		p.skipWhitespace()

		if p.curTokenIs(token.LPAREN) {
			p.nextToken()
			p.skipWhitespace()
			doStmt.Cond = p.parseExpression(LOWEST)
			p.skipWhitespace()
			if p.curTokenIs(token.RPAREN) {
				p.nextToken()
			}
		}
	}

	p.skipWhitespace()
	if p.curTokenIs(token.SEMICOLON) {
		doStmt.Semicolon = p.curPos()
		p.nextToken()
	}

	return doStmt
}

func (p *Parser) parseForStmt() *ast.ForStmt {
	forStmt := &ast.ForStmt{ForPos: p.curPos()}
	p.nextToken() // skip for
	p.skipWhitespace()

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		p.skipWhitespace()

		// Init
		if !p.curTokenIs(token.SEMICOLON) {
			forStmt.Init = p.parseExpressionList()
		}
		p.skipWhitespace()
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}

		p.skipWhitespace()

		// Condition
		if !p.curTokenIs(token.SEMICOLON) {
			forStmt.Cond = p.parseExpressionList()
		}
		p.skipWhitespace()
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}

		p.skipWhitespace()

		// Loop
		if !p.curTokenIs(token.RPAREN) {
			forStmt.Loop = p.parseExpressionList()
		}
		p.skipWhitespace()
		if p.curTokenIs(token.RPAREN) {
			p.nextToken()
		}
	}

	p.skipWhitespace()

	// Body
	if p.curTokenIs(token.COLON) {
		forStmt.UseAltSyntax = true
		p.nextToken()
		forStmt.Body = p.parseAltSyntaxBody(token.T_ENDFOR)
		if p.curTokenIs(token.T_ENDFOR) {
			p.nextToken()
			p.skipWhitespace()
			if p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}
	} else {
		forStmt.Body = p.parseStatementBody()
	}

	return forStmt
}

func (p *Parser) parseExpressionList() []ast.Expr {
	var exprs []ast.Expr
	for {
		exprs = append(exprs, p.parseExpression(LOWEST))
		p.skipWhitespace()
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			p.skipWhitespace()
		} else {
			break
		}
	}
	return exprs
}

func (p *Parser) parseForeachStmt() *ast.ForeachStmt {
	foreachStmt := &ast.ForeachStmt{ForeachPos: p.curPos()}
	p.nextToken() // skip foreach
	p.skipWhitespace()

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		p.skipWhitespace()

		// Expression
		foreachStmt.Expr = p.parseExpression(LOWEST)
		p.skipWhitespace()

		// as
		if p.curTokenIs(token.T_AS) {
			p.nextToken()
			p.skipWhitespace()
		}

		// Check for reference
		if p.curTokenIs(token.AMPERSAND) {
			foreachStmt.ByRef = true
			p.nextToken()
			p.skipWhitespace()
		}

		// Key or value
		first := p.parseExpression(LOWEST)
		p.skipWhitespace()

		if p.curTokenIs(token.T_DOUBLE_ARROW) {
			foreachStmt.KeyVar = first
			p.nextToken()
			p.skipWhitespace()

			if p.curTokenIs(token.AMPERSAND) {
				foreachStmt.ByRef = true
				p.nextToken()
				p.skipWhitespace()
			}

			foreachStmt.ValueVar = p.parseExpression(LOWEST)
		} else {
			foreachStmt.ValueVar = first
		}

		p.skipWhitespace()
		if p.curTokenIs(token.RPAREN) {
			p.nextToken()
		}
	}

	p.skipWhitespace()

	// Body
	if p.curTokenIs(token.COLON) {
		foreachStmt.UseAltSyntax = true
		p.nextToken()
		foreachStmt.Body = p.parseAltSyntaxBody(token.T_ENDFOREACH)
		if p.curTokenIs(token.T_ENDFOREACH) {
			p.nextToken()
			p.skipWhitespace()
			if p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}
	} else {
		foreachStmt.Body = p.parseStatementBody()
	}

	return foreachStmt
}

func (p *Parser) parseSwitchStmt() *ast.SwitchStmt {
	switchStmt := &ast.SwitchStmt{SwitchPos: p.curPos()}
	p.nextToken() // skip switch
	p.skipWhitespace()

	// Condition
	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		p.skipWhitespace()
		switchStmt.Cond = p.parseExpression(LOWEST)
		p.skipWhitespace()
		if p.curTokenIs(token.RPAREN) {
			p.nextToken()
		}
	}

	p.skipWhitespace()

	// Body
	if p.curTokenIs(token.COLON) {
		switchStmt.UseAltSyntax = true
		p.nextToken()
	} else if p.curTokenIs(token.LBRACE) {
		switchStmt.Lbrace = p.curPos()
		p.nextToken()
	}

	endToken := token.RBRACE
	if switchStmt.UseAltSyntax {
		endToken = token.T_ENDSWITCH
	}

	for !p.curTokenIs(endToken) && !p.curTokenIs(token.EOF) {
		p.skipWhitespace()
		if p.curTokenIs(endToken) {
			break
		}

		caseClause := &ast.CaseClause{}

		if p.curTokenIs(token.T_CASE) {
			caseClause.CasePos = p.curPos()
			p.nextToken()
			p.skipWhitespace()
			caseClause.Cond = p.parseExpression(LOWEST)
		} else if p.curTokenIs(token.T_DEFAULT) {
			caseClause.CasePos = p.curPos()
			p.nextToken()
		}

		p.skipWhitespace()
		if p.curTokenIs(token.COLON) || p.curTokenIs(token.SEMICOLON) {
			caseClause.Separator = p.curPos()
			p.nextToken()
		}

		// Parse case statements
		for !p.curTokenIs(token.T_CASE) && !p.curTokenIs(token.T_DEFAULT) &&
			!p.curTokenIs(endToken) && !p.curTokenIs(token.EOF) {
			p.skipWhitespace()
			if p.curTokenIs(token.T_CASE) || p.curTokenIs(token.T_DEFAULT) || p.curTokenIs(endToken) {
				break
			}
			stmt := p.parseStatement()
			if stmt != nil {
				caseClause.Stmts = append(caseClause.Stmts, stmt)
			}
		}

		switchStmt.Cases = append(switchStmt.Cases, caseClause)
	}

	switchStmt.Rbrace = p.curPos()
	p.nextToken()

	if switchStmt.UseAltSyntax {
		p.skipWhitespace()
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	return switchStmt
}

func (p *Parser) parseTryStmt() *ast.TryStmt {
	tryStmt := &ast.TryStmt{TryPos: p.curPos()}
	p.nextToken() // skip try
	p.skipWhitespace()

	// Body
	if p.curTokenIs(token.LBRACE) {
		tryStmt.Body = p.parseBlockStmt()
	}

	p.skipWhitespace()

	// Catch clauses
	for p.curTokenIs(token.T_CATCH) {
		catch := &ast.CatchClause{CatchPos: p.curPos()}
		p.nextToken()
		p.skipWhitespace()

		if p.curTokenIs(token.LPAREN) {
			p.nextToken()
			p.skipWhitespace()

			// Exception types (can be union)
			for {
				if p.curTokenIs(token.T_STRING) || p.curTokenIs(token.T_NAME_QUALIFIED) {
					catch.Types = append(catch.Types, &ast.Ident{
						NamePos: p.curPos(),
						Name:    p.curToken.Literal,
					})
					p.nextToken()
				}
				p.skipWhitespace()
				if p.curTokenIs(token.PIPE) {
					p.nextToken()
					p.skipWhitespace()
				} else {
					break
				}
			}

			p.skipWhitespace()

			// Variable (optional in PHP 8.0+)
			if p.curTokenIs(token.T_VARIABLE) {
				catch.Var = p.parseVariable().(*ast.Variable)
			}

			p.skipWhitespace()
			if p.curTokenIs(token.RPAREN) {
				p.nextToken()
			}
		}

		p.skipWhitespace()

		if p.curTokenIs(token.LBRACE) {
			catch.Body = p.parseBlockStmt()
		}

		tryStmt.Catches = append(tryStmt.Catches, catch)
		p.skipWhitespace()
	}

	// Finally
	if p.curTokenIs(token.T_FINALLY) {
		finally := &ast.FinallyClause{FinallyPos: p.curPos()}
		p.nextToken()
		p.skipWhitespace()

		if p.curTokenIs(token.LBRACE) {
			finally.Body = p.parseBlockStmt()
		}

		tryStmt.Finally = finally
	}

	return tryStmt
}

func (p *Parser) parseThrowStmt() *ast.ThrowStmt {
	throwStmt := &ast.ThrowStmt{ThrowPos: p.curPos()}
	p.nextToken() // skip throw
	p.skipWhitespace()

	throwStmt.Expr = p.parseExpression(LOWEST)

	p.skipWhitespace()
	if p.curTokenIs(token.SEMICOLON) {
		throwStmt.Semicolon = p.curPos()
		p.nextToken()
	}

	return throwStmt
}

func (p *Parser) parseReturnStmt() *ast.ReturnStmt {
	returnStmt := &ast.ReturnStmt{ReturnPos: p.curPos()}
	p.nextToken() // skip return
	p.skipWhitespace()

	if !p.curTokenIs(token.SEMICOLON) && !p.curTokenIs(token.EOF) {
		returnStmt.Result = p.parseExpression(LOWEST)
	}

	p.skipWhitespace()
	if p.curTokenIs(token.SEMICOLON) {
		returnStmt.Semicolon = p.curPos()
		p.nextToken()
	}

	return returnStmt
}

func (p *Parser) parseBreakStmt() *ast.BreakStmt {
	breakStmt := &ast.BreakStmt{BreakPos: p.curPos()}
	p.nextToken() // skip break
	p.skipWhitespace()

	if !p.curTokenIs(token.SEMICOLON) && !p.curTokenIs(token.EOF) {
		breakStmt.Num = p.parseExpression(LOWEST)
	}

	p.skipWhitespace()
	if p.curTokenIs(token.SEMICOLON) {
		breakStmt.Semicolon = p.curPos()
		p.nextToken()
	}

	return breakStmt
}

func (p *Parser) parseContinueStmt() *ast.ContinueStmt {
	continueStmt := &ast.ContinueStmt{ContinuePos: p.curPos()}
	p.nextToken() // skip continue
	p.skipWhitespace()

	if !p.curTokenIs(token.SEMICOLON) && !p.curTokenIs(token.EOF) {
		continueStmt.Num = p.parseExpression(LOWEST)
	}

	p.skipWhitespace()
	if p.curTokenIs(token.SEMICOLON) {
		continueStmt.Semicolon = p.curPos()
		p.nextToken()
	}

	return continueStmt
}

func (p *Parser) parseGotoStmt() *ast.GotoStmt {
	gotoStmt := &ast.GotoStmt{GotoPos: p.curPos()}
	p.nextToken() // skip goto
	p.skipWhitespace()

	if p.curTokenIs(token.T_STRING) {
		gotoStmt.Label = &ast.Ident{
			NamePos: p.curPos(),
			Name:    p.curToken.Literal,
		}
		p.nextToken()
	}

	p.skipWhitespace()
	if p.curTokenIs(token.SEMICOLON) {
		gotoStmt.Semicolon = p.curPos()
		p.nextToken()
	}

	return gotoStmt
}

func (p *Parser) parseEchoStmt() *ast.EchoStmt {
	echoStmt := &ast.EchoStmt{EchoPos: p.curPos()}
	p.nextToken() // skip echo
	p.skipWhitespace()

	for {
		echoStmt.Exprs = append(echoStmt.Exprs, p.parseExpression(LOWEST))
		p.skipWhitespace()
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			p.skipWhitespace()
		} else {
			break
		}
	}

	if p.curTokenIs(token.SEMICOLON) {
		echoStmt.Semicolon = p.curPos()
		p.nextToken()
	}

	return echoStmt
}

func (p *Parser) parseGlobalStmt() *ast.GlobalStmt {
	globalStmt := &ast.GlobalStmt{GlobalPos: p.curPos()}
	p.nextToken() // skip global
	p.skipWhitespace()

	for {
		if p.curTokenIs(token.T_VARIABLE) {
			globalStmt.Vars = append(globalStmt.Vars, p.parseVariable().(*ast.Variable))
		}
		p.skipWhitespace()
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			p.skipWhitespace()
		} else {
			break
		}
	}

	if p.curTokenIs(token.SEMICOLON) {
		globalStmt.Semicolon = p.curPos()
		p.nextToken()
	}

	return globalStmt
}

func (p *Parser) parseStaticVarStmt() *ast.StaticVarStmt {
	staticStmt := &ast.StaticVarStmt{StaticPos: p.curPos()}
	p.nextToken() // skip static
	p.skipWhitespace()

	// Check if this is actually a static property access or static function
	if p.curTokenIs(token.T_PAAMAYIM_NEKUDOTAYIM) || p.curTokenIs(token.T_FUNCTION) || p.curTokenIs(token.T_FN) {
		// This is a static expression, not a static variable declaration
		// We need to backtrack - but since we can't easily do that,
		// let's handle this as an expression statement
		return nil
	}

	for {
		staticVar := &ast.StaticVar{}
		if p.curTokenIs(token.T_VARIABLE) {
			staticVar.Var = p.parseVariable().(*ast.Variable)
		}

		p.skipWhitespace()
		if p.curTokenIs(token.EQUALS) {
			p.nextToken()
			p.skipWhitespace()
			staticVar.Default = p.parseExpression(LOWEST)
		}

		staticStmt.Vars = append(staticStmt.Vars, staticVar)
		p.skipWhitespace()

		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			p.skipWhitespace()
		} else {
			break
		}
	}

	if p.curTokenIs(token.SEMICOLON) {
		staticStmt.Semicolon = p.curPos()
		p.nextToken()
	}

	return staticStmt
}

func (p *Parser) parseUnsetStmt() *ast.UnsetStmt {
	unsetStmt := &ast.UnsetStmt{UnsetPos: p.curPos()}
	p.nextToken() // skip unset
	p.skipWhitespace()

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
			p.skipWhitespace()
			unsetStmt.Vars = append(unsetStmt.Vars, p.parseExpression(LOWEST))
			p.skipWhitespace()
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
			}
		}
		unsetStmt.Rparen = p.curPos()
		p.nextToken()
	}

	p.skipWhitespace()
	if p.curTokenIs(token.SEMICOLON) {
		unsetStmt.Semicolon = p.curPos()
		p.nextToken()
	}

	return unsetStmt
}

func (p *Parser) parseDeclareStmt() *ast.DeclareStmt {
	declareStmt := &ast.DeclareStmt{DeclarePos: p.curPos()}
	p.nextToken() // skip declare
	p.skipWhitespace()

	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
			p.skipWhitespace()

			directive := &ast.DeclareDirective{}
			if p.curTokenIs(token.T_STRING) {
				directive.Name = &ast.Ident{
					NamePos: p.curPos(),
					Name:    p.curToken.Literal,
				}
				p.nextToken()
			}

			p.skipWhitespace()
			if p.curTokenIs(token.EQUALS) {
				p.nextToken()
				p.skipWhitespace()
				directive.Value = p.parseExpression(LOWEST)
			}

			declareStmt.Directives = append(declareStmt.Directives, directive)
			p.skipWhitespace()
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
			}
		}
		p.nextToken() // skip )
	}

	p.skipWhitespace()

	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	} else if p.curTokenIs(token.LBRACE) {
		declareStmt.Body = p.parseBlockStmt()
	} else if p.curTokenIs(token.COLON) {
		p.nextToken()
		declareStmt.Body = p.parseAltSyntaxBody(token.T_ENDDECLARE)
		if p.curTokenIs(token.T_ENDDECLARE) {
			p.nextToken()
			p.skipWhitespace()
			if p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}
	}

	return declareStmt
}
