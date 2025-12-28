<?php
/**
 * Output Buffering Example
 *
 * Demonstrates ob_start, ob_get_contents, ob_end_clean, etc.
 * Useful for capturing output, creating templates, and more.
 */

// Basic output buffering
echo "=== Basic Output Buffering ===\n\n";

ob_start();
echo "This is captured in the buffer\n";
echo "So is this line\n";
$captured = ob_get_contents();
ob_end_clean();

echo "Captured output:\n";
echo str_repeat("-", 30) . "\n";
echo $captured;
echo str_repeat("-", 30) . "\n\n";

// Using ob_get_clean (combines get_contents + end_clean)
echo "=== Using ob_get_clean ===\n\n";

ob_start();
echo "Line 1\n";
echo "Line 2\n";
echo "Line 3\n";
$output = ob_get_clean();

echo "Total lines: " . count(explode("\n", trim($output))) . "\n";
echo "Character count: " . strlen($output) . "\n\n";

// Nested buffering
echo "=== Nested Buffering ===\n\n";

ob_start();
echo "Outer buffer content\n";

ob_start();
echo "Inner buffer content\n";
$inner = ob_get_clean();

echo "After inner: level = " . ob_get_level() . "\n";
$outer = ob_get_clean();

echo "Inner captured: " . trim($inner) . "\n";
echo "Outer captured: " . trim($outer) . "\n\n";

// Practical example: Simple template rendering
echo "=== Simple Template Engine ===\n\n";

function render(string $template, array $data): string {
    ob_start();

    // Extract variables into local scope
    extract($data);

    // Simple template processing (replacing {{var}} patterns)
    $output = $template;
    foreach ($data as $key => $value) {
        if (is_scalar($value)) {
            $output = str_replace("{{" . $key . "}}", $value, $output);
        }
    }

    echo $output;
    return ob_get_clean();
}

$template = "Hello, {{name}}!\nYou have {{count}} new messages.\nYour role: {{role}}\n";

$html = render($template, [
    'name' => 'Alice',
    'count' => 5,
    'role' => 'Administrator'
]);

echo "Rendered template:\n";
echo str_repeat("-", 30) . "\n";
echo $html;
echo str_repeat("-", 30) . "\n\n";

// Practical example: Generating formatted output
echo "=== Formatted Table Generator ===\n\n";

function generateTable(array $headers, array $rows): string {
    ob_start();

    // Calculate column widths
    $widths = [];
    foreach ($headers as $i => $header) {
        $widths[$i] = strlen($header);
    }
    foreach ($rows as $row) {
        foreach ($row as $i => $cell) {
            $widths[$i] = max($widths[$i] ?? 0, strlen((string)$cell));
        }
    }

    // Generate separator
    $separator = "+";
    foreach ($widths as $width) {
        $separator .= str_repeat("-", $width + 2) . "+";
    }

    // Print table
    echo $separator . "\n";

    // Headers
    echo "|";
    foreach ($headers as $i => $header) {
        echo " " . str_pad($header, $widths[$i]) . " |";
    }
    echo "\n";
    echo $separator . "\n";

    // Rows
    foreach ($rows as $row) {
        echo "|";
        foreach ($row as $i => $cell) {
            echo " " . str_pad((string)$cell, $widths[$i]) . " |";
        }
        echo "\n";
    }
    echo $separator . "\n";

    return ob_get_clean();
}

$table = generateTable(
    ['ID', 'Name', 'Email', 'Status'],
    [
        [1, 'Alice', 'alice@example.com', 'Active'],
        [2, 'Bob', 'bob@example.com', 'Inactive'],
        [3, 'Charlie', 'charlie@example.com', 'Active'],
    ]
);

echo $table;
