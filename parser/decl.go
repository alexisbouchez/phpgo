package parser

import (
	"github.com/alexisbouchez/phpgo/ast"
	"github.com/alexisbouchez/phpgo/token"
)

func (p *Parser) parseNamespaceDecl() *ast.NamespaceDecl {
	ns := &ast.NamespaceDecl{NamespacePos: p.curPos()}
	p.nextToken() // skip namespace
	p.skipWhitespace()

	// Parse namespace name
	if p.curTokenIs(token.T_STRING) || p.curTokenIs(token.T_NAME_QUALIFIED) {
		ns.Name = &ast.Ident{
			NamePos: p.curPos(),
			Name:    p.curToken.Literal,
		}
		p.nextToken()
	}

	p.skipWhitespace()

	// Bracketed namespace
	if p.curTokenIs(token.LBRACE) {
		ns.Bracketed = true
		ns.Lbrace = p.curPos()
		p.nextToken()

		for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
			p.skipWhitespace()
			if p.curTokenIs(token.RBRACE) {
				break
			}
			stmt := p.parseStatement()
			if stmt != nil {
				ns.Stmts = append(ns.Stmts, stmt)
			}
		}

		ns.Rbrace = p.curPos()
		p.nextToken()
	} else if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return ns
}

func (p *Parser) parseUseDecl() *ast.UseDecl {
	use := &ast.UseDecl{UsePos: p.curPos()}
	p.nextToken() // skip use
	p.skipWhitespace()

	// Check for function or const
	if p.curTokenIs(token.T_FUNCTION) {
		use.Type = token.T_FUNCTION
		p.nextToken()
		p.skipWhitespace()
	} else if p.curTokenIs(token.T_CONST) {
		use.Type = token.T_CONST
		p.nextToken()
		p.skipWhitespace()
	}

	// Parse use clauses
	for {
		clause := &ast.UseClause{}

		if p.curTokenIs(token.T_STRING) || p.curTokenIs(token.T_NAME_QUALIFIED) ||
			p.curTokenIs(token.T_NAME_FULLY_QUALIFIED) {
			clause.Name = &ast.Ident{
				NamePos: p.curPos(),
				Name:    p.curToken.Literal,
			}
			p.nextToken()
		}

		p.skipWhitespace()

		// Check for alias
		if p.curTokenIs(token.T_AS) {
			p.nextToken()
			p.skipWhitespace()
			if p.curTokenIs(token.T_STRING) {
				clause.Alias = &ast.Ident{
					NamePos: p.curPos(),
					Name:    p.curToken.Literal,
				}
				p.nextToken()
			}
		}

		use.Uses = append(use.Uses, clause)
		p.skipWhitespace()

		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			p.skipWhitespace()
		} else {
			break
		}
	}

	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return use
}

func (p *Parser) parseConstDecl() *ast.ConstDecl {
	constDecl := &ast.ConstDecl{ConstPos: p.curPos()}
	p.nextToken() // skip const
	p.skipWhitespace()

	for {
		spec := &ast.ConstSpec{}
		if p.curTokenIs(token.T_STRING) {
			spec.Name = &ast.Ident{
				NamePos: p.curPos(),
				Name:    p.curToken.Literal,
			}
			p.nextToken()
		}

		p.skipWhitespace()
		if p.curTokenIs(token.EQUALS) {
			p.nextToken()
			p.skipWhitespace()
			spec.Value = p.parseExpression(LOWEST)
		}

		constDecl.Consts = append(constDecl.Consts, spec)
		p.skipWhitespace()

		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			p.skipWhitespace()
		} else {
			break
		}
	}

	if p.curTokenIs(token.SEMICOLON) {
		constDecl.Semicolon = p.curPos()
		p.nextToken()
	}

	return constDecl
}

func (p *Parser) parseFunctionDecl() *ast.FunctionDecl {
	fn := &ast.FunctionDecl{FuncPos: p.curPos()}
	p.nextToken() // skip function
	p.skipWhitespace()

	// Check for reference
	if p.curTokenIs(token.AMPERSAND) {
		fn.ByRef = true
		p.nextToken()
		p.skipWhitespace()
	}

	// Function name
	if p.curTokenIs(token.T_STRING) {
		fn.Name = &ast.Ident{
			NamePos: p.curPos(),
			Name:    p.curToken.Literal,
		}
		p.nextToken()
	}

	p.skipWhitespace()

	// Parameters
	if p.curTokenIs(token.LPAREN) {
		fn.Params = p.parseParameterList()
	}

	p.skipWhitespace()

	// Return type
	if p.curTokenIs(token.COLON) {
		p.nextToken()
		p.skipWhitespace()
		fn.ReturnType = p.parseTypeExpr()
	}

	p.skipWhitespace()

	// Body
	if p.curTokenIs(token.LBRACE) {
		fn.Body = p.parseBlockStmt()
	}

	return fn
}

func (p *Parser) parseParameterList() []*ast.Parameter {
	var params []*ast.Parameter
	p.nextToken() // skip (

	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		p.skipWhitespace()
		if p.curTokenIs(token.RPAREN) {
			break
		}

		param := &ast.Parameter{}

		// Attributes
		if p.curTokenIs(token.T_ATTRIBUTE) {
			param.Attrs = p.parseAttributeGroups()
			p.skipWhitespace()
		}

		// Visibility (for constructor promotion)
		if p.curTokenIs(token.T_PUBLIC) || p.curTokenIs(token.T_PROTECTED) || p.curTokenIs(token.T_PRIVATE) {
			param.Visibility = p.curToken.Type
			p.nextToken()
			p.skipWhitespace()
		}

		// Readonly
		if p.curTokenIs(token.T_READONLY) {
			param.Readonly = true
			p.nextToken()
			p.skipWhitespace()
		}

		// Type
		if p.isTypeName() {
			param.Type = p.parseTypeExpr()
			p.skipWhitespace()
		}

		// Reference
		if p.curTokenIs(token.AMPERSAND) {
			param.ByRef = true
			p.nextToken()
			p.skipWhitespace()
		}

		// Variadic
		if p.curTokenIs(token.T_ELLIPSIS) {
			param.Variadic = true
			p.nextToken()
			p.skipWhitespace()
		}

		// Variable
		if p.curTokenIs(token.T_VARIABLE) {
			param.Var = p.parseVariable().(*ast.Variable)
		}

		p.skipWhitespace()

		// Default value
		if p.curTokenIs(token.EQUALS) {
			p.nextToken()
			p.skipWhitespace()
			param.Default = p.parseExpression(LOWEST)
		}

		params = append(params, param)
		p.skipWhitespace()

		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	if p.curTokenIs(token.RPAREN) {
		p.nextToken()
	}

	return params
}

func (p *Parser) parseTypeExpr() *ast.TypeExpr {
	typeExpr := &ast.TypeExpr{StartPos: p.curPos()}

	// Nullable
	if p.curTokenIs(token.QUESTION) {
		typeExpr.Nullable = true
		p.nextToken()
		p.skipWhitespace()
	}

	typeExpr.Type = p.parseType()
	return typeExpr
}

func (p *Parser) parseType() ast.Type {
	// Parse first type
	first := p.parseSimpleType()
	p.skipWhitespace()

	// Check for union or intersection
	if p.curTokenIs(token.PIPE) {
		union := &ast.UnionType{Types: []ast.Type{first}}
		for p.curTokenIs(token.PIPE) {
			p.nextToken()
			p.skipWhitespace()
			union.Types = append(union.Types, p.parseSimpleType())
			p.skipWhitespace()
		}
		return union
	}

	if p.curTokenIs(token.AMPERSAND) {
		inter := &ast.IntersectionType{Types: []ast.Type{first}}
		for p.curTokenIs(token.AMPERSAND) {
			p.nextToken()
			p.skipWhitespace()
			inter.Types = append(inter.Types, p.parseSimpleType())
			p.skipWhitespace()
		}
		return inter
	}

	return first
}

func (p *Parser) parseSimpleType() *ast.SimpleType {
	typ := &ast.SimpleType{
		NamePos: p.curPos(),
		Name:    p.curToken.Literal,
	}
	p.nextToken()
	return typ
}

func (p *Parser) isTypeName() bool {
	switch p.curToken.Type {
	case token.T_STRING, token.T_NAME_QUALIFIED, token.T_NAME_FULLY_QUALIFIED,
		token.T_ARRAY, token.T_CALLABLE, token.QUESTION:
		return true
	default:
		return false
	}
}

func (p *Parser) parseClassDecl(modifiers *ast.ClassModifiers) *ast.ClassDecl {
	class := &ast.ClassDecl{
		ClassPos:  p.curPos(),
		Modifiers: modifiers,
	}
	if class.Modifiers == nil {
		class.Modifiers = &ast.ClassModifiers{}
	}

	p.nextToken() // skip class
	p.skipWhitespace()

	// Class name
	if p.curTokenIs(token.T_STRING) {
		class.Name = &ast.Ident{
			NamePos: p.curPos(),
			Name:    p.curToken.Literal,
		}
		p.nextToken()
	}

	p.skipWhitespace()

	// Extends
	if p.curTokenIs(token.T_EXTENDS) {
		p.nextToken()
		p.skipWhitespace()
		class.Extends = p.parseExpression(LOWEST)
		p.skipWhitespace()
	}

	// Implements
	if p.curTokenIs(token.T_IMPLEMENTS) {
		p.nextToken()
		p.skipWhitespace()
		for {
			class.Implements = append(class.Implements, p.parseExpression(LOWEST))
			p.skipWhitespace()
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
				p.skipWhitespace()
			} else {
				break
			}
		}
	}

	p.skipWhitespace()

	// Body
	if p.curTokenIs(token.LBRACE) {
		class.Lbrace = p.curPos()
		p.nextToken()

		for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
			p.skipWhitespace()
			if p.curTokenIs(token.RBRACE) {
				break
			}

			member := p.parseClassMember()
			if member != nil {
				class.Members = append(class.Members, member)
			}
		}

		class.Rbrace = p.curPos()
		p.nextToken()
	}

	return class
}

func (p *Parser) parseAbstractClass() ast.Stmt {
	p.nextToken() // skip abstract
	p.skipWhitespace()
	if p.curTokenIs(token.T_CLASS) {
		return p.parseClassDecl(&ast.ClassModifiers{Abstract: true})
	}
	return nil
}

func (p *Parser) parseFinalClass() ast.Stmt {
	p.nextToken() // skip final
	p.skipWhitespace()
	if p.curTokenIs(token.T_CLASS) {
		return p.parseClassDecl(&ast.ClassModifiers{Final: true})
	}
	return nil
}

func (p *Parser) parseReadonlyClass() ast.Stmt {
	p.nextToken() // skip readonly
	p.skipWhitespace()
	if p.curTokenIs(token.T_CLASS) {
		return p.parseClassDecl(&ast.ClassModifiers{Readonly: true})
	}
	return nil
}

func (p *Parser) parseClassMember() ast.ClassMember {
	var attrs []*ast.AttributeGroup
	if p.curTokenIs(token.T_ATTRIBUTE) {
		attrs = p.parseAttributeGroups()
		p.skipWhitespace()
	}

	// Parse modifiers
	modifiers := &ast.PropertyModifiers{}
	methodMods := &ast.MethodModifiers{}
	constMods := &ast.ConstModifiers{}

	for {
		switch p.curToken.Type {
		case token.T_PUBLIC:
			modifiers.Public = true
			methodMods.Public = true
			constMods.Public = true
		case token.T_PROTECTED:
			modifiers.Protected = true
			methodMods.Protected = true
			constMods.Protected = true
		case token.T_PRIVATE:
			modifiers.Private = true
			methodMods.Private = true
			constMods.Private = true
		case token.T_STATIC:
			modifiers.Static = true
			methodMods.Static = true
		case token.T_READONLY:
			modifiers.Readonly = true
		case token.T_ABSTRACT:
			methodMods.Abstract = true
		case token.T_FINAL:
			methodMods.Final = true
			constMods.Final = true
		default:
			goto parseBody
		}
		p.nextToken()
		p.skipWhitespace()
	}

parseBody:
	switch p.curToken.Type {
	case token.T_CONST:
		return p.parseClassConstDecl(attrs, constMods)
	case token.T_FUNCTION:
		return p.parseMethodDecl(attrs, methodMods)
	case token.T_USE:
		return p.parseTraitUseDecl()
	case token.T_VARIABLE:
		return p.parsePropertyDecl(attrs, modifiers, nil)
	default:
		// Could be a typed property
		if p.isTypeName() {
			typeExpr := p.parseTypeExpr()
			p.skipWhitespace()
			if p.curTokenIs(token.T_VARIABLE) {
				return p.parsePropertyDecl(attrs, modifiers, typeExpr)
			}
		}
		p.nextToken()
		return nil
	}
}

func (p *Parser) parseClassConstDecl(attrs []*ast.AttributeGroup, modifiers *ast.ConstModifiers) *ast.ClassConstDecl {
	constDecl := &ast.ClassConstDecl{
		Attrs:     attrs,
		Modifiers: modifiers,
		ConstPos:  p.curPos(),
	}
	p.nextToken() // skip const
	p.skipWhitespace()

	for {
		spec := &ast.ConstSpec{}
		if p.curTokenIs(token.T_STRING) {
			spec.Name = &ast.Ident{
				NamePos: p.curPos(),
				Name:    p.curToken.Literal,
			}
			p.nextToken()
		}

		p.skipWhitespace()
		if p.curTokenIs(token.EQUALS) {
			p.nextToken()
			p.skipWhitespace()
			spec.Value = p.parseExpression(LOWEST)
		}

		constDecl.Consts = append(constDecl.Consts, spec)
		p.skipWhitespace()

		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			p.skipWhitespace()
		} else {
			break
		}
	}

	if p.curTokenIs(token.SEMICOLON) {
		constDecl.Semicolon = p.curPos()
		p.nextToken()
	}

	return constDecl
}

func (p *Parser) parseMethodDecl(attrs []*ast.AttributeGroup, modifiers *ast.MethodModifiers) *ast.MethodDecl {
	method := &ast.MethodDecl{
		Attrs:     attrs,
		Modifiers: modifiers,
		FuncPos:   p.curPos(),
	}
	p.nextToken() // skip function
	p.skipWhitespace()

	// Reference
	if p.curTokenIs(token.AMPERSAND) {
		method.ByRef = true
		p.nextToken()
		p.skipWhitespace()
	}

	// Name
	if p.curTokenIs(token.T_STRING) {
		method.Name = &ast.Ident{
			NamePos: p.curPos(),
			Name:    p.curToken.Literal,
		}
		p.nextToken()
	}

	p.skipWhitespace()

	// Parameters
	if p.curTokenIs(token.LPAREN) {
		method.Params = p.parseParameterList()
	}

	p.skipWhitespace()

	// Return type
	if p.curTokenIs(token.COLON) {
		p.nextToken()
		p.skipWhitespace()
		method.ReturnType = p.parseTypeExpr()
	}

	p.skipWhitespace()

	// Body (or semicolon for abstract/interface)
	if p.curTokenIs(token.LBRACE) {
		method.Body = p.parseBlockStmt()
	} else if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return method
}

func (p *Parser) parsePropertyDecl(attrs []*ast.AttributeGroup, modifiers *ast.PropertyModifiers, typeExpr *ast.TypeExpr) *ast.PropertyDecl {
	prop := &ast.PropertyDecl{
		Attrs:     attrs,
		Modifiers: modifiers,
		Type:      typeExpr,
	}

	for {
		item := &ast.PropertyItem{}
		if p.curTokenIs(token.T_VARIABLE) {
			item.Var = p.parseVariable().(*ast.Variable)
		}

		p.skipWhitespace()

		// Default value
		if p.curTokenIs(token.EQUALS) {
			p.nextToken()
			p.skipWhitespace()
			item.Default = p.parseExpression(LOWEST)
		}

		prop.Props = append(prop.Props, item)
		p.skipWhitespace()

		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			p.skipWhitespace()
		} else {
			break
		}
	}

	if p.curTokenIs(token.SEMICOLON) {
		prop.Semicolon = p.curPos()
		p.nextToken()
	}

	return prop
}

func (p *Parser) parseTraitUseDecl() *ast.TraitUseDecl {
	use := &ast.TraitUseDecl{UsePos: p.curPos()}
	p.nextToken() // skip use
	p.skipWhitespace()

	// Parse trait names
	for {
		if p.curTokenIs(token.T_STRING) || p.curTokenIs(token.T_NAME_QUALIFIED) {
			use.Traits = append(use.Traits, &ast.Ident{
				NamePos: p.curPos(),
				Name:    p.curToken.Literal,
			})
			p.nextToken()
		}

		p.skipWhitespace()
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
			p.skipWhitespace()
		} else {
			break
		}
	}

	// Adaptations
	if p.curTokenIs(token.LBRACE) {
		p.nextToken()
		for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
			p.skipWhitespace()
			if p.curTokenIs(token.RBRACE) {
				break
			}
			// Parse adaptation
			p.nextToken()
		}
		p.nextToken() // skip }
	} else if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return use
}

func (p *Parser) parseInterfaceDecl() *ast.InterfaceDecl {
	iface := &ast.InterfaceDecl{InterfacePos: p.curPos()}
	p.nextToken() // skip interface
	p.skipWhitespace()

	// Name
	if p.curTokenIs(token.T_STRING) {
		iface.Name = &ast.Ident{
			NamePos: p.curPos(),
			Name:    p.curToken.Literal,
		}
		p.nextToken()
	}

	p.skipWhitespace()

	// Extends
	if p.curTokenIs(token.T_EXTENDS) {
		p.nextToken()
		p.skipWhitespace()
		for {
			iface.Extends = append(iface.Extends, p.parseExpression(LOWEST))
			p.skipWhitespace()
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
				p.skipWhitespace()
			} else {
				break
			}
		}
	}

	p.skipWhitespace()

	// Body
	if p.curTokenIs(token.LBRACE) {
		iface.Lbrace = p.curPos()
		p.nextToken()

		for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
			p.skipWhitespace()
			if p.curTokenIs(token.RBRACE) {
				break
			}

			member := p.parseClassMember()
			if member != nil {
				iface.Members = append(iface.Members, member)
			}
		}

		iface.Rbrace = p.curPos()
		p.nextToken()
	}

	return iface
}

func (p *Parser) parseTraitDecl() *ast.TraitDecl {
	trait := &ast.TraitDecl{TraitPos: p.curPos()}
	p.nextToken() // skip trait
	p.skipWhitespace()

	// Name
	if p.curTokenIs(token.T_STRING) {
		trait.Name = &ast.Ident{
			NamePos: p.curPos(),
			Name:    p.curToken.Literal,
		}
		p.nextToken()
	}

	p.skipWhitespace()

	// Body
	if p.curTokenIs(token.LBRACE) {
		trait.Lbrace = p.curPos()
		p.nextToken()

		for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
			p.skipWhitespace()
			if p.curTokenIs(token.RBRACE) {
				break
			}

			member := p.parseClassMember()
			if member != nil {
				trait.Members = append(trait.Members, member)
			}
		}

		trait.Rbrace = p.curPos()
		p.nextToken()
	}

	return trait
}

func (p *Parser) parseEnumDecl() *ast.EnumDecl {
	enum := &ast.EnumDecl{EnumPos: p.curPos()}
	p.nextToken() // skip enum
	p.skipWhitespace()

	// Name
	if p.curTokenIs(token.T_STRING) {
		enum.Name = &ast.Ident{
			NamePos: p.curPos(),
			Name:    p.curToken.Literal,
		}
		p.nextToken()
	}

	p.skipWhitespace()

	// Backing type
	if p.curTokenIs(token.COLON) {
		p.nextToken()
		p.skipWhitespace()
		enum.BackingType = p.parseTypeExpr()
	}

	p.skipWhitespace()

	// Implements
	if p.curTokenIs(token.T_IMPLEMENTS) {
		p.nextToken()
		p.skipWhitespace()
		for {
			enum.Implements = append(enum.Implements, p.parseExpression(LOWEST))
			p.skipWhitespace()
			if p.curTokenIs(token.COMMA) {
				p.nextToken()
				p.skipWhitespace()
			} else {
				break
			}
		}
	}

	p.skipWhitespace()

	// Body
	if p.curTokenIs(token.LBRACE) {
		enum.Lbrace = p.curPos()
		p.nextToken()

		for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
			p.skipWhitespace()
			if p.curTokenIs(token.RBRACE) {
				break
			}

			// Parse case or method
			if p.curTokenIs(token.T_CASE) {
				enum.Members = append(enum.Members, p.parseEnumCase())
			} else {
				member := p.parseClassMember()
				if member != nil {
					enum.Members = append(enum.Members, member)
				}
			}
		}

		enum.Rbrace = p.curPos()
		p.nextToken()
	}

	return enum
}

func (p *Parser) parseEnumCase() *ast.EnumCaseDecl {
	caseDecl := &ast.EnumCaseDecl{CasePos: p.curPos()}
	p.nextToken() // skip case
	p.skipWhitespace()

	// Name
	if p.curTokenIs(token.T_STRING) {
		caseDecl.Name = &ast.Ident{
			NamePos: p.curPos(),
			Name:    p.curToken.Literal,
		}
		p.nextToken()
	}

	p.skipWhitespace()

	// Value
	if p.curTokenIs(token.EQUALS) {
		p.nextToken()
		p.skipWhitespace()
		caseDecl.Value = p.parseExpression(LOWEST)
	}

	p.skipWhitespace()

	if p.curTokenIs(token.SEMICOLON) {
		caseDecl.Semicolon = p.curPos()
		p.nextToken()
	}

	return caseDecl
}
