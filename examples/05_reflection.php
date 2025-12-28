<?php
/**
 * Reflection API Example
 *
 * Demonstrates introspection of classes, methods, and properties at runtime.
 */

class Calculator {
    private float $lastResult = 0;

    public function add(float $a, float $b): float {
        $this->lastResult = $a + $b;
        return $this->lastResult;
    }

    public function subtract(float $a, float $b): float {
        $this->lastResult = $a - $b;
        return $this->lastResult;
    }

    public function multiply(float $a, float $b): float {
        $this->lastResult = $a * $b;
        return $this->lastResult;
    }

    protected function divide(float $a, float $b): float {
        if ($b === 0.0) {
            throw new Exception("Division by zero");
        }
        $this->lastResult = $a / $b;
        return $this->lastResult;
    }

    public function getLastResult(): float {
        return $this->lastResult;
    }
}

// Reflect on the class
$reflection = new ReflectionClass(Calculator::class);

echo "Class: " . $reflection->getName() . "\n";
echo str_repeat("=", 50) . "\n\n";

// Properties
echo "Properties:\n";
echo str_repeat("-", 30) . "\n";
foreach ($reflection->getProperties() as $property) {
    $visibility = $property->isPrivate() ? "private" :
                 ($property->isProtected() ? "protected" : "public");
    echo sprintf("  %-12s $%s\n", $visibility, $property->getName());
}

// Methods
echo "\nMethods:\n";
echo str_repeat("-", 30) . "\n";
foreach ($reflection->getMethods() as $method) {
    $visibility = $method->isPrivate() ? "private" :
                 ($method->isProtected() ? "protected" : "public");

    // Get parameters
    $params = [];
    foreach ($method->getParameters() as $param) {
        $paramStr = "";
        if ($param->hasType()) {
            $paramStr .= $param->getType() . " ";
        }
        $paramStr .= "$" . $param->getName();
        if ($param->isOptional() && $param->isDefaultValueAvailable()) {
            $paramStr .= " = " . var_export($param->getDefaultValue(), true);
        }
        $params[] = $paramStr;
    }

    $returnType = $method->hasReturnType() ? ": " . $method->getReturnType() : "";

    echo sprintf("  %-12s %s(%s)%s\n",
        $visibility,
        $method->getName(),
        implode(", ", $params),
        $returnType
    );
}

// Dynamic method invocation
echo "\nDynamic Invocation:\n";
echo str_repeat("-", 30) . "\n";

$calc = new Calculator();
$addMethod = $reflection->getMethod('add');

$result = $addMethod->invoke($calc, 5, 3);
echo "add(5, 3) = $result\n";

// Invoke with named parameters via invokeArgs
$multiplyMethod = $reflection->getMethod('multiply');
$result = $multiplyMethod->invokeArgs($calc, [4, 7]);
echo "multiply(4, 7) = $result\n";

// Accessing private property via reflection
$lastResultProp = $reflection->getProperty('lastResult');
$lastResultProp->setAccessible(true);
echo "lastResult (private): " . $lastResultProp->getValue($calc) . "\n";
