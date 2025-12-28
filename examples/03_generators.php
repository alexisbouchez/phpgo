<?php
/**
 * Generators Example
 *
 * Demonstrates yield, yield from, and practical generator use cases.
 */

// Simple range generator
function xrange(int $start, int $end, int $step = 1) {
    for ($i = $start; $i <= $end; $i += $step) {
        yield $i;
    }
}

echo "Range 1-5:\n";
foreach (xrange(1, 5) as $num) {
    echo $num . " ";
}
echo "\n\n";

// Generator with keys
function keyValuePairs() {
    yield "name" => "John";
    yield "age" => 30;
    yield "city" => "NYC";
}

echo "Key-value pairs:\n";
foreach (keyValuePairs() as $key => $value) {
    echo "$key: $value\n";
}
echo "\n";

// Fibonacci generator
function fibonacci(int $limit) {
    $a = 0;
    $b = 1;
    while ($a <= $limit) {
        yield $a;
        $temp = $a;
        $a = $b;
        $b = $temp + $b;
    }
}

echo "Fibonacci up to 100:\n";
foreach (fibonacci(100) as $num) {
    echo $num . " ";
}
echo "\n\n";

// Yield from - delegating to another generator
function inner() {
    yield 1;
    yield 2;
    yield 3;
}

function outer() {
    yield 0;
    yield from inner();
    yield 4;
}

echo "Yield from (0, inner 1-3, 4):\n";
foreach (outer() as $num) {
    echo $num . " ";
}
echo "\n\n";

// Practical: Reading chunks of data
function chunkData(array $data, int $chunkSize) {
    $chunk = [];
    foreach ($data as $item) {
        $chunk[] = $item;
        if (count($chunk) === $chunkSize) {
            yield $chunk;
            $chunk = [];
        }
    }
    if (!empty($chunk)) {
        yield $chunk;
    }
}

$items = range(1, 10);
echo "Chunked data (size 3):\n";
foreach (chunkData($items, 3) as $chunk) {
    echo "[" . implode(", ", $chunk) . "]\n";
}
