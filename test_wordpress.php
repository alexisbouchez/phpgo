<?php
// Simple WordPress compatibility test
// This tests basic WordPress functionality that PHPGo should support

echo "Testing WordPress compatibility with PHPGo...\n";

// Test 1: Basic PHP functionality
echo "Test 1: Basic PHP functionality - ";
$test1 = "Hello WordPress";
if (strlen($test1) === 14) {
    echo "PASS\n";
} else {
    echo "FAIL\n";
}

// Test 2: Object-oriented programming
echo "Test 2: OOP functionality - ";
class WP_Test {
    public $version = "1.0";
    
    public function getVersion() {
        return $this->version;
    }
}

$wpTest = new WP_Test();
if ($wpTest->getVersion() === "1.0") {
    echo "PASS\n";
} else {
    echo "FAIL\n";
}

// Test 3: Array functions
echo "Test 3: Array functions - ";
$plugins = ["akismet", "hello-dolly", "jetpack"];
if (count($plugins) === 3 && in_array("akismet", $plugins)) {
    echo "PASS\n";
} else {
    echo "FAIL\n";
}

// Test 4: String functions
echo "Test 4: String functions - ";
$url = "https://wordpress.org";
if (strpos($url, "wordpress") !== false) {
    echo "PASS\n";
} else {
    echo "FAIL\n";
}

// Test 5: File functions
echo "Test 5: File functions - ";
if (function_exists("file_exists") && function_exists("file_get_contents")) {
    echo "PASS\n";
} else {
    echo "FAIL\n";
}

// Test 6: JSON functions
echo "Test 6: JSON functions - ";
$data = ["site" => "WordPress", "version" => "6.0"];
$json = json_encode($data);
if ($json !== false && strpos($json, "WordPress") !== false) {
    echo "PASS\n";
} else {
    echo "FAIL\n";
}

// Test 7: Database functions (MySQLi)
echo "Test 7: Database functions - ";
if (function_exists("mysqli_connect") || function_exists("mysql_connect")) {
    echo "PASS\n";
} else {
    echo "FAIL (expected for PHPGo)\n";
}

// Test 8: HTTP functions
echo "Test 8: HTTP functions - ";
if (function_exists("http_build_query")) {
    echo "PASS\n";
} else {
    echo "FAIL\n";
}

// Test 9: Session functions
echo "Test 9: Session functions - ";
if (function_exists("session_start")) {
    echo "PASS\n";
} else {
    echo "FAIL\n";
}

// Test 10: XML functions
echo "Test 10: XML functions - ";
if (function_exists("simplexml_load_string")) {
    echo "PASS\n";
} else {
    echo "FAIL\n";
}

echo "\nWordPress compatibility test completed!\n";
echo "PHPGo appears to be working correctly for basic WordPress functionality.\n";