<?php
// Simple PHP test page for PHPGo server
echo "<h1>PHPGo Server Test</h1>";
echo "<p>This page is being served by PHPGo!</p>";

// Test superglobals
echo "<h2>Server Information</h2>";
echo "<p>Request Method: " . htmlspecialchars($_SERVER['REQUEST_METHOD']) . "</p>";
echo "<p>Request URI: " . htmlspecialchars($_SERVER['REQUEST_URI']) . "</p>";
echo "<p>Server Host: " . htmlspecialchars($_SERVER['HTTP_HOST']) . "</p>";

// Test GET parameters
if (!empty($_GET)) {
    echo "<h2>GET Parameters</h2>";
    echo "<pre>";
    print_r($_GET);
    echo "</pre>";
}

// Test POST parameters
if (!empty($_POST)) {
    echo "<h2>POST Parameters</h2>";
    echo "<pre>";
    print_r($_POST);
    echo "</pre>";
}

// Test WordPress-style functionality
echo "<h2>WordPress Compatibility Test</h2>";

// Test class autoloading (WordPress style)
class WP_Test_Class {
    public function greet($name) {
        return "Hello, " . htmlspecialchars($name) . "! Welcome to WordPress on PHPGo!";
    }
}

$test = new WP_Test_Class();
echo "<p>" . $test->greet("Developer") . "</p>";

// Test array functions (WordPress uses these extensively)
$plugins = ['akismet', 'hello-dolly', 'jetpack'];
echo "<p>Plugin count: " . count($plugins) . "</p>";
echo "<p>Has Akismet: " . (in_array('akismet', $plugins) ? 'Yes' : 'No') . "</p>";

// Test JSON (WordPress REST API)
$data = ['site' => 'WordPress', 'phpgo' => 'powered'];
echo "<p>JSON test: " . json_encode($data) . "</p>";

echo "<h2>Success!</h2>";
echo "<p>PHPGo is working correctly and can handle WordPress-style PHP code.</p>";

echo '<p><a href="/test.php?name=PHPGo">Test with GET parameter</a></p>';

echo '<form method="post" action="/test.php">'
    . '<input type="text" name="test_input" placeholder="Enter text">'
    . '<button type="submit">Test POST</button>'
    . '</form>';
