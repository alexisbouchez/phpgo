<?php
// PHPGo WordPress Handler
// This script acts as a bridge between Apache and PHPGo

// Get the requested PHP file
$requested_file = $_SERVER['DOCUMENT_ROOT'] . $_SERVER['REQUEST_URI'];

// Check if the file exists and is a PHP file
if (file_exists($requested_file) && pathinfo($requested_file, PATHINFO_EXTENSION) === 'php') {
    // Set up environment for PHPGo
    putenv('PHPGO_REQUEST_URI=' . $_SERVER['REQUEST_URI']);
    putenv('PHPGO_DOCUMENT_ROOT=' . $_SERVER['DOCUMENT_ROOT']);
    
    // Forward all server variables
    foreach ($_SERVER as $key => $value) {
        putenv('PHPGO_SERVER_' . $key . '=' . $value);
    }
    
    // Forward all GET/POST/COOKIE data
    if (!empty($_GET)) {
        putenv('PHPGO_GET=' . json_encode($_GET));
    }
    if (!empty($_POST)) {
        putenv('PHPGO_POST=' . json_encode($_POST));
    }
    if (!empty($_COOKIE)) {
        putenv('PHPGO_COOKIE=' . json_encode($_COOKIE));
    }
    
    // Execute the PHP file using PHPGo
    $command = '/app/phpgo ' . escapeshellarg($requested_file);
    
    // Capture output
    ob_start();
    system($command, $return_code);
    $output = ob_get_clean();
    
    // Handle headers if needed
    if (strpos($output, 'HTTP/') === 0) {
        // Extract headers and body
        list($headers, $body) = explode("\r\n\r\n", $output, 2);
        $header_lines = explode("\r\n", $headers);
        
        foreach ($header_lines as $header) {
            if (!empty($header)) {
                header($header, true);
            }
        }
        
        echo $body;
    } else {
        // No headers, just output the content
        echo $output;
    }
    
    exit($return_code);
} else {
    // File doesn't exist or isn't PHP, let Apache handle it normally
    return false;
}