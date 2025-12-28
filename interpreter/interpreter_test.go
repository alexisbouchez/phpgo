package interpreter

import (
	"strings"
	"testing"

	"github.com/alexisbouchez/phpgo/runtime"
)

// Helper to run PHP code and get result
func eval(input string) runtime.Value {
	interp := New()
	return interp.Eval(input)
}

// Helper to run PHP code and capture output
func evalOutput(input string) string {
	interp := New()
	interp.Eval(input)
	return interp.Output()
}

// ----------------------------------------------------------------------------
// Literals

func TestEvalIntegerLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php 5;`, 5},
		{`<?php 10;`, 10},
		{`<?php 0;`, 0},
		{`<?php -5;`, -5},
		{`<?php 123456789;`, 123456789},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

func TestEvalFloatLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`<?php 5.5;`, 5.5},
		{`<?php 0.1;`, 0.1},
		{`<?php 3.14159;`, 3.14159},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testFloatValue(t, result, tt.expected)
	}
}

func TestEvalStringLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`<?php "hello";`, "hello"},
		{`<?php 'world';`, "world"},
		{`<?php "";`, ""},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testStringValue(t, result, tt.expected)
	}
}

func TestEvalBoolLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`<?php true;`, true},
		{`<?php false;`, false},
		{`<?php TRUE;`, true},
		{`<?php FALSE;`, false},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testBoolValue(t, result, tt.expected)
	}
}

func TestEvalNull(t *testing.T) {
	tests := []string{
		`<?php null;`,
		`<?php NULL;`,
	}

	for _, input := range tests {
		result := eval(input)
		if _, ok := result.(*runtime.Null); !ok {
			t.Errorf("expected Null, got %T", result)
		}
	}
}

// ----------------------------------------------------------------------------
// Variables

func TestEvalVariable(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $x = 5; $x;`, 5},
		{`<?php $x = 5; $y = $x; $y;`, 5},
		{`<?php $x = 5; $y = 10; $x + $y;`, 15},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Arithmetic

func TestEvalArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php 1 + 2;`, 3},
		{`<?php 5 - 3;`, 2},
		{`<?php 3 * 4;`, 12},
		{`<?php 10 / 2;`, 5},
		{`<?php 10 % 3;`, 1},
		{`<?php 2 ** 3;`, 8},
		{`<?php -5 + 10;`, 5},
		{`<?php (1 + 2) * 3;`, 9},
		{`<?php 2 + 3 * 4;`, 14},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

func TestEvalFloatArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`<?php 1.5 + 2.5;`, 4.0},
		{`<?php 5.0 / 2.0;`, 2.5},
		{`<?php 3.0 * 1.5;`, 4.5},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testFloatValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// String operations

func TestEvalStringConcat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`<?php "hello" . " " . "world";`, "hello world"},
		{`<?php "a" . "b" . "c";`, "abc"},
		{`<?php $x = "foo"; $x . "bar";`, "foobar"},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testStringValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Comparison

func TestEvalComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`<?php 1 == 1;`, true},
		{`<?php 1 == 2;`, false},
		{`<?php 1 != 2;`, true},
		{`<?php 1 < 2;`, true},
		{`<?php 2 > 1;`, true},
		{`<?php 1 <= 1;`, true},
		{`<?php 1 >= 1;`, true},
		{`<?php 1 === 1;`, true},
		{`<?php 1 === "1";`, false},
		{`<?php 1 == "1";`, true},
		{`<?php 1 !== "1";`, true},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testBoolValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Logical

func TestEvalLogical(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`<?php true && true;`, true},
		{`<?php true && false;`, false},
		{`<?php false || true;`, true},
		{`<?php false || false;`, false},
		{`<?php !true;`, false},
		{`<?php !false;`, true},
		{`<?php true and true;`, true},
		{`<?php true or false;`, true},
		{`<?php true xor false;`, true},
		{`<?php true xor true;`, false},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testBoolValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Increment/Decrement

func TestEvalIncDec(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $x = 5; ++$x;`, 6},
		{`<?php $x = 5; --$x;`, 4},
		{`<?php $x = 5; $x++;`, 5},  // Post-increment returns original
		{`<?php $x = 5; $x--; $x;`, 4},
		{`<?php $x = 5; ++$x; ++$x;`, 7},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Compound assignment

func TestEvalCompoundAssignment(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $x = 5; $x += 3; $x;`, 8},
		{`<?php $x = 10; $x -= 4; $x;`, 6},
		{`<?php $x = 3; $x *= 4; $x;`, 12},
		{`<?php $x = 20; $x /= 4; $x;`, 5},
		{`<?php $x = 10; $x %= 3; $x;`, 1},
		{`<?php $x = 2; $x **= 3; $x;`, 8},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Ternary and null coalesce

func TestEvalTernary(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php true ? 1 : 2;`, 1},
		{`<?php false ? 1 : 2;`, 2},
		{`<?php 1 > 0 ? 10 : 20;`, 10},
		{`<?php 1 < 0 ? 10 : 20;`, 20},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

func TestEvalElvis(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php 5 ?: 10;`, 5},
		{`<?php 0 ?: 10;`, 10},
		{`<?php "" ?: 10;`, 10},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

func TestEvalNullCoalesce(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php null ?? 10;`, 10},
		{`<?php 5 ?? 10;`, 5},
		{`<?php $undefined ?? 10;`, 10},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Arrays

func TestEvalArrayLiteral(t *testing.T) {
	input := `<?php [1, 2, 3];`
	result := eval(input)

	arr, ok := result.(*runtime.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", result)
	}

	if arr.Len() != 3 {
		t.Errorf("expected 3 elements, got %d", arr.Len())
	}
}

func TestEvalArrayAccess(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $arr = [1, 2, 3]; $arr[0];`, 1},
		{`<?php $arr = [1, 2, 3]; $arr[2];`, 3},
		{`<?php $arr = ["a" => 1, "b" => 2]; $arr["a"];`, 1},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

func TestEvalArrayAssignment(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $arr = []; $arr[0] = 5; $arr[0];`, 5},
		{`<?php $arr = [1, 2]; $arr[1] = 10; $arr[1];`, 10},
		{`<?php $arr = []; $arr[] = 1; $arr[] = 2; $arr[1];`, 2},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// If statements

func TestEvalIfStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $x = 10; if ($x > 5) { $x = 1; } $x;`, 1},
		{`<?php $x = 10; if ($x < 5) { $x = 1; } $x;`, 10},
		{`<?php $x = 10; if ($x > 5) { $x = 1; } else { $x = 2; } $x;`, 1},
		{`<?php $x = 3; if ($x > 5) { $x = 1; } else { $x = 2; } $x;`, 2},
		{`<?php $x = 5; if ($x > 10) { $x = 1; } elseif ($x > 3) { $x = 2; } else { $x = 3; } $x;`, 2},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// While loop

func TestEvalWhileLoop(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $x = 0; while ($x < 5) { $x++; } $x;`, 5},
		{`<?php $sum = 0; $i = 1; while ($i <= 5) { $sum += $i; $i++; } $sum;`, 15},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// For loop

func TestEvalForLoop(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $sum = 0; for ($i = 1; $i <= 5; $i++) { $sum += $i; } $sum;`, 15},
		{`<?php $x = 0; for ($i = 0; $i < 10; $i++) { $x++; } $x;`, 10},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Foreach loop

func TestEvalForeachLoop(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $sum = 0; foreach ([1, 2, 3, 4, 5] as $v) { $sum += $v; } $sum;`, 15},
		{`<?php $arr = ["a" => 1, "b" => 2]; $sum = 0; foreach ($arr as $k => $v) { $sum += $v; } $sum;`, 3},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Break and continue

func TestEvalBreak(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $x = 0; while (true) { $x++; if ($x >= 5) break; } $x;`, 5},
		{`<?php $x = 0; for ($i = 0; $i < 10; $i++) { if ($i == 5) break; $x++; } $x;`, 5},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

func TestEvalContinue(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $sum = 0; for ($i = 1; $i <= 5; $i++) { if ($i == 3) continue; $sum += $i; } $sum;`, 12},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Switch

func TestEvalSwitch(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $x = 2; switch ($x) { case 1: $r = 10; break; case 2: $r = 20; break; default: $r = 0; } $r;`, 20},
		{`<?php $x = 5; switch ($x) { case 1: $r = 10; break; default: $r = 0; } $r;`, 0},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Functions

func TestEvalFunctionDeclaration(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php function add($a, $b) { return $a + $b; } add(2, 3);`, 5},
		{`<?php function factorial($n) { if ($n <= 1) return 1; return $n * factorial($n - 1); } factorial(5);`, 120},
		{`<?php function greet($name = "World") { return $name; } greet();`, 0}, // Will be string
	}

	for i, tt := range tests {
		result := eval(tt.input)
		if i == 2 {
			testStringValue(t, result, "World")
		} else {
			testIntegerValue(t, result, tt.expected)
		}
	}
}

func TestEvalClosure(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $add = function($a, $b) { return $a + $b; }; $add(2, 3);`, 5},
		{`<?php $x = 10; $add = function($a) use ($x) { return $a + $x; }; $add(5);`, 15},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

func TestEvalArrowFunction(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $add = fn($a, $b) => $a + $b; $add(2, 3);`, 5},
		{`<?php $x = 10; $add = fn($a) => $a + $x; $add(5);`, 15},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Echo and print

func TestEvalEcho(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`<?php echo "hello";`, "hello"},
		{`<?php echo "hello", " ", "world";`, "hello world"},
		{`<?php echo 42;`, "42"},
		{`<?php $x = "test"; echo $x;`, "test"},
	}

	for _, tt := range tests {
		output := evalOutput(tt.input)
		if output != tt.expected {
			t.Errorf("input %q: expected output %q, got %q", tt.input, tt.expected, output)
		}
	}
}

func TestEvalPrint(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`<?php print "hello";`, "hello"},
		{`<?php print 42;`, "42"},
	}

	for _, tt := range tests {
		output := evalOutput(tt.input)
		if output != tt.expected {
			t.Errorf("input %q: expected output %q, got %q", tt.input, tt.expected, output)
		}
	}
}

// ----------------------------------------------------------------------------
// Classes

func TestEvalClassDeclaration(t *testing.T) {
	input := `<?php
class Point {
    public $x;
    public $y;

    public function __construct($x, $y) {
        $this->x = $x;
        $this->y = $y;
    }

    public function sum() {
        return $this->x + $this->y;
    }
}

$p = new Point(3, 4);
$p->sum();
`
	result := eval(input)
	testIntegerValue(t, result, 7)
}

func TestEvalClassInheritance(t *testing.T) {
	input := `<?php
class Animal {
    public function speak() {
        return "...";
    }
}

class Dog extends Animal {
    public function speak() {
        return "Woof!";
    }
}

$dog = new Dog();
$dog->speak();
`
	result := eval(input)
	testStringValue(t, result, "Woof!")
}

func TestEvalStaticMembers(t *testing.T) {
	input := `<?php
class Counter {
    public static $count = 0;

    public static function increment() {
        self::$count++;
        return self::$count;
    }
}

Counter::increment();
Counter::increment();
Counter::$count;
`
	result := eval(input)
	testIntegerValue(t, result, 2)
}

// ----------------------------------------------------------------------------
// Match expression

func TestEvalMatch(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php $x = 2; match($x) { 1 => 10, 2 => 20, default => 0 };`, 20},
		{`<?php match(true) { true => 1, false => 0 };`, 1},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Try/catch

func TestEvalTryCatch(t *testing.T) {
	input := `<?php
try {
    throw new Exception("error");
    $x = 1;
} catch (Exception $e) {
    $x = 2;
}
$x;
`
	result := eval(input)
	testIntegerValue(t, result, 2)
}

// ----------------------------------------------------------------------------
// Built-in functions

func TestEvalBuiltinStrlen(t *testing.T) {
	input := `<?php strlen("hello");`
	result := eval(input)
	testIntegerValue(t, result, 5)
}

func TestEvalBuiltinCount(t *testing.T) {
	input := `<?php count([1, 2, 3, 4, 5]);`
	result := eval(input)
	testIntegerValue(t, result, 5)
}

func TestEvalBuiltinArrayPush(t *testing.T) {
	input := `<?php $arr = [1, 2]; array_push($arr, 3); count($arr);`
	result := eval(input)
	testIntegerValue(t, result, 3)
}

func TestEvalBuiltinIsset(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`<?php $x = 1; isset($x);`, true},
		{`<?php isset($undefined);`, false},
		{`<?php $x = null; isset($x);`, false},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testBoolValue(t, result, tt.expected)
	}
}

func TestEvalBuiltinEmpty(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`<?php empty("");`, true},
		{`<?php empty(0);`, true},
		{`<?php empty([]);`, true},
		{`<?php empty("hello");`, false},
		{`<?php empty([1]);`, false},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testBoolValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Complex program

func TestEvalComplexProgram(t *testing.T) {
	input := `<?php
function fibonacci($n) {
    if ($n <= 1) {
        return $n;
    }
    return fibonacci($n - 1) + fibonacci($n - 2);
}

$result = 0;
for ($i = 0; $i < 10; $i++) {
    $result += fibonacci($i);
}
$result;
`
	result := eval(input)
	// fib(0)+fib(1)+...+fib(9) = 0+1+1+2+3+5+8+13+21+34 = 88
	testIntegerValue(t, result, 88)
}

// ----------------------------------------------------------------------------
// New Built-in Functions Tests

func TestEvalBuiltinPregMatch(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`<?php preg_match("/hello/", "hello world");`, 1},
		{`<?php preg_match("/foo/", "hello world");`, 0},
		{`<?php preg_match("/\d+/", "abc123def");`, 1},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testIntegerValue(t, result, tt.expected)
	}
}

func TestEvalBuiltinPregReplace(t *testing.T) {
	input := `<?php preg_replace("/world/", "PHP", "hello world");`
	result := eval(input)
	testStringValue(t, result, "hello PHP")
}

func TestEvalBuiltinJsonEncode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`<?php json_encode(["a" => 1, "b" => 2]);`, `{"a":1,"b":2}`},
		{`<?php json_encode([1, 2, 3]);`, `[1,2,3]`},
		{`<?php json_encode("hello");`, `"hello"`},
		{`<?php json_encode(42);`, `42`},
		{`<?php json_encode(true);`, `true`},
		{`<?php json_encode(null);`, `null`},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testStringValue(t, result, tt.expected)
	}
}

func TestEvalBuiltinJsonDecode(t *testing.T) {
	input := `<?php $data = json_decode('{"name":"John","age":30}', true); $data["name"];`
	result := eval(input)
	testStringValue(t, result, "John")
}

func TestEvalBuiltinMd5(t *testing.T) {
	input := `<?php md5("hello");`
	result := eval(input)
	testStringValue(t, result, "5d41402abc4b2a76b9719d911017c592")
}

func TestEvalBuiltinSha1(t *testing.T) {
	input := `<?php sha1("hello");`
	result := eval(input)
	testStringValue(t, result, "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d")
}

func TestEvalBuiltinBase64(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`<?php base64_encode("hello");`, "aGVsbG8="},
		{`<?php base64_decode("aGVsbG8=");`, "hello"},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testStringValue(t, result, tt.expected)
	}
}

func TestEvalBuiltinTime(t *testing.T) {
	input := `<?php time();`
	result := eval(input)
	if _, ok := result.(*runtime.Int); !ok {
		t.Errorf("expected Int, got %T", result)
	}
}

func TestEvalBuiltinDate(t *testing.T) {
	input := `<?php date("Y");`
	result := eval(input)
	if _, ok := result.(*runtime.String); !ok {
		t.Errorf("expected String, got %T", result)
	}
}

func TestEvalBuiltinStrContains(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`<?php str_contains("hello world", "world");`, true},
		{`<?php str_contains("hello world", "foo");`, false},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testBoolValue(t, result, tt.expected)
	}
}

func TestEvalBuiltinStrStartsWith(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`<?php str_starts_with("hello world", "hello");`, true},
		{`<?php str_starts_with("hello world", "world");`, false},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testBoolValue(t, result, tt.expected)
	}
}

func TestEvalBuiltinStrEndsWith(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`<?php str_ends_with("hello world", "world");`, true},
		{`<?php str_ends_with("hello world", "hello");`, false},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testBoolValue(t, result, tt.expected)
	}
}

func TestEvalBuiltinHtmlspecialchars(t *testing.T) {
	input := `<?php htmlspecialchars("<div>Hello & World</div>");`
	result := eval(input)
	testStringValue(t, result, "&lt;div&gt;Hello &amp; World&lt;/div&gt;")
}

func TestEvalBuiltinArrayCombine(t *testing.T) {
	input := `<?php $arr = array_combine(["a", "b", "c"], [1, 2, 3]); $arr["b"];`
	result := eval(input)
	testIntegerValue(t, result, 2)
}

func TestEvalBuiltinArrayChunk(t *testing.T) {
	input := `<?php $chunks = array_chunk([1, 2, 3, 4, 5], 2); count($chunks);`
	result := eval(input)
	testIntegerValue(t, result, 3)
}

func TestEvalBuiltinArrayDiff(t *testing.T) {
	input := `<?php $diff = array_diff([1, 2, 3, 4], [2, 4]); count($diff);`
	result := eval(input)
	testIntegerValue(t, result, 2)
}

func TestEvalBuiltinArrayFill(t *testing.T) {
	input := `<?php $arr = array_fill(0, 3, "x"); count($arr);`
	result := eval(input)
	testIntegerValue(t, result, 3)
}

func TestEvalBuiltinNumberFormat(t *testing.T) {
	input := `<?php number_format(1234567.891, 2);`
	result := eval(input)
	testStringValue(t, result, "1,234,567.89")
}

func TestEvalBuiltinDirname(t *testing.T) {
	input := `<?php dirname("/path/to/file.txt");`
	result := eval(input)
	testStringValue(t, result, "/path/to")
}

func TestEvalBuiltinBasename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`<?php basename("/path/to/file.txt");`, "file.txt"},
		{`<?php basename("/path/to/file.txt", ".txt");`, "file"},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testStringValue(t, result, tt.expected)
	}
}

func TestEvalBuiltinPathinfo(t *testing.T) {
	input := `<?php $info = pathinfo("/path/to/file.txt"); $info["filename"];`
	result := eval(input)
	testStringValue(t, result, "file")
}

func TestEvalBuiltinMath(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`<?php sin(0);`, 0.0},
		{`<?php cos(0);`, 1.0},
		{`<?php exp(0);`, 1.0},
		{`<?php log(1);`, 0.0},
	}

	for _, tt := range tests {
		result := eval(tt.input)
		testFloatValue(t, result, tt.expected)
	}
}

// ----------------------------------------------------------------------------
// Helper functions

func testIntegerValue(t *testing.T, val runtime.Value, expected int64) {
	t.Helper()
	intVal, ok := val.(*runtime.Int)
	if !ok {
		t.Errorf("expected Int, got %T (%v)", val, val)
		return
	}
	if intVal.Value != expected {
		t.Errorf("expected %d, got %d", expected, intVal.Value)
	}
}

func testFloatValue(t *testing.T, val runtime.Value, expected float64) {
	t.Helper()
	floatVal, ok := val.(*runtime.Float)
	if !ok {
		// Could be an Int
		if intVal, ok := val.(*runtime.Int); ok {
			if float64(intVal.Value) != expected {
				t.Errorf("expected %f, got %d", expected, intVal.Value)
			}
			return
		}
		t.Errorf("expected Float, got %T (%v)", val, val)
		return
	}
	if floatVal.Value != expected {
		t.Errorf("expected %f, got %f", expected, floatVal.Value)
	}
}

func testStringValue(t *testing.T, val runtime.Value, expected string) {
	t.Helper()
	strVal, ok := val.(*runtime.String)
	if !ok {
		t.Errorf("expected String, got %T (%v)", val, val)
		return
	}
	if strVal.Value != expected {
		t.Errorf("expected %q, got %q", expected, strVal.Value)
	}
}

func testBoolValue(t *testing.T, val runtime.Value, expected bool) {
	t.Helper()
	boolVal, ok := val.(*runtime.Bool)
	if !ok {
		t.Errorf("expected Bool, got %T (%v)", val, val)
		return
	}
	if boolVal.Value != expected {
		t.Errorf("expected %v, got %v", expected, boolVal.Value)
	}
}

// Suppress unused import warning
var _ = strings.Contains

// ----------------------------------------------------------------------------
// Extended OOP Tests

func TestEvalClassInheritanceWithConstructor(t *testing.T) {
	input := `<?php
	class Animal {
		public $name;
		public function __construct($name) {
			$this->name = $name;
		}
		public function speak() {
			return "Some sound";
		}
	}
	class Dog extends Animal {
		public function speak() {
			return $this->name . " barks!";
		}
	}
	$dog = new Dog("Rex");
	echo $dog->speak();
	`
	expected := "Rex barks!"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalParentKeyword(t *testing.T) {
	input := `<?php
	class Animal {
		public function speak() {
			return "Animal speaks";
		}
	}
	class Dog extends Animal {
		public function speak() {
			return parent::speak() . " and Dog barks";
		}
	}
	$dog = new Dog();
	echo $dog->speak();
	`
	expected := "Animal speaks and Dog barks"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalMethodDefaults(t *testing.T) {
	input := `<?php
	class Greeter {
		public function greet($name = "World") {
			return "Hello, " . $name . "!";
		}
	}
	$g = new Greeter();
	echo $g->greet();
	echo " ";
	echo $g->greet("PHP");
	`
	expected := "Hello, World! Hello, PHP!"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalTrait(t *testing.T) {
	input := `<?php
	trait Greeting {
		public function greet() {
			return "Hello!";
		}
	}
	class Person {
		use Greeting;
	}
	$p = new Person();
	echo $p->greet();
	`
	expected := "Hello!"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalTraitWithProperty(t *testing.T) {
	input := `<?php
	trait Counter {
		public $count = 0;
		public function increment() {
			$this->count++;
		}
	}
	class MyCounter {
		use Counter;
	}
	$c = new MyCounter();
	$c->increment();
	$c->increment();
	echo $c->count;
	`
	expected := "2"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalInterface(t *testing.T) {
	input := `<?php
	interface Speakable {
		public function speak();
	}
	class Dog implements Speakable {
		public function speak() {
			return "Woof!";
		}
	}
	$d = new Dog();
	echo $d instanceof Speakable ? "yes" : "no";
	`
	expected := "yes"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalMagicCall(t *testing.T) {
	input := `<?php
	class Magic {
		public function __call($name, $args) {
			return "Called " . $name . " with " . count($args) . " args";
		}
	}
	$m = new Magic();
	echo $m->foo(1, 2, 3);
	`
	expected := "Called foo with 3 args"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalMagicGetSet(t *testing.T) {
	input := `<?php
	class PropertyBag {
		private $data = [];
		public function __get($name) {
			return isset($this->data[$name]) ? $this->data[$name] : null;
		}
		public function __set($name, $value) {
			$this->data[$name] = $value;
		}
	}
	$bag = new PropertyBag();
	$bag->foo = "bar";
	echo $bag->foo;
	`
	expected := "bar"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalToString(t *testing.T) {
	input := `<?php
	class Person {
		private $name;
		public function __construct($name) {
			$this->name = $name;
		}
		public function __toString() {
			return "Person: " . $this->name;
		}
	}
	$p = new Person("John");
	echo $p;
	`
	expected := "Person: John"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalStaticProperties(t *testing.T) {
	input := `<?php
	class Counter {
		public static $count = 0;
		public static function increment() {
			self::$count++;
		}
	}
	Counter::increment();
	Counter::increment();
	Counter::increment();
	echo Counter::$count;
	`
	expected := "3"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalClassConstants(t *testing.T) {
	input := `<?php
	class Math {
		const PI = 3.14159;
		const E = 2.71828;
	}
	echo Math::PI . " and " . Math::E;
	`
	expected := "3.14159 and 2.71828"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalInstanceof(t *testing.T) {
	input := `<?php
	class Animal {}
	class Dog extends Animal {}
	$d = new Dog();
	$r1 = $d instanceof Dog ? "1" : "0";
	$r2 = $d instanceof Animal ? "1" : "0";
	echo $r1 . $r2;
	`
	expected := "11"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalVisibilityPublic(t *testing.T) {
	input := `<?php
	class Foo {
		public $value = "public";
		public function getValue() {
			return $this->value;
		}
	}
	$f = new Foo();
	echo $f->value;
	`
	expected := "public"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalSelfKeyword(t *testing.T) {
	input := `<?php
	class Counter {
		private static $count = 0;
		public function increment() {
			self::$count++;
		}
		public function getCount() {
			return self::$count;
		}
	}
	$c = new Counter();
	$c->increment();
	$c->increment();
	echo $c->getCount();
	`
	expected := "2"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalVariadicFunction(t *testing.T) {
	input := `<?php
	function sum(...$nums) {
		$total = 0;
		foreach ($nums as $n) {
			$total += $n;
		}
		return $total;
	}
	echo sum(1, 2, 3, 4, 5);
	`
	expected := "15"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalSpreadOperator(t *testing.T) {
	input := `<?php
	function add($a, $b, $c) {
		return $a + $b + $c;
	}
	$args = [1, 2, 3];
	echo add(...$args);
	`
	expected := "6"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalSpreadInArray(t *testing.T) {
	input := `<?php
	$a = [1, 2];
	$b = [3, 4];
	$c = [...$a, ...$b, 5];
	echo count($c);
	`
	expected := "5"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalListDestructuring(t *testing.T) {
	input := `<?php
	$arr = [1, 2, 3];
	list($a, $b, $c) = $arr;
	echo $a . $b . $c;
	`
	expected := "123"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalShortArrayDestructuring(t *testing.T) {
	input := `<?php
	$arr = [10, 20, 30];
	[$a, $b, $c] = $arr;
	echo $a + $b + $c;
	`
	expected := "60"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalMagicInvoke(t *testing.T) {
	input := `<?php
	class CallableClass {
		private $value;
		public function __construct($v) {
			$this->value = $v;
		}
		public function __invoke($x) {
			return $this->value * $x;
		}
	}
	$obj = new CallableClass(5);
	echo $obj(10);
	`
	expected := "50"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalMagicIsset(t *testing.T) {
	input := `<?php
	class MagicContainer {
		private $data = [];
		public function __set($name, $value) {
			$this->data[$name] = $value;
		}
		public function __get($name) {
			return $this->data[$name];
		}
		public function __isset($name) {
			return isset($this->data[$name]);
		}
	}
	$obj = new MagicContainer();
	$obj->foo = "bar";
	$r1 = isset($obj->foo) ? "1" : "0";
	$r2 = isset($obj->missing) ? "1" : "0";
	echo $r1 . $r2;
	`
	expected := "10"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalMagicUnset(t *testing.T) {
	input := `<?php
	class MagicContainer {
		private $data = [];
		public function __set($name, $value) {
			$this->data[$name] = $value;
		}
		public function __get($name) {
			return isset($this->data[$name]) ? $this->data[$name] : "gone";
		}
		public function __isset($name) {
			return isset($this->data[$name]);
		}
		public function __unset($name) {
			unset($this->data[$name]);
		}
	}
	$obj = new MagicContainer();
	$obj->foo = "bar";
	unset($obj->foo);
	echo $obj->foo;
	`
	expected := "gone"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalAbstractClass(t *testing.T) {
	input := `<?php
	abstract class Animal {
		abstract public function speak();
	}
	class Dog extends Animal {
		public function speak() {
			return "Woof!";
		}
	}
	$dog = new Dog();
	echo $dog->speak();
	`
	expected := "Woof!"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalAbstractClassCannotInstantiate(t *testing.T) {
	input := `<?php
	abstract class Animal {
		abstract public function speak();
	}
	$a = new Animal();
	`
	interp := New()
	result := interp.Eval(input)
	errVal, ok := result.(*runtime.Error)
	if !ok || !strings.Contains(errVal.Message, "cannot instantiate abstract class") {
		t.Errorf("expected error about abstract class, got %v", result)
	}
}

func TestEvalAbstractMethodMustImplement(t *testing.T) {
	input := `<?php
	abstract class Animal {
		abstract public function speak();
	}
	class Dog extends Animal {
		// Missing speak() implementation
	}
	`
	interp := New()
	result := interp.Eval(input)
	errVal, ok := result.(*runtime.Error)
	if !ok || !strings.Contains(errVal.Message, "must implement") {
		t.Errorf("expected error about unimplemented abstract method, got %v", result)
	}
}

func TestEvalFinalClass(t *testing.T) {
	input := `<?php
	final class Singleton {
		public function getValue() {
			return 42;
		}
	}
	$s = new Singleton();
	echo $s->getValue();
	`
	expected := "42"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalFinalClassCannotExtend(t *testing.T) {
	input := `<?php
	final class Singleton {
		public function getValue() {
			return 42;
		}
	}
	class Extended extends Singleton {
	}
	`
	interp := New()
	result := interp.Eval(input)
	errVal, ok := result.(*runtime.Error)
	if !ok || !strings.Contains(errVal.Message, "cannot extend final class") {
		t.Errorf("expected error about final class, got %v", result)
	}
}

func TestEvalFinalMethod(t *testing.T) {
	input := `<?php
	class Base {
		final public function locked() {
			return "locked";
		}
	}
	class Child extends Base {
		public function other() {
			return $this->locked();
		}
	}
	$c = new Child();
	echo $c->other();
	`
	expected := "locked"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalFinalMethodCannotOverride(t *testing.T) {
	input := `<?php
	class Base {
		final public function locked() {
			return "locked";
		}
	}
	class Child extends Base {
		public function locked() {
			return "overridden";
		}
	}
	`
	interp := New()
	result := interp.Eval(input)
	errVal, ok := result.(*runtime.Error)
	if !ok || !strings.Contains(errVal.Message, "cannot override final method") {
		t.Errorf("expected error about final method, got %v", result)
	}
}

func TestEvalNullSafePropertyAccess(t *testing.T) {
	input := `<?php
	class User {
		public $name = "Alice";
	}
	$user = null;
	$name = $user?->name;
	echo $name === null ? "null" : $name;
	`
	expected := "null"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalNullSafePropertyAccessWithValue(t *testing.T) {
	input := `<?php
	class User {
		public $name = "Alice";
	}
	$user = new User();
	$name = $user?->name;
	echo $name;
	`
	expected := "Alice"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalNullSafeMethodCall(t *testing.T) {
	input := `<?php
	class User {
		public function getName() {
			return "Bob";
		}
	}
	$user = null;
	$name = $user?->getName();
	echo $name === null ? "null" : $name;
	`
	expected := "null"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalNullSafeMethodCallWithValue(t *testing.T) {
	input := `<?php
	class User {
		public function getName() {
			return "Bob";
		}
	}
	$user = new User();
	$name = $user?->getName();
	echo $name;
	`
	expected := "Bob"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalNullSafeChained(t *testing.T) {
	input := `<?php
	class Profile {
		public $email = "test@example.com";
	}
	class User {
		public $profile;
		public function __construct() {
			$this->profile = new Profile();
		}
	}
	$user = null;
	$email = $user?->profile?->email;
	echo $email === null ? "null" : $email;
	`
	expected := "null"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalExit(t *testing.T) {
	input := `<?php
	echo "before";
	exit;
	echo "after";
	`
	expected := "before"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalExitWithCode(t *testing.T) {
	input := `<?php
	echo "start";
	exit(1);
	echo "end";
	`
	interp := New()
	result := interp.Eval(input)
	if exitVal, ok := result.(*runtime.Exit); !ok || exitVal.Status != 1 {
		t.Errorf("expected exit with status 1, got %v", result)
	}
	if interp.Output() != "start" {
		t.Errorf("expected output 'start', got %q", interp.Output())
	}
}

func TestEvalDieWithMessage(t *testing.T) {
	input := `<?php
	echo "hello ";
	die("goodbye");
	echo "world";
	`
	expected := "hello goodbye"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalNamedArguments(t *testing.T) {
	input := `<?php
	function greet($name, $greeting = "Hello") {
		return "$greeting, $name!";
	}
	echo greet(name: "World");
	`
	expected := "Hello, World!"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalNamedArgumentsReorder(t *testing.T) {
	input := `<?php
	function sub($a, $b) {
		return $a - $b;
	}
	echo sub(b: 3, a: 10);
	`
	expected := "7"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalNamedArgumentsSkipDefaults(t *testing.T) {
	input := `<?php
	function test($a, $b = 2, $c = 3) {
		return $a + $b + $c;
	}
	echo test(1, c: 10);
	`
	expected := "13"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalNamedArgumentsMethod(t *testing.T) {
	input := `<?php
	class Calculator {
		public function add($x, $y) {
			return $x + $y;
		}
	}
	$calc = new Calculator();
	echo $calc->add(y: 30, x: 12);
	`
	expected := "42"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// ----------------------------------------------------------------------------
// Generators

func TestEvalGeneratorBasic(t *testing.T) {
	input := `<?php
	function gen() {
		yield 1;
		yield 2;
		yield 3;
	}
	foreach (gen() as $v) {
		echo $v;
	}
	`
	expected := "123"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalGeneratorWithKeys(t *testing.T) {
	input := `<?php
	function gen() {
		yield "a" => 1;
		yield "b" => 2;
	}
	foreach (gen() as $k => $v) {
		echo $k . $v;
	}
	`
	expected := "a1b2"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalGeneratorYieldFrom(t *testing.T) {
	input := `<?php
	function inner() {
		yield 2;
		yield 3;
	}
	function outer() {
		yield 1;
		yield from inner();
		yield 4;
	}
	foreach (outer() as $v) {
		echo $v;
	}
	`
	expected := "1234"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalGeneratorYieldFromArray(t *testing.T) {
	input := `<?php
	function gen() {
		yield 1;
		yield from [2, 3, 4];
		yield 5;
	}
	foreach (gen() as $v) {
		echo $v;
	}
	`
	expected := "12345"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// ----------------------------------------------------------------------------
// Constructor Property Promotion

func TestEvalConstructorPropertyPromotion(t *testing.T) {
	input := `<?php
	class Point {
		public function __construct(
			public int $x,
			public int $y
		) {}
	}
	$p = new Point(3, 4);
	echo $p->x . "," . $p->y;
	`
	expected := "3,4"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalConstructorPropertyPromotionMixed(t *testing.T) {
	input := `<?php
	class User {
		public function __construct(
			public string $name,
			private string $email,
			$normalParam
		) {
			echo "Normal: " . $normalParam;
		}
	}
	$u = new User("John", "john@example.com", "test");
	echo " Name: " . $u->name;
	`
	expected := "Normal: test Name: John"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalConstructorPropertyPromotionDefaults(t *testing.T) {
	input := `<?php
	class Config {
		public function __construct(
			public string $env = "production",
			public bool $debug = false
		) {}
	}
	$c = new Config();
	echo $c->env;
	`
	expected := "production"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// ----------------------------------------------------------------------------
// Namespace Resolution

func TestEvalNamespaceBasic(t *testing.T) {
	input := `<?php
	namespace App\Models;

	class User {
		public function getName() {
			return "John";
		}
	}

	$u = new User();
	echo $u->getName();
	`
	expected := "John"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalNamespaceUseClass(t *testing.T) {
	input := `<?php
	namespace App\Models;

	class User {
		public string $name = "Alice";
	}

	namespace App\Controllers;

	use App\Models\User;

	$u = new User();
	echo $u->name;
	`
	expected := "Alice"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalNamespaceUseAlias(t *testing.T) {
	input := `<?php
	namespace App\Models;

	class User {
		public string $name = "Bob";
	}

	namespace App\Controllers;

	use App\Models\User as U;

	$u = new U();
	echo $u->name;
	`
	expected := "Bob"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalNamespaceFunction(t *testing.T) {
	input := `<?php
	namespace App\Helpers;

	function greet($name) {
		return "Hello, " . $name;
	}

	echo greet("World");
	`
	expected := "Hello, World"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalNamespaceFQN(t *testing.T) {
	input := `<?php
	namespace App\Models;

	class User {
		public string $name = "Charlie";
	}

	namespace App\Controllers;

	$u = new \App\Models\User();
	echo $u->name;
	`
	expected := "Charlie"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// ----------------------------------------------------------------------------
// call_user_func / call_user_func_array

func TestEvalCallUserFuncString(t *testing.T) {
	input := `<?php
	function greet($name) {
		return "Hello, " . $name;
	}
	echo call_user_func('greet', 'World');
	`
	expected := "Hello, World"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalCallUserFuncClosure(t *testing.T) {
	input := `<?php
	$fn = function($x, $y) {
		return $x + $y;
	};
	echo call_user_func($fn, 3, 4);
	`
	expected := "7"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalCallUserFuncMethod(t *testing.T) {
	input := `<?php
	class Calculator {
		public function add($a, $b) {
			return $a + $b;
		}
	}
	$calc = new Calculator();
	echo call_user_func([$calc, 'add'], 10, 20);
	`
	expected := "30"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalCallUserFuncStaticMethod(t *testing.T) {
	input := `<?php
	class Math {
		public static function multiply($a, $b) {
			return $a * $b;
		}
	}
	echo call_user_func(['Math', 'multiply'], 5, 6);
	`
	expected := "30"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalCallUserFuncArray(t *testing.T) {
	input := `<?php
	function sum($a, $b, $c) {
		return $a + $b + $c;
	}
	echo call_user_func_array('sum', [1, 2, 3]);
	`
	expected := "6"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEvalCallUserFuncArrayMethod(t *testing.T) {
	input := `<?php
	class Greeter {
		private $prefix;
		public function __construct($prefix) {
			$this->prefix = $prefix;
		}
		public function greet($name) {
			return $this->prefix . " " . $name;
		}
	}
	$g = new Greeter("Hello");
	echo call_user_func_array([$g, 'greet'], ['Alice']);
	`
	expected := "Hello Alice"
	result := evalOutput(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
