package interpreter

import (
	"fmt"
	"strings"

	"github.com/alexisbouchez/phpgo/ast"
	"github.com/alexisbouchez/phpgo/runtime"
)

// ReflectionClass wraps a runtime.Class for reflection
type ReflectionClass struct {
	Class *runtime.Class
}

func (r *ReflectionClass) Type() string     { return "object" }
func (r *ReflectionClass) ToBool() bool     { return true }
func (r *ReflectionClass) ToInt() int64     { return 1 }
func (r *ReflectionClass) ToFloat() float64 { return 1.0 }
func (r *ReflectionClass) ToString() string { return "ReflectionClass" }
func (r *ReflectionClass) Inspect() string  { return fmt.Sprintf("object(ReflectionClass)#%p", r) }

// ReflectionMethod wraps a runtime.Method for reflection
type ReflectionMethod struct {
	Class  *runtime.Class
	Method *runtime.Method
}

func (r *ReflectionMethod) Type() string     { return "object" }
func (r *ReflectionMethod) ToBool() bool     { return true }
func (r *ReflectionMethod) ToInt() int64     { return 1 }
func (r *ReflectionMethod) ToFloat() float64 { return 1.0 }
func (r *ReflectionMethod) ToString() string { return "ReflectionMethod" }
func (r *ReflectionMethod) Inspect() string  { return fmt.Sprintf("object(ReflectionMethod)#%p", r) }

// ReflectionProperty wraps a runtime.PropertyDef for reflection
type ReflectionProperty struct {
	Class    *runtime.Class
	Property *runtime.PropertyDef
}

func (r *ReflectionProperty) Type() string     { return "object" }
func (r *ReflectionProperty) ToBool() bool     { return true }
func (r *ReflectionProperty) ToInt() int64     { return 1 }
func (r *ReflectionProperty) ToFloat() float64 { return 1.0 }
func (r *ReflectionProperty) ToString() string { return "ReflectionProperty" }
func (r *ReflectionProperty) Inspect() string  { return fmt.Sprintf("object(ReflectionProperty)#%p", r) }

// ReflectionFunction wraps a runtime.Function for reflection
type ReflectionFunction struct {
	Function *runtime.Function
	Name     string
}

func (r *ReflectionFunction) Type() string     { return "object" }
func (r *ReflectionFunction) ToBool() bool     { return true }
func (r *ReflectionFunction) ToInt() int64     { return 1 }
func (r *ReflectionFunction) ToFloat() float64 { return 1.0 }
func (r *ReflectionFunction) ToString() string { return "ReflectionFunction" }
func (r *ReflectionFunction) Inspect() string  { return fmt.Sprintf("object(ReflectionFunction)#%p", r) }

// ReflectionParameter wraps a function parameter for reflection
type ReflectionParameter struct {
	Function     *runtime.Function
	Method       *runtime.Method
	ParamName    string
	ParamIndex   int
	DefaultValue runtime.Value
	HasDefault   bool
}

func (r *ReflectionParameter) Type() string     { return "object" }
func (r *ReflectionParameter) ToBool() bool     { return true }
func (r *ReflectionParameter) ToInt() int64     { return 1 }
func (r *ReflectionParameter) ToFloat() float64 { return 1.0 }
func (r *ReflectionParameter) ToString() string { return "ReflectionParameter" }
func (r *ReflectionParameter) Inspect() string  { return fmt.Sprintf("object(ReflectionParameter)#%p", r) }

// handleReflectionNew handles instantiation of Reflection* classes
func (i *Interpreter) handleReflectionNew(className string, args []runtime.Value) runtime.Value {
	switch className {
	case "ReflectionClass":
		return i.newReflectionClass(args)
	case "ReflectionMethod":
		return i.newReflectionMethod(args)
	case "ReflectionProperty":
		return i.newReflectionProperty(args)
	case "ReflectionFunction":
		return i.newReflectionFunction(args)
	default:
		return nil
	}
}

func (i *Interpreter) newReflectionClass(args []runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewError("ReflectionClass::__construct() expects exactly 1 parameter")
	}

	var className string
	switch v := args[0].(type) {
	case *runtime.String:
		className = v.Value
	case *runtime.Object:
		className = v.Class.Name
	default:
		className = args[0].ToString()
	}

	class, ok := i.env.GetClass(className)
	if !ok {
		return runtime.NewError(fmt.Sprintf("Class %s does not exist", className))
	}

	return &ReflectionClass{Class: class}
}

func (i *Interpreter) newReflectionMethod(args []runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewError("ReflectionMethod::__construct() expects exactly 2 parameters")
	}

	var className string
	switch v := args[0].(type) {
	case *runtime.String:
		className = v.Value
	case *runtime.Object:
		className = v.Class.Name
	case *ReflectionClass:
		className = v.Class.Name
	default:
		className = args[0].ToString()
	}

	methodName := args[1].ToString()

	class, ok := i.env.GetClass(className)
	if !ok {
		return runtime.NewError(fmt.Sprintf("Class %s does not exist", className))
	}

	method, exists := class.Methods[methodName]
	if !exists {
		// Check parent classes
		method, class = i.findMethod(class, methodName)
		if method == nil {
			return runtime.NewError(fmt.Sprintf("Method %s::%s() does not exist", className, methodName))
		}
	}

	return &ReflectionMethod{Class: class, Method: method}
}

func (i *Interpreter) newReflectionProperty(args []runtime.Value) runtime.Value {
	if len(args) < 2 {
		return runtime.NewError("ReflectionProperty::__construct() expects exactly 2 parameters")
	}

	var className string
	switch v := args[0].(type) {
	case *runtime.String:
		className = v.Value
	case *runtime.Object:
		className = v.Class.Name
	case *ReflectionClass:
		className = v.Class.Name
	default:
		className = args[0].ToString()
	}

	propName := args[1].ToString()

	class, ok := i.env.GetClass(className)
	if !ok {
		return runtime.NewError(fmt.Sprintf("Class %s does not exist", className))
	}

	prop, exists := class.Properties[propName]
	if !exists {
		return runtime.NewError(fmt.Sprintf("Property %s::$%s does not exist", className, propName))
	}

	return &ReflectionProperty{Class: class, Property: prop}
}

func (i *Interpreter) newReflectionFunction(args []runtime.Value) runtime.Value {
	if len(args) < 1 {
		return runtime.NewError("ReflectionFunction::__construct() expects exactly 1 parameter")
	}

	funcName := args[0].ToString()
	fn, ok := i.env.GetFunction(funcName)
	if !ok {
		return runtime.NewError(fmt.Sprintf("Function %s() does not exist", funcName))
	}

	return &ReflectionFunction{Function: fn, Name: funcName}
}

// callReflectionMethod handles method calls on Reflection* objects
func (i *Interpreter) callReflectionMethod(obj runtime.Value, methodName string, args []runtime.Value) runtime.Value {
	switch r := obj.(type) {
	case *ReflectionClass:
		return i.callReflectionClassMethod(r, methodName, args)
	case *ReflectionMethod:
		return i.callReflectionMethodMethod(r, methodName, args)
	case *ReflectionProperty:
		return i.callReflectionPropertyMethod(r, methodName, args)
	case *ReflectionFunction:
		return i.callReflectionFunctionMethod(r, methodName, args)
	case *ReflectionParameter:
		return i.callReflectionParameterMethod(r, methodName, args)
	}
	return runtime.NULL
}

// ReflectionClass methods
func (i *Interpreter) callReflectionClassMethod(r *ReflectionClass, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "getName":
		return runtime.NewString(r.Class.Name)
	case "getShortName":
		parts := strings.Split(r.Class.Name, "\\")
		return runtime.NewString(parts[len(parts)-1])
	case "getNamespaceName":
		parts := strings.Split(r.Class.Name, "\\")
		if len(parts) > 1 {
			return runtime.NewString(strings.Join(parts[:len(parts)-1], "\\"))
		}
		return runtime.NewString("")
	case "isAbstract":
		return runtime.NewBool(r.Class.IsAbstract)
	case "isFinal":
		return runtime.NewBool(r.Class.IsFinal)
	case "isInterface":
		return runtime.FALSE // We don't track this separately yet
	case "isInstantiable":
		return runtime.NewBool(!r.Class.IsAbstract)
	case "hasMethod":
		if len(args) < 1 {
			return runtime.FALSE
		}
		_, exists := r.Class.Methods[args[0].ToString()]
		return runtime.NewBool(exists)
	case "hasProperty":
		if len(args) < 1 {
			return runtime.FALSE
		}
		_, exists := r.Class.Properties[args[0].ToString()]
		return runtime.NewBool(exists)
	case "hasConstant":
		if len(args) < 1 {
			return runtime.FALSE
		}
		_, exists := r.Class.Constants[args[0].ToString()]
		return runtime.NewBool(exists)
	case "getMethod":
		if len(args) < 1 {
			return runtime.NewError("ReflectionClass::getMethod() expects exactly 1 parameter")
		}
		methodName := args[0].ToString()
		method, exists := r.Class.Methods[methodName]
		if !exists {
			method, _ = i.findMethod(r.Class, methodName)
			if method == nil {
				return runtime.NewError(fmt.Sprintf("Method %s does not exist", methodName))
			}
		}
		return &ReflectionMethod{Class: r.Class, Method: method}
	case "getMethods":
		arr := runtime.NewArray()
		for name, method := range r.Class.Methods {
			rm := &ReflectionMethod{Class: r.Class, Method: method}
			arr.Set(runtime.NewString(name), rm)
		}
		return arr
	case "getProperty":
		if len(args) < 1 {
			return runtime.NewError("ReflectionClass::getProperty() expects exactly 1 parameter")
		}
		propName := args[0].ToString()
		prop, exists := r.Class.Properties[propName]
		if !exists {
			return runtime.NewError(fmt.Sprintf("Property %s does not exist", propName))
		}
		return &ReflectionProperty{Class: r.Class, Property: prop}
	case "getProperties":
		arr := runtime.NewArray()
		for name, prop := range r.Class.Properties {
			rp := &ReflectionProperty{Class: r.Class, Property: prop}
			arr.Set(runtime.NewString(name), rp)
		}
		return arr
	case "getConstants":
		arr := runtime.NewArray()
		for name, val := range r.Class.Constants {
			arr.Set(runtime.NewString(name), val)
		}
		return arr
	case "getConstant":
		if len(args) < 1 {
			return runtime.FALSE
		}
		if val, exists := r.Class.Constants[args[0].ToString()]; exists {
			return val
		}
		return runtime.FALSE
	case "getParentClass":
		if r.Class.Parent != nil {
			return &ReflectionClass{Class: r.Class.Parent}
		}
		return runtime.FALSE
	case "getInterfaceNames":
		arr := runtime.NewArray()
		for _, iface := range r.Class.Interfaces {
			arr.Set(nil, runtime.NewString(iface.Name))
		}
		return arr
	case "implementsInterface":
		if len(args) < 1 {
			return runtime.FALSE
		}
		ifaceName := args[0].ToString()
		for _, iface := range r.Class.Interfaces {
			if iface.Name == ifaceName {
				return runtime.TRUE
			}
		}
		return runtime.FALSE
	case "newInstance":
		return i.createInstance(r.Class, args)
	case "newInstanceArgs":
		if len(args) > 0 {
			if arr, ok := args[0].(*runtime.Array); ok {
				var newArgs []runtime.Value
				for _, k := range arr.Keys {
					newArgs = append(newArgs, arr.Elements[k])
				}
				return i.createInstance(r.Class, newArgs)
			}
		}
		return i.createInstance(r.Class, nil)
	case "newInstanceWithoutConstructor":
		obj := runtime.NewObject(r.Class)
		// Initialize properties with defaults
		for name, prop := range r.Class.Properties {
			if !prop.IsStatic && prop.Default != nil {
				obj.Properties[name] = prop.Default
			}
		}
		return obj
	case "isSubclassOf":
		if len(args) < 1 {
			return runtime.FALSE
		}
		parentName := args[0].ToString()
		parent := r.Class.Parent
		for parent != nil {
			if parent.Name == parentName {
				return runtime.TRUE
			}
			parent = parent.Parent
		}
		return runtime.FALSE
	default:
		return runtime.NewError(fmt.Sprintf("Call to undefined method ReflectionClass::%s()", methodName))
	}
}

// ReflectionMethod methods
func (i *Interpreter) callReflectionMethodMethod(r *ReflectionMethod, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "getName":
		return runtime.NewString(r.Method.Name)
	case "getDeclaringClass":
		return &ReflectionClass{Class: r.Class}
	case "isPublic":
		return runtime.NewBool(r.Method.IsPublic)
	case "isProtected":
		return runtime.NewBool(r.Method.IsProtected)
	case "isPrivate":
		return runtime.NewBool(r.Method.IsPrivate)
	case "isStatic":
		return runtime.NewBool(r.Method.IsStatic)
	case "isAbstract":
		return runtime.NewBool(r.Method.IsAbstract)
	case "isFinal":
		return runtime.NewBool(r.Method.IsFinal)
	case "isConstructor":
		return runtime.NewBool(r.Method.Name == "__construct")
	case "isDestructor":
		return runtime.NewBool(r.Method.Name == "__destruct")
	case "getNumberOfParameters":
		return runtime.NewInt(int64(len(r.Method.Params)))
	case "getNumberOfRequiredParameters":
		required := 0
		for idx := range r.Method.Params {
			if idx >= len(r.Method.Defaults) || r.Method.Defaults[idx] == nil {
				required++
			}
		}
		return runtime.NewInt(int64(required))
	case "getParameters":
		arr := runtime.NewArray()
		for idx, param := range r.Method.Params {
			var defaultVal runtime.Value
			hasDefault := false
			if idx < len(r.Method.Defaults) && r.Method.Defaults[idx] != nil {
				defaultVal = r.Method.Defaults[idx]
				hasDefault = true
			}
			rp := &ReflectionParameter{
				Method:       r.Method,
				ParamName:    param,
				ParamIndex:   idx,
				DefaultValue: defaultVal,
				HasDefault:   hasDefault,
			}
			arr.Set(nil, rp)
		}
		return arr
	case "invoke":
		if len(args) < 1 {
			return runtime.NewError("ReflectionMethod::invoke() expects at least 1 parameter")
		}
		obj, ok := args[0].(*runtime.Object)
		if !ok {
			return runtime.NewError("ReflectionMethod::invoke() expects parameter 1 to be object")
		}
		return i.callMethod(obj, r.Method.Name, args[1:])
	case "invokeArgs":
		if len(args) < 2 {
			return runtime.NewError("ReflectionMethod::invokeArgs() expects exactly 2 parameters")
		}
		obj, ok := args[0].(*runtime.Object)
		if !ok {
			return runtime.NewError("ReflectionMethod::invokeArgs() expects parameter 1 to be object")
		}
		var methodArgs []runtime.Value
		if arr, ok := args[1].(*runtime.Array); ok {
			for _, k := range arr.Keys {
				methodArgs = append(methodArgs, arr.Elements[k])
			}
		}
		return i.callMethod(obj, r.Method.Name, methodArgs)
	case "setAccessible":
		// In PHP 8+, this is a no-op but we accept it for compatibility
		return runtime.NULL
	default:
		return runtime.NewError(fmt.Sprintf("Call to undefined method ReflectionMethod::%s()", methodName))
	}
}

// ReflectionProperty methods
func (i *Interpreter) callReflectionPropertyMethod(r *ReflectionProperty, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "getName":
		return runtime.NewString(r.Property.Name)
	case "getDeclaringClass":
		return &ReflectionClass{Class: r.Class}
	case "isPublic":
		return runtime.NewBool(r.Property.IsPublic)
	case "isProtected":
		return runtime.NewBool(r.Property.IsProtected)
	case "isPrivate":
		return runtime.NewBool(r.Property.IsPrivate)
	case "isStatic":
		return runtime.NewBool(r.Property.IsStatic)
	case "isReadOnly":
		return runtime.NewBool(r.Property.IsReadonly)
	case "hasDefaultValue":
		return runtime.NewBool(r.Property.Default != nil)
	case "getDefaultValue":
		if r.Property.Default != nil {
			return r.Property.Default
		}
		return runtime.NULL
	case "getValue":
		if r.Property.IsStatic {
			if val, ok := r.Class.StaticProps[r.Property.Name]; ok {
				return val
			}
			return runtime.NULL
		}
		if len(args) < 1 {
			return runtime.NewError("ReflectionProperty::getValue() expects exactly 1 parameter for non-static property")
		}
		if obj, ok := args[0].(*runtime.Object); ok {
			return obj.GetProperty(r.Property.Name)
		}
		return runtime.NULL
	case "setValue":
		if r.Property.IsStatic {
			if len(args) < 1 {
				return runtime.NewError("ReflectionProperty::setValue() expects at least 1 parameter")
			}
			if r.Class.StaticProps == nil {
				r.Class.StaticProps = make(map[string]runtime.Value)
			}
			r.Class.StaticProps[r.Property.Name] = args[0]
			return runtime.NULL
		}
		if len(args) < 2 {
			return runtime.NewError("ReflectionProperty::setValue() expects exactly 2 parameters for non-static property")
		}
		if obj, ok := args[0].(*runtime.Object); ok {
			obj.SetProperty(r.Property.Name, args[1])
		}
		return runtime.NULL
	case "setAccessible":
		// In PHP 8+, this is a no-op but we accept it for compatibility
		return runtime.NULL
	default:
		return runtime.NewError(fmt.Sprintf("Call to undefined method ReflectionProperty::%s()", methodName))
	}
}

// ReflectionFunction methods
func (i *Interpreter) callReflectionFunctionMethod(r *ReflectionFunction, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "getName":
		return runtime.NewString(r.Name)
	case "getShortName":
		parts := strings.Split(r.Name, "\\")
		return runtime.NewString(parts[len(parts)-1])
	case "getNamespaceName":
		parts := strings.Split(r.Name, "\\")
		if len(parts) > 1 {
			return runtime.NewString(strings.Join(parts[:len(parts)-1], "\\"))
		}
		return runtime.NewString("")
	case "getNumberOfParameters":
		return runtime.NewInt(int64(len(r.Function.Params)))
	case "getNumberOfRequiredParameters":
		required := 0
		for idx := range r.Function.Params {
			if idx >= len(r.Function.Defaults) || r.Function.Defaults[idx] == nil {
				required++
			}
		}
		return runtime.NewInt(int64(required))
	case "getParameters":
		arr := runtime.NewArray()
		for idx, param := range r.Function.Params {
			var defaultVal runtime.Value
			hasDefault := false
			if idx < len(r.Function.Defaults) && r.Function.Defaults[idx] != nil {
				defaultVal = r.Function.Defaults[idx]
				hasDefault = true
			}
			rp := &ReflectionParameter{
				Function:     r.Function,
				ParamName:    param,
				ParamIndex:   idx,
				DefaultValue: defaultVal,
				HasDefault:   hasDefault,
			}
			arr.Set(nil, rp)
		}
		return arr
	case "isVariadic":
		return runtime.NewBool(r.Function.Variadic)
	case "invoke":
		return i.callUserFunction(r.Function, args)
	case "invokeArgs":
		if len(args) > 0 {
			if arr, ok := args[0].(*runtime.Array); ok {
				var funcArgs []runtime.Value
				for _, k := range arr.Keys {
					funcArgs = append(funcArgs, arr.Elements[k])
				}
				return i.callUserFunction(r.Function, funcArgs)
			}
		}
		return i.callUserFunction(r.Function, nil)
	default:
		return runtime.NewError(fmt.Sprintf("Call to undefined method ReflectionFunction::%s()", methodName))
	}
}

// ReflectionParameter methods
func (i *Interpreter) callReflectionParameterMethod(r *ReflectionParameter, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "getName":
		return runtime.NewString(r.ParamName)
	case "getPosition":
		return runtime.NewInt(int64(r.ParamIndex))
	case "isOptional":
		return runtime.NewBool(r.HasDefault)
	case "hasDefaultValue", "isDefaultValueAvailable":
		return runtime.NewBool(r.HasDefault)
	case "getDefaultValue":
		if r.HasDefault {
			return r.DefaultValue
		}
		return runtime.NewError("Parameter has no default value")
	case "isVariadic":
		if r.Function != nil {
			return runtime.NewBool(r.Function.Variadic && r.ParamIndex == len(r.Function.Params)-1)
		}
		if r.Method != nil {
			return runtime.NewBool(r.Method.Variadic && r.ParamIndex == len(r.Method.Params)-1)
		}
		return runtime.FALSE
	case "allowsNull":
		// We don't track nullability yet, return true for optional params
		return runtime.NewBool(r.HasDefault)
	default:
		return runtime.NewError(fmt.Sprintf("Call to undefined method ReflectionParameter::%s()", methodName))
	}
}

// createInstance creates a new instance of a class with constructor
func (i *Interpreter) createInstance(class *runtime.Class, args []runtime.Value) runtime.Value {
	if class.IsAbstract {
		return runtime.NewError(fmt.Sprintf("Cannot instantiate abstract class %s", class.Name))
	}

	obj := runtime.NewObject(class)

	// Initialize properties with defaults
	for name, prop := range class.Properties {
		if !prop.IsStatic && prop.Default != nil {
			obj.Properties[name] = prop.Default
		}
	}

	// Set up __toString callback if method exists
	if _, hasToString := class.Methods["__toString"]; hasToString {
		obj.SetToStringCallback(i.createToStringCallback())
	}

	// Call constructor if exists
	if constructor, exists := class.Methods["__construct"]; exists {
		i.callMethodOnObject(obj, constructor, args)
	}

	return obj
}

// isReflectionClass checks if a class name is a Reflection class
func isReflectionClass(name string) bool {
	switch name {
	case "ReflectionClass", "ReflectionMethod", "ReflectionProperty", "ReflectionFunction", "ReflectionParameter":
		return true
	}
	return false
}

// callMethodOnObject calls a method on an object with evaluated arguments
func (i *Interpreter) callMethodOnObject(obj *runtime.Object, method *runtime.Method, args []runtime.Value) runtime.Value {
	return i.invokeMethodWithArgs(obj, method, obj.Class, args)
}

// callMethod calls a method by name on an object
func (i *Interpreter) callMethod(obj *runtime.Object, methodName string, args []runtime.Value) runtime.Value {
	method, foundClass := i.findMethod(obj.Class, methodName)
	if method == nil {
		return runtime.NewError(fmt.Sprintf("Method %s::%s() does not exist", obj.Class.Name, methodName))
	}
	return i.invokeMethodWithArgs(obj, method, foundClass, args)
}

// callUserFunction calls a user-defined function with evaluated arguments
func (i *Interpreter) callUserFunction(fn *runtime.Function, args []runtime.Value) runtime.Value {
	env := runtime.NewEnclosedEnvironment(fn.Env)
	oldEnv := i.env
	oldFuncArgs := i.currentFuncArgs
	i.env = env
	i.currentFuncArgs = args

	// Bind parameters
	for idx, param := range fn.Params {
		if idx < len(args) {
			env.Set(param, args[idx])
		} else if idx < len(fn.Defaults) && fn.Defaults[idx] != nil {
			env.Set(param, fn.Defaults[idx])
		}
	}

	// Handle variadic
	if fn.Variadic && len(fn.Params) > 0 {
		lastParam := fn.Params[len(fn.Params)-1]
		variadicArgs := runtime.NewArray()
		for idx := len(fn.Params) - 1; idx < len(args); idx++ {
			variadicArgs.Set(nil, args[idx])
		}
		env.Set(lastParam, variadicArgs)
	}

	var result runtime.Value = runtime.NULL
	if block, ok := fn.Body.(*ast.BlockStmt); ok {
		result = i.evalBlock(block)
	}

	i.env = oldEnv
	i.currentFuncArgs = oldFuncArgs

	if ret, ok := result.(*runtime.ReturnValue); ok {
		return ret.Value
	}
	return result
}
