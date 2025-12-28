package runtime

import (
	"fmt"
	"strconv"
	"strings"
)

// Value represents a PHP runtime value.
type Value interface {
	Type() string
	// Type coercion
	ToBool() bool
	ToInt() int64
	ToFloat() float64
	ToString() string
	// For debugging
	Inspect() string
}

// ----------------------------------------------------------------------------
// Null

type Null struct{}

var NULL = &Null{}

func (n *Null) Type() string    { return "NULL" }
func (n *Null) ToBool() bool    { return false }
func (n *Null) ToInt() int64    { return 0 }
func (n *Null) ToFloat() float64 { return 0.0 }
func (n *Null) ToString() string { return "" }
func (n *Null) Inspect() string { return "NULL" }

// ----------------------------------------------------------------------------
// Bool

type Bool struct {
	Value bool
}

var TRUE = &Bool{Value: true}
var FALSE = &Bool{Value: false}

func NewBool(v bool) *Bool {
	if v {
		return TRUE
	}
	return FALSE
}

func (b *Bool) Type() string { return "boolean" }
func (b *Bool) ToBool() bool { return b.Value }
func (b *Bool) ToInt() int64 {
	if b.Value {
		return 1
	}
	return 0
}
func (b *Bool) ToFloat() float64 {
	if b.Value {
		return 1.0
	}
	return 0.0
}
func (b *Bool) ToString() string {
	if b.Value {
		return "1"
	}
	return ""
}
func (b *Bool) Inspect() string {
	if b.Value {
		return "bool(true)"
	}
	return "bool(false)"
}

// ----------------------------------------------------------------------------
// Int

type Int struct {
	Value int64
}

func NewInt(v int64) *Int {
	return &Int{Value: v}
}

func (i *Int) Type() string     { return "integer" }
func (i *Int) ToBool() bool     { return i.Value != 0 }
func (i *Int) ToInt() int64     { return i.Value }
func (i *Int) ToFloat() float64 { return float64(i.Value) }
func (i *Int) ToString() string { return strconv.FormatInt(i.Value, 10) }
func (i *Int) Inspect() string  { return fmt.Sprintf("int(%d)", i.Value) }

// ----------------------------------------------------------------------------
// Float

type Float struct {
	Value float64
}

func NewFloat(v float64) *Float {
	return &Float{Value: v}
}

func (f *Float) Type() string { return "double" }
func (f *Float) ToBool() bool { return f.Value != 0.0 }
func (f *Float) ToInt() int64 { return int64(f.Value) }
func (f *Float) ToFloat() float64 { return f.Value }
func (f *Float) ToString() string {
	s := strconv.FormatFloat(f.Value, 'G', -1, 64)
	// PHP always shows decimal point for floats
	if !strings.Contains(s, ".") && !strings.Contains(s, "E") {
		s += ".0"
	}
	return s
}
func (f *Float) Inspect() string { return fmt.Sprintf("float(%s)", f.ToString()) }

// ----------------------------------------------------------------------------
// String

type String struct {
	Value string
}

func NewString(v string) *String {
	return &String{Value: v}
}

func (s *String) Type() string { return "string" }
func (s *String) ToBool() bool { return s.Value != "" && s.Value != "0" }
func (s *String) ToInt() int64 {
	// PHP string to int conversion: parse leading numeric part
	v, _ := strconv.ParseInt(strings.TrimSpace(s.Value), 10, 64)
	return v
}
func (s *String) ToFloat() float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s.Value), 64)
	return v
}
func (s *String) ToString() string { return s.Value }
func (s *String) Inspect() string  { return fmt.Sprintf("string(%d) %q", len(s.Value), s.Value) }

// ----------------------------------------------------------------------------
// Array

type Array struct {
	Elements map[Value]Value
	Keys     []Value // Maintain insertion order
	NextIndex int64  // For auto-indexing
}

func NewArray() *Array {
	return &Array{
		Elements:  make(map[Value]Value),
		Keys:      make([]Value, 0),
		NextIndex: 0,
	}
}

func (a *Array) Type() string { return "array" }
func (a *Array) ToBool() bool { return len(a.Elements) > 0 }
func (a *Array) ToInt() int64 {
	if len(a.Elements) > 0 {
		return 1
	}
	return 0
}
func (a *Array) ToFloat() float64 { return float64(a.ToInt()) }
func (a *Array) ToString() string { return "Array" }
func (a *Array) Inspect() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("array(%d) {\n", len(a.Elements)))
	for _, k := range a.Keys {
		v := a.Elements[k]
		sb.WriteString(fmt.Sprintf("  [%s] => %s\n", keyString(k), v.Inspect()))
	}
	sb.WriteString("}")
	return sb.String()
}

func (a *Array) Get(key Value) Value {
	// Direct lookup first
	if v, ok := a.Elements[key]; ok {
		return v
	}
	// Value-based lookup for Int and String keys
	for k, v := range a.Elements {
		if keysEqual(key, k) {
			return v
		}
	}
	return NULL
}

// keysEqual compares array keys by value
func keysEqual(a, b Value) bool {
	switch av := a.(type) {
	case *Int:
		if bv, ok := b.(*Int); ok {
			return av.Value == bv.Value
		}
	case *String:
		if bv, ok := b.(*String); ok {
			return av.Value == bv.Value
		}
	}
	return a == b
}

func (a *Array) Set(key Value, val Value) {
	if key == nil {
		// Auto-index
		key = NewInt(a.NextIndex)
		a.NextIndex++
	}

	// Check if key already exists (by value)
	existingKey := a.findKey(key)
	if existingKey != nil {
		a.Elements[existingKey] = val
	} else {
		a.Keys = append(a.Keys, key)
		a.Elements[key] = val
	}

	// Update NextIndex if integer key
	if intKey, ok := key.(*Int); ok {
		if intKey.Value >= a.NextIndex {
			a.NextIndex = intKey.Value + 1
		}
	}
}

// findKey finds an existing key that matches by value
func (a *Array) findKey(key Value) Value {
	for _, k := range a.Keys {
		if keysEqual(key, k) {
			return k
		}
	}
	return nil
}

func (a *Array) Len() int {
	return len(a.Elements)
}

// Unset removes an element from the array by key.
func (a *Array) Unset(key Value) {
	existingKey := a.findKey(key)
	if existingKey == nil {
		return
	}
	delete(a.Elements, existingKey)
	// Remove from keys slice
	for idx, k := range a.Keys {
		if keysEqual(k, existingKey) {
			a.Keys = append(a.Keys[:idx], a.Keys[idx+1:]...)
			break
		}
	}
}

func keyString(k Value) string {
	switch v := k.(type) {
	case *Int:
		return strconv.FormatInt(v.Value, 10)
	case *String:
		return fmt.Sprintf("%q", v.Value)
	default:
		return v.Inspect()
	}
}

// ----------------------------------------------------------------------------
// Object

type Object struct {
	Class      *Class
	Properties map[string]Value
	toStringFn func(*Object) string // Callback for __toString, set by interpreter
}

func NewObject(class *Class) *Object {
	return &Object{
		Class:      class,
		Properties: make(map[string]Value),
	}
}

func (o *Object) Type() string { return "object" }
func (o *Object) ToBool() bool { return true }
func (o *Object) ToInt() int64 { return 1 }
func (o *Object) ToFloat() float64 { return 1.0 }
func (o *Object) ToString() string {
	// Check for __toString method via callback
	if o.toStringFn != nil {
		return o.toStringFn(o)
	}
	return fmt.Sprintf("Object(%s)", o.Class.Name)
}

// SetToStringCallback sets the callback for __toString
func (o *Object) SetToStringCallback(fn func(*Object) string) {
	o.toStringFn = fn
}
func (o *Object) Inspect() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("object(%s)#%p {\n", o.Class.Name, o))
	for k, v := range o.Properties {
		sb.WriteString(fmt.Sprintf("  [%q] => %s\n", k, v.Inspect()))
	}
	sb.WriteString("}")
	return sb.String()
}

func (o *Object) GetProperty(name string) Value {
	if v, ok := o.Properties[name]; ok {
		return v
	}
	return NULL
}

func (o *Object) SetProperty(name string, val Value) {
	o.Properties[name] = val
}

// ----------------------------------------------------------------------------
// Class (for object creation)

// AttributeInstance represents an applied attribute
type AttributeInstance struct {
	Name      string  // Fully qualified attribute class name
	Arguments []Value // Constructor arguments
}

type Class struct {
	Name        string
	Parent      *Class
	Interfaces  []*Interface
	Properties  map[string]*PropertyDef
	StaticProps map[string]Value // Runtime values for static properties
	Methods     map[string]*Method
	Constants   map[string]Value
	IsAbstract  bool
	IsFinal     bool
	Attributes  []*AttributeInstance
}

type PropertyDef struct {
	Name       string
	Default    Value
	IsPublic   bool
	IsProtected bool
	IsPrivate  bool
	IsStatic   bool
	IsReadonly bool
	Attributes []*AttributeInstance
}

// PromotedParam represents a constructor property promotion
type PromotedParam struct {
	Name        string
	IsPublic    bool
	IsProtected bool
	IsPrivate   bool
	Readonly    bool
}

type Method struct {
	Name           string
	Params         []string
	ParamTypes     []string // Type hints for parameters (empty string = no type)
	ParamNullable  []bool   // Whether parameter allows null
	Defaults       []Value  // Default values for parameters
	Variadic       bool     // Last param is variadic (...$args)
	PromotedParams []PromotedParam
	Body           interface{} // Will be *ast.BlockStmt
	IsPublic       bool
	IsProtected    bool
	IsPrivate      bool
	IsStatic       bool
	IsAbstract     bool
	IsFinal        bool
	ReturnType     string // Return type hint
	ReturnNullable bool   // Whether return allows null
	Attributes     []*AttributeInstance
}

type Interface struct {
	Name    string
	Methods map[string]*Method
}

// Trait represents a PHP trait
type Trait struct {
	Name       string
	Properties map[string]*PropertyDef
	Methods    map[string]*Method
}

// ----------------------------------------------------------------------------
// Function

type Function struct {
	Name           string
	Params         []string
	ParamTypes     []string // Type hints for parameters (empty string = no type)
	ParamNullable  []bool   // Whether parameter allows null
	Defaults       []Value  // Default values for each parameter (nil if no default)
	Variadic       bool     // Last param is variadic (...$args)
	IsGenerator    bool     // Function contains yield
	Body           interface{} // *ast.BlockStmt
	Env            *Environment
	ReturnType     string // Return type hint
	ReturnNullable bool   // Whether return allows null
	Attributes     []*AttributeInstance
}

func (f *Function) Type() string    { return "object" } // Closure is an object in PHP
func (f *Function) ToBool() bool    { return true }
func (f *Function) ToInt() int64    { return 1 }
func (f *Function) ToFloat() float64 { return 1.0 }
func (f *Function) ToString() string { return "" } // Cannot convert closure to string
func (f *Function) Inspect() string {
	if f.Name != "" {
		return fmt.Sprintf("function %s(%s)", f.Name, strings.Join(f.Params, ", "))
	}
	return fmt.Sprintf("Closure(%s)", strings.Join(f.Params, ", "))
}

// ----------------------------------------------------------------------------
// Builtin function

type BuiltinFunc func(args ...Value) Value

type Builtin struct {
	Name string
	Fn   BuiltinFunc
}

func (b *Builtin) Type() string     { return "object" }
func (b *Builtin) ToBool() bool     { return true }
func (b *Builtin) ToInt() int64     { return 1 }
func (b *Builtin) ToFloat() float64 { return 1.0 }
func (b *Builtin) ToString() string { return "" }
func (b *Builtin) Inspect() string  { return fmt.Sprintf("builtin(%s)", b.Name) }

// ----------------------------------------------------------------------------
// Reference (for pass-by-reference)

type Reference struct {
	Value *Value
}

func NewReference(v *Value) *Reference {
	return &Reference{Value: v}
}

func (r *Reference) Type() string     { return (*r.Value).Type() }
func (r *Reference) ToBool() bool     { return (*r.Value).ToBool() }
func (r *Reference) ToInt() int64     { return (*r.Value).ToInt() }
func (r *Reference) ToFloat() float64 { return (*r.Value).ToFloat() }
func (r *Reference) ToString() string { return (*r.Value).ToString() }
func (r *Reference) Inspect() string  { return fmt.Sprintf("&%s", (*r.Value).Inspect()) }

// Dereference returns the actual value
func (r *Reference) Deref() Value {
	return *r.Value
}

// Set updates the referenced value
func (r *Reference) Set(v Value) {
	*r.Value = v
}

// ----------------------------------------------------------------------------
// Return value (for control flow)

type ReturnValue struct {
	Value Value
}

func (r *ReturnValue) Type() string     { return r.Value.Type() }
func (r *ReturnValue) ToBool() bool     { return r.Value.ToBool() }
func (r *ReturnValue) ToInt() int64     { return r.Value.ToInt() }
func (r *ReturnValue) ToFloat() float64 { return r.Value.ToFloat() }
func (r *ReturnValue) ToString() string { return r.Value.ToString() }
func (r *ReturnValue) Inspect() string  { return r.Value.Inspect() }

// ----------------------------------------------------------------------------
// Break/Continue (for loop control)

type Break struct {
	Levels int
}

func (b *Break) Type() string     { return "break" }
func (b *Break) ToBool() bool     { return false }
func (b *Break) ToInt() int64     { return 0 }
func (b *Break) ToFloat() float64 { return 0 }
func (b *Break) ToString() string { return "" }
func (b *Break) Inspect() string  { return fmt.Sprintf("break(%d)", b.Levels) }

type Continue struct {
	Levels int
}

func (c *Continue) Type() string     { return "continue" }
func (c *Continue) ToBool() bool     { return false }
func (c *Continue) ToInt() int64     { return 0 }
func (c *Continue) ToFloat() float64 { return 0 }
func (c *Continue) ToString() string { return "" }
func (c *Continue) Inspect() string  { return fmt.Sprintf("continue(%d)", c.Levels) }

// ----------------------------------------------------------------------------
// Exit (for exit/die)

type Exit struct {
	Status int
	Message string
}

func (e *Exit) Type() string     { return "exit" }
func (e *Exit) ToBool() bool     { return false }
func (e *Exit) ToInt() int64     { return int64(e.Status) }
func (e *Exit) ToFloat() float64 { return float64(e.Status) }
func (e *Exit) ToString() string { return e.Message }
func (e *Exit) Inspect() string  { return fmt.Sprintf("exit(%d)", e.Status) }

// ----------------------------------------------------------------------------
// Generator

type Generator struct {
	Keys     []Value
	Values   []Value
	Position int
}

func NewGenerator() *Generator {
	return &Generator{Position: 0}
}

func (g *Generator) Type() string     { return "Generator" }
func (g *Generator) ToBool() bool     { return true }
func (g *Generator) ToInt() int64     { return 0 }
func (g *Generator) ToFloat() float64 { return 0 }
func (g *Generator) ToString() string { return "Generator" }
func (g *Generator) Inspect() string  { return "Generator" }

func (g *Generator) Add(key, value Value) {
	g.Keys = append(g.Keys, key)
	g.Values = append(g.Values, value)
}

func (g *Generator) Valid() bool {
	return g.Position < len(g.Values)
}

func (g *Generator) Current() Value {
	if g.Position < len(g.Values) {
		return g.Values[g.Position]
	}
	return NULL
}

func (g *Generator) Key() Value {
	if g.Position < len(g.Keys) {
		return g.Keys[g.Position]
	}
	return NULL
}

func (g *Generator) Next() {
	g.Position++
}

func (g *Generator) Rewind() {
	g.Position = 0
}

// Yield is a signal value returned when yield is encountered
type Yield struct {
	Key   Value
	Value Value
}

func (y *Yield) Type() string     { return "yield" }
func (y *Yield) ToBool() bool     { return false }
func (y *Yield) ToInt() int64     { return 0 }
func (y *Yield) ToFloat() float64 { return 0 }
func (y *Yield) ToString() string { return "" }
func (y *Yield) Inspect() string  { return "yield" }

// ----------------------------------------------------------------------------
// Error

type Error struct {
	Message string
	Level   string // E_ERROR, E_WARNING, E_NOTICE, etc.
}

func NewError(msg string) *Error {
	return &Error{Message: msg, Level: "E_ERROR"}
}

func (e *Error) Type() string     { return "error" }
func (e *Error) ToBool() bool     { return false }
func (e *Error) ToInt() int64     { return 0 }
func (e *Error) ToFloat() float64 { return 0 }
func (e *Error) ToString() string { return e.Message }
func (e *Error) Inspect() string  { return fmt.Sprintf("%s: %s", e.Level, e.Message) }

// ----------------------------------------------------------------------------
// Resource (for file handles, etc.)

type Resource struct {
	ResType string      // "stream", "curl", etc.
	Handle  interface{} // Actual resource (e.g., *os.File)
	ID      int64       // Resource ID
}

func NewResource(resType string, handle interface{}, id int64) *Resource {
	return &Resource{ResType: resType, Handle: handle, ID: id}
}

func (r *Resource) Type() string     { return "resource" }
func (r *Resource) ToBool() bool     { return true }
func (r *Resource) ToInt() int64     { return r.ID }
func (r *Resource) ToFloat() float64 { return float64(r.ID) }
func (r *Resource) ToString() string { return fmt.Sprintf("Resource id #%d", r.ID) }
func (r *Resource) Inspect() string  { return fmt.Sprintf("resource(%s) #%d", r.ResType, r.ID) }

// ----------------------------------------------------------------------------
// Exception

type Exception struct {
	Class    *Class
	Message  string
	Code     int64
	Previous *Exception
}

func NewException(msg string) *Exception {
	return &Exception{Message: msg}
}

func (e *Exception) Type() string     { return "object" }
func (e *Exception) ToBool() bool     { return true }
func (e *Exception) ToInt() int64     { return 0 }
func (e *Exception) ToFloat() float64 { return 0 }
func (e *Exception) ToString() string { return e.Message }
func (e *Exception) Inspect() string {
	className := "Exception"
	if e.Class != nil {
		className = e.Class.Name
	}
	return fmt.Sprintf("%s: %s", className, e.Message)
}

// ----------------------------------------------------------------------------
// Helper functions

// IsEqual compares two values using PHP's == semantics (type juggling)
func IsEqual(a, b Value) bool {
	// NULL comparisons
	if _, ok := a.(*Null); ok {
		return !b.ToBool()
	}
	if _, ok := b.(*Null); ok {
		return !a.ToBool()
	}

	// Same types - compare directly
	switch av := a.(type) {
	case *Bool:
		return av.Value == b.ToBool()
	case *Int:
		switch bv := b.(type) {
		case *Int:
			return av.Value == bv.Value
		case *Float:
			return float64(av.Value) == bv.Value
		case *String:
			return av.Value == b.ToInt()
		default:
			return av.Value == b.ToInt()
		}
	case *Float:
		return av.Value == b.ToFloat()
	case *String:
		switch bv := b.(type) {
		case *String:
			return av.Value == bv.Value
		default:
			return av.Value == b.ToString()
		}
	case *Array:
		bArr, ok := b.(*Array)
		if !ok {
			return false
		}
		if len(av.Elements) != len(bArr.Elements) {
			return false
		}
		for k, v := range av.Elements {
			if bv, exists := bArr.Elements[k]; !exists || !IsEqual(v, bv) {
				return false
			}
		}
		return true
	case *Object:
		bObj, ok := b.(*Object)
		if !ok {
			return false
		}
		return av == bObj // Same instance
	}
	return false
}

// IsIdentical compares two values using PHP's === semantics (no type juggling)
func IsIdentical(a, b Value) bool {
	if a.Type() != b.Type() {
		return false
	}

	switch av := a.(type) {
	case *Null:
		_, ok := b.(*Null)
		return ok
	case *Bool:
		bv, ok := b.(*Bool)
		return ok && av.Value == bv.Value
	case *Int:
		bv, ok := b.(*Int)
		return ok && av.Value == bv.Value
	case *Float:
		bv, ok := b.(*Float)
		return ok && av.Value == bv.Value
	case *String:
		bv, ok := b.(*String)
		return ok && av.Value == bv.Value
	case *Array:
		bv, ok := b.(*Array)
		if !ok || len(av.Elements) != len(bv.Elements) {
			return false
		}
		for i, k := range av.Keys {
			if i >= len(bv.Keys) || !IsIdentical(k, bv.Keys[i]) {
				return false
			}
			if !IsIdentical(av.Elements[k], bv.Elements[bv.Keys[i]]) {
				return false
			}
		}
		return true
	case *Object:
		bv, ok := b.(*Object)
		return ok && av == bv // Same instance
	}
	return false
}

// Compare returns -1, 0, or 1 for spaceship operator
func Compare(a, b Value) int {
	af := a.ToFloat()
	bf := b.ToFloat()
	if af < bf {
		return -1
	}
	if af > bf {
		return 1
	}
	return 0
}
