package interpreter

import (
	"fmt"

	"github.com/alexisbouchez/phpgo/ast"
	"github.com/alexisbouchez/phpgo/runtime"
)

// isSplDataStructure checks if a class name is an SPL data structure
func isSplDataStructure(name string) bool {
	switch name {
	case "SplFixedArray", "SplDoublyLinkedList", "SplStack", "SplQueue",
		"SplHeap", "SplMinHeap", "SplMaxHeap", "SplPriorityQueue", "SplObjectStorage":
		return true
	}
	return false
}

// handleSplStaticCall handles static method calls on SPL classes
func (i *Interpreter) handleSplStaticCall(className, methodName string, args []runtime.Value) runtime.Value {
	switch className {
	case "SplFixedArray":
		switch methodName {
		case "fromArray":
			if len(args) < 1 {
				return runtime.NewError("SplFixedArray::fromArray() expects at least 1 parameter")
			}
			arr, ok := args[0].(*runtime.Array)
			if !ok {
				return runtime.NewError("SplFixedArray::fromArray() expects parameter 1 to be array")
			}
			size := int64(arr.Len())
			spl := NewSplFixedArray(size)
			// Copy array elements in order
			for idx, key := range arr.Keys {
				spl.elements[idx] = arr.Elements[key]
			}
			return spl
		}
	}
	return runtime.NewError(fmt.Sprintf("undefined static method: %s::%s", className, methodName))
}

// handleSplNew creates a new SPL data structure object
func (i *Interpreter) handleSplNew(className string, args []runtime.Value) runtime.Value {
	switch className {
	case "SplFixedArray":
		size := int64(0)
		if len(args) > 0 {
			size = args[0].ToInt()
		}
		if size < 0 {
			return runtime.NewError("array size cannot be less than zero")
		}
		return NewSplFixedArray(size)
	case "SplDoublyLinkedList":
		return NewSplDoublyLinkedList()
	case "SplStack":
		return NewSplStack()
	case "SplQueue":
		return NewSplQueue()
	case "SplHeap":
		return runtime.NewError("Cannot instantiate abstract class SplHeap")
	case "SplMinHeap":
		return NewSplMinHeap()
	case "SplMaxHeap":
		return NewSplMaxHeap()
	case "SplPriorityQueue":
		return NewSplPriorityQueue()
	case "SplObjectStorage":
		return NewSplObjectStorage()
	}
	return runtime.NewError(fmt.Sprintf("unknown SPL class: %s", className))
}

// SplFixedArrayObject represents a native SplFixedArray
type SplFixedArrayObject struct {
	elements []runtime.Value
	size     int64
}

func NewSplFixedArray(size int64) *SplFixedArrayObject {
	elements := make([]runtime.Value, size)
	for i := range elements {
		elements[i] = runtime.NULL
	}
	return &SplFixedArrayObject{elements: elements, size: size}
}

func (s *SplFixedArrayObject) Type() string     { return "object" }
func (s *SplFixedArrayObject) ToBool() bool     { return true }
func (s *SplFixedArrayObject) ToInt() int64     { return 1 }
func (s *SplFixedArrayObject) ToFloat() float64 { return 1.0 }
func (s *SplFixedArrayObject) ToString() string { return "SplFixedArray" }
func (s *SplFixedArrayObject) Inspect() string  { return fmt.Sprintf("object(SplFixedArray)#%p (%d)", s, s.size) }

// SplDoublyLinkedListObject represents a native SplDoublyLinkedList
type SplDoublyLinkedListObject struct {
	elements []runtime.Value
	mode     int64 // IT_MODE_LIFO=2, IT_MODE_FIFO=0, IT_MODE_DELETE=1, IT_MODE_KEEP=0
	position int
}

func NewSplDoublyLinkedList() *SplDoublyLinkedListObject {
	return &SplDoublyLinkedListObject{
		elements: make([]runtime.Value, 0),
		mode:     0, // IT_MODE_FIFO | IT_MODE_KEEP
		position: 0,
	}
}

func (s *SplDoublyLinkedListObject) Type() string     { return "object" }
func (s *SplDoublyLinkedListObject) ToBool() bool     { return len(s.elements) > 0 }
func (s *SplDoublyLinkedListObject) ToInt() int64     { return int64(len(s.elements)) }
func (s *SplDoublyLinkedListObject) ToFloat() float64 { return float64(len(s.elements)) }
func (s *SplDoublyLinkedListObject) ToString() string { return "SplDoublyLinkedList" }
func (s *SplDoublyLinkedListObject) Inspect() string {
	return fmt.Sprintf("object(SplDoublyLinkedList)#%p (%d)", s, len(s.elements))
}

// SplStackObject represents a native SplStack (LIFO)
type SplStackObject struct {
	*SplDoublyLinkedListObject
}

func NewSplStack() *SplStackObject {
	dll := NewSplDoublyLinkedList()
	dll.mode = 2 // IT_MODE_LIFO
	return &SplStackObject{SplDoublyLinkedListObject: dll}
}

func (s *SplStackObject) ToString() string { return "SplStack" }
func (s *SplStackObject) Inspect() string {
	return fmt.Sprintf("object(SplStack)#%p (%d)", s, len(s.elements))
}

// SplQueueObject represents a native SplQueue (FIFO)
type SplQueueObject struct {
	*SplDoublyLinkedListObject
}

func NewSplQueue() *SplQueueObject {
	dll := NewSplDoublyLinkedList()
	dll.mode = 0 // IT_MODE_FIFO
	return &SplQueueObject{SplDoublyLinkedListObject: dll}
}

func (s *SplQueueObject) ToString() string { return "SplQueue" }
func (s *SplQueueObject) Inspect() string {
	return fmt.Sprintf("object(SplQueue)#%p (%d)", s, len(s.elements))
}

// SplHeapObject represents an abstract SplHeap
type SplHeapObject struct {
	elements   []runtime.Value
	position   int
	isMaxHeap  bool
	interpreter *Interpreter
}

func NewSplMinHeap() *SplHeapObject {
	return &SplHeapObject{
		elements:  make([]runtime.Value, 0),
		position:  0,
		isMaxHeap: false,
	}
}

func NewSplMaxHeap() *SplHeapObject {
	return &SplHeapObject{
		elements:  make([]runtime.Value, 0),
		position:  0,
		isMaxHeap: true,
	}
}

func (s *SplHeapObject) Type() string     { return "object" }
func (s *SplHeapObject) ToBool() bool     { return len(s.elements) > 0 }
func (s *SplHeapObject) ToInt() int64     { return int64(len(s.elements)) }
func (s *SplHeapObject) ToFloat() float64 { return float64(len(s.elements)) }
func (s *SplHeapObject) ToString() string {
	if s.isMaxHeap {
		return "SplMaxHeap"
	}
	return "SplMinHeap"
}
func (s *SplHeapObject) Inspect() string {
	return fmt.Sprintf("object(%s)#%p (%d)", s.ToString(), s, len(s.elements))
}

// Heap operations
func (s *SplHeapObject) insert(value runtime.Value) {
	s.elements = append(s.elements, value)
	s.heapifyUp(len(s.elements) - 1)
}

func (s *SplHeapObject) extract() runtime.Value {
	if len(s.elements) == 0 {
		return runtime.NewError("Can't extract from an empty heap")
	}
	result := s.elements[0]
	last := len(s.elements) - 1
	s.elements[0] = s.elements[last]
	s.elements = s.elements[:last]
	if len(s.elements) > 0 {
		s.heapifyDown(0)
	}
	return result
}

func (s *SplHeapObject) top() runtime.Value {
	if len(s.elements) == 0 {
		return runtime.NewError("Can't peek at an empty heap")
	}
	return s.elements[0]
}

func (s *SplHeapObject) compare(a, b runtime.Value) int {
	// Default comparison for numeric values
	aNum := a.ToFloat()
	bNum := b.ToFloat()
	if aNum < bNum {
		return -1
	} else if aNum > bNum {
		return 1
	}
	return 0
}

func (s *SplHeapObject) heapifyUp(index int) {
	for index > 0 {
		parent := (index - 1) / 2
		cmp := s.compare(s.elements[index], s.elements[parent])
		// For max heap: child > parent means swap
		// For min heap: child < parent means swap
		shouldSwap := (s.isMaxHeap && cmp > 0) || (!s.isMaxHeap && cmp < 0)
		if !shouldSwap {
			break
		}
		s.elements[index], s.elements[parent] = s.elements[parent], s.elements[index]
		index = parent
	}
}

func (s *SplHeapObject) heapifyDown(index int) {
	for {
		best := index
		left := 2*index + 1
		right := 2*index + 2

		if left < len(s.elements) {
			cmp := s.compare(s.elements[left], s.elements[best])
			if (s.isMaxHeap && cmp > 0) || (!s.isMaxHeap && cmp < 0) {
				best = left
			}
		}
		if right < len(s.elements) {
			cmp := s.compare(s.elements[right], s.elements[best])
			if (s.isMaxHeap && cmp > 0) || (!s.isMaxHeap && cmp < 0) {
				best = right
			}
		}
		if best == index {
			break
		}
		s.elements[index], s.elements[best] = s.elements[best], s.elements[index]
		index = best
	}
}

// SplPriorityQueueObject represents a native SplPriorityQueue
type SplPriorityQueueObject struct {
	elements   []priorityQueueElement
	position   int
	extractFlag int64 // EXTR_DATA=1, EXTR_PRIORITY=2, EXTR_BOTH=3
}

type priorityQueueElement struct {
	data     runtime.Value
	priority runtime.Value
}

func NewSplPriorityQueue() *SplPriorityQueueObject {
	return &SplPriorityQueueObject{
		elements:    make([]priorityQueueElement, 0),
		position:    0,
		extractFlag: 1, // EXTR_DATA
	}
}

func (s *SplPriorityQueueObject) Type() string     { return "object" }
func (s *SplPriorityQueueObject) ToBool() bool     { return len(s.elements) > 0 }
func (s *SplPriorityQueueObject) ToInt() int64     { return int64(len(s.elements)) }
func (s *SplPriorityQueueObject) ToFloat() float64 { return float64(len(s.elements)) }
func (s *SplPriorityQueueObject) ToString() string { return "SplPriorityQueue" }
func (s *SplPriorityQueueObject) Inspect() string {
	return fmt.Sprintf("object(SplPriorityQueue)#%p (%d)", s, len(s.elements))
}

func (s *SplPriorityQueueObject) insert(data, priority runtime.Value) {
	s.elements = append(s.elements, priorityQueueElement{data: data, priority: priority})
	s.heapifyUp(len(s.elements) - 1)
}

func (s *SplPriorityQueueObject) extract() runtime.Value {
	if len(s.elements) == 0 {
		return runtime.NewError("Can't extract from an empty heap")
	}
	result := s.elements[0]
	last := len(s.elements) - 1
	s.elements[0] = s.elements[last]
	s.elements = s.elements[:last]
	if len(s.elements) > 0 {
		s.heapifyDown(0)
	}

	switch s.extractFlag {
	case 1: // EXTR_DATA
		return result.data
	case 2: // EXTR_PRIORITY
		return result.priority
	case 3: // EXTR_BOTH
		arr := runtime.NewArray()
		arr.Set(runtime.NewString("data"), result.data)
		arr.Set(runtime.NewString("priority"), result.priority)
		return arr
	}
	return result.data
}

func (s *SplPriorityQueueObject) top() runtime.Value {
	if len(s.elements) == 0 {
		return runtime.NewError("Can't peek at an empty heap")
	}
	result := s.elements[0]
	switch s.extractFlag {
	case 1:
		return result.data
	case 2:
		return result.priority
	case 3:
		arr := runtime.NewArray()
		arr.Set(runtime.NewString("data"), result.data)
		arr.Set(runtime.NewString("priority"), result.priority)
		return arr
	}
	return result.data
}

func (s *SplPriorityQueueObject) compare(a, b priorityQueueElement) int {
	aNum := a.priority.ToFloat()
	bNum := b.priority.ToFloat()
	if aNum < bNum {
		return -1
	} else if aNum > bNum {
		return 1
	}
	return 0
}

func (s *SplPriorityQueueObject) heapifyUp(index int) {
	for index > 0 {
		parent := (index - 1) / 2
		cmp := s.compare(s.elements[index], s.elements[parent])
		// Max heap for priority queue (higher priority first)
		if cmp <= 0 {
			break
		}
		s.elements[index], s.elements[parent] = s.elements[parent], s.elements[index]
		index = parent
	}
}

func (s *SplPriorityQueueObject) heapifyDown(index int) {
	for {
		best := index
		left := 2*index + 1
		right := 2*index + 2

		if left < len(s.elements) && s.compare(s.elements[left], s.elements[best]) > 0 {
			best = left
		}
		if right < len(s.elements) && s.compare(s.elements[right], s.elements[best]) > 0 {
			best = right
		}
		if best == index {
			break
		}
		s.elements[index], s.elements[best] = s.elements[best], s.elements[index]
		index = best
	}
}

// SplObjectStorageObject represents a native SplObjectStorage
type SplObjectStorageObject struct {
	objects  map[string]runtime.Value // object hash -> object
	infos    map[string]runtime.Value // object hash -> associated info
	keys     []string                 // maintain insertion order
	position int
}

func NewSplObjectStorage() *SplObjectStorageObject {
	return &SplObjectStorageObject{
		objects:  make(map[string]runtime.Value),
		infos:    make(map[string]runtime.Value),
		keys:     make([]string, 0),
		position: 0,
	}
}

func (s *SplObjectStorageObject) Type() string     { return "object" }
func (s *SplObjectStorageObject) ToBool() bool     { return len(s.objects) > 0 }
func (s *SplObjectStorageObject) ToInt() int64     { return int64(len(s.objects)) }
func (s *SplObjectStorageObject) ToFloat() float64 { return float64(len(s.objects)) }
func (s *SplObjectStorageObject) ToString() string { return "SplObjectStorage" }
func (s *SplObjectStorageObject) Inspect() string {
	return fmt.Sprintf("object(SplObjectStorage)#%p (%d)", s, len(s.objects))
}

// evalForeachSplFixedArray handles foreach for SplFixedArray
func (i *Interpreter) evalForeachSplFixedArray(s *ast.ForeachStmt, spl *SplFixedArrayObject) runtime.Value {
	for idx := int64(0); idx < spl.size; idx++ {
		// Set key variable if present
		if s.KeyVar != nil {
			keyName := s.KeyVar.(*ast.Variable).Name.(*ast.Ident).Name
			i.env.Set(keyName, runtime.NewInt(idx))
		}

		// Set value variable
		valName := s.ValueVar.(*ast.Variable).Name.(*ast.Ident).Name
		i.env.Set(valName, spl.elements[idx])

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

// evalForeachSplDoublyLinkedList handles foreach for SplDoublyLinkedList, SplStack, SplQueue
func (i *Interpreter) evalForeachSplDoublyLinkedList(s *ast.ForeachStmt, spl *SplDoublyLinkedListObject) runtime.Value {
	// Determine iteration order based on mode
	isLIFO := spl.mode&2 == 2

	if isLIFO {
		// LIFO: iterate from end to beginning
		for idx := len(spl.elements) - 1; idx >= 0; idx-- {
			if s.KeyVar != nil {
				keyName := s.KeyVar.(*ast.Variable).Name.(*ast.Ident).Name
				i.env.Set(keyName, runtime.NewInt(int64(idx)))
			}
			valName := s.ValueVar.(*ast.Variable).Name.(*ast.Ident).Name
			i.env.Set(valName, spl.elements[idx])

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
	} else {
		// FIFO: iterate from beginning to end
		for idx := 0; idx < len(spl.elements); idx++ {
			if s.KeyVar != nil {
				keyName := s.KeyVar.(*ast.Variable).Name.(*ast.Ident).Name
				i.env.Set(keyName, runtime.NewInt(int64(idx)))
			}
			valName := s.ValueVar.(*ast.Variable).Name.(*ast.Ident).Name
			i.env.Set(valName, spl.elements[idx])

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
	}
	return runtime.NULL
}

// callSplMethod handles method calls on SPL data structure objects
func (i *Interpreter) callSplMethod(obj runtime.Value, methodName string, args []runtime.Value) runtime.Value {
	switch o := obj.(type) {
	case *SplFixedArrayObject:
		return i.callSplFixedArrayMethod(o, methodName, args)
	case *SplDoublyLinkedListObject:
		return i.callSplDoublyLinkedListMethod(o, methodName, args)
	case *SplStackObject:
		return i.callSplStackMethod(o, methodName, args)
	case *SplQueueObject:
		return i.callSplQueueMethod(o, methodName, args)
	case *SplHeapObject:
		return i.callSplHeapMethod(o, methodName, args)
	case *SplPriorityQueueObject:
		return i.callSplPriorityQueueMethod(o, methodName, args)
	case *SplObjectStorageObject:
		return i.callSplObjectStorageMethod(o, methodName, args)
	}
	return runtime.NewError("unknown SPL object type")
}

func (i *Interpreter) callSplFixedArrayMethod(s *SplFixedArrayObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "getSize", "count":
		return runtime.NewInt(s.size)
	case "setSize":
		if len(args) < 1 {
			return runtime.NewError("SplFixedArray::setSize() expects exactly 1 parameter")
		}
		newSize := args[0].ToInt()
		if newSize < 0 {
			return runtime.NewError("array size cannot be less than zero")
		}
		newElements := make([]runtime.Value, newSize)
		for i := range newElements {
			if int64(i) < s.size && i < len(s.elements) {
				newElements[i] = s.elements[i]
			} else {
				newElements[i] = runtime.NULL
			}
		}
		s.elements = newElements
		s.size = newSize
		return runtime.NULL
	case "toArray":
		arr := runtime.NewArray()
		for idx, val := range s.elements {
			arr.Set(runtime.NewInt(int64(idx)), val)
		}
		return arr
	case "offsetExists":
		if len(args) < 1 {
			return runtime.FALSE
		}
		idx := args[0].ToInt()
		return runtime.NewBool(idx >= 0 && idx < s.size)
	case "offsetGet":
		if len(args) < 1 {
			return runtime.NULL
		}
		idx := args[0].ToInt()
		if idx < 0 || idx >= s.size {
			return runtime.NULL
		}
		return s.elements[idx]
	case "offsetSet":
		if len(args) < 2 {
			return runtime.NULL
		}
		idx := args[0].ToInt()
		if idx < 0 || idx >= s.size {
			return runtime.NewError("index invalid or out of range")
		}
		s.elements[idx] = args[1]
		return runtime.NULL
	case "offsetUnset":
		if len(args) < 1 {
			return runtime.NULL
		}
		idx := args[0].ToInt()
		if idx >= 0 && idx < s.size {
			s.elements[idx] = runtime.NULL
		}
		return runtime.NULL
	case "rewind":
		return runtime.NULL
	case "current":
		return runtime.NULL
	case "key":
		return runtime.NULL
	case "next":
		return runtime.NULL
	case "valid":
		return runtime.FALSE
	}
	return runtime.NewError(fmt.Sprintf("undefined method: SplFixedArray::%s", methodName))
}

func (i *Interpreter) callSplDoublyLinkedListMethod(s *SplDoublyLinkedListObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "push":
		if len(args) < 1 {
			return runtime.NewError("SplDoublyLinkedList::push() expects exactly 1 parameter")
		}
		s.elements = append(s.elements, args[0])
		return runtime.NULL
	case "pop":
		if len(s.elements) == 0 {
			return runtime.NewError("Can't pop from an empty datastructure")
		}
		val := s.elements[len(s.elements)-1]
		s.elements = s.elements[:len(s.elements)-1]
		return val
	case "shift":
		if len(s.elements) == 0 {
			return runtime.NewError("Can't shift from an empty datastructure")
		}
		val := s.elements[0]
		s.elements = s.elements[1:]
		return val
	case "unshift":
		if len(args) < 1 {
			return runtime.NewError("SplDoublyLinkedList::unshift() expects exactly 1 parameter")
		}
		s.elements = append([]runtime.Value{args[0]}, s.elements...)
		return runtime.NULL
	case "top":
		if len(s.elements) == 0 {
			return runtime.NewError("Can't peek at an empty datastructure")
		}
		return s.elements[len(s.elements)-1]
	case "bottom":
		if len(s.elements) == 0 {
			return runtime.NewError("Can't peek at an empty datastructure")
		}
		return s.elements[0]
	case "count":
		return runtime.NewInt(int64(len(s.elements)))
	case "isEmpty":
		return runtime.NewBool(len(s.elements) == 0)
	case "setIteratorMode":
		if len(args) >= 1 {
			s.mode = args[0].ToInt()
		}
		return runtime.NULL
	case "getIteratorMode":
		return runtime.NewInt(s.mode)
	case "rewind":
		if s.mode&2 == 2 { // LIFO
			s.position = len(s.elements) - 1
		} else {
			s.position = 0
		}
		return runtime.NULL
	case "current":
		if s.position < 0 || s.position >= len(s.elements) {
			return runtime.FALSE
		}
		return s.elements[s.position]
	case "key":
		return runtime.NewInt(int64(s.position))
	case "next":
		if s.mode&2 == 2 { // LIFO
			s.position--
		} else {
			s.position++
		}
		return runtime.NULL
	case "valid":
		return runtime.NewBool(s.position >= 0 && s.position < len(s.elements))
	case "offsetExists":
		if len(args) < 1 {
			return runtime.FALSE
		}
		idx := args[0].ToInt()
		return runtime.NewBool(idx >= 0 && idx < int64(len(s.elements)))
	case "offsetGet":
		if len(args) < 1 {
			return runtime.NULL
		}
		idx := args[0].ToInt()
		if idx < 0 || idx >= int64(len(s.elements)) {
			return runtime.NULL
		}
		return s.elements[idx]
	case "offsetSet":
		if len(args) < 2 {
			return runtime.NULL
		}
		idx := args[0].ToInt()
		if idx < 0 || idx >= int64(len(s.elements)) {
			return runtime.NewError("index invalid or out of range")
		}
		s.elements[idx] = args[1]
		return runtime.NULL
	case "offsetUnset":
		if len(args) < 1 {
			return runtime.NULL
		}
		idx := int(args[0].ToInt())
		if idx >= 0 && idx < len(s.elements) {
			s.elements = append(s.elements[:idx], s.elements[idx+1:]...)
		}
		return runtime.NULL
	}
	return runtime.NewError(fmt.Sprintf("undefined method: SplDoublyLinkedList::%s", methodName))
}

func (i *Interpreter) callSplStackMethod(s *SplStackObject, methodName string, args []runtime.Value) runtime.Value {
	// SplStack inherits from SplDoublyLinkedList but with LIFO mode
	return i.callSplDoublyLinkedListMethod(s.SplDoublyLinkedListObject, methodName, args)
}

func (i *Interpreter) callSplQueueMethod(s *SplQueueObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "enqueue":
		if len(args) < 1 {
			return runtime.NewError("SplQueue::enqueue() expects exactly 1 parameter")
		}
		s.elements = append(s.elements, args[0])
		return runtime.NULL
	case "dequeue":
		if len(s.elements) == 0 {
			return runtime.NewError("Can't shift from an empty datastructure")
		}
		val := s.elements[0]
		s.elements = s.elements[1:]
		return val
	default:
		return i.callSplDoublyLinkedListMethod(s.SplDoublyLinkedListObject, methodName, args)
	}
}

func (i *Interpreter) callSplHeapMethod(s *SplHeapObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "insert":
		if len(args) < 1 {
			return runtime.NewError("SplHeap::insert() expects exactly 1 parameter")
		}
		s.insert(args[0])
		return runtime.TRUE
	case "extract":
		return s.extract()
	case "top":
		return s.top()
	case "count":
		return runtime.NewInt(int64(len(s.elements)))
	case "isEmpty":
		return runtime.NewBool(len(s.elements) == 0)
	case "rewind":
		s.position = 0
		return runtime.NULL
	case "current":
		if s.position >= len(s.elements) {
			return runtime.FALSE
		}
		// Create a copy to extract without modifying
		if len(s.elements) == 0 {
			return runtime.FALSE
		}
		return s.elements[0]
	case "key":
		return runtime.NewInt(int64(s.position))
	case "next":
		s.position++
		return runtime.NULL
	case "valid":
		return runtime.NewBool(s.position < len(s.elements) && len(s.elements) > 0)
	case "isCorrupted":
		return runtime.FALSE
	case "recoverFromCorruption":
		return runtime.TRUE
	}
	return runtime.NewError(fmt.Sprintf("undefined method: %s::%s", s.ToString(), methodName))
}

func (i *Interpreter) callSplPriorityQueueMethod(s *SplPriorityQueueObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "insert":
		if len(args) < 2 {
			return runtime.NewError("SplPriorityQueue::insert() expects exactly 2 parameters")
		}
		s.insert(args[0], args[1])
		return runtime.TRUE
	case "extract":
		return s.extract()
	case "top":
		return s.top()
	case "count":
		return runtime.NewInt(int64(len(s.elements)))
	case "isEmpty":
		return runtime.NewBool(len(s.elements) == 0)
	case "setExtractFlags":
		if len(args) >= 1 {
			s.extractFlag = args[0].ToInt()
		}
		return runtime.NULL
	case "getExtractFlags":
		return runtime.NewInt(s.extractFlag)
	case "rewind":
		s.position = 0
		return runtime.NULL
	case "current":
		if len(s.elements) == 0 {
			return runtime.FALSE
		}
		result := s.elements[0]
		switch s.extractFlag {
		case 1:
			return result.data
		case 2:
			return result.priority
		case 3:
			arr := runtime.NewArray()
			arr.Set(runtime.NewString("data"), result.data)
			arr.Set(runtime.NewString("priority"), result.priority)
			return arr
		}
		return result.data
	case "key":
		return runtime.NewInt(int64(s.position))
	case "next":
		s.position++
		return runtime.NULL
	case "valid":
		return runtime.NewBool(s.position < len(s.elements) && len(s.elements) > 0)
	case "isCorrupted":
		return runtime.FALSE
	case "recoverFromCorruption":
		return runtime.TRUE
	}
	return runtime.NewError(fmt.Sprintf("undefined method: SplPriorityQueue::%s", methodName))
}

func (i *Interpreter) callSplObjectStorageMethod(s *SplObjectStorageObject, methodName string, args []runtime.Value) runtime.Value {
	switch methodName {
	case "attach":
		if len(args) < 1 {
			return runtime.NewError("SplObjectStorage::attach() expects at least 1 parameter")
		}
		obj := args[0]
		hash := fmt.Sprintf("%p", obj)
		if _, exists := s.objects[hash]; !exists {
			s.keys = append(s.keys, hash)
		}
		s.objects[hash] = obj
		if len(args) >= 2 {
			s.infos[hash] = args[1]
		} else {
			s.infos[hash] = runtime.NULL
		}
		return runtime.NULL
	case "detach":
		if len(args) < 1 {
			return runtime.NULL
		}
		hash := fmt.Sprintf("%p", args[0])
		delete(s.objects, hash)
		delete(s.infos, hash)
		for i, k := range s.keys {
			if k == hash {
				s.keys = append(s.keys[:i], s.keys[i+1:]...)
				break
			}
		}
		return runtime.NULL
	case "contains":
		if len(args) < 1 {
			return runtime.FALSE
		}
		hash := fmt.Sprintf("%p", args[0])
		_, exists := s.objects[hash]
		return runtime.NewBool(exists)
	case "count":
		return runtime.NewInt(int64(len(s.objects)))
	case "getInfo":
		if s.position >= 0 && s.position < len(s.keys) {
			hash := s.keys[s.position]
			if info, ok := s.infos[hash]; ok {
				return info
			}
		}
		return runtime.NULL
	case "setInfo":
		if len(args) >= 1 && s.position >= 0 && s.position < len(s.keys) {
			hash := s.keys[s.position]
			s.infos[hash] = args[0]
		}
		return runtime.NULL
	case "rewind":
		s.position = 0
		return runtime.NULL
	case "current":
		if s.position >= 0 && s.position < len(s.keys) {
			hash := s.keys[s.position]
			if obj, ok := s.objects[hash]; ok {
				return obj
			}
		}
		return runtime.FALSE
	case "key":
		return runtime.NewInt(int64(s.position))
	case "next":
		s.position++
		return runtime.NULL
	case "valid":
		return runtime.NewBool(s.position >= 0 && s.position < len(s.keys))
	case "getHash":
		if len(args) < 1 {
			return runtime.NewError("SplObjectStorage::getHash() expects exactly 1 parameter")
		}
		return runtime.NewString(fmt.Sprintf("%p", args[0]))
	case "offsetExists":
		if len(args) < 1 {
			return runtime.FALSE
		}
		hash := fmt.Sprintf("%p", args[0])
		_, exists := s.objects[hash]
		return runtime.NewBool(exists)
	case "offsetGet":
		if len(args) < 1 {
			return runtime.NULL
		}
		hash := fmt.Sprintf("%p", args[0])
		if info, ok := s.infos[hash]; ok {
			return info
		}
		return runtime.NULL
	case "offsetSet":
		if len(args) < 2 {
			return runtime.NULL
		}
		hash := fmt.Sprintf("%p", args[0])
		if _, exists := s.objects[hash]; !exists {
			s.keys = append(s.keys, hash)
			s.objects[hash] = args[0]
		}
		s.infos[hash] = args[1]
		return runtime.NULL
	case "offsetUnset":
		if len(args) < 1 {
			return runtime.NULL
		}
		hash := fmt.Sprintf("%p", args[0])
		delete(s.objects, hash)
		delete(s.infos, hash)
		for i, k := range s.keys {
			if k == hash {
				s.keys = append(s.keys[:i], s.keys[i+1:]...)
				break
			}
		}
		return runtime.NULL
	}
	return runtime.NewError(fmt.Sprintf("undefined method: SplObjectStorage::%s", methodName))
}
