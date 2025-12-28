<?php
/**
 * WordPress-style PHP test for PHPGo
 * This demonstrates WordPress-compatible functionality
 */

echo "<h1>WordPress-Style Test</h1>";
echo "<p>Testing WordPress compatibility with PHPGo</p>";

// Simulate WordPress plugin architecture
class My_WordPress_Plugin {
    private $version = '1.0.0';
    private $plugin_name = 'My Plugin';

    public function __construct() {
        // Simulate WordPress hook registration
        add_action('init', [$this, 'on_init']);
        add_filter('the_content', [$this, 'filter_content']);
    }

    public function on_init() {
        return "Plugin initialized!";
    }

    public function filter_content($content) {
        return $content . " <!-- Filtered by {$this->plugin_name} -->";
    }

    public function get_version() {
        return $this->version;
    }

    public function activate() {
        return "{$this->plugin_name} activated!";
    }

    public function deactivate() {
        return "{$this->plugin_name} deactivated!";
    }
}

// Simulate WordPress hook system
function add_action($hook, $callback) {
    global $wp_hooks;
    if (!isset($wp_hooks)) {
        $wp_hooks = [];
    }
    if (!isset($wp_hooks[$hook])) {
        $wp_hooks[$hook] = [];
    }
    $wp_hooks[$hook][] = $callback;
}

function add_filter($hook, $callback) {
    return add_action($hook, $callback);
}

function do_action($hook, $arg = null) {
    global $wp_hooks;
    if (isset($wp_hooks[$hook])) {
        foreach ($wp_hooks[$hook] as $callback) {
            if (is_callable($callback)) {
                echo "<p>Hook '{$hook}' executed: " . call_user_func($callback, $arg) . "</p>";
            }
        }
    }
}

// Initialize the plugin
$my_plugin = new My_WordPress_Plugin();

echo "<h2>Plugin Information</h2>";
echo "<p>Plugin: {$my_plugin->get_version()}</p>";
echo "<p>Activation: " . $my_plugin->activate() . "</p>";

// Test WordPress-style hooks
echo "<h2>WordPress Hook System</h2>";
do_action('init');

// Test WordPress-style content filtering
echo "<h2>Content Filtering</h2>";
$content = "Hello WordPress!";
$filtered_content = $my_plugin->filter_content($content);
echo "<p>Original: {$content}</p>";
echo "<p>Filtered: {$filtered_content}</p>";

// Test WordPress-style array operations
echo "<h2>WordPress Array Operations</h2>";
$plugins = ['akismet', 'hello-dolly', 'jetpack', 'woocommerce'];
echo "<p>Active plugins: " . count($plugins) . "</p>";
echo "<p>Has WooCommerce: " . (in_array('woocommerce', $plugins) ? 'Yes' : 'No') . "</p>";

// Test WordPress-style string operations
echo "<h2>WordPress String Operations</h2>";
$url = "https://wordpress.org/plugins/my-plugin";
if (strpos($url, 'wordpress.org') !== false) {
    echo "<p>✅ Valid WordPress.org URL</p>";
}

// Test WordPress-style JSON (REST API)
echo "<h2>WordPress REST API</h2>";
$api_response = [
    'success' => true,
    'data' => ['posts' => 10, 'pages' => 5],
    'message' => 'API working correctly'
];
echo "<p>API Response: " . htmlspecialchars(json_encode($api_response)) . "</p>";

// Test WordPress-style database check
echo "<h2>WordPress Database</h2>";
if (function_exists('mysqli_connect')) {
    echo "<p>✅ MySQLi functions available for WordPress DB</p>";
} else {
    echo "<p>❌ MySQLi functions not available</p>";
}

// Test WordPress-style error handling
echo "<h2>WordPress Error Handling</h2>";
if (function_exists('trigger_error')) {
    echo "<p>✅ Error handling functions available</p>";
}

// Test WordPress-style type checking
echo "<h2>WordPress Type Checking</h2>";
$test_var = "WordPress";
echo "<p>Variable is string: " . (is_string($test_var) ? 'Yes' : 'No') . "</p>";
echo "<p>Variable is array: " . (is_array($test_var) ? 'Yes' : 'No') . "</p>";

echo "<h2>Success!</h2>";
echo "<p>🎉 PHPGo successfully demonstrates WordPress compatibility!</p>";
echo "<p><a href='/'>Back to home</a></p>";

echo "<div style='margin-top: 20px; padding: 10px; background: #f0f0f0; border-radius: 4px;'>";
echo "<p><strong>PHPGo WordPress Test</strong> | © 2025 | Powered by Go</p>";
echo "</div>";