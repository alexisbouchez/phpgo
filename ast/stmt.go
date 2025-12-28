package ast

import "github.com/alexisbouchez/phpgo/token"

// ----------------------------------------------------------------------------
// Statements

// BadStmt represents a syntactically invalid statement.
type BadStmt struct {
	From, To Position
}

// EmptyStmt represents an empty statement (;).
type EmptyStmt struct {
	Semicolon Position
}

// ExprStmt represents an expression statement.
type ExprStmt struct {
	Expr      Expr
	Semicolon Position
}

// BlockStmt represents a block of statements.
type BlockStmt struct {
	Lbrace Position
	Stmts  []Stmt
	Rbrace Position
}

// IfStmt represents an if statement.
type IfStmt struct {
	IfPos     Position
	Cond      Expr
	Body      Stmt
	ElseIfs   []*ElseIfClause
	Else      *ElseClause
	EndIf     Position // For alternative syntax
	UseAltSyntax bool
}

// ElseIfClause represents an elseif clause.
type ElseIfClause struct {
	ElseIfPos Position
	Cond      Expr
	Body      Stmt
}

// ElseClause represents an else clause.
type ElseClause struct {
	ElsePos Position
	Body    Stmt
}

// SwitchStmt represents a switch statement.
type SwitchStmt struct {
	SwitchPos    Position
	Cond         Expr
	Lbrace       Position
	Cases        []*CaseClause
	Rbrace       Position
	UseAltSyntax bool
}

// CaseClause represents a case or default clause.
type CaseClause struct {
	CasePos   Position
	Cond      Expr // nil for default
	Separator Position // : or ;
	Stmts     []Stmt
}

// WhileStmt represents a while statement.
type WhileStmt struct {
	WhilePos     Position
	Cond         Expr
	Body         Stmt
	UseAltSyntax bool
}

// DoWhileStmt represents a do-while statement.
type DoWhileStmt struct {
	DoPos     Position
	Body      Stmt
	WhilePos  Position
	Cond      Expr
	Semicolon Position
}

// ForStmt represents a for statement.
type ForStmt struct {
	ForPos       Position
	Init         []Expr
	Cond         []Expr
	Loop         []Expr
	Body         Stmt
	UseAltSyntax bool
}

// ForeachStmt represents a foreach statement.
type ForeachStmt struct {
	ForeachPos   Position
	Expr         Expr
	KeyVar       Expr // nil if no key
	ValueVar     Expr
	ByRef        bool
	Body         Stmt
	UseAltSyntax bool
}

// BreakStmt represents a break statement.
type BreakStmt struct {
	BreakPos  Position
	Num       Expr // nil for break without number
	Semicolon Position
}

// ContinueStmt represents a continue statement.
type ContinueStmt struct {
	ContinuePos Position
	Num         Expr // nil for continue without number
	Semicolon   Position
}

// ReturnStmt represents a return statement.
type ReturnStmt struct {
	ReturnPos Position
	Result    Expr // nil for empty return
	Semicolon Position
}

// GotoStmt represents a goto statement.
type GotoStmt struct {
	GotoPos   Position
	Label     *Ident
	Semicolon Position
}

// LabelStmt represents a label statement.
type LabelStmt struct {
	Label *Ident
	Colon Position
}

// TryStmt represents a try-catch-finally statement.
type TryStmt struct {
	TryPos   Position
	Body     *BlockStmt
	Catches  []*CatchClause
	Finally  *FinallyClause
}

// CatchClause represents a catch clause.
type CatchClause struct {
	CatchPos Position
	Types    []Expr // Can be union: Exception|Error
	Var      *Variable // nil in PHP 8.0+ non-capturing catches
	Body     *BlockStmt
}

// FinallyClause represents a finally clause.
type FinallyClause struct {
	FinallyPos Position
	Body       *BlockStmt
}

// ThrowStmt represents a throw statement.
type ThrowStmt struct {
	ThrowPos  Position
	Expr      Expr
	Semicolon Position
}

// EchoStmt represents an echo statement.
type EchoStmt struct {
	EchoPos   Position
	Exprs     []Expr
	Semicolon Position
}

// GlobalStmt represents a global statement.
type GlobalStmt struct {
	GlobalPos Position
	Vars      []*Variable
	Semicolon Position
}

// StaticVarStmt represents a static variable declaration.
type StaticVarStmt struct {
	StaticPos Position
	Vars      []*StaticVar
	Semicolon Position
}

// StaticVar represents a single static variable.
type StaticVar struct {
	Var     *Variable
	Default Expr // nil if no default
}

// UnsetStmt represents an unset statement.
type UnsetStmt struct {
	UnsetPos  Position
	Vars      []Expr
	Rparen    Position
	Semicolon Position
}

// DeclareStmt represents a declare statement.
type DeclareStmt struct {
	DeclarePos Position
	Directives []*DeclareDirective
	Body       Stmt // nil for declare(); format
}

// DeclareDirective represents a declare directive.
type DeclareDirective struct {
	Name  *Ident
	Value Expr
}

// InlineHTMLStmt represents inline HTML outside PHP tags.
type InlineHTMLStmt struct {
	Start Position
	Value string
}

// HaltCompilerStmt represents __halt_compiler().
type HaltCompilerStmt struct {
	HaltPos Position
}

// Statement node implementations
func (*BadStmt) stmtNode()          {}
func (*EmptyStmt) stmtNode()        {}
func (*ExprStmt) stmtNode()         {}
func (*BlockStmt) stmtNode()        {}
func (*IfStmt) stmtNode()           {}
func (*SwitchStmt) stmtNode()       {}
func (*WhileStmt) stmtNode()        {}
func (*DoWhileStmt) stmtNode()      {}
func (*ForStmt) stmtNode()          {}
func (*ForeachStmt) stmtNode()      {}
func (*BreakStmt) stmtNode()        {}
func (*ContinueStmt) stmtNode()     {}
func (*ReturnStmt) stmtNode()       {}
func (*GotoStmt) stmtNode()         {}
func (*LabelStmt) stmtNode()        {}
func (*TryStmt) stmtNode()          {}
func (*ThrowStmt) stmtNode()        {}
func (*EchoStmt) stmtNode()         {}
func (*GlobalStmt) stmtNode()       {}
func (*StaticVarStmt) stmtNode()    {}
func (*UnsetStmt) stmtNode()        {}
func (*DeclareStmt) stmtNode()      {}
func (*InlineHTMLStmt) stmtNode()   {}
func (*HaltCompilerStmt) stmtNode() {}

// Pos implementations for statements
func (s *BadStmt) Pos() Position          { return s.From }
func (s *EmptyStmt) Pos() Position        { return s.Semicolon }
func (s *ExprStmt) Pos() Position         { return s.Expr.Pos() }
func (s *BlockStmt) Pos() Position        { return s.Lbrace }
func (s *IfStmt) Pos() Position           { return s.IfPos }
func (s *SwitchStmt) Pos() Position       { return s.SwitchPos }
func (s *WhileStmt) Pos() Position        { return s.WhilePos }
func (s *DoWhileStmt) Pos() Position      { return s.DoPos }
func (s *ForStmt) Pos() Position          { return s.ForPos }
func (s *ForeachStmt) Pos() Position      { return s.ForeachPos }
func (s *BreakStmt) Pos() Position        { return s.BreakPos }
func (s *ContinueStmt) Pos() Position     { return s.ContinuePos }
func (s *ReturnStmt) Pos() Position       { return s.ReturnPos }
func (s *GotoStmt) Pos() Position         { return s.GotoPos }
func (s *LabelStmt) Pos() Position        { return s.Label.Pos() }
func (s *TryStmt) Pos() Position          { return s.TryPos }
func (s *ThrowStmt) Pos() Position        { return s.ThrowPos }
func (s *EchoStmt) Pos() Position         { return s.EchoPos }
func (s *GlobalStmt) Pos() Position       { return s.GlobalPos }
func (s *StaticVarStmt) Pos() Position    { return s.StaticPos }
func (s *UnsetStmt) Pos() Position        { return s.UnsetPos }
func (s *DeclareStmt) Pos() Position      { return s.DeclarePos }
func (s *InlineHTMLStmt) Pos() Position   { return s.Start }
func (s *HaltCompilerStmt) Pos() Position { return s.HaltPos }

// End implementations for statements
func (s *BadStmt) End() Position          { return s.To }
func (s *EmptyStmt) End() Position        { return s.Semicolon }
func (s *ExprStmt) End() Position         { return s.Semicolon }
func (s *BlockStmt) End() Position        { return s.Rbrace }
func (s *IfStmt) End() Position           {
	if s.UseAltSyntax { return s.EndIf }
	if s.Else != nil { return s.Else.Body.End() }
	if len(s.ElseIfs) > 0 { return s.ElseIfs[len(s.ElseIfs)-1].Body.End() }
	return s.Body.End()
}
func (s *SwitchStmt) End() Position       { return s.Rbrace }
func (s *WhileStmt) End() Position        { return s.Body.End() }
func (s *DoWhileStmt) End() Position      { return s.Semicolon }
func (s *ForStmt) End() Position          { return s.Body.End() }
func (s *ForeachStmt) End() Position      { return s.Body.End() }
func (s *BreakStmt) End() Position        { return s.Semicolon }
func (s *ContinueStmt) End() Position     { return s.Semicolon }
func (s *ReturnStmt) End() Position       { return s.Semicolon }
func (s *GotoStmt) End() Position         { return s.Semicolon }
func (s *LabelStmt) End() Position        { return s.Colon }
func (s *TryStmt) End() Position {
	if s.Finally != nil { return s.Finally.Body.End() }
	if len(s.Catches) > 0 { return s.Catches[len(s.Catches)-1].Body.End() }
	return s.Body.End()
}
func (s *ThrowStmt) End() Position        { return s.Semicolon }
func (s *EchoStmt) End() Position         { return s.Semicolon }
func (s *GlobalStmt) End() Position       { return s.Semicolon }
func (s *StaticVarStmt) End() Position    { return s.Semicolon }
func (s *UnsetStmt) End() Position        { return s.Semicolon }
func (s *DeclareStmt) End() Position      { if s.Body != nil { return s.Body.End() }; return s.DeclarePos }
func (s *InlineHTMLStmt) End() Position   { return Position{Offset: s.Start.Offset + len(s.Value)} }
func (s *HaltCompilerStmt) End() Position { return s.HaltPos }

// ----------------------------------------------------------------------------
// Types

// TypeExpr represents a type expression.
type TypeExpr struct {
	StartPos Position
	Nullable bool
	Type     Type
}

// Type is the interface for type nodes.
type Type interface {
	Node
	typeNode()
}

// SimpleType represents a simple type (int, string, ClassName).
type SimpleType struct {
	NamePos Position
	Name    string
}

// UnionType represents a union type (A|B|C).
type UnionType struct {
	Types []Type
}

// IntersectionType represents an intersection type (A&B).
type IntersectionType struct {
	Types []Type
}

func (*SimpleType) typeNode()       {}
func (*UnionType) typeNode()        {}
func (*IntersectionType) typeNode() {}

func (t *SimpleType) Pos() Position       { return t.NamePos }
func (t *UnionType) Pos() Position        { return t.Types[0].Pos() }
func (t *IntersectionType) Pos() Position { return t.Types[0].Pos() }

func (t *SimpleType) End() Position       { return Position{Offset: t.NamePos.Offset + len(t.Name)} }
func (t *UnionType) End() Position        { return t.Types[len(t.Types)-1].End() }
func (t *IntersectionType) End() Position { return t.Types[len(t.Types)-1].End() }

func (t *TypeExpr) Pos() Position { return t.StartPos }
func (t *TypeExpr) End() Position { return t.Type.End() }

// ----------------------------------------------------------------------------
// Parameters

// Parameter represents a function/method parameter.
type Parameter struct {
	Attrs      []*AttributeGroup
	Visibility token.Token // For constructor promotion
	Readonly   bool
	Type       *TypeExpr
	ByRef      bool
	Variadic   bool
	Var        *Variable
	Default    Expr
}

func (p *Parameter) Pos() Position { return p.Var.Pos() }
func (p *Parameter) End() Position {
	if p.Default != nil { return p.Default.End() }
	return p.Var.End()
}

// ----------------------------------------------------------------------------
// Attributes

// Attribute represents a single attribute.
type Attribute struct {
	Name *Ident
	Args *ArgumentList
}

// AttributeGroup represents #[Attr1, Attr2].
type AttributeGroup struct {
	HashBracket Position
	Attrs       []*Attribute
	Rbrack      Position
}

func (a *Attribute) Pos() Position      { return a.Name.Pos() }
func (a *Attribute) End() Position      { if a.Args != nil { return a.Args.Rparen }; return a.Name.End() }
func (g *AttributeGroup) Pos() Position { return g.HashBracket }
func (g *AttributeGroup) End() Position { return g.Rbrack }
