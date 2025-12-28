<?php
declare(strict_types=1);

/**
 * Strict Types Example
 *
 * Demonstrates strict type checking with declare(strict_types=1).
 * With strict mode, PHP will throw TypeError for type mismatches.
 */

function addNumbers(int $a, int $b): int {
    return $a + $b;
}

function greet(string $name): string {
    return "Hello, $name!";
}

function calculateAverage(array $numbers): float {
    if (count($numbers) === 0) {
        return 0.0;
    }
    return array_sum($numbers) / count($numbers);
}

// Nullable types
function findUser(?int $id): ?string {
    if ($id === null) {
        return null;
    }
    $users = [
        1 => "Alice",
        2 => "Bob",
        3 => "Charlie"
    ];
    return $users[$id] ?? null;
}

// Union types simulation with multiple checks
function processValue(mixed $value): string {
    if (is_int($value)) {
        return "Integer: $value";
    } elseif (is_string($value)) {
        return "String: $value";
    } elseif (is_array($value)) {
        return "Array with " . count($value) . " elements";
    }
    return "Unknown type";
}

// Valid calls with correct types
echo "Valid operations:\n";
echo str_repeat("-", 30) . "\n";
echo "addNumbers(5, 3) = " . addNumbers(5, 3) . "\n";
echo greet("World") . "\n";
echo "Average of [1,2,3,4,5] = " . calculateAverage([1, 2, 3, 4, 5]) . "\n";
echo "findUser(1) = " . findUser(1) . "\n";
echo "findUser(null) = " . var_export(findUser(null), true) . "\n";

echo "\nProcessing different types:\n";
echo str_repeat("-", 30) . "\n";
echo processValue(42) . "\n";
echo processValue("hello") . "\n";
echo processValue([1, 2, 3]) . "\n";

// Return type enforcement
function divideNumbers(float $a, float $b): float {
    return $a / $b;
}

echo "\nDivision with floats:\n";
echo str_repeat("-", 30) . "\n";
echo "divideNumbers(10.0, 3.0) = " . divideNumbers(10.0, 3.0) . "\n";
