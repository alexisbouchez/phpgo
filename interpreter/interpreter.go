package interpreter

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alexisbouchez/phpgo/ast"
	"github.com/alexisbouchez/phpgo/parser"
	"github.com/alexisbouchez/phpgo/runtime"
	"github.com/alexisbouchez/phpgo/token"
)

// Interpreter executes PHP code.
type Interpreter struct {
	env           *runtime.Environment
	output        strings.Builder
	staticVars    *runtime.StaticVars
	currentClass  string          // Current class context for self/parent/static
	currentThis   *runtime.Object // Current object for method calls
	includedFiles map[string]bool // Track files included with _once
	currentDir    string          // Current directory for relative paths
}

// New creates a new interpreter.
func New() *Interpreter {
	env := runtime.NewEnvironment()
	env.InitSuperglobals()
	cwd, _ := os.Getwd()
	i := &Interpreter{
		env:           env,
		staticVars:    runtime.NewStaticVars(),
		includedFiles: make(map[string]bool),
		currentDir:    cwd,
	}
	i.registerBuiltins()
	return i
}

// Eval parses and executes PHP code.
func (i *Interpreter) Eval(input string) runtime.Value {
	file := parser.ParseString(input)
	return i.evalFile(file)
}

// Output returns the captured output.
func (i *Interpreter) Output() string {
	return i.output.String()
}

func (i *Interpreter) evalFile(file *ast.File) runtime.Value {
	var result runtime.Value = runtime.NULL
	for _, stmt := range file.Stmts {
		result = i.evalStmt(stmt)
		// Check for return/break/continue
		switch result.(type) {
		case *runtime.ReturnValue:
			return result.(*runtime.ReturnValue).Value
		case *runtime.Error:
			return result
		}
	}
	return result
}

// ----------------------------------------------------------------------------
// Statement evaluation

func (i *Interpreter) evalStmt(stmt ast.Stmt) runtime.Value {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		return i.evalExpr(s.Expr)
	case *ast.EchoStmt:
		return i.evalEcho(s)
	case *ast.IfStmt:
		return i.evalIf(s)
	case *ast.WhileStmt:
		return i.evalWhile(s)
	case *ast.DoWhileStmt:
		return i.evalDoWhile(s)
	case *ast.ForStmt:
		return i.evalFor(s)
	case *ast.ForeachStmt:
		return i.evalForeach(s)
	case *ast.SwitchStmt:
		return i.evalSwitch(s)
	case *ast.BreakStmt:
		levels := 1
		if s.Num != nil {
			levels = int(i.evalExpr(s.Num).ToInt())
		}
		return &runtime.Break{Levels: levels}
	case *ast.ContinueStmt:
		levels := 1
		if s.Num != nil {
			levels = int(i.evalExpr(s.Num).ToInt())
		}
		return &runtime.Continue{Levels: levels}
	case *ast.ReturnStmt:
		var val runtime.Value = runtime.NULL
		if s.Result != nil {
			val = i.evalExpr(s.Result)
		}
		return &runtime.ReturnValue{Value: val}
	case *ast.BlockStmt:
		return i.evalBlock(s)
	case *ast.TryStmt:
		return i.evalTry(s)
	case *ast.ThrowStmt:
		return i.evalThrow(s)
	case *ast.GlobalStmt:
		for _, v := range s.Vars {
			name := v.Name.(*ast.Ident).Name
			i.env.ImportGlobal(name)
		}
		return runtime.NULL
	case *ast.StaticVarStmt:
		return i.evalStatic(s)
	case *ast.UnsetStmt:
		for _, v := range s.Vars {
			if varExpr, ok := v.(*ast.Variable); ok {
				name := varExpr.Name.(*ast.Ident).Name
				i.env.Unset(name)
			} else if propExpr, ok := v.(*ast.PropertyFetchExpr); ok {
				// Property unset - check for __unset
				objVal := i.evalExpr(propExpr.Object)
				if obj, ok := objVal.(*runtime.Object); ok {
					propName := propExpr.Property.(*ast.Ident).Name
					// Check if property exists
					if _, exists := obj.Properties[propName]; exists {
						delete(obj.Properties, propName)
						continue
					}
					// Check for __unset magic method
					if unsetMethod, _ := i.findMethod(obj.Class, "__unset"); unsetMethod != nil {
						i.callMagicGetSet(obj, unsetMethod, propName, nil)
					}
				}
			} else if arrExpr, ok := v.(*ast.ArrayAccessExpr); ok {
				// Array element unset
				arrVal := i.evalExpr(arrExpr.Array)
				if arr, ok := arrVal.(*runtime.Array); ok {
					if arrExpr.Index != nil {
						key := i.evalExpr(arrExpr.Index)
						arr.Unset(key)
					}
				}
			}
		}
		return runtime.NULL
	case *ast.FunctionDecl:
		return i.evalFunctionDecl(s)
	case *ast.ClassDecl:
		return i.evalClassDecl(s)
	case *ast.InterfaceDecl:
		return i.evalInterfaceDecl(s)
	case *ast.TraitDecl:
		return i.evalTraitDecl(s)
	case *ast.EnumDecl:
		return i.evalEnumDecl(s)
	case *ast.NamespaceDecl:
		// For now, just evaluate the statements
		for _, stmt := range s.Stmts {
			i.evalStmt(stmt)
		}
		return runtime.NULL
	case *ast.UseDecl:
		// Use declarations are handled at compile time
		return runtime.NULL
	case *ast.ConstDecl:
		return i.evalConstDecl(s)
	default:
		return runtime.NewError(fmt.Sprintf("unknown statement type: %T", stmt))
	}
}

func (i *Interpreter) evalBlock(block *ast.BlockStmt) runtime.Value {
	var result runtime.Value = runtime.NULL
	for _, stmt := range block.Stmts {
		result = i.evalStmt(stmt)
		switch result.(type) {
		case *runtime.ReturnValue, *runtime.Break, *runtime.Continue, *runtime.Exception:
			return result
		}
	}
	return result
}

func (i *Interpreter) evalEcho(s *ast.EchoStmt) runtime.Value {
	for _, expr := range s.Exprs {
		val := i.evalExpr(expr)
		i.output.WriteString(val.ToString())
	}
	return runtime.NULL
}

func (i *Interpreter) evalIf(s *ast.IfStmt) runtime.Value {
	cond := i.evalExpr(s.Cond)
	if cond.ToBool() {
		return i.evalStmt(s.Body)
	}

	for _, elseif := range s.ElseIfs {
		cond := i.evalExpr(elseif.Cond)
		if cond.ToBool() {
			return i.evalStmt(elseif.Body)
		}
	}

	if s.Else != nil {
		return i.evalStmt(s.Else.Body)
	}

	return runtime.NULL
}

func (i *Interpreter) evalWhile(s *ast.WhileStmt) runtime.Value {
	for {
		cond := i.evalExpr(s.Cond)
		if !cond.ToBool() {
			break
		}
		result := i.evalStmt(s.Body)
		switch r := result.(type) {
		case *runtime.Break:
			if r.Levels <= 1 {
				return runtime.NULL
			}
			return &runtime.Break{Levels: r.Levels - 1}
		case *runtime.Continue:
			if r.Levels <= 1 {
				continue
			}
			return &runtime.Continue{Levels: r.Levels - 1}
		case *runtime.ReturnValue:
			return result
		}
	}
	return runtime.NULL
}

func (i *Interpreter) evalDoWhile(s *ast.DoWhileStmt) runtime.Value {
	for {
		result := i.evalStmt(s.Body)
		switch r := result.(type) {
		case *runtime.Break:
			if r.Levels <= 1 {
				return runtime.NULL
			}
			return &runtime.Break{Levels: r.Levels - 1}
		case *runtime.Continue:
			if r.Levels <= 1 {
				// Continue in do-while checks condition
			}
		case *runtime.ReturnValue:
			return result
		}

		cond := i.evalExpr(s.Cond)
		if !cond.ToBool() {
			break
		}
	}
	return runtime.NULL
}

func (i *Interpreter) evalFor(s *ast.ForStmt) runtime.Value {
	// Init
	for _, init := range s.Init {
		i.evalExpr(init)
	}

	for {
		// Condition
		if len(s.Cond) > 0 {
			cond := i.evalExpr(s.Cond[len(s.Cond)-1])
			if !cond.ToBool() {
				break
			}
		}

		// Body
		result := i.evalStmt(s.Body)
		switch r := result.(type) {
		case *runtime.Break:
			if r.Levels <= 1 {
				return runtime.NULL
			}
			return &runtime.Break{Levels: r.Levels - 1}
		case *runtime.Continue:
			if r.Levels <= 1 {
				// Fall through to loop
			}
		case *runtime.ReturnValue:
			return result
		}

		// Loop
		for _, loop := range s.Loop {
			i.evalExpr(loop)
		}
	}
	return runtime.NULL
}

func (i *Interpreter) evalForeach(s *ast.ForeachStmt) runtime.Value {
	arr := i.evalExpr(s.Expr)
	arrVal, ok := arr.(*runtime.Array)
	if !ok {
		return runtime.NewError("foreach requires an array")
	}

	for _, key := range arrVal.Keys {
		val := arrVal.Elements[key]

		// Set key variable if present
		if s.KeyVar != nil {
			keyName := s.KeyVar.(*ast.Variable).Name.(*ast.Ident).Name
			i.env.Set(keyName, key)
		}

		// Set value variable
		valName := s.ValueVar.(*ast.Variable).Name.(*ast.Ident).Name
		i.env.Set(valName, val)

		// Execute body
		result := i.evalStmt(s.Body)
		switch r := result.(type) {
		case *runtime.Break:
			if r.Levels <= 1 {
				return runtime.NULL
			}
			return &runtime.Break{Levels: r.Levels - 1}
		case *runtime.Continue:
			if r.Levels <= 1 {
				continue
			}
			return &runtime.Continue{Levels: r.Levels - 1}
		case *runtime.ReturnValue:
			return result
		}
	}
	return runtime.NULL
}

func (i *Interpreter) evalSwitch(s *ast.SwitchStmt) runtime.Value {
	switchVal := i.evalExpr(s.Cond)
	matched := false

	for _, caseStmt := range s.Cases {
		if caseStmt.Cond == nil {
			// Default case - only match if nothing else matched
			if !matched {
				matched = true
			}
		} else if !matched {
			caseVal := i.evalExpr(caseStmt.Cond)
			if runtime.IsEqual(switchVal, caseVal) {
				matched = true
			}
		}

		if matched {
			for _, stmt := range caseStmt.Stmts {
				result := i.evalStmt(stmt)
				switch r := result.(type) {
				case *runtime.Break:
					if r.Levels <= 1 {
						return runtime.NULL
					}
					return &runtime.Break{Levels: r.Levels - 1}
				case *runtime.Continue:
					return result
				case *runtime.ReturnValue:
					return result
				}
			}
		}
	}
	return runtime.NULL
}

func (i *Interpreter) evalTry(s *ast.TryStmt) runtime.Value {
	result := i.evalBlock(s.Body)

	if exc, ok := result.(*runtime.Exception); ok {
		// Find matching catch
		for _, catch := range s.Catches {
			// For now, catch all exceptions
			// Set the exception variable
			if catch.Var != nil {
				varName := catch.Var.Name.(*ast.Ident).Name
				i.env.Set(varName, exc)
			}
			result = i.evalBlock(catch.Body)
			break
		}
	}

	if s.Finally != nil {
		i.evalBlock(s.Finally.Body)
	}

	return result
}

func (i *Interpreter) evalThrow(s *ast.ThrowStmt) runtime.Value {
	val := i.evalExpr(s.Expr)
	if exc, ok := val.(*runtime.Exception); ok {
		return exc
	}
	if obj, ok := val.(*runtime.Object); ok {
		return &runtime.Exception{
			Class:   obj.Class,
			Message: obj.GetProperty("message").ToString(),
		}
	}
	return &runtime.Exception{Message: val.ToString()}
}

func (i *Interpreter) evalStatic(s *ast.StaticVarStmt) runtime.Value {
	// Static variables are handled per-function
	// For now, just initialize them
	for _, sv := range s.Vars {
		name := sv.Var.Name.(*ast.Ident).Name
		if _, exists := i.env.Get(name); !exists {
			var val runtime.Value = runtime.NULL
			if sv.Default != nil {
				val = i.evalExpr(sv.Default)
			}
			i.env.Set(name, val)
		}
	}
	return runtime.NULL
}

// ----------------------------------------------------------------------------
// Expression evaluation

func (i *Interpreter) evalExpr(expr ast.Expr) runtime.Value {
	switch e := expr.(type) {
	case *ast.Literal:
		return i.evalLiteral(e)
	case *ast.Variable:
		return i.evalVariable(e)
	case *ast.Ident:
		return i.evalIdent(e)
	case *ast.BinaryExpr:
		return i.evalBinary(e)
	case *ast.UnaryExpr:
		return i.evalUnary(e)
	case *ast.AssignExpr:
		return i.evalAssign(e)
	case *ast.TernaryExpr:
		return i.evalTernary(e)
	case *ast.CoalesceExpr:
		return i.evalNullCoalesce(e)
	case *ast.CallExpr:
		return i.evalCall(e)
	case *ast.MethodCallExpr:
		return i.evalMethodCall(e)
	case *ast.StaticCallExpr:
		return i.evalStaticCall(e)
	case *ast.PropertyFetchExpr:
		return i.evalPropertyAccess(e)
	case *ast.StaticPropertyFetchExpr:
		return i.evalStaticProperty(e)
	case *ast.ArrayAccessExpr:
		return i.evalArrayAccess(e)
	case *ast.ArrayExpr:
		return i.evalArray(e)
	case *ast.NewExpr:
		return i.evalNew(e)
	case *ast.CloneExpr:
		return i.evalClone(e)
	case *ast.ClosureExpr:
		return i.evalClosure(e)
	case *ast.ArrowFuncExpr:
		return i.evalArrowFunc(e)
	case *ast.MatchExpr:
		return i.evalMatch(e)
	case *ast.InstanceofExpr:
		return i.evalInstanceof(e)
	case *ast.CastExpr:
		return i.evalCast(e)
	case *ast.IssetExpr:
		return i.evalIsset(e)
	case *ast.EmptyExpr:
		return i.evalEmpty(e)
	case *ast.PostfixExpr:
		return i.evalIncDec(e)
	case *ast.PrintExpr:
		val := i.evalExpr(e.Expr)
		i.output.WriteString(val.ToString())
		return runtime.NewInt(1)
	case *ast.ErrorSuppressExpr:
		// Suppress errors and evaluate
		return i.evalExpr(e.Expr)
	case *ast.ParenExpr:
		return i.evalExpr(e.X)
	case *ast.EncapsedStringExpr:
		return i.evalEncapsedString(e)
	case *ast.ClassConstFetchExpr:
		return i.evalConstantAccess(e)
	case *ast.IncludeExpr:
		return i.evalInclude(e)
	case *ast.ListExpr:
		// list() on its own doesn't make sense, it's used in assignment
		return runtime.NULL
	default:
		return runtime.NewError(fmt.Sprintf("unknown expression type: %T", expr))
	}
}

func (i *Interpreter) evalLiteral(lit *ast.Literal) runtime.Value {
	switch lit.Kind {
	case token.T_LNUMBER:
		val, _ := strconv.ParseInt(lit.Value, 0, 64)
		return runtime.NewInt(val)
	case token.T_DNUMBER:
		val, _ := strconv.ParseFloat(lit.Value, 64)
		return runtime.NewFloat(val)
	case token.T_CONSTANT_ENCAPSED_STRING:
		// Remove quotes
		s := lit.Value
		if len(s) >= 2 {
			s = s[1 : len(s)-1]
		}
		return runtime.NewString(s)
	default:
		return runtime.NewString(lit.Value)
	}
}

func (i *Interpreter) evalVariable(v *ast.Variable) runtime.Value {
	name := v.Name.(*ast.Ident).Name
	val, ok := i.env.Get(name)
	if !ok {
		return runtime.NULL
	}
	return val
}

func (i *Interpreter) evalIdent(ident *ast.Ident) runtime.Value {
	name := strings.ToLower(ident.Name)
	switch name {
	case "true":
		return runtime.TRUE
	case "false":
		return runtime.FALSE
	case "null":
		return runtime.NULL
	}

	// Check for constant
	if val, ok := i.env.GetConstant(ident.Name); ok {
		return val
	}

	return runtime.NewString(ident.Name)
}

func (i *Interpreter) evalBinary(e *ast.BinaryExpr) runtime.Value {
	left := i.evalExpr(e.Left)
	right := i.evalExpr(e.Right)

	switch e.Op {
	// Arithmetic
	case token.PLUS:
		return i.addValues(left, right)
	case token.MINUS:
		return i.subtractValues(left, right)
	case token.ASTERISK:
		return i.multiplyValues(left, right)
	case token.SLASH:
		return i.divideValues(left, right)
	case token.PERCENT:
		return runtime.NewInt(left.ToInt() % right.ToInt())
	case token.T_POW:
		return i.powerValues(left, right)

	// String
	case token.DOT:
		return runtime.NewString(left.ToString() + right.ToString())

	// Comparison
	case token.T_IS_EQUAL:
		return runtime.NewBool(runtime.IsEqual(left, right))
	case token.T_IS_NOT_EQUAL:
		return runtime.NewBool(!runtime.IsEqual(left, right))
	case token.T_IS_IDENTICAL:
		return runtime.NewBool(runtime.IsIdentical(left, right))
	case token.T_IS_NOT_IDENTICAL:
		return runtime.NewBool(!runtime.IsIdentical(left, right))
	case token.LESS:
		return runtime.NewBool(left.ToFloat() < right.ToFloat())
	case token.GREATER:
		return runtime.NewBool(left.ToFloat() > right.ToFloat())
	case token.T_IS_SMALLER_OR_EQUAL:
		return runtime.NewBool(left.ToFloat() <= right.ToFloat())
	case token.T_IS_GREATER_OR_EQUAL:
		return runtime.NewBool(left.ToFloat() >= right.ToFloat())
	case token.T_SPACESHIP:
		return runtime.NewInt(int64(runtime.Compare(left, right)))

	// Logical
	case token.T_BOOLEAN_AND, token.T_LOGICAL_AND:
		return runtime.NewBool(left.ToBool() && right.ToBool())
	case token.T_BOOLEAN_OR, token.T_LOGICAL_OR:
		return runtime.NewBool(left.ToBool() || right.ToBool())
	case token.T_LOGICAL_XOR:
		return runtime.NewBool(left.ToBool() != right.ToBool())

	// Bitwise
	case token.AMPERSAND:
		return runtime.NewInt(left.ToInt() & right.ToInt())
	case token.PIPE:
		return runtime.NewInt(left.ToInt() | right.ToInt())
	case token.CARET:
		return runtime.NewInt(left.ToInt() ^ right.ToInt())
	case token.T_SL:
		return runtime.NewInt(left.ToInt() << uint(right.ToInt()))
	case token.T_SR:
		return runtime.NewInt(left.ToInt() >> uint(right.ToInt()))

	default:
		return runtime.NewError(fmt.Sprintf("unknown binary operator: %v", e.Op))
	}
}

func (i *Interpreter) addValues(left, right runtime.Value) runtime.Value {
	_, leftFloat := left.(*runtime.Float)
	_, rightFloat := right.(*runtime.Float)
	if leftFloat || rightFloat {
		return runtime.NewFloat(left.ToFloat() + right.ToFloat())
	}
	return runtime.NewInt(left.ToInt() + right.ToInt())
}

func (i *Interpreter) subtractValues(left, right runtime.Value) runtime.Value {
	_, leftFloat := left.(*runtime.Float)
	_, rightFloat := right.(*runtime.Float)
	if leftFloat || rightFloat {
		return runtime.NewFloat(left.ToFloat() - right.ToFloat())
	}
	return runtime.NewInt(left.ToInt() - right.ToInt())
}

func (i *Interpreter) multiplyValues(left, right runtime.Value) runtime.Value {
	_, leftFloat := left.(*runtime.Float)
	_, rightFloat := right.(*runtime.Float)
	if leftFloat || rightFloat {
		return runtime.NewFloat(left.ToFloat() * right.ToFloat())
	}
	return runtime.NewInt(left.ToInt() * right.ToInt())
}

func (i *Interpreter) divideValues(left, right runtime.Value) runtime.Value {
	if right.ToFloat() == 0 {
		return runtime.NewError("Division by zero")
	}
	result := left.ToFloat() / right.ToFloat()
	if result == float64(int64(result)) {
		return runtime.NewInt(int64(result))
	}
	return runtime.NewFloat(result)
}

func (i *Interpreter) powerValues(left, right runtime.Value) runtime.Value {
	base := left.ToFloat()
	exp := right.ToInt()
	result := 1.0
	for j := int64(0); j < exp; j++ {
		result *= base
	}
	if result == float64(int64(result)) {
		return runtime.NewInt(int64(result))
	}
	return runtime.NewFloat(result)
}

func (i *Interpreter) evalUnary(e *ast.UnaryExpr) runtime.Value {
	operand := i.evalExpr(e.X)

	switch e.Op {
	case token.MINUS:
		if _, ok := operand.(*runtime.Float); ok {
			return runtime.NewFloat(-operand.ToFloat())
		}
		return runtime.NewInt(-operand.ToInt())
	case token.PLUS:
		if _, ok := operand.(*runtime.Float); ok {
			return runtime.NewFloat(operand.ToFloat())
		}
		return runtime.NewInt(operand.ToInt())
	case token.EXCLAMATION:
		return runtime.NewBool(!operand.ToBool())
	case token.TILDE:
		return runtime.NewInt(^operand.ToInt())
	case token.T_INC:
		return i.evalPreIncDec(e.X, true)
	case token.T_DEC:
		return i.evalPreIncDec(e.X, false)
	default:
		return runtime.NewError(fmt.Sprintf("unknown unary operator: %v", e.Op))
	}
}

func (i *Interpreter) evalPreIncDec(expr ast.Expr, inc bool) runtime.Value {
	if v, ok := expr.(*ast.Variable); ok {
		name := v.Name.(*ast.Ident).Name
		val, _ := i.env.Get(name)
		var newVal runtime.Value
		if inc {
			newVal = runtime.NewInt(val.ToInt() + 1)
		} else {
			newVal = runtime.NewInt(val.ToInt() - 1)
		}
		i.env.Set(name, newVal)
		return newVal
	}
	return runtime.NULL
}

func (i *Interpreter) evalIncDec(e *ast.PostfixExpr) runtime.Value {
	if v, ok := e.X.(*ast.Variable); ok {
		name := v.Name.(*ast.Ident).Name
		val, _ := i.env.Get(name)
		oldVal := val.ToInt()

		var newVal runtime.Value
		if e.Op == token.T_INC {
			newVal = runtime.NewInt(oldVal + 1)
		} else {
			newVal = runtime.NewInt(oldVal - 1)
		}
		i.env.Set(name, newVal)

		// PostfixExpr is always post-increment/decrement, returns old value
		return runtime.NewInt(oldVal)
	}

	// Handle property increment ($obj->prop++)
	if pf, ok := e.X.(*ast.PropertyFetchExpr); ok {
		obj := i.evalExpr(pf.Object)
		if objVal, ok := obj.(*runtime.Object); ok {
			propName := pf.Property.(*ast.Ident).Name
			val := objVal.GetProperty(propName)
			oldVal := val.ToInt()

			var newVal runtime.Value
			if e.Op == token.T_INC {
				newVal = runtime.NewInt(oldVal + 1)
			} else {
				newVal = runtime.NewInt(oldVal - 1)
			}
			objVal.SetProperty(propName, newVal)

			return runtime.NewInt(oldVal)
		}
	}

	// Handle static property increment
	if sp, ok := e.X.(*ast.StaticPropertyFetchExpr); ok {
		var className string
		switch c := sp.Class.(type) {
		case *ast.Ident:
			className = c.Name
			if className == "self" || className == "static" {
				className = i.currentClass
			}
		default:
			className = i.evalExpr(c).ToString()
		}

		class, ok := i.env.GetClass(className)
		if !ok {
			return runtime.NewError(fmt.Sprintf("undefined class: %s", className))
		}

		propName := sp.Property.(*ast.Variable).Name.(*ast.Ident).Name
		val := class.StaticProps[propName]
		if val == nil {
			val = runtime.NewInt(0)
		}
		oldVal := val.ToInt()

		var newVal runtime.Value
		if e.Op == token.T_INC {
			newVal = runtime.NewInt(oldVal + 1)
		} else {
			newVal = runtime.NewInt(oldVal - 1)
		}
		class.StaticProps[propName] = newVal

		return runtime.NewInt(oldVal)
	}

	return runtime.NULL
}

func (i *Interpreter) evalAssign(e *ast.AssignExpr) runtime.Value {
	val := i.evalExpr(e.Value)

	switch e.Op {
	case token.EQUALS:
		// Simple assignment
	case token.T_PLUS_EQUAL:
		left := i.evalExpr(e.Var)
		val = i.addValues(left, val)
	case token.T_MINUS_EQUAL:
		left := i.evalExpr(e.Var)
		val = i.subtractValues(left, val)
	case token.T_MUL_EQUAL:
		left := i.evalExpr(e.Var)
		val = i.multiplyValues(left, val)
	case token.T_DIV_EQUAL:
		left := i.evalExpr(e.Var)
		val = i.divideValues(left, val)
	case token.T_MOD_EQUAL:
		left := i.evalExpr(e.Var)
		val = runtime.NewInt(left.ToInt() % val.ToInt())
	case token.T_POW_EQUAL:
		left := i.evalExpr(e.Var)
		val = i.powerValues(left, val)
	case token.T_CONCAT_EQUAL:
		left := i.evalExpr(e.Var)
		val = runtime.NewString(left.ToString() + val.ToString())
	case token.T_AND_EQUAL:
		left := i.evalExpr(e.Var)
		val = runtime.NewInt(left.ToInt() & val.ToInt())
	case token.T_OR_EQUAL:
		left := i.evalExpr(e.Var)
		val = runtime.NewInt(left.ToInt() | val.ToInt())
	case token.T_XOR_EQUAL:
		left := i.evalExpr(e.Var)
		val = runtime.NewInt(left.ToInt() ^ val.ToInt())
	case token.T_SL_EQUAL:
		left := i.evalExpr(e.Var)
		val = runtime.NewInt(left.ToInt() << uint(val.ToInt()))
	case token.T_SR_EQUAL:
		left := i.evalExpr(e.Var)
		val = runtime.NewInt(left.ToInt() >> uint(val.ToInt()))
	case token.T_COALESCE_EQUAL:
		left := i.evalExpr(e.Var)
		if _, ok := left.(*runtime.Null); !ok {
			return left
		}
	}

	return i.assignTo(e.Var, val)
}

func (i *Interpreter) assignTo(target ast.Expr, val runtime.Value) runtime.Value {
	switch t := target.(type) {
	case *ast.Variable:
		name := t.Name.(*ast.Ident).Name
		i.env.Set(name, val)
	case *ast.ArrayAccessExpr:
		arr := i.evalExpr(t.Array)
		if arrVal, ok := arr.(*runtime.Array); ok {
			if t.Index == nil {
				// $arr[] = val
				arrVal.Set(nil, val)
			} else {
				key := i.evalExpr(t.Index)
				arrVal.Set(key, val)
			}
		}
	case *ast.PropertyFetchExpr:
		obj := i.evalExpr(t.Object)
		if objVal, ok := obj.(*runtime.Object); ok {
			propName := t.Property.(*ast.Ident).Name

			// Check if property is defined in class
			if _, exists := objVal.Class.Properties[propName]; exists {
				objVal.SetProperty(propName, val)
			} else if _, exists := objVal.Properties[propName]; exists {
				// Dynamic property already exists
				objVal.SetProperty(propName, val)
			} else {
				// Check for __set magic method
				if method, _ := i.findMethod(objVal.Class, "__set"); method != nil {
					i.callMagicGetSet(objVal, method, propName, val)
				} else {
					// Allow dynamic properties
					objVal.SetProperty(propName, val)
				}
			}
		}
	case *ast.StaticPropertyFetchExpr:
		var className string
		switch c := t.Class.(type) {
		case *ast.Ident:
			className = c.Name
			if className == "self" || className == "static" {
				className = i.currentClass
			}
		}
		if class, ok := i.env.GetClass(className); ok {
			propName := t.Property.(*ast.Variable).Name.(*ast.Ident).Name
			class.StaticProps[propName] = val
		}
	case *ast.ListExpr:
		// Destructuring assignment: list($a, $b) = $arr or [$a, $b] = $arr
		if arrVal, ok := val.(*runtime.Array); ok {
			for idx, item := range t.Items {
				if item == nil || item.Value == nil {
					continue // Skip empty positions
				}
				var itemVal runtime.Value = runtime.NULL
				if item.Key != nil {
					// Keyed destructuring: ["a" => $a]
					key := i.evalExpr(item.Key)
					itemVal = arrVal.Get(key)
				} else {
					// Indexed destructuring
					itemVal = arrVal.Get(runtime.NewInt(int64(idx)))
				}
				i.assignTo(item.Value, itemVal)
			}
		}
	case *ast.ArrayExpr:
		// Short array destructuring syntax: [$a, $b] = $arr
		if arrVal, ok := val.(*runtime.Array); ok {
			for idx, item := range t.Items {
				if item == nil || item.Value == nil {
					continue // Skip empty positions
				}
				var itemVal runtime.Value = runtime.NULL
				if item.Key != nil {
					// Keyed destructuring: ["a" => $a]
					key := i.evalExpr(item.Key)
					itemVal = arrVal.Get(key)
				} else {
					// Indexed destructuring
					itemVal = arrVal.Get(runtime.NewInt(int64(idx)))
				}
				i.assignTo(item.Value, itemVal)
			}
		}
	}
	return val
}

func (i *Interpreter) evalTernary(e *ast.TernaryExpr) runtime.Value {
	cond := i.evalExpr(e.Cond)
	if e.Then == nil {
		// Elvis operator: $a ?: $b
		if cond.ToBool() {
			return cond
		}
		return i.evalExpr(e.Else)
	}
	if cond.ToBool() {
		return i.evalExpr(e.Then)
	}
	return i.evalExpr(e.Else)
}

func (i *Interpreter) evalNullCoalesce(e *ast.CoalesceExpr) runtime.Value {
	left := i.evalExpr(e.Left)
	if _, ok := left.(*runtime.Null); !ok {
		return left
	}
	return i.evalExpr(e.Right)
}

func (i *Interpreter) evalCall(e *ast.CallExpr) runtime.Value {
	// Get function name
	var funcName string
	switch fn := e.Func.(type) {
	case *ast.Ident:
		funcName = fn.Name
	case *ast.Variable:
		// Variable function call
		val := i.evalExpr(fn)
		if closure, ok := val.(*runtime.Function); ok {
			return i.callFunction(closure, e.Args)
		}
		// Check for __invoke on object
		if objVal, ok := val.(*runtime.Object); ok {
			if invokeMethod, foundClass := i.findMethod(objVal.Class, "__invoke"); invokeMethod != nil {
				return i.invokeMethod(objVal, invokeMethod, foundClass, e.Args)
			}
		}
		funcName = val.ToString()
	default:
		// Could be a closure
		val := i.evalExpr(e.Func)
		if closure, ok := val.(*runtime.Function); ok {
			return i.callFunction(closure, e.Args)
		}
		if builtin, ok := val.(*runtime.Builtin); ok {
			args := i.evalArgs(e.Args)
			return builtin.Fn(args...)
		}
		// Check for __invoke magic method on object
		if objVal, ok := val.(*runtime.Object); ok {
			if invokeMethod, foundClass := i.findMethod(objVal.Class, "__invoke"); invokeMethod != nil {
				return i.invokeMethod(objVal, invokeMethod, foundClass, e.Args)
			}
		}
		return runtime.NewError(fmt.Sprintf("cannot call %T", val))
	}

	// Check for builtin first
	if builtin := i.getBuiltin(funcName); builtin != nil {
		args := i.evalArgs(e.Args)
		return builtin(args...)
	}

	// Check for user function
	if fn, ok := i.env.GetFunction(funcName); ok {
		return i.callFunction(fn, e.Args)
	}

	return runtime.NewError(fmt.Sprintf("undefined function: %s", funcName))
}

func (i *Interpreter) evalArgs(args *ast.ArgumentList) []runtime.Value {
	if args == nil {
		return nil
	}
	var result []runtime.Value
	for _, arg := range args.Args {
		val := i.evalExpr(arg.Value)
		if arg.Unpack {
			// Spread operator: ...$array
			if arr, ok := val.(*runtime.Array); ok {
				for _, k := range arr.Keys {
					result = append(result, arr.Elements[k])
				}
			}
		} else {
			result = append(result, val)
		}
	}
	return result
}

func (i *Interpreter) callFunction(fn *runtime.Function, args *ast.ArgumentList) runtime.Value {
	// Create new environment
	env := runtime.NewEnclosedEnvironment(fn.Env)
	oldEnv := i.env
	i.env = env

	// Bind parameters
	argVals := i.evalArgsInEnv(oldEnv, args)
	numParams := len(fn.Params)

	for idx, param := range fn.Params {
		// Check if this is the variadic param (last one and fn.Variadic is true)
		isVariadicParam := fn.Variadic && idx == numParams-1

		if isVariadicParam {
			// Collect remaining args into an array
			variadicArr := runtime.NewArray()
			for j := idx; j < len(argVals); j++ {
				variadicArr.Set(nil, argVals[j])
			}
			env.Set(param, variadicArr)
		} else if idx < len(argVals) {
			env.Set(param, argVals[idx])
		} else if fn.Defaults != nil && idx < len(fn.Defaults) && fn.Defaults[idx] != nil {
			// Use default value
			env.Set(param, fn.Defaults[idx])
		}
	}

	// Execute body
	var result runtime.Value = runtime.NULL
	if block, ok := fn.Body.(*ast.BlockStmt); ok {
		result = i.evalBlock(block)
	}

	// Restore environment
	i.env = oldEnv

	// Unwrap return value
	if ret, ok := result.(*runtime.ReturnValue); ok {
		return ret.Value
	}
	return result
}

func (i *Interpreter) evalArgsInEnv(env *runtime.Environment, args *ast.ArgumentList) []runtime.Value {
	if args == nil {
		return nil
	}
	oldEnv := i.env
	i.env = env
	var result []runtime.Value
	for _, arg := range args.Args {
		val := i.evalExpr(arg.Value)
		if arg.Unpack {
			// Spread operator: ...$array
			if arr, ok := val.(*runtime.Array); ok {
				for _, k := range arr.Keys {
					result = append(result, arr.Elements[k])
				}
			}
		} else {
			result = append(result, val)
		}
	}
	i.env = oldEnv
	return result
}

func (i *Interpreter) evalMethodCall(e *ast.MethodCallExpr) runtime.Value {
	obj := i.evalExpr(e.Object)
	// Null safe operator: return null if object is null
	if e.NullSafe {
		if _, isNull := obj.(*runtime.Null); isNull {
			return runtime.NULL
		}
	}
	objVal, ok := obj.(*runtime.Object)
	if !ok {
		// Check for magic __call
		return runtime.NewError("method call on non-object")
	}

	methodName := e.Method.(*ast.Ident).Name

	// Look up method in class hierarchy
	method, foundClass := i.findMethod(objVal.Class, methodName)
	if method == nil {
		// Check for __call magic method
		if callMethod, _ := i.findMethod(objVal.Class, "__call"); callMethod != nil {
			return i.callMagicCall(objVal, callMethod, methodName, e.Args)
		}
		return runtime.NewError(fmt.Sprintf("undefined method: %s::%s", objVal.Class.Name, methodName))
	}

	// Check visibility
	var callerClass *runtime.Class
	if i.currentClass != "" {
		callerClass, _ = i.env.GetClass(i.currentClass)
	}
	if !i.checkMethodVisibility(method, callerClass, foundClass) {
		visibility := "private"
		if method.IsProtected {
			visibility = "protected"
		}
		return runtime.NewError(fmt.Sprintf("cannot access %s method %s::%s", visibility, foundClass.Name, methodName))
	}

	// Create environment with $this
	env := runtime.NewEnclosedEnvironment(i.env)
	env.Set("this", objVal)

	oldEnv := i.env
	oldClass := i.currentClass
	oldThis := i.currentThis
	i.env = env
	i.currentClass = foundClass.Name
	i.currentThis = objVal

	// Bind parameters
	argVals := i.evalArgsInEnv(oldEnv, e.Args)
	numParams := len(method.Params)
	for idx, param := range method.Params {
		isVariadicParam := method.Variadic && idx == numParams-1

		if isVariadicParam {
			variadicArr := runtime.NewArray()
			for j := idx; j < len(argVals); j++ {
				variadicArr.Set(nil, argVals[j])
			}
			env.Set(param, variadicArr)
		} else if idx < len(argVals) {
			env.Set(param, argVals[idx])
		} else if method.Defaults != nil && idx < len(method.Defaults) && method.Defaults[idx] != nil {
			env.Set(param, method.Defaults[idx])
		}
	}

	// Execute body
	var result runtime.Value = runtime.NULL
	if block, ok := method.Body.(*ast.BlockStmt); ok {
		result = i.evalBlock(block)
	}

	// Restore environment
	i.env = oldEnv
	i.currentClass = oldClass
	i.currentThis = oldThis

	// Unwrap return value
	if ret, ok := result.(*runtime.ReturnValue); ok {
		return ret.Value
	}
	return result
}

// findMethod looks up a method in the class hierarchy
func (i *Interpreter) findMethod(class *runtime.Class, name string) (*runtime.Method, *runtime.Class) {
	if method, ok := class.Methods[name]; ok {
		return method, class
	}
	if class.Parent != nil {
		return i.findMethod(class.Parent, name)
	}
	return nil, nil
}

// checkMethodVisibility checks if a method is accessible from the current context
func (i *Interpreter) checkMethodVisibility(method *runtime.Method, callerClass, methodClass *runtime.Class) bool {
	if method.IsPublic {
		return true
	}
	if callerClass == nil {
		return false // No class context, can only access public
	}
	if method.IsPrivate {
		return callerClass.Name == methodClass.Name
	}
	if method.IsProtected {
		// Check if callerClass is same or subclass of methodClass
		c := callerClass
		for c != nil {
			if c.Name == methodClass.Name {
				return true
			}
			c = c.Parent
		}
		// Or if methodClass is subclass of callerClass
		c = methodClass
		for c != nil {
			if c.Name == callerClass.Name {
				return true
			}
			c = c.Parent
		}
	}
	return false
}

// checkPropertyVisibility checks if a property is accessible from the current context
func (i *Interpreter) checkPropertyVisibility(prop *runtime.PropertyDef, callerClass, propClass *runtime.Class) bool {
	if prop.IsPublic {
		return true
	}
	if callerClass == nil {
		return false // No class context, can only access public
	}
	if prop.IsPrivate {
		return callerClass.Name == propClass.Name
	}
	if prop.IsProtected {
		// Check if callerClass is same or subclass of propClass
		c := callerClass
		for c != nil {
			if c.Name == propClass.Name {
				return true
			}
			c = c.Parent
		}
		// Or if propClass is subclass of callerClass
		c = propClass
		for c != nil {
			if c.Name == callerClass.Name {
				return true
			}
			c = c.Parent
		}
	}
	return false
}

// callMagicCall invokes the __call magic method
func (i *Interpreter) callMagicCall(obj *runtime.Object, method *runtime.Method, name string, args *ast.ArgumentList) runtime.Value {
	env := runtime.NewEnclosedEnvironment(i.env)
	env.Set("this", obj)

	oldEnv := i.env
	i.env = env

	// __call receives method name and array of arguments
	argVals := i.evalArgsInEnv(oldEnv, args)
	argsArray := runtime.NewArray()
	for _, arg := range argVals {
		argsArray.Set(nil, arg)
	}

	if len(method.Params) >= 1 {
		env.Set(method.Params[0], runtime.NewString(name))
	}
	if len(method.Params) >= 2 {
		env.Set(method.Params[1], argsArray)
	}

	var result runtime.Value = runtime.NULL
	if block, ok := method.Body.(*ast.BlockStmt); ok {
		result = i.evalBlock(block)
	}

	i.env = oldEnv

	if ret, ok := result.(*runtime.ReturnValue); ok {
		return ret.Value
	}
	return result
}

// invokeMethod calls a method on an object (used for __invoke and similar)
func (i *Interpreter) invokeMethod(obj *runtime.Object, method *runtime.Method, foundClass *runtime.Class, args *ast.ArgumentList) runtime.Value {
	env := runtime.NewEnclosedEnvironment(i.env)
	env.Set("this", obj)

	oldEnv := i.env
	oldClass := i.currentClass
	oldThis := i.currentThis
	i.env = env
	i.currentClass = foundClass.Name
	i.currentThis = obj

	// Bind parameters
	argVals := i.evalArgsInEnv(oldEnv, args)
	numParams := len(method.Params)
	for idx, param := range method.Params {
		isVariadicParam := method.Variadic && idx == numParams-1
		if isVariadicParam {
			variadicArr := runtime.NewArray()
			for j := idx; j < len(argVals); j++ {
				variadicArr.Set(nil, argVals[j])
			}
			env.Set(param, variadicArr)
		} else if idx < len(argVals) {
			env.Set(param, argVals[idx])
		} else if method.Defaults != nil && idx < len(method.Defaults) && method.Defaults[idx] != nil {
			env.Set(param, method.Defaults[idx])
		}
	}

	var result runtime.Value = runtime.NULL
	if block, ok := method.Body.(*ast.BlockStmt); ok {
		result = i.evalBlock(block)
	}

	i.env = oldEnv
	i.currentClass = oldClass
	i.currentThis = oldThis

	if ret, ok := result.(*runtime.ReturnValue); ok {
		return ret.Value
	}
	return result
}

func (i *Interpreter) evalStaticCall(e *ast.StaticCallExpr) runtime.Value {
	var className string
	var isParentCall bool
	switch c := e.Class.(type) {
	case *ast.Ident:
		className = c.Name
		// Handle self/static/parent
		if className == "self" || className == "static" {
			className = i.currentClass
		} else if className == "parent" {
			isParentCall = true
			// Get parent class
			if i.currentClass == "" {
				return runtime.NewError("Cannot use 'parent' when not in a class")
			}
			currentClassObj, ok := i.env.GetClass(i.currentClass)
			if !ok || currentClassObj.Parent == nil {
				return runtime.NewError("Cannot use 'parent' - class has no parent")
			}
			className = currentClassObj.Parent.Name
		}
	default:
		className = i.evalExpr(c).ToString()
	}

	class, ok := i.env.GetClass(className)
	if !ok {
		return runtime.NewError(fmt.Sprintf("undefined class: %s", className))
	}

	methodName := e.Method.(*ast.Ident).Name
	method, ok := class.Methods[methodName]
	if !ok {
		return runtime.NewError(fmt.Sprintf("undefined static method: %s::%s", className, methodName))
	}

	// Create environment
	env := runtime.NewEnclosedEnvironment(i.env)
	oldEnv := i.env
	oldClass := i.currentClass
	i.env = env
	i.currentClass = className

	// For parent calls on non-static methods, pass $this
	if isParentCall && i.currentThis != nil {
		env.Set("this", i.currentThis)
	}

	// Bind parameters
	argVals := i.evalArgsInEnv(oldEnv, e.Args)
	for idx, param := range method.Params {
		if idx < len(argVals) {
			env.Set(param, argVals[idx])
		} else if method.Defaults != nil && idx < len(method.Defaults) && method.Defaults[idx] != nil {
			env.Set(param, method.Defaults[idx])
		}
	}

	// Execute body
	var result runtime.Value = runtime.NULL
	if block, ok := method.Body.(*ast.BlockStmt); ok {
		result = i.evalBlock(block)
	}

	i.env = oldEnv
	i.currentClass = oldClass

	if ret, ok := result.(*runtime.ReturnValue); ok {
		return ret.Value
	}
	return result
}

func (i *Interpreter) evalPropertyAccess(e *ast.PropertyFetchExpr) runtime.Value {
	obj := i.evalExpr(e.Object)
	// Null safe operator: return null if object is null
	if e.NullSafe {
		if _, isNull := obj.(*runtime.Null); isNull {
			return runtime.NULL
		}
	}
	if objVal, ok := obj.(*runtime.Object); ok {
		propName := e.Property.(*ast.Ident).Name

		// Check visibility for defined properties
		if propDef, exists := objVal.Class.Properties[propName]; exists {
			var callerClass *runtime.Class
			if i.currentClass != "" {
				callerClass, _ = i.env.GetClass(i.currentClass)
			}
			if !i.checkPropertyVisibility(propDef, callerClass, objVal.Class) {
				visibility := "private"
				if propDef.IsProtected {
					visibility = "protected"
				}
				return runtime.NewError(fmt.Sprintf("cannot access %s property %s::$%s", visibility, objVal.Class.Name, propName))
			}
		}

		// Check if property exists
		if val := objVal.GetProperty(propName); val != runtime.NULL {
			return val
		}

		// Check for __get magic method
		if method, _ := i.findMethod(objVal.Class, "__get"); method != nil {
			return i.callMagicGetSet(objVal, method, propName, nil)
		}

		return runtime.NULL
	}
	return runtime.NULL
}

// createToStringCallback creates a callback function for __toString
func (i *Interpreter) createToStringCallback() func(*runtime.Object) string {
	return func(obj *runtime.Object) string {
		method, _ := i.findMethod(obj.Class, "__toString")
		if method == nil {
			return fmt.Sprintf("Object(%s)", obj.Class.Name)
		}

		env := runtime.NewEnclosedEnvironment(i.env)
		env.Set("this", obj)

		oldEnv := i.env
		oldClass := i.currentClass
		oldThis := i.currentThis
		i.env = env
		i.currentClass = obj.Class.Name
		i.currentThis = obj

		var result runtime.Value = runtime.NULL
		if block, ok := method.Body.(*ast.BlockStmt); ok {
			result = i.evalBlock(block)
		}

		i.env = oldEnv
		i.currentClass = oldClass
		i.currentThis = oldThis

		if ret, ok := result.(*runtime.ReturnValue); ok {
			return ret.Value.ToString()
		}
		return result.ToString()
	}
}

// callMagicGetSet invokes __get or __set magic methods
func (i *Interpreter) callMagicGetSet(obj *runtime.Object, method *runtime.Method, propName string, value runtime.Value) runtime.Value {
	env := runtime.NewEnclosedEnvironment(i.env)
	env.Set("this", obj)

	oldEnv := i.env
	oldClass := i.currentClass
	oldThis := i.currentThis
	i.env = env
	i.currentClass = obj.Class.Name
	i.currentThis = obj

	// __get receives property name, __set receives name and value
	if len(method.Params) >= 1 {
		env.Set(method.Params[0], runtime.NewString(propName))
	}
	if value != nil && len(method.Params) >= 2 {
		env.Set(method.Params[1], value)
	}

	var result runtime.Value = runtime.NULL
	if block, ok := method.Body.(*ast.BlockStmt); ok {
		result = i.evalBlock(block)
	}

	i.env = oldEnv
	i.currentClass = oldClass
	i.currentThis = oldThis

	if ret, ok := result.(*runtime.ReturnValue); ok {
		return ret.Value
	}
	return result
}

func (i *Interpreter) evalStaticProperty(e *ast.StaticPropertyFetchExpr) runtime.Value {
	var className string
	switch c := e.Class.(type) {
	case *ast.Ident:
		className = c.Name
		// Handle self/static/parent
		if className == "self" || className == "static" {
			className = i.currentClass
		}
	default:
		className = i.evalExpr(c).ToString()
	}

	class, ok := i.env.GetClass(className)
	if !ok {
		return runtime.NewError(fmt.Sprintf("undefined class: %s", className))
	}

	propName := e.Property.(*ast.Variable).Name.(*ast.Ident).Name
	if val, ok := class.StaticProps[propName]; ok {
		return val
	}
	return runtime.NULL
}

func (i *Interpreter) evalArrayAccess(e *ast.ArrayAccessExpr) runtime.Value {
	arr := i.evalExpr(e.Array)
	if arrVal, ok := arr.(*runtime.Array); ok {
		if e.Index == nil {
			return runtime.NULL
		}
		key := i.evalExpr(e.Index)
		return arrVal.Get(key)
	}
	if strVal, ok := arr.(*runtime.String); ok {
		idx := i.evalExpr(e.Index).ToInt()
		if idx >= 0 && idx < int64(len(strVal.Value)) {
			return runtime.NewString(string(strVal.Value[idx]))
		}
		return runtime.NewString("")
	}
	return runtime.NULL
}

func (i *Interpreter) evalArray(e *ast.ArrayExpr) runtime.Value {
	arr := runtime.NewArray()
	for _, item := range e.Items {
		val := i.evalExpr(item.Value)
		if item.Unpack {
			// Spread operator: ...$array
			if srcArr, ok := val.(*runtime.Array); ok {
				for _, k := range srcArr.Keys {
					arr.Set(nil, srcArr.Elements[k])
				}
			}
		} else if item.Key != nil {
			key := i.evalExpr(item.Key)
			arr.Set(key, val)
		} else {
			arr.Set(nil, val)
		}
	}
	return arr
}

func (i *Interpreter) evalNew(e *ast.NewExpr) runtime.Value {
	var className string
	switch c := e.Class.(type) {
	case *ast.Ident:
		className = c.Name
	default:
		className = i.evalExpr(c).ToString()
	}

	// Special case for Exception
	if className == "Exception" {
		args := i.evalArgs(e.Args)
		msg := ""
		if len(args) > 0 {
			msg = args[0].ToString()
		}
		return &runtime.Exception{Message: msg}
	}

	class, ok := i.env.GetClass(className)
	if !ok {
		return runtime.NewError(fmt.Sprintf("undefined class: %s", className))
	}

	// Cannot instantiate abstract class
	if class.IsAbstract {
		return runtime.NewError(fmt.Sprintf("cannot instantiate abstract class %s", className))
	}

	obj := runtime.NewObject(class)

	// Set up __toString callback if method exists
	if _, hasToString := class.Methods["__toString"]; hasToString {
		obj.SetToStringCallback(i.createToStringCallback())
	}

	// Initialize properties with defaults
	for name, prop := range class.Properties {
		if prop.Default != nil {
			obj.SetProperty(name, prop.Default)
		}
	}

	// Call constructor if exists
	if constructor, ok := class.Methods["__construct"]; ok {
		env := runtime.NewEnclosedEnvironment(i.env)
		env.Set("this", obj)
		oldEnv := i.env
		i.env = env

		argVals := i.evalArgsInEnv(oldEnv, e.Args)
		for idx, param := range constructor.Params {
			if idx < len(argVals) {
				env.Set(param, argVals[idx])
			}
		}

		if block, ok := constructor.Body.(*ast.BlockStmt); ok {
			i.evalBlock(block)
		}

		i.env = oldEnv
	}

	return obj
}

func (i *Interpreter) evalClone(e *ast.CloneExpr) runtime.Value {
	obj := i.evalExpr(e.Expr)
	if objVal, ok := obj.(*runtime.Object); ok {
		clone := runtime.NewObject(objVal.Class)
		for k, v := range objVal.Properties {
			clone.Properties[k] = v
		}
		// Set up __toString callback if method exists
		if _, hasToString := objVal.Class.Methods["__toString"]; hasToString {
			clone.SetToStringCallback(i.createToStringCallback())
		}
		return clone
	}
	return runtime.NULL
}

func (i *Interpreter) evalClosure(e *ast.ClosureExpr) runtime.Value {
	params := make([]string, len(e.Params))
	for idx, p := range e.Params {
		params[idx] = p.Var.Name.(*ast.Ident).Name
	}

	// Create environment for closure
	closureEnv := runtime.NewEnclosedEnvironment(i.env)

	// Handle use clause - capture variables
	if len(e.Uses) > 0 {
		for _, use := range e.Uses {
			varName := use.Var.Name.(*ast.Ident).Name
			if val, ok := i.env.Get(varName); ok {
				closureEnv.Set(varName, val)
			}
		}
	}

	fn := &runtime.Function{
		Params: params,
		Body:   e.Body,
		Env:    closureEnv,
	}

	return fn
}

func (i *Interpreter) evalArrowFunc(e *ast.ArrowFuncExpr) runtime.Value {
	params := make([]string, len(e.Params))
	for idx, p := range e.Params {
		params[idx] = p.Var.Name.(*ast.Ident).Name
	}

	// Arrow functions capture outer scope automatically
	return &runtime.Function{
		Params: params,
		Body:   &ast.BlockStmt{Stmts: []ast.Stmt{&ast.ReturnStmt{Result: e.Body}}},
		Env:    i.env,
	}
}

func (i *Interpreter) evalMatch(e *ast.MatchExpr) runtime.Value {
	subject := i.evalExpr(e.Cond)

	for _, arm := range e.Arms {
		if arm.Conds == nil {
			// Default arm
			return i.evalExpr(arm.Body)
		}
		for _, cond := range arm.Conds {
			condVal := i.evalExpr(cond)
			if runtime.IsIdentical(subject, condVal) {
				return i.evalExpr(arm.Body)
			}
		}
	}

	return runtime.NewError("unhandled match case")
}

func (i *Interpreter) evalInstanceof(e *ast.InstanceofExpr) runtime.Value {
	obj := i.evalExpr(e.Expr)
	objVal, ok := obj.(*runtime.Object)
	if !ok {
		return runtime.FALSE
	}

	var className string
	switch c := e.Class.(type) {
	case *ast.Ident:
		className = c.Name
	default:
		className = i.evalExpr(c).ToString()
	}

	// Check class hierarchy
	class := objVal.Class
	for class != nil {
		if class.Name == className {
			return runtime.TRUE
		}
		// Check implemented interfaces
		for _, iface := range class.Interfaces {
			if iface.Name == className {
				return runtime.TRUE
			}
		}
		class = class.Parent
	}

	return runtime.FALSE
}

func (i *Interpreter) evalCast(e *ast.CastExpr) runtime.Value {
	val := i.evalExpr(e.X)

	switch e.Type {
	case token.T_INT_CAST:
		return runtime.NewInt(val.ToInt())
	case token.T_DOUBLE_CAST:
		return runtime.NewFloat(val.ToFloat())
	case token.T_STRING_CAST:
		return runtime.NewString(val.ToString())
	case token.T_BOOL_CAST:
		return runtime.NewBool(val.ToBool())
	case token.T_ARRAY_CAST:
		if arr, ok := val.(*runtime.Array); ok {
			return arr
		}
		arr := runtime.NewArray()
		arr.Set(runtime.NewInt(0), val)
		return arr
	case token.T_OBJECT_CAST:
		if obj, ok := val.(*runtime.Object); ok {
			return obj
		}
		// Create stdClass
		class := &runtime.Class{Name: "stdClass", Properties: make(map[string]*runtime.PropertyDef), Methods: make(map[string]*runtime.Method)}
		obj := runtime.NewObject(class)
		if arr, ok := val.(*runtime.Array); ok {
			for _, k := range arr.Keys {
				obj.SetProperty(k.ToString(), arr.Elements[k])
			}
		} else {
			obj.SetProperty("scalar", val)
		}
		return obj
	case token.T_UNSET_CAST:
		return runtime.NULL
	}

	return val
}

func (i *Interpreter) evalIsset(e *ast.IssetExpr) runtime.Value {
	for _, v := range e.Vars {
		if varExpr, ok := v.(*ast.Variable); ok {
			name := varExpr.Name.(*ast.Ident).Name
			if !i.env.Isset(name) {
				return runtime.FALSE
			}
		} else if propExpr, ok := v.(*ast.PropertyFetchExpr); ok {
			// Property access - check for __isset
			objVal := i.evalExpr(propExpr.Object)
			if obj, ok := objVal.(*runtime.Object); ok {
				propName := propExpr.Property.(*ast.Ident).Name
				// Check if property exists
				if _, exists := obj.Properties[propName]; exists {
					continue
				}
				// Check for __isset magic method
				if issetMethod, _ := i.findMethod(obj.Class, "__isset"); issetMethod != nil {
					result := i.callMagicGetSet(obj, issetMethod, propName, nil)
					if !result.ToBool() {
						return runtime.FALSE
					}
					continue
				}
				return runtime.FALSE
			}
			return runtime.FALSE
		} else {
			val := i.evalExpr(v)
			if _, ok := val.(*runtime.Null); ok {
				return runtime.FALSE
			}
		}
	}
	return runtime.TRUE
}

func (i *Interpreter) evalEmpty(e *ast.EmptyExpr) runtime.Value {
	val := i.evalExpr(e.Expr)
	return runtime.NewBool(!val.ToBool())
}

func (i *Interpreter) evalEncapsedString(e *ast.EncapsedStringExpr) runtime.Value {
	var sb strings.Builder
	for _, part := range e.Parts {
		val := i.evalExpr(part)
		sb.WriteString(val.ToString())
	}
	return runtime.NewString(sb.String())
}

func (i *Interpreter) evalConstantAccess(e *ast.ClassConstFetchExpr) runtime.Value {
	var className string
	switch c := e.Class.(type) {
	case *ast.Ident:
		className = c.Name
	default:
		className = i.evalExpr(c).ToString()
	}

	class, ok := i.env.GetClass(className)
	if !ok {
		return runtime.NewError(fmt.Sprintf("undefined class: %s", className))
	}

	constName := e.Const.Name
	if val, ok := class.Constants[constName]; ok {
		return val
	}

	return runtime.NewError(fmt.Sprintf("undefined class constant: %s::%s", className, constName))
}

// ----------------------------------------------------------------------------
// Declaration evaluation

func (i *Interpreter) evalFunctionDecl(s *ast.FunctionDecl) runtime.Value {
	params := make([]string, len(s.Params))
	defaults := make([]runtime.Value, len(s.Params))
	variadic := false
	for idx, p := range s.Params {
		params[idx] = p.Var.Name.(*ast.Ident).Name
		if p.Default != nil {
			defaults[idx] = i.evalExpr(p.Default)
		}
		if p.Variadic {
			variadic = true
		}
	}

	fn := &runtime.Function{
		Name:     s.Name.Name,
		Params:   params,
		Defaults: defaults,
		Variadic: variadic,
		Body:     s.Body,
		Env:      i.env,
	}

	i.env.DefineFunction(s.Name.Name, fn)
	return runtime.NULL
}

func (i *Interpreter) evalClassDecl(s *ast.ClassDecl) runtime.Value {
	class := &runtime.Class{
		Name:        s.Name.Name,
		Properties:  make(map[string]*runtime.PropertyDef),
		StaticProps: make(map[string]runtime.Value),
		Methods:     make(map[string]*runtime.Method),
		Constants:   make(map[string]runtime.Value),
		IsAbstract:  s.Modifiers != nil && s.Modifiers.Abstract,
		IsFinal:     s.Modifiers != nil && s.Modifiers.Final,
	}

	// Handle extends
	if s.Extends != nil {
		parentName := s.Extends.(*ast.Ident).Name
		if parent, ok := i.env.GetClass(parentName); ok {
			// Cannot extend final class
			if parent.IsFinal {
				return runtime.NewError(fmt.Sprintf("cannot extend final class %s", parentName))
			}
			class.Parent = parent
			// Copy parent methods
			for name, method := range parent.Methods {
				class.Methods[name] = method
			}
			// Copy parent properties
			for name, prop := range parent.Properties {
				class.Properties[name] = prop
			}
		}
	}

	// Handle implements
	for _, impl := range s.Implements {
		var ifaceName string
		switch iface := impl.(type) {
		case *ast.Ident:
			ifaceName = iface.Name
		default:
			ifaceName = i.evalExpr(iface).ToString()
		}
		if iface, ok := i.env.GetInterface(ifaceName); ok {
			class.Interfaces = append(class.Interfaces, iface)
		}
	}

	// Process members - first pass for trait uses
	for _, member := range s.Members {
		if traitUse, ok := member.(*ast.TraitUseDecl); ok {
			for _, traitExpr := range traitUse.Traits {
				var traitName string
				switch t := traitExpr.(type) {
				case *ast.Ident:
					traitName = t.Name
				default:
					traitName = i.evalExpr(t).ToString()
				}

				trait, ok := i.env.GetTrait(traitName)
				if !ok {
					continue // Skip unknown traits
				}

				// Copy trait methods to class
				for name, method := range trait.Methods {
					// Check for alias/insteadof adaptations
					aliasName := name
					shouldInclude := true
					for _, adaptation := range traitUse.Adaptations {
						if adaptation.Method != nil && adaptation.Method.Name == name {
							if adaptation.Insteadof != nil {
								shouldInclude = true
							}
							if adaptation.Alias != nil {
								aliasName = adaptation.Alias.Name
							}
						}
					}
					if shouldInclude {
						class.Methods[aliasName] = method
					}
				}

				// Copy trait properties to class
				for name, prop := range trait.Properties {
					if _, exists := class.Properties[name]; !exists {
						class.Properties[name] = prop
					}
				}
			}
		}
	}

	// Process members - second pass for regular members
	for _, member := range s.Members {
		switch m := member.(type) {
		case *ast.TraitUseDecl:
			// Already handled above
			continue
		case *ast.PropertyDecl:
			for _, prop := range m.Props {
				propName := prop.Var.Name.(*ast.Ident).Name
				isStatic := m.Modifiers != nil && m.Modifiers.Static
				propDef := &runtime.PropertyDef{
					Name:        propName,
					IsPublic:    m.Modifiers == nil || m.Modifiers.Public,
					IsProtected: m.Modifiers != nil && m.Modifiers.Protected,
					IsPrivate:   m.Modifiers != nil && m.Modifiers.Private,
					IsStatic:    isStatic,
					IsReadonly:  m.Modifiers != nil && m.Modifiers.Readonly,
				}
				if prop.Default != nil {
					propDef.Default = i.evalExpr(prop.Default)
				}
				class.Properties[propName] = propDef
				// Initialize static properties
				if isStatic {
					if propDef.Default != nil {
						class.StaticProps[propName] = propDef.Default
					} else {
						class.StaticProps[propName] = runtime.NULL
					}
				}
			}

		case *ast.MethodDecl:
			// Check if overriding a final method
			if existingMethod, exists := class.Methods[m.Name.Name]; exists {
				if existingMethod.IsFinal {
					return runtime.NewError(fmt.Sprintf("cannot override final method %s::%s", class.Name, m.Name.Name))
				}
			}
			params := make([]string, len(m.Params))
			defaults := make([]runtime.Value, len(m.Params))
			variadic := false
			for idx, p := range m.Params {
				params[idx] = p.Var.Name.(*ast.Ident).Name
				if p.Default != nil {
					defaults[idx] = i.evalExpr(p.Default)
				}
				if p.Variadic {
					variadic = true
				}
			}
			method := &runtime.Method{
				Name:        m.Name.Name,
				Params:      params,
				Defaults:    defaults,
				Variadic:    variadic,
				Body:        m.Body,
				IsPublic:    m.Modifiers == nil || m.Modifiers.Public,
				IsProtected: m.Modifiers != nil && m.Modifiers.Protected,
				IsPrivate:   m.Modifiers != nil && m.Modifiers.Private,
				IsStatic:    m.Modifiers != nil && m.Modifiers.Static,
				IsAbstract:  m.Modifiers != nil && m.Modifiers.Abstract,
				IsFinal:     m.Modifiers != nil && m.Modifiers.Final,
			}
			class.Methods[m.Name.Name] = method

		case *ast.ClassConstDecl:
			for _, c := range m.Consts {
				class.Constants[c.Name.Name] = i.evalExpr(c.Value)
			}
		}
	}

	// Verify all abstract methods are implemented (for non-abstract classes)
	if !class.IsAbstract {
		// Check parent abstract methods
		if class.Parent != nil {
			if err := i.checkAbstractMethods(class, class.Parent); err != nil {
				return err
			}
		}
		// Check interface methods
		for _, iface := range class.Interfaces {
			for methodName := range iface.Methods {
				if method, exists := class.Methods[methodName]; !exists || method.IsAbstract {
					return runtime.NewError(fmt.Sprintf("class %s must implement method %s::%s", class.Name, iface.Name, methodName))
				}
			}
		}
	}

	i.env.DefineClass(s.Name.Name, class)
	return runtime.NULL
}

// checkAbstractMethods checks that all abstract methods from parent are implemented
func (i *Interpreter) checkAbstractMethods(class, parent *runtime.Class) runtime.Value {
	for name, method := range parent.Methods {
		if method.IsAbstract {
			if impl, exists := class.Methods[name]; !exists || impl.IsAbstract {
				return runtime.NewError(fmt.Sprintf("class %s must implement abstract method %s::%s", class.Name, parent.Name, name))
			}
		}
	}
	// Check parent's parent
	if parent.Parent != nil {
		return i.checkAbstractMethods(class, parent.Parent)
	}
	return nil
}

func (i *Interpreter) evalInterfaceDecl(s *ast.InterfaceDecl) runtime.Value {
	iface := &runtime.Interface{
		Name:    s.Name.Name,
		Methods: make(map[string]*runtime.Method),
	}

	// Process members (interface methods are all abstract/public)
	for _, member := range s.Members {
		if m, ok := member.(*ast.MethodDecl); ok {
			params := make([]string, len(m.Params))
			for idx, p := range m.Params {
				params[idx] = p.Var.Name.(*ast.Ident).Name
			}
			method := &runtime.Method{
				Name:       m.Name.Name,
				Params:     params,
				Body:       nil, // Interface methods have no body
				IsPublic:   true,
				IsAbstract: true,
			}
			iface.Methods[m.Name.Name] = method
		}
	}

	i.env.DefineInterface(s.Name.Name, iface)
	return runtime.NULL
}

func (i *Interpreter) evalTraitDecl(s *ast.TraitDecl) runtime.Value {
	trait := &runtime.Trait{
		Name:       s.Name.Name,
		Properties: make(map[string]*runtime.PropertyDef),
		Methods:    make(map[string]*runtime.Method),
	}

	// Process members
	for _, member := range s.Members {
		switch m := member.(type) {
		case *ast.PropertyDecl:
			for _, prop := range m.Props {
				propName := prop.Var.Name.(*ast.Ident).Name
				propDef := &runtime.PropertyDef{
					Name:       propName,
					IsPublic:   m.Modifiers == nil || m.Modifiers.Public,
					IsProtected: m.Modifiers != nil && m.Modifiers.Protected,
					IsPrivate:  m.Modifiers != nil && m.Modifiers.Private,
					IsStatic:   m.Modifiers != nil && m.Modifiers.Static,
				}
				if prop.Default != nil {
					propDef.Default = i.evalExpr(prop.Default)
				}
				trait.Properties[propName] = propDef
			}

		case *ast.MethodDecl:
			params := make([]string, len(m.Params))
			defaults := make([]runtime.Value, len(m.Params))
			for idx, p := range m.Params {
				params[idx] = p.Var.Name.(*ast.Ident).Name
				if p.Default != nil {
					defaults[idx] = i.evalExpr(p.Default)
				}
			}
			method := &runtime.Method{
				Name:       m.Name.Name,
				Params:     params,
				Defaults:   defaults,
				Body:       m.Body,
				IsPublic:   m.Modifiers == nil || m.Modifiers.Public,
				IsProtected: m.Modifiers != nil && m.Modifiers.Protected,
				IsPrivate:  m.Modifiers != nil && m.Modifiers.Private,
				IsStatic:   m.Modifiers != nil && m.Modifiers.Static,
				IsAbstract: m.Modifiers != nil && m.Modifiers.Abstract,
				IsFinal:    m.Modifiers != nil && m.Modifiers.Final,
			}
			trait.Methods[m.Name.Name] = method
		}
	}

	i.env.DefineTrait(s.Name.Name, trait)
	return runtime.NULL
}

func (i *Interpreter) evalEnumDecl(s *ast.EnumDecl) runtime.Value {
	// Enums are similar to classes
	class := &runtime.Class{
		Name:       s.Name.Name,
		Properties: make(map[string]*runtime.PropertyDef),
		Methods:    make(map[string]*runtime.Method),
		Constants:  make(map[string]runtime.Value),
	}

	// Process enum cases
	for _, member := range s.Members {
		if caseDecl, ok := member.(*ast.EnumCaseDecl); ok {
			// Each case is a constant
			obj := runtime.NewObject(class)
			obj.SetProperty("name", runtime.NewString(caseDecl.Name.Name))
			if caseDecl.Value != nil {
				obj.SetProperty("value", i.evalExpr(caseDecl.Value))
			}
			class.Constants[caseDecl.Name.Name] = obj
		}
	}

	i.env.DefineClass(s.Name.Name, class)
	return runtime.NULL
}

func (i *Interpreter) evalConstDecl(s *ast.ConstDecl) runtime.Value {
	for _, c := range s.Consts {
		val := i.evalExpr(c.Value)
		i.env.DefineConstant(c.Name.Name, val)
	}
	return runtime.NULL
}

// evalInclude handles include, include_once, require, require_once
func (i *Interpreter) evalInclude(e *ast.IncludeExpr) runtime.Value {
	pathVal := i.evalExpr(e.Expr)
	path := pathVal.ToString()

	// Resolve relative paths
	if !filepath.IsAbs(path) {
		path = filepath.Join(i.currentDir, path)
	}

	// Get absolute path for tracking
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	// Check for _once variants
	isOnce := e.Type == token.T_INCLUDE_ONCE || e.Type == token.T_REQUIRE_ONCE
	if isOnce {
		if i.includedFiles[absPath] {
			return runtime.TRUE // Already included
		}
	}

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		if e.Type == token.T_REQUIRE || e.Type == token.T_REQUIRE_ONCE {
			return runtime.NewError(fmt.Sprintf("require: failed to open '%s'", path))
		}
		// include just returns false on failure
		return runtime.FALSE
	}

	// Mark as included for _once variants
	if isOnce {
		i.includedFiles[absPath] = true
	}

	// Save current directory and set to included file's directory
	oldDir := i.currentDir
	i.currentDir = filepath.Dir(absPath)

	// Parse and execute
	file := parser.ParseString(string(content))
	result := i.evalFile(file)

	// Restore directory
	i.currentDir = oldDir

	return result
}
