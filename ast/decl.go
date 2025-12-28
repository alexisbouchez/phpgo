package ast

import "github.com/alexisbouchez/phpgo/token"

// ----------------------------------------------------------------------------
// Declarations

// BadDecl represents a syntactically invalid declaration.
type BadDecl struct {
	From, To Position
}

// NamespaceDecl represents a namespace declaration.
type NamespaceDecl struct {
	NamespacePos Position
	Name         Expr // nil for global namespace
	Lbrace       Position // For bracketed namespace
	Stmts        []Stmt
	Rbrace       Position
	Bracketed    bool
}

// UseDecl represents a use declaration.
type UseDecl struct {
	UsePos Position
	Type   token.Token // T_USE, T_FUNCTION, T_CONST or 0 for class
	Uses   []*UseClause
}

// UseClause represents a single use clause.
type UseClause struct {
	Type  token.Token // For grouped imports
	Name  Expr
	Alias *Ident
}

// ConstDecl represents a constant declaration.
type ConstDecl struct {
	ConstPos  Position
	Consts    []*ConstSpec
	Semicolon Position
}

// ConstSpec represents a single constant.
type ConstSpec struct {
	Name  *Ident
	Value Expr
}

// FunctionDecl represents a function declaration.
type FunctionDecl struct {
	Attrs      []*AttributeGroup
	FuncPos    Position
	ByRef      bool
	Name       *Ident
	Params     []*Parameter
	ReturnType *TypeExpr
	Body       *BlockStmt
}

// ClassDecl represents a class declaration.
type ClassDecl struct {
	Attrs      []*AttributeGroup
	Modifiers  *ClassModifiers
	ClassPos   Position
	Name       *Ident
	Extends    Expr
	Implements []Expr
	Lbrace     Position
	Members    []ClassMember
	Rbrace     Position
}

// ClassModifiers represents class modifiers (abstract, final, readonly).
type ClassModifiers struct {
	Abstract bool
	Final    bool
	Readonly bool
}

// InterfaceDecl represents an interface declaration.
type InterfaceDecl struct {
	Attrs        []*AttributeGroup
	InterfacePos Position
	Name         *Ident
	Extends      []Expr
	Lbrace       Position
	Members      []ClassMember
	Rbrace       Position
}

// TraitDecl represents a trait declaration.
type TraitDecl struct {
	Attrs    []*AttributeGroup
	TraitPos Position
	Name     *Ident
	Lbrace   Position
	Members  []ClassMember
	Rbrace   Position
}

// EnumDecl represents an enum declaration.
type EnumDecl struct {
	Attrs      []*AttributeGroup
	EnumPos    Position
	Name       *Ident
	BackingType *TypeExpr
	Implements []Expr
	Lbrace     Position
	Members    []ClassMember
	Rbrace     Position
}

// ClassMember is the interface for class members.
type ClassMember interface {
	Node
	classMemberNode()
}

// PropertyDecl represents a property declaration.
type PropertyDecl struct {
	Attrs      []*AttributeGroup
	Modifiers  *PropertyModifiers
	Type       *TypeExpr
	Props      []*PropertyItem
	Semicolon  Position
}

// PropertyModifiers represents property modifiers.
type PropertyModifiers struct {
	Public    bool
	Protected bool
	Private   bool
	Static    bool
	Readonly  bool
	PublicSet    bool // public(set)
	ProtectedSet bool // protected(set)
	PrivateSet   bool // private(set)
}

// PropertyItem represents a single property.
type PropertyItem struct {
	Var     *Variable
	Default Expr
	Hooks   *PropertyHooks
}

// PropertyHooks represents property hooks (get/set).
type PropertyHooks struct {
	Lbrace Position
	Get    *PropertyHook
	Set    *PropertyHook
	Rbrace Position
}

// PropertyHook represents a single property hook.
type PropertyHook struct {
	Attrs  []*AttributeGroup
	ByRef  bool
	Name   *Ident
	Params []*Parameter // For set hook
	Body   Stmt         // BlockStmt or ExprStmt (=> expr;)
}

// MethodDecl represents a method declaration.
type MethodDecl struct {
	Attrs      []*AttributeGroup
	Modifiers  *MethodModifiers
	FuncPos    Position
	ByRef      bool
	Name       *Ident
	Params     []*Parameter
	ReturnType *TypeExpr
	Body       *BlockStmt // nil for abstract/interface methods
}

// MethodModifiers represents method modifiers.
type MethodModifiers struct {
	Public    bool
	Protected bool
	Private   bool
	Static    bool
	Abstract  bool
	Final     bool
}

// ClassConstDecl represents a class constant declaration.
type ClassConstDecl struct {
	Attrs     []*AttributeGroup
	Modifiers *ConstModifiers
	ConstPos  Position
	Consts    []*ConstSpec
	Semicolon Position
}

// ConstModifiers represents constant modifiers.
type ConstModifiers struct {
	Public    bool
	Protected bool
	Private   bool
	Final     bool
}

// TraitUseDecl represents a trait use declaration in a class.
type TraitUseDecl struct {
	UsePos      Position
	Traits      []Expr
	Adaptations []*TraitAdaptation
}

// TraitAdaptation represents a trait adaptation.
type TraitAdaptation struct {
	Trait      Expr
	Method     *Ident
	Insteadof  []Expr // For insteadof
	Alias      *Ident // For as
	Visibility token.Token
}

// EnumCaseDecl represents an enum case declaration.
type EnumCaseDecl struct {
	Attrs    []*AttributeGroup
	CasePos  Position
	Name     *Ident
	Value    Expr
	Semicolon Position
}

// Class member implementations
func (*PropertyDecl) classMemberNode()   {}
func (*MethodDecl) classMemberNode()     {}
func (*ClassConstDecl) classMemberNode() {}
func (*TraitUseDecl) classMemberNode()   {}
func (*EnumCaseDecl) classMemberNode()   {}

// Declaration implementations
func (*BadDecl) declNode()       {}
func (*NamespaceDecl) declNode() {}
func (*UseDecl) declNode()       {}
func (*ConstDecl) declNode()     {}
func (*FunctionDecl) declNode()  {}
func (*ClassDecl) declNode()     {}
func (*InterfaceDecl) declNode() {}
func (*TraitDecl) declNode()     {}
func (*EnumDecl) declNode()      {}

// Statement implementations for declarations
func (*NamespaceDecl) stmtNode() {}
func (*UseDecl) stmtNode()       {}
func (*ConstDecl) stmtNode()     {}
func (*FunctionDecl) stmtNode()  {}
func (*ClassDecl) stmtNode()     {}
func (*InterfaceDecl) stmtNode() {}
func (*TraitDecl) stmtNode()     {}
func (*EnumDecl) stmtNode()      {}

// Pos implementations for declarations
func (d *BadDecl) Pos() Position       { return d.From }
func (d *NamespaceDecl) Pos() Position { return d.NamespacePos }
func (d *UseDecl) Pos() Position       { return d.UsePos }
func (d *ConstDecl) Pos() Position     { return d.ConstPos }
func (d *FunctionDecl) Pos() Position  { return d.FuncPos }
func (d *ClassDecl) Pos() Position     { return d.ClassPos }
func (d *InterfaceDecl) Pos() Position { return d.InterfacePos }
func (d *TraitDecl) Pos() Position     { return d.TraitPos }
func (d *EnumDecl) Pos() Position      { return d.EnumPos }

// Pos implementations for class members
func (m *PropertyDecl) Pos() Position   { return m.Props[0].Var.Pos() }
func (m *MethodDecl) Pos() Position     { return m.FuncPos }
func (m *ClassConstDecl) Pos() Position { return m.ConstPos }
func (m *TraitUseDecl) Pos() Position   { return m.UsePos }
func (m *EnumCaseDecl) Pos() Position   { return m.CasePos }

// End implementations for declarations
func (d *BadDecl) End() Position       { return d.To }
func (d *NamespaceDecl) End() Position { if d.Bracketed { return d.Rbrace }; return d.NamespacePos }
func (d *UseDecl) End() Position       { return d.UsePos }
func (d *ConstDecl) End() Position     { return d.Semicolon }
func (d *FunctionDecl) End() Position  { return d.Body.End() }
func (d *ClassDecl) End() Position     { return d.Rbrace }
func (d *InterfaceDecl) End() Position { return d.Rbrace }
func (d *TraitDecl) End() Position     { return d.Rbrace }
func (d *EnumDecl) End() Position      { return d.Rbrace }

// End implementations for class members
func (m *PropertyDecl) End() Position   { return m.Semicolon }
func (m *MethodDecl) End() Position     { if m.Body != nil { return m.Body.End() }; return m.FuncPos }
func (m *ClassConstDecl) End() Position { return m.Semicolon }
func (m *TraitUseDecl) End() Position   { return m.UsePos }
func (m *EnumCaseDecl) End() Position   { return m.Semicolon }

// ----------------------------------------------------------------------------
// File

// File represents a PHP source file.
type File struct {
	Name    string
	Stmts   []Stmt
	OpenTag Position
}

func (f *File) Pos() Position { return f.OpenTag }
func (f *File) End() Position {
	if len(f.Stmts) > 0 {
		return f.Stmts[len(f.Stmts)-1].End()
	}
	return f.OpenTag
}
