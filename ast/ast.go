// Package ast defines the Abstract Syntax Tree nodes for PHP.
package ast

import (
	"github.com/alexisbouchez/phpgo/token"
)

// Position represents a position in source code.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Node is the base interface for all AST nodes.
type Node interface {
	Pos() Position
	End() Position
}

// Expr is the interface for all expression nodes.
type Expr interface {
	Node
	exprNode()
}

// Stmt is the interface for all statement nodes.
type Stmt interface {
	Node
	stmtNode()
}

// Decl is the interface for all declaration nodes.
type Decl interface {
	Node
	declNode()
}

// ----------------------------------------------------------------------------
// Expressions

// BadExpr represents a syntactically invalid expression.
type BadExpr struct {
	From, To Position
}

// Ident represents an identifier.
type Ident struct {
	NamePos Position
	Name    string
}

// Variable represents a PHP variable ($name).
type Variable struct {
	DollarPos Position
	Name      Expr // Can be Ident for simple vars, or Expr for ${expr}
}

// Literal represents a literal value (int, float, string).
type Literal struct {
	ValuePos Position
	Kind     token.Token // T_LNUMBER, T_DNUMBER, T_CONSTANT_ENCAPSED_STRING
	Value    string
}

// ArrayExpr represents an array literal.
type ArrayExpr struct {
	Lbrack   Position
	Items    []*ArrayItem
	Rbrack   Position
	IsShort  bool // [] vs array()
}

// ArrayItem represents a single array element.
type ArrayItem struct {
	Key      Expr // nil for value-only items
	Value    Expr
	ByRef    bool // &$value
	Unpack   bool // ...$arr
}

// BinaryExpr represents a binary expression.
type BinaryExpr struct {
	Left  Expr
	OpPos Position
	Op    token.Token
	Right Expr
}

// UnaryExpr represents a unary expression.
type UnaryExpr struct {
	OpPos Position
	Op    token.Token
	X     Expr
}

// PostfixExpr represents a postfix expression (++, --).
type PostfixExpr struct {
	X     Expr
	OpPos Position
	Op    token.Token
}

// TernaryExpr represents a ternary expression (cond ? then : else).
type TernaryExpr struct {
	Cond      Expr
	Question  Position
	Then      Expr // nil for Elvis operator (?:)
	Colon     Position
	Else      Expr
}

// CoalesceExpr represents a null coalescing expression (??).
type CoalesceExpr struct {
	Left  Expr
	OpPos Position
	Right Expr
}

// InstanceofExpr represents an instanceof expression.
type InstanceofExpr struct {
	Expr   Expr
	OpPos  Position
	Class  Expr
}

// CastExpr represents a type cast expression.
type CastExpr struct {
	CastPos Position
	Type    token.Token // T_INT_CAST, T_DOUBLE_CAST, etc.
	X       Expr
}

// CloneExpr represents a clone expression.
type CloneExpr struct {
	ClonePos Position
	Expr     Expr
}

// NewExpr represents a new expression.
type NewExpr struct {
	NewPos Position
	Class  Expr
	Args   *ArgumentList
}

// CallExpr represents a function or method call.
type CallExpr struct {
	Func Expr
	Args *ArgumentList
}

// ArgumentList represents a list of arguments.
type ArgumentList struct {
	Lparen Position
	Args   []*Argument
	Rparen Position
}

// Argument represents a single argument.
type Argument struct {
	Name   *Ident // Named argument (nil for positional)
	Value  Expr
	Unpack bool // ...$args
}

// MethodCallExpr represents a method call ($obj->method()).
type MethodCallExpr struct {
	Object   Expr
	Arrow    Position
	NullSafe bool // ?->
	Method   Expr
	Args     *ArgumentList
}

// StaticCallExpr represents a static method call (Class::method()).
type StaticCallExpr struct {
	Class       Expr
	DoubleColon Position
	Method      Expr
	Args        *ArgumentList
}

// PropertyFetchExpr represents property access ($obj->prop).
type PropertyFetchExpr struct {
	Object   Expr
	Arrow    Position
	NullSafe bool // ?->
	Property Expr
}

// StaticPropertyFetchExpr represents static property access (Class::$prop).
type StaticPropertyFetchExpr struct {
	Class       Expr
	DoubleColon Position
	Property    Expr
}

// ClassConstFetchExpr represents class constant access (Class::CONST).
type ClassConstFetchExpr struct {
	Class       Expr
	DoubleColon Position
	Const       *Ident
}

// ArrayAccessExpr represents array access ($arr[key]).
type ArrayAccessExpr struct {
	Array  Expr
	Lbrack Position
	Index  Expr // nil for $arr[]
	Rbrack Position
}

// EncapsedStringExpr represents a double-quoted string with interpolation.
type EncapsedStringExpr struct {
	OpenQuote  Position
	Parts      []Expr
	CloseQuote Position
}

// HeredocExpr represents a heredoc/nowdoc string.
type HeredocExpr struct {
	StartPos Position
	Label    string
	IsNowdoc bool
	Parts    []Expr
	EndPos   Position
}

// ClosureExpr represents an anonymous function.
type ClosureExpr struct {
	Static     bool
	FuncPos    Position
	ByRef      bool
	Params     []*Parameter
	Uses       []*ClosureUse
	ReturnType *TypeExpr
	Body       *BlockStmt
}

// ClosureUse represents a use clause variable.
type ClosureUse struct {
	ByRef bool
	Var   *Variable
}

// ArrowFuncExpr represents an arrow function (fn($x) => $x).
type ArrowFuncExpr struct {
	Static     bool
	FnPos      Position
	ByRef      bool
	Params     []*Parameter
	ReturnType *TypeExpr
	Arrow      Position
	Body       Expr
}

// YieldExpr represents a yield expression.
type YieldExpr struct {
	YieldPos Position
	Key      Expr // nil for yield without key
	Value    Expr
}

// YieldFromExpr represents a yield from expression.
type YieldFromExpr struct {
	YieldPos Position
	Expr     Expr
}

// ThrowExpr represents a throw expression.
type ThrowExpr struct {
	ThrowPos Position
	Expr     Expr
}

// PrintExpr represents a print expression.
type PrintExpr struct {
	PrintPos Position
	Expr     Expr
}

// IncludeExpr represents include/require expressions.
type IncludeExpr struct {
	IncludePos Position
	Type       token.Token // T_INCLUDE, T_INCLUDE_ONCE, T_REQUIRE, T_REQUIRE_ONCE
	Expr       Expr
}

// IssetExpr represents an isset() expression.
type IssetExpr struct {
	IssetPos Position
	Vars     []Expr
	Rparen   Position
}

// EmptyExpr represents an empty() expression.
type EmptyExpr struct {
	EmptyPos Position
	Expr     Expr
	Rparen   Position
}

// EvalExpr represents an eval() expression.
type EvalExpr struct {
	EvalPos Position
	Expr    Expr
	Rparen  Position
}

// ExitExpr represents exit/die expressions.
type ExitExpr struct {
	ExitPos Position
	Expr    Expr // nil for exit without argument
}

// ListExpr represents a list() expression.
type ListExpr struct {
	ListPos Position
	Items   []*ArrayItem
	Rparen  Position
	IsShort bool // [] vs list()
}

// MatchExpr represents a match expression.
type MatchExpr struct {
	MatchPos Position
	Cond     Expr
	Lbrace   Position
	Arms     []*MatchArm
	Rbrace   Position
}

// MatchArm represents a single match arm.
type MatchArm struct {
	Conds   []Expr // nil for default
	Arrow   Position
	Body    Expr
}

// AssignExpr represents an assignment expression.
type AssignExpr struct {
	Var   Expr
	OpPos Position
	Op    token.Token // =, +=, -=, etc.
	Value Expr
}

// AssignRefExpr represents a reference assignment ($a = &$b).
type AssignRefExpr struct {
	Var      Expr
	Equals   Position
	AmpPos   Position
	Value    Expr
}

// ErrorSuppressExpr represents the error suppression operator (@).
type ErrorSuppressExpr struct {
	AtPos Position
	Expr  Expr
}

// ShellExecExpr represents backtick string execution.
type ShellExecExpr struct {
	OpenTick  Position
	Parts     []Expr
	CloseTick Position
}

// MagicConstExpr represents magic constants (__LINE__, etc.).
type MagicConstExpr struct {
	ConstPos Position
	Kind     token.Token
}

// ParenExpr represents a parenthesized expression.
type ParenExpr struct {
	Lparen Position
	X      Expr
	Rparen Position
}

// Expression node implementations
func (*BadExpr) exprNode()               {}
func (*Ident) exprNode()                 {}
func (*Variable) exprNode()              {}
func (*Literal) exprNode()               {}
func (*ArrayExpr) exprNode()             {}
func (*BinaryExpr) exprNode()            {}
func (*UnaryExpr) exprNode()             {}
func (*PostfixExpr) exprNode()           {}
func (*TernaryExpr) exprNode()           {}
func (*CoalesceExpr) exprNode()          {}
func (*InstanceofExpr) exprNode()        {}
func (*CastExpr) exprNode()              {}
func (*CloneExpr) exprNode()             {}
func (*NewExpr) exprNode()               {}
func (*CallExpr) exprNode()              {}
func (*MethodCallExpr) exprNode()        {}
func (*StaticCallExpr) exprNode()        {}
func (*PropertyFetchExpr) exprNode()     {}
func (*StaticPropertyFetchExpr) exprNode() {}
func (*ClassConstFetchExpr) exprNode()   {}
func (*ArrayAccessExpr) exprNode()       {}
func (*EncapsedStringExpr) exprNode()    {}
func (*HeredocExpr) exprNode()           {}
func (*ClosureExpr) exprNode()           {}
func (*ArrowFuncExpr) exprNode()         {}
func (*YieldExpr) exprNode()             {}
func (*YieldFromExpr) exprNode()         {}
func (*ThrowExpr) exprNode()             {}
func (*PrintExpr) exprNode()             {}
func (*IncludeExpr) exprNode()           {}
func (*IssetExpr) exprNode()             {}
func (*EmptyExpr) exprNode()             {}
func (*EvalExpr) exprNode()              {}
func (*ExitExpr) exprNode()              {}
func (*ListExpr) exprNode()              {}
func (*MatchExpr) exprNode()             {}
func (*AssignExpr) exprNode()            {}
func (*AssignRefExpr) exprNode()         {}
func (*ErrorSuppressExpr) exprNode()     {}
func (*ShellExecExpr) exprNode()         {}
func (*MagicConstExpr) exprNode()        {}
func (*ParenExpr) exprNode()             {}

// Pos implementations for expressions
func (x *BadExpr) Pos() Position               { return x.From }
func (x *Ident) Pos() Position                 { return x.NamePos }
func (x *Variable) Pos() Position              { return x.DollarPos }
func (x *Literal) Pos() Position               { return x.ValuePos }
func (x *ArrayExpr) Pos() Position             { return x.Lbrack }
func (x *BinaryExpr) Pos() Position            { return x.Left.Pos() }
func (x *UnaryExpr) Pos() Position             { return x.OpPos }
func (x *PostfixExpr) Pos() Position           { return x.X.Pos() }
func (x *TernaryExpr) Pos() Position           { return x.Cond.Pos() }
func (x *CoalesceExpr) Pos() Position          { return x.Left.Pos() }
func (x *InstanceofExpr) Pos() Position        { return x.Expr.Pos() }
func (x *CastExpr) Pos() Position              { return x.CastPos }
func (x *CloneExpr) Pos() Position             { return x.ClonePos }
func (x *NewExpr) Pos() Position               { return x.NewPos }
func (x *CallExpr) Pos() Position              { return x.Func.Pos() }
func (x *MethodCallExpr) Pos() Position        { return x.Object.Pos() }
func (x *StaticCallExpr) Pos() Position        { return x.Class.Pos() }
func (x *PropertyFetchExpr) Pos() Position     { return x.Object.Pos() }
func (x *StaticPropertyFetchExpr) Pos() Position { return x.Class.Pos() }
func (x *ClassConstFetchExpr) Pos() Position   { return x.Class.Pos() }
func (x *ArrayAccessExpr) Pos() Position       { return x.Array.Pos() }
func (x *EncapsedStringExpr) Pos() Position    { return x.OpenQuote }
func (x *HeredocExpr) Pos() Position           { return x.StartPos }
func (x *ClosureExpr) Pos() Position           { return x.FuncPos }
func (x *ArrowFuncExpr) Pos() Position         { return x.FnPos }
func (x *YieldExpr) Pos() Position             { return x.YieldPos }
func (x *YieldFromExpr) Pos() Position         { return x.YieldPos }
func (x *ThrowExpr) Pos() Position             { return x.ThrowPos }
func (x *PrintExpr) Pos() Position             { return x.PrintPos }
func (x *IncludeExpr) Pos() Position           { return x.IncludePos }
func (x *IssetExpr) Pos() Position             { return x.IssetPos }
func (x *EmptyExpr) Pos() Position             { return x.EmptyPos }
func (x *EvalExpr) Pos() Position              { return x.EvalPos }
func (x *ExitExpr) Pos() Position              { return x.ExitPos }
func (x *ListExpr) Pos() Position              { return x.ListPos }
func (x *MatchExpr) Pos() Position             { return x.MatchPos }
func (x *AssignExpr) Pos() Position            { return x.Var.Pos() }
func (x *AssignRefExpr) Pos() Position         { return x.Var.Pos() }
func (x *ErrorSuppressExpr) Pos() Position     { return x.AtPos }
func (x *ShellExecExpr) Pos() Position         { return x.OpenTick }
func (x *MagicConstExpr) Pos() Position        { return x.ConstPos }
func (x *ParenExpr) Pos() Position             { return x.Lparen }

// End implementations for expressions
func (x *BadExpr) End() Position               { return x.To }
func (x *Ident) End() Position                 { return Position{Offset: x.NamePos.Offset + len(x.Name)} }
func (x *Variable) End() Position              { return x.Name.End() }
func (x *Literal) End() Position               { return Position{Offset: x.ValuePos.Offset + len(x.Value)} }
func (x *ArrayExpr) End() Position             { return x.Rbrack }
func (x *BinaryExpr) End() Position            { return x.Right.End() }
func (x *UnaryExpr) End() Position             { return x.X.End() }
func (x *PostfixExpr) End() Position           { return x.OpPos }
func (x *TernaryExpr) End() Position           { return x.Else.End() }
func (x *CoalesceExpr) End() Position          { return x.Right.End() }
func (x *InstanceofExpr) End() Position        { return x.Class.End() }
func (x *CastExpr) End() Position              { return x.X.End() }
func (x *CloneExpr) End() Position             { return x.Expr.End() }
func (x *NewExpr) End() Position               { if x.Args != nil { return x.Args.Rparen }; return x.Class.End() }
func (x *CallExpr) End() Position              { return x.Args.Rparen }
func (x *MethodCallExpr) End() Position        { return x.Args.Rparen }
func (x *StaticCallExpr) End() Position        { return x.Args.Rparen }
func (x *PropertyFetchExpr) End() Position     { return x.Property.End() }
func (x *StaticPropertyFetchExpr) End() Position { return x.Property.End() }
func (x *ClassConstFetchExpr) End() Position   { return x.Const.End() }
func (x *ArrayAccessExpr) End() Position       { return x.Rbrack }
func (x *EncapsedStringExpr) End() Position    { return x.CloseQuote }
func (x *HeredocExpr) End() Position           { return x.EndPos }
func (x *ClosureExpr) End() Position           { return x.Body.End() }
func (x *ArrowFuncExpr) End() Position         { return x.Body.End() }
func (x *YieldExpr) End() Position             { if x.Value != nil { return x.Value.End() }; return x.YieldPos }
func (x *YieldFromExpr) End() Position         { return x.Expr.End() }
func (x *ThrowExpr) End() Position             { return x.Expr.End() }
func (x *PrintExpr) End() Position             { return x.Expr.End() }
func (x *IncludeExpr) End() Position           { return x.Expr.End() }
func (x *IssetExpr) End() Position             { return x.Rparen }
func (x *EmptyExpr) End() Position             { return x.Rparen }
func (x *EvalExpr) End() Position              { return x.Rparen }
func (x *ExitExpr) End() Position              { if x.Expr != nil { return x.Expr.End() }; return x.ExitPos }
func (x *ListExpr) End() Position              { return x.Rparen }
func (x *MatchExpr) End() Position             { return x.Rbrace }
func (x *AssignExpr) End() Position            { return x.Value.End() }
func (x *AssignRefExpr) End() Position         { return x.Value.End() }
func (x *ErrorSuppressExpr) End() Position     { return x.Expr.End() }
func (x *ShellExecExpr) End() Position         { return x.CloseTick }
func (x *MagicConstExpr) End() Position        { return x.ConstPos }
func (x *ParenExpr) End() Position             { return x.Rparen }
