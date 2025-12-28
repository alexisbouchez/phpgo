package parser

import (
	"testing"

	"github.com/alexisbouchez/phpgo/ast"
	"github.com/alexisbouchez/phpgo/token"
)

func TestParseSimpleVariable(t *testing.T) {
	input := `<?php $x;`
	file := ParseString(input)

	if len(file.Stmts) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(file.Stmts))
	}

	stmt, ok := file.Stmts[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", file.Stmts[0])
	}

	v, ok := stmt.Expr.(*ast.Variable)
	if !ok {
		t.Fatalf("expected Variable, got %T", stmt.Expr)
	}

	ident, ok := v.Name.(*ast.Ident)
	if !ok {
		t.Fatalf("expected Ident, got %T", v.Name)
	}

	if ident.Name != "x" {
		t.Errorf("expected 'x', got %q", ident.Name)
	}
}

func TestParseIntegerLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`<?php 123;`, "123"},
		{`<?php 0xFF;`, "0xFF"},
		{`<?php 0b1010;`, "0b1010"},
		{`<?php 0o755;`, "0o755"},
		{`<?php 1_000_000;`, "1_000_000"},
	}

	for _, tt := range tests {
		file := ParseString(tt.input)
		if len(file.Stmts) != 1 {
			t.Fatalf("expected 1 statement, got %d", len(file.Stmts))
		}

		stmt := file.Stmts[0].(*ast.ExprStmt)
		lit, ok := stmt.Expr.(*ast.Literal)
		if !ok {
			t.Fatalf("expected Literal, got %T", stmt.Expr)
		}
		if lit.Value != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, lit.Value)
		}
	}
}

func TestParseFloatLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`<?php 1.5;`, "1.5"},
		{`<?php .5;`, ".5"},
		{`<?php 1e5;`, "1e5"},
		{`<?php 1.5e-10;`, "1.5e-10"},
	}

	for _, tt := range tests {
		file := ParseString(tt.input)
		stmt := file.Stmts[0].(*ast.ExprStmt)
		lit, ok := stmt.Expr.(*ast.Literal)
		if !ok {
			t.Fatalf("expected Literal, got %T", stmt.Expr)
		}
		if lit.Value != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, lit.Value)
		}
	}
}

func TestParseStringLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`<?php 'hello';`, "'hello'"},
		{`<?php "world";`, `"world"`},
	}

	for _, tt := range tests {
		file := ParseString(tt.input)
		stmt := file.Stmts[0].(*ast.ExprStmt)
		lit, ok := stmt.Expr.(*ast.Literal)
		if !ok {
			t.Fatalf("expected Literal, got %T", stmt.Expr)
		}
		if lit.Value != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, lit.Value)
		}
	}
}

func TestParseBinaryExpressions(t *testing.T) {
	tests := []struct {
		input string
		op    token.Token
	}{
		{`<?php 1 + 2;`, token.PLUS},
		{`<?php 1 - 2;`, token.MINUS},
		{`<?php 1 * 2;`, token.ASTERISK},
		{`<?php 1 / 2;`, token.SLASH},
		{`<?php 1 % 2;`, token.PERCENT},
		{`<?php 1 ** 2;`, token.T_POW},
		{`<?php 1 == 2;`, token.T_IS_EQUAL},
		{`<?php 1 === 2;`, token.T_IS_IDENTICAL},
		{`<?php 1 != 2;`, token.T_IS_NOT_EQUAL},
		{`<?php 1 !== 2;`, token.T_IS_NOT_IDENTICAL},
		{`<?php 1 < 2;`, token.LESS},
		{`<?php 1 <= 2;`, token.T_IS_SMALLER_OR_EQUAL},
		{`<?php 1 > 2;`, token.GREATER},
		{`<?php 1 >= 2;`, token.T_IS_GREATER_OR_EQUAL},
		{`<?php 1 <=> 2;`, token.T_SPACESHIP},
		{`<?php 1 && 2;`, token.T_BOOLEAN_AND},
		{`<?php 1 || 2;`, token.T_BOOLEAN_OR},
		{`<?php 1 . 2;`, token.DOT},
		{`<?php 1 & 2;`, token.AMPERSAND},
		{`<?php 1 | 2;`, token.PIPE},
		{`<?php 1 ^ 2;`, token.CARET},
		{`<?php 1 << 2;`, token.T_SL},
		{`<?php 1 >> 2;`, token.T_SR},
	}

	for _, tt := range tests {
		file := ParseString(tt.input)
		stmt := file.Stmts[0].(*ast.ExprStmt)
		bin, ok := stmt.Expr.(*ast.BinaryExpr)
		if !ok {
			t.Fatalf("input %q: expected BinaryExpr, got %T", tt.input, stmt.Expr)
		}
		if bin.Op != tt.op {
			t.Errorf("input %q: expected op %v, got %v", tt.input, tt.op, bin.Op)
		}
	}
}

func TestParseUnaryExpressions(t *testing.T) {
	tests := []struct {
		input string
		op    string
	}{
		{`<?php -$x;`, "-"},
		{`<?php +$x;`, "+"},
		{`<?php !$x;`, "!"},
		{`<?php ~$x;`, "~"},
		{`<?php ++$x;`, "++"},
		{`<?php --$x;`, "--"},
	}

	for _, tt := range tests {
		file := ParseString(tt.input)
		stmt := file.Stmts[0].(*ast.ExprStmt)
		un, ok := stmt.Expr.(*ast.UnaryExpr)
		if !ok {
			t.Fatalf("input %q: expected UnaryExpr, got %T", tt.input, stmt.Expr)
		}
		_ = un
	}
}

func TestParsePostfixExpressions(t *testing.T) {
	tests := []string{
		`<?php $x++;`,
		`<?php $x--;`,
	}

	for _, input := range tests {
		file := ParseString(input)
		stmt := file.Stmts[0].(*ast.ExprStmt)
		_, ok := stmt.Expr.(*ast.PostfixExpr)
		if !ok {
			t.Fatalf("input %q: expected PostfixExpr, got %T", input, stmt.Expr)
		}
	}
}

func TestParseAssignment(t *testing.T) {
	input := `<?php $x = 1;`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	assign, ok := stmt.Expr.(*ast.AssignExpr)
	if !ok {
		t.Fatalf("expected AssignExpr, got %T", stmt.Expr)
	}

	v, ok := assign.Var.(*ast.Variable)
	if !ok {
		t.Fatalf("expected Variable, got %T", assign.Var)
	}
	_ = v

	lit, ok := assign.Value.(*ast.Literal)
	if !ok {
		t.Fatalf("expected Literal, got %T", assign.Value)
	}
	if lit.Value != "1" {
		t.Errorf("expected '1', got %q", lit.Value)
	}
}

func TestParseCompoundAssignment(t *testing.T) {
	tests := []struct {
		input string
		op    string
	}{
		{`<?php $x += 1;`, "+="},
		{`<?php $x -= 1;`, "-="},
		{`<?php $x *= 1;`, "*="},
		{`<?php $x /= 1;`, "/="},
		{`<?php $x .= 1;`, ".="},
		{`<?php $x ??= 1;`, "??="},
	}

	for _, tt := range tests {
		file := ParseString(tt.input)
		stmt := file.Stmts[0].(*ast.ExprStmt)
		assign, ok := stmt.Expr.(*ast.AssignExpr)
		if !ok {
			t.Fatalf("input %q: expected AssignExpr, got %T", tt.input, stmt.Expr)
		}
		_ = assign
	}
}

func TestParseTernary(t *testing.T) {
	input := `<?php $x ? 1 : 2;`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	tern, ok := stmt.Expr.(*ast.TernaryExpr)
	if !ok {
		t.Fatalf("expected TernaryExpr, got %T", stmt.Expr)
	}

	if tern.Then == nil {
		t.Error("expected Then expression")
	}
}

func TestParseElvis(t *testing.T) {
	input := `<?php $x ?: $y;`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	tern, ok := stmt.Expr.(*ast.TernaryExpr)
	if !ok {
		t.Fatalf("expected TernaryExpr, got %T", stmt.Expr)
	}

	if tern.Then != nil {
		t.Error("Elvis operator should have nil Then")
	}
}

func TestParseNullCoalesce(t *testing.T) {
	input := `<?php $x ?? $y;`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	coal, ok := stmt.Expr.(*ast.CoalesceExpr)
	if !ok {
		t.Fatalf("expected CoalesceExpr, got %T", stmt.Expr)
	}
	_ = coal
}

func TestParseFunctionCall(t *testing.T) {
	input := `<?php foo(1, 2, 3);`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	call, ok := stmt.Expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", stmt.Expr)
	}

	if len(call.Args.Args) != 3 {
		t.Errorf("expected 3 args, got %d", len(call.Args.Args))
	}
}

func TestParseMethodCall(t *testing.T) {
	input := `<?php $obj->method();`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	call, ok := stmt.Expr.(*ast.MethodCallExpr)
	if !ok {
		t.Fatalf("expected MethodCallExpr, got %T", stmt.Expr)
	}

	if call.NullSafe {
		t.Error("expected non-nullsafe call")
	}
}

func TestParseNullsafeMethodCall(t *testing.T) {
	input := `<?php $obj?->method();`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	call, ok := stmt.Expr.(*ast.MethodCallExpr)
	if !ok {
		t.Fatalf("expected MethodCallExpr, got %T", stmt.Expr)
	}

	if !call.NullSafe {
		t.Error("expected nullsafe call")
	}
}

func TestParseStaticCall(t *testing.T) {
	input := `<?php Foo::bar();`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	call, ok := stmt.Expr.(*ast.StaticCallExpr)
	if !ok {
		t.Fatalf("expected StaticCallExpr, got %T", stmt.Expr)
	}
	_ = call
}

func TestParseArrayAccess(t *testing.T) {
	input := `<?php $arr[0];`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	access, ok := stmt.Expr.(*ast.ArrayAccessExpr)
	if !ok {
		t.Fatalf("expected ArrayAccessExpr, got %T", stmt.Expr)
	}
	_ = access
}

func TestParsePropertyAccess(t *testing.T) {
	input := `<?php $obj->prop;`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	prop, ok := stmt.Expr.(*ast.PropertyFetchExpr)
	if !ok {
		t.Fatalf("expected PropertyFetchExpr, got %T", stmt.Expr)
	}
	_ = prop
}

func TestParseArrayLiteral(t *testing.T) {
	tests := []struct {
		input string
		count int
	}{
		{`<?php [];`, 0},
		{`<?php [1, 2, 3];`, 3},
		{`<?php ['a' => 1, 'b' => 2];`, 2},
		{`<?php array(1, 2);`, 2},
	}

	for _, tt := range tests {
		file := ParseString(tt.input)
		stmt := file.Stmts[0].(*ast.ExprStmt)
		arr, ok := stmt.Expr.(*ast.ArrayExpr)
		if !ok {
			t.Fatalf("input %q: expected ArrayExpr, got %T", tt.input, stmt.Expr)
		}
		if len(arr.Items) != tt.count {
			t.Errorf("input %q: expected %d items, got %d", tt.input, tt.count, len(arr.Items))
		}
	}
}

func TestParseNew(t *testing.T) {
	input := `<?php new Foo();`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	new_, ok := stmt.Expr.(*ast.NewExpr)
	if !ok {
		t.Fatalf("expected NewExpr, got %T", stmt.Expr)
	}
	_ = new_
}

func TestParseCast(t *testing.T) {
	tests := []string{
		`<?php (int) $x;`,
		`<?php (float) $x;`,
		`<?php (string) $x;`,
		`<?php (array) $x;`,
		`<?php (object) $x;`,
		`<?php (bool) $x;`,
	}

	for _, input := range tests {
		file := ParseString(input)
		stmt := file.Stmts[0].(*ast.ExprStmt)
		cast, ok := stmt.Expr.(*ast.CastExpr)
		if !ok {
			t.Fatalf("input %q: expected CastExpr, got %T", input, stmt.Expr)
		}
		_ = cast
	}
}

func TestParseClone(t *testing.T) {
	input := `<?php clone $obj;`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	clone, ok := stmt.Expr.(*ast.CloneExpr)
	if !ok {
		t.Fatalf("expected CloneExpr, got %T", stmt.Expr)
	}
	_ = clone
}

func TestParseIfStatement(t *testing.T) {
	input := `<?php if ($x) { echo $x; }`
	file := ParseString(input)

	if_, ok := file.Stmts[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", file.Stmts[0])
	}
	_ = if_
}

func TestParseIfElseStatement(t *testing.T) {
	input := `<?php if ($x) { echo 1; } else { echo 2; }`
	file := ParseString(input)

	if_, ok := file.Stmts[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", file.Stmts[0])
	}

	if if_.Else == nil {
		t.Error("expected Else clause")
	}
}

func TestParseIfElseifElse(t *testing.T) {
	input := `<?php if ($x) { echo 1; } elseif ($y) { echo 2; } else { echo 3; }`
	file := ParseString(input)

	if_, ok := file.Stmts[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", file.Stmts[0])
	}

	if len(if_.ElseIfs) != 1 {
		t.Errorf("expected 1 elseif, got %d", len(if_.ElseIfs))
	}
}

func TestParseWhileStatement(t *testing.T) {
	input := `<?php while ($x) { $x--; }`
	file := ParseString(input)

	while, ok := file.Stmts[0].(*ast.WhileStmt)
	if !ok {
		t.Fatalf("expected WhileStmt, got %T", file.Stmts[0])
	}
	_ = while
}

func TestParseDoWhileStatement(t *testing.T) {
	input := `<?php do { $x++; } while ($x < 10);`
	file := ParseString(input)

	do, ok := file.Stmts[0].(*ast.DoWhileStmt)
	if !ok {
		t.Fatalf("expected DoWhileStmt, got %T", file.Stmts[0])
	}
	_ = do
}

func TestParseForStatement(t *testing.T) {
	input := `<?php for ($i = 0; $i < 10; $i++) { echo $i; }`
	file := ParseString(input)

	for_, ok := file.Stmts[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", file.Stmts[0])
	}
	_ = for_
}

func TestParseForeachStatement(t *testing.T) {
	input := `<?php foreach ($arr as $val) { echo $val; }`
	file := ParseString(input)

	foreach, ok := file.Stmts[0].(*ast.ForeachStmt)
	if !ok {
		t.Fatalf("expected ForeachStmt, got %T", file.Stmts[0])
	}
	_ = foreach
}

func TestParseForeachWithKey(t *testing.T) {
	input := `<?php foreach ($arr as $key => $val) { echo $key; }`
	file := ParseString(input)

	foreach, ok := file.Stmts[0].(*ast.ForeachStmt)
	if !ok {
		t.Fatalf("expected ForeachStmt, got %T", file.Stmts[0])
	}

	if foreach.KeyVar == nil {
		t.Error("expected key variable")
	}
}

func TestParseSwitchStatement(t *testing.T) {
	input := `<?php switch ($x) { case 1: echo 1; break; default: echo 0; }`
	file := ParseString(input)

	switch_, ok := file.Stmts[0].(*ast.SwitchStmt)
	if !ok {
		t.Fatalf("expected SwitchStmt, got %T", file.Stmts[0])
	}

	if len(switch_.Cases) != 2 {
		t.Errorf("expected 2 cases, got %d", len(switch_.Cases))
	}
}

func TestParseTryCatch(t *testing.T) {
	input := `<?php try { foo(); } catch (Exception $e) { bar(); }`
	file := ParseString(input)

	try, ok := file.Stmts[0].(*ast.TryStmt)
	if !ok {
		t.Fatalf("expected TryStmt, got %T", file.Stmts[0])
	}

	if len(try.Catches) != 1 {
		t.Errorf("expected 1 catch, got %d", len(try.Catches))
	}
}

func TestParseTryCatchFinally(t *testing.T) {
	input := `<?php try { foo(); } catch (Exception $e) { bar(); } finally { cleanup(); }`
	file := ParseString(input)

	try, ok := file.Stmts[0].(*ast.TryStmt)
	if !ok {
		t.Fatalf("expected TryStmt, got %T", file.Stmts[0])
	}

	if try.Finally == nil {
		t.Error("expected finally clause")
	}
}

func TestParseReturn(t *testing.T) {
	tests := []struct {
		input   string
		hasExpr bool
	}{
		{`<?php return;`, false},
		{`<?php return 1;`, true},
		{`<?php return $x;`, true},
	}

	for _, tt := range tests {
		file := ParseString(tt.input)
		ret, ok := file.Stmts[0].(*ast.ReturnStmt)
		if !ok {
			t.Fatalf("input %q: expected ReturnStmt, got %T", tt.input, file.Stmts[0])
		}
		if (ret.Result != nil) != tt.hasExpr {
			t.Errorf("input %q: hasExpr mismatch", tt.input)
		}
	}
}

func TestParseEcho(t *testing.T) {
	input := `<?php echo 1, 2, 3;`
	file := ParseString(input)

	echo, ok := file.Stmts[0].(*ast.EchoStmt)
	if !ok {
		t.Fatalf("expected EchoStmt, got %T", file.Stmts[0])
	}

	if len(echo.Exprs) != 3 {
		t.Errorf("expected 3 expressions, got %d", len(echo.Exprs))
	}
}

func TestParseFunctionDeclaration(t *testing.T) {
	input := `<?php function foo($a, $b) { return $a + $b; }`
	file := ParseString(input)

	fn, ok := file.Stmts[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected FunctionDecl, got %T", file.Stmts[0])
	}

	if fn.Name.Name != "foo" {
		t.Errorf("expected 'foo', got %q", fn.Name.Name)
	}

	if len(fn.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(fn.Params))
	}
}

func TestParseFunctionWithReturnType(t *testing.T) {
	input := `<?php function foo(): int { return 1; }`
	file := ParseString(input)

	fn, ok := file.Stmts[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("expected FunctionDecl, got %T", file.Stmts[0])
	}

	if fn.ReturnType == nil {
		t.Error("expected return type")
	}
}

func TestParseClassDeclaration(t *testing.T) {
	input := `<?php class Foo { }`
	file := ParseString(input)

	class, ok := file.Stmts[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", file.Stmts[0])
	}

	if class.Name.Name != "Foo" {
		t.Errorf("expected 'Foo', got %q", class.Name.Name)
	}
}

func TestParseClassWithExtends(t *testing.T) {
	input := `<?php class Foo extends Bar { }`
	file := ParseString(input)

	class, ok := file.Stmts[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", file.Stmts[0])
	}

	if class.Extends == nil {
		t.Error("expected extends")
	}
}

func TestParseClassWithImplements(t *testing.T) {
	input := `<?php class Foo implements Bar, Baz { }`
	file := ParseString(input)

	class, ok := file.Stmts[0].(*ast.ClassDecl)
	if !ok {
		t.Fatalf("expected ClassDecl, got %T", file.Stmts[0])
	}

	if len(class.Implements) != 2 {
		t.Errorf("expected 2 implements, got %d", len(class.Implements))
	}
}

func TestParseClassProperty(t *testing.T) {
	input := `<?php class Foo { public $bar; }`
	file := ParseString(input)

	class := file.Stmts[0].(*ast.ClassDecl)
	if len(class.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(class.Members))
	}

	prop, ok := class.Members[0].(*ast.PropertyDecl)
	if !ok {
		t.Fatalf("expected PropertyDecl, got %T", class.Members[0])
	}
	_ = prop
}

func TestParseClassMethod(t *testing.T) {
	input := `<?php class Foo { public function bar() { } }`
	file := ParseString(input)

	class := file.Stmts[0].(*ast.ClassDecl)
	if len(class.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(class.Members))
	}

	method, ok := class.Members[0].(*ast.MethodDecl)
	if !ok {
		t.Fatalf("expected MethodDecl, got %T", class.Members[0])
	}
	_ = method
}

func TestParseClassConstant(t *testing.T) {
	input := `<?php class Foo { const BAR = 1; }`
	file := ParseString(input)

	class := file.Stmts[0].(*ast.ClassDecl)
	if len(class.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(class.Members))
	}

	const_, ok := class.Members[0].(*ast.ClassConstDecl)
	if !ok {
		t.Fatalf("expected ClassConstDecl, got %T", class.Members[0])
	}
	_ = const_
}

func TestParseInterface(t *testing.T) {
	input := `<?php interface Foo { public function bar(); }`
	file := ParseString(input)

	iface, ok := file.Stmts[0].(*ast.InterfaceDecl)
	if !ok {
		t.Fatalf("expected InterfaceDecl, got %T", file.Stmts[0])
	}

	if iface.Name.Name != "Foo" {
		t.Errorf("expected 'Foo', got %q", iface.Name.Name)
	}
}

func TestParseTrait(t *testing.T) {
	input := `<?php trait Foo { public function bar() { } }`
	file := ParseString(input)

	trait, ok := file.Stmts[0].(*ast.TraitDecl)
	if !ok {
		t.Fatalf("expected TraitDecl, got %T", file.Stmts[0])
	}

	if trait.Name.Name != "Foo" {
		t.Errorf("expected 'Foo', got %q", trait.Name.Name)
	}
}

func TestParseEnum(t *testing.T) {
	input := `<?php enum Status { case Active; case Inactive; }`
	file := ParseString(input)

	enum, ok := file.Stmts[0].(*ast.EnumDecl)
	if !ok {
		t.Fatalf("expected EnumDecl, got %T", file.Stmts[0])
	}

	if enum.Name.Name != "Status" {
		t.Errorf("expected 'Status', got %q", enum.Name.Name)
	}

	if len(enum.Members) != 2 {
		t.Errorf("expected 2 cases, got %d", len(enum.Members))
	}
}

func TestParseEnumWithBackingType(t *testing.T) {
	input := `<?php enum Status: string { case Active = 'active'; }`
	file := ParseString(input)

	enum, ok := file.Stmts[0].(*ast.EnumDecl)
	if !ok {
		t.Fatalf("expected EnumDecl, got %T", file.Stmts[0])
	}

	if enum.BackingType == nil {
		t.Error("expected backing type")
	}
}

func TestParseNamespace(t *testing.T) {
	input := `<?php namespace App\Controllers;`
	file := ParseString(input)

	ns, ok := file.Stmts[0].(*ast.NamespaceDecl)
	if !ok {
		t.Fatalf("expected NamespaceDecl, got %T", file.Stmts[0])
	}
	_ = ns
}

func TestParseUseDeclaration(t *testing.T) {
	input := `<?php use App\Models\User;`
	file := ParseString(input)

	use, ok := file.Stmts[0].(*ast.UseDecl)
	if !ok {
		t.Fatalf("expected UseDecl, got %T", file.Stmts[0])
	}

	if len(use.Uses) != 1 {
		t.Errorf("expected 1 use, got %d", len(use.Uses))
	}
}

func TestParseUseWithAlias(t *testing.T) {
	input := `<?php use App\Models\User as U;`
	file := ParseString(input)

	use := file.Stmts[0].(*ast.UseDecl)
	if use.Uses[0].Alias == nil {
		t.Error("expected alias")
	}
}

func TestParseClosure(t *testing.T) {
	input := `<?php $fn = function($x) { return $x * 2; };`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	assign := stmt.Expr.(*ast.AssignExpr)
	closure, ok := assign.Value.(*ast.ClosureExpr)
	if !ok {
		t.Fatalf("expected ClosureExpr, got %T", assign.Value)
	}
	_ = closure
}

func TestParseClosureWithUse(t *testing.T) {
	input := `<?php $fn = function($x) use ($y) { return $x + $y; };`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	assign := stmt.Expr.(*ast.AssignExpr)
	closure, ok := assign.Value.(*ast.ClosureExpr)
	if !ok {
		t.Fatalf("expected ClosureExpr, got %T", assign.Value)
	}

	if len(closure.Uses) != 1 {
		t.Errorf("expected 1 use, got %d", len(closure.Uses))
	}
}

func TestParseArrowFunction(t *testing.T) {
	input := `<?php $fn = fn($x) => $x * 2;`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	assign := stmt.Expr.(*ast.AssignExpr)
	arrow, ok := assign.Value.(*ast.ArrowFuncExpr)
	if !ok {
		t.Fatalf("expected ArrowFuncExpr, got %T", assign.Value)
	}
	_ = arrow
}

func TestParseMatch(t *testing.T) {
	input := `<?php match($x) { 1 => 'one', 2 => 'two', default => 'other' };`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	match, ok := stmt.Expr.(*ast.MatchExpr)
	if !ok {
		t.Fatalf("expected MatchExpr, got %T", stmt.Expr)
	}

	if len(match.Arms) != 3 {
		t.Errorf("expected 3 arms, got %d", len(match.Arms))
	}
}

func TestParseInstanceof(t *testing.T) {
	input := `<?php $x instanceof Foo;`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	instanceof, ok := stmt.Expr.(*ast.InstanceofExpr)
	if !ok {
		t.Fatalf("expected InstanceofExpr, got %T", stmt.Expr)
	}
	_ = instanceof
}

func TestParsePrecedence(t *testing.T) {
	// 1 + 2 * 3 should be 1 + (2 * 3)
	input := `<?php 1 + 2 * 3;`
	file := ParseString(input)

	stmt := file.Stmts[0].(*ast.ExprStmt)
	bin, ok := stmt.Expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", stmt.Expr)
	}

	// The top-level should be +
	_, isRight := bin.Right.(*ast.BinaryExpr)
	if !isRight {
		t.Error("expected right side to be BinaryExpr (* has higher precedence)")
	}
}

func TestParseComplexProgram(t *testing.T) {
	input := `<?php
namespace App\Controllers;

use App\Models\User;
use Illuminate\Http\Request;

class UserController
{
    private UserRepository $repo;

    public function __construct(UserRepository $repo)
    {
        $this->repo = $repo;
    }

    public function index(): array
    {
        return $this->repo->findAll();
    }

    public function show(int $id): ?User
    {
        return $this->repo->find($id);
    }
}
`

	file := ParseString(input)
	if file == nil {
		t.Fatal("ParseString returned nil")
	}

	// Should have namespace, 2 use statements, and 1 class
	if len(file.Stmts) < 1 {
		t.Errorf("expected at least 1 statement, got %d", len(file.Stmts))
	}
}
