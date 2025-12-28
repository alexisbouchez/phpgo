<?php
/**
 * Basic Hello World Example
 *
 * This demonstrates the simplest PHP program.
 */

echo "Hello, World!\n";

// Variables and string interpolation
$name = "PHP Developer";
echo "Welcome, $name!\n";

// Basic arithmetic
$a = 10;
$b = 5;
echo "Sum: " . ($a + $b) . "\n";
echo "Product: " . ($a * $b) . "\n";

// Arrays
$fruits = ["apple", "banana", "orange"];
echo "Fruits: " . implode(", ", $fruits) . "\n";

// Associative arrays
$person = [
    "name" => "John",
    "age" => 30,
    "city" => "New York"
];
echo "Person: " . $person["name"] . " is " . $person["age"] . " years old\n";
