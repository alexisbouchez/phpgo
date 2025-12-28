<?php
/**
 * Test Superglobals
 */

echo "=== Testing Superglobals ===\n\n";

// Test $_SERVER
echo "\_SERVER variables:\n";
echo "PHP_SELF: " . ($_SERVER['PHP_SELF'] ?? 'NOT SET') . "\n";
echo "SCRIPT_NAME: " . ($_SERVER['SCRIPT_NAME'] ?? 'NOT SET') . "\n";
echo "DOCUMENT_ROOT: " . ($_SERVER['DOCUMENT_ROOT'] ?? 'NOT SET') . "\n";
echo "REQUEST_METHOD: " . ($_SERVER['REQUEST_METHOD'] ?? 'NOT SET') . "\n";
echo "REQUEST_URI: " . ($_SERVER['REQUEST_URI'] ?? 'NOT SET') . "\n";
echo "QUERY_STRING: " . ($_SERVER['QUERY_STRING'] ?? 'NOT SET') . "\n";
echo "HTTP_HOST: " . ($_SERVER['HTTP_HOST'] ?? 'NOT SET') . "\n";
echo "REMOTE_ADDR: " . ($_SERVER['REMOTE_ADDR'] ?? 'NOT SET') . "\n\n";

// Test $_GET
echo "\_GET variables:\n";
if (!empty($_GET)) {
    foreach ($_GET as $key => $value) {
        echo "$key: $value\n";
    }
} else {
    echo "(empty)\n";
}
echo "\n";

// Test $_POST
echo "\_POST variables:\n";
if (!empty($_POST)) {
    foreach ($_POST as $key => $value) {
        echo "$key: $value\n";
    }
} else {
    echo "(empty)\n";
}
echo "\n";

// Test $_COOKIE
echo "\_COOKIE variables:\n";
if (!empty($_COOKIE)) {
    foreach ($_COOKIE as $key => $value) {
        echo "$key: $value\n";
    }
} else {
    echo "(empty)\n";
}
echo "\n";

// Test $_REQUEST
echo "\_REQUEST variables:\n";
if (!empty($_REQUEST)) {
    foreach ($_REQUEST as $key => $value) {
        echo "$key: $value\n";
    }
} else {
    echo "(empty)\n";
}
echo "\n";

// Test $_FILES
echo "\_FILES variables:\n";
if (!empty($_FILES)) {
    foreach ($_FILES as $key => $file) {
        echo "$key: " . $file['name'] . " (" . $file['size'] . " bytes)\n";
    }
} else {
    echo "(empty)\n";
}

echo "\n=== Superglobals Test Complete ===\n";
