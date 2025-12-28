package runtime

// Environment represents a variable scope.
type Environment struct {
	store      map[string]Value
	outer      *Environment
	functions  map[string]*Function
	classes    map[string]*Class
	traits     map[string]*Trait
	interfaces map[string]*Interface
	constants  map[string]Value
	global     *Environment
}

// NewEnvironment creates a new global environment.
func NewEnvironment() *Environment {
	env := &Environment{
		store:      make(map[string]Value),
		functions:  make(map[string]*Function),
		classes:    make(map[string]*Class),
		traits:     make(map[string]*Trait),
		interfaces: make(map[string]*Interface),
		constants:  make(map[string]Value),
	}
	env.global = env
	return env
}

// NewEnclosedEnvironment creates a new environment enclosed by an outer one.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := &Environment{
		store:      make(map[string]Value),
		outer:      outer,
		functions:  outer.functions,  // Share function registry
		classes:    outer.classes,    // Share class registry
		traits:     outer.traits,     // Share trait registry
		interfaces: outer.interfaces, // Share interface registry
		constants:  outer.constants,  // Share constants
		global:     outer.global,
	}
	return env
}

// Get retrieves a variable from the current scope or outer scopes.
func (e *Environment) Get(name string) (Value, bool) {
	val, ok := e.store[name]
	if !ok && e.outer != nil {
		// Look up in outer scope (for closures with captured variables)
		return e.outer.Get(name)
	}
	return val, ok
}

// Set sets a variable in the current scope.
func (e *Environment) Set(name string, val Value) Value {
	e.store[name] = val
	return val
}

// GetRef gets a reference to a variable (for pass-by-reference).
func (e *Environment) GetRef(name string) *Value {
	if _, ok := e.store[name]; !ok {
		e.store[name] = NULL
	}
	val := e.store[name]
	return &val
}

// SetRef sets a variable by reference.
func (e *Environment) SetRef(name string, ref *Value) {
	e.store[name] = *ref
}

// GetAllVariables returns all variables in the current scope.
func (e *Environment) GetAllVariables() map[string]Value {
	result := make(map[string]Value)
	for k, v := range e.store {
		result[k] = v
	}
	return result
}

// GetAllConstants returns all constants.
func (e *Environment) GetAllConstants() map[string]Value {
	result := make(map[string]Value)
	for k, v := range e.constants {
		result[k] = v
	}
	return result
}

// Isset checks if a variable is set and not null.
func (e *Environment) Isset(name string) bool {
	val, ok := e.store[name]
	if !ok {
		return false
	}
	_, isNull := val.(*Null)
	return !isNull
}

// Unset removes a variable from the current scope.
func (e *Environment) Unset(name string) {
	delete(e.store, name)
}

// Global returns the global environment.
func (e *Environment) Global() *Environment {
	return e.global
}

// GetGlobal gets a variable from the global scope.
func (e *Environment) GetGlobal(name string) (Value, bool) {
	return e.global.Get(name)
}

// SetGlobal sets a variable in the global scope.
func (e *Environment) SetGlobal(name string, val Value) {
	e.global.Set(name, val)
}

// ImportGlobal imports a global variable into the current scope.
func (e *Environment) ImportGlobal(name string) {
	if val, ok := e.global.store[name]; ok {
		e.store[name] = val
	} else {
		e.store[name] = NULL
	}
}

// ----------------------------------------------------------------------------
// Functions

// DefineFunction defines a function.
func (e *Environment) DefineFunction(name string, fn *Function) {
	e.functions[name] = fn
}

// GetFunction retrieves a function by name.
func (e *Environment) GetFunction(name string) (*Function, bool) {
	fn, ok := e.functions[name]
	return fn, ok
}

// ----------------------------------------------------------------------------
// Classes

// DefineClass defines a class.
func (e *Environment) DefineClass(name string, class *Class) {
	e.classes[name] = class
}

// GetClass retrieves a class by name.
func (e *Environment) GetClass(name string) (*Class, bool) {
	class, ok := e.classes[name]
	return class, ok
}

// ----------------------------------------------------------------------------
// Traits

// DefineTrait defines a trait.
func (e *Environment) DefineTrait(name string, trait *Trait) {
	e.traits[name] = trait
}

// GetTrait retrieves a trait by name.
func (e *Environment) GetTrait(name string) (*Trait, bool) {
	trait, ok := e.traits[name]
	return trait, ok
}

// ----------------------------------------------------------------------------
// Interfaces

// DefineInterface defines an interface.
func (e *Environment) DefineInterface(name string, iface *Interface) {
	e.interfaces[name] = iface
}

// GetInterface retrieves an interface by name.
func (e *Environment) GetInterface(name string) (*Interface, bool) {
	iface, ok := e.interfaces[name]
	return iface, ok
}

// ----------------------------------------------------------------------------
// Constants

// DefineConstant defines a constant.
func (e *Environment) DefineConstant(name string, val Value) bool {
	if _, exists := e.constants[name]; exists {
		return false // Already defined
	}
	e.constants[name] = val
	return true
}

// GetConstant retrieves a constant by name.
func (e *Environment) GetConstant(name string) (Value, bool) {
	val, ok := e.constants[name]
	return val, ok
}

// ----------------------------------------------------------------------------
// Superglobals

// InitSuperglobals initializes PHP superglobals.
func (e *Environment) InitSuperglobals() {
	// $_SERVER
	server := NewArray()
	e.global.Set("_SERVER", server)

	// $_GET
	get := NewArray()
	e.global.Set("_GET", get)

	// $_POST
	post := NewArray()
	e.global.Set("_POST", post)

	// $_REQUEST
	request := NewArray()
	e.global.Set("_REQUEST", request)

	// $_COOKIE
	cookie := NewArray()
	e.global.Set("_COOKIE", cookie)

	// $_SESSION
	session := NewArray()
	e.global.Set("_SESSION", session)

	// $_FILES
	files := NewArray()
	e.global.Set("_FILES", files)

	// $_ENV
	envArr := NewArray()
	e.global.Set("_ENV", envArr)

	// $GLOBALS - this is a reference to all global variables
	globals := NewArray()
	e.global.Set("GLOBALS", globals)
}

// ----------------------------------------------------------------------------
// Static variables

type StaticVars struct {
	vars map[string]map[string]Value // function -> variable -> value
}

func NewStaticVars() *StaticVars {
	return &StaticVars{
		vars: make(map[string]map[string]Value),
	}
}

func (s *StaticVars) Get(funcName, varName string) (Value, bool) {
	if fn, ok := s.vars[funcName]; ok {
		if val, ok := fn[varName]; ok {
			return val, true
		}
	}
	return nil, false
}

func (s *StaticVars) Set(funcName, varName string, val Value) {
	if _, ok := s.vars[funcName]; !ok {
		s.vars[funcName] = make(map[string]Value)
	}
	s.vars[funcName][varName] = val
}
