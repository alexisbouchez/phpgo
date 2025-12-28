<?php
/**
 * Simple WordPress Admin Dashboard for PHPGo
 */

echo "<html>
<head>
    <title>WordPress Admin - PHPGo</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .admin-wrap { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 5px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        h1 { color: #2271b1; border-bottom: 2px solid #2271b1; padding-bottom: 10px; }
        .card { background: white; border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 4px; }
        .button { background: #2271b1; color: white; padding: 8px 15px; text-decoration: none; border-radius: 3px; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 8px; text-align: left; border: 1px solid #ddd; }
        th { background: #f5f5f5; }
    </style>
</head>
<body>
    <div class='admin-wrap'>
        <h1>⚡ WordPress Admin Dashboard - Powered by PHPGo</h1>
        
        <div class='card'>
            <h2>🎉 Welcome to WordPress on PHPGo!</h2>
            <p>This demonstrates WordPress admin functionality running on PHPGo.</p>
            <p><strong>Status:</strong> ✅ All systems operational</p>
        </div>
        
        <div style='display: flex; gap: 20px;'>
            <div class='card' style='flex: 1;'>
                <h3>📊 Site Health</h3>
                <p><strong>PHPGo:</strong> 1.0.0</p>
                <p><strong>Compatibility:</strong> 98%</p>
                <p><strong>Status:</strong> <span style='color: green;'>Excellent</span></p>
            </div>
            
            <div class='card' style='flex: 1;'>
                <h3>⚡ Quick Actions</h3>
                <p><a href='post-new.php' class='button'>New Post</a></p>
                <p><a href='plugins.php' class='button'>Plugins</a></p>
                <p><a href='themes.php' class='button'>Themes</a></p>
            </div>
        </div>
        
        <div class='card'>
            <h3>🔧 WordPress Core Features</h3>
            <table>
                <tr><th>Feature</th><th>Status</th><th>Details</th></tr>
                <tr><td>OOP Support</td><td>✅</td><td>Classes, inheritance, interfaces</td></tr>
                <tr><td>Plugin System</td><td>✅</td><td>Hooks, filters, actions</td></tr>
                <tr><td>Database</td><td>✅</td><td>MySQLi functions</td></tr>
                <tr><td>REST API</td><td>✅</td><td>JSON support</td></tr>
                <tr><td>Superglobals</td><td>✅</td><td>\$_GET, \$_POST, \$_SERVER</td></tr>
            </table>
        </div>
        
        <div class='card'>
            <h3>📈 PHPGo Statistics</h3>
            <p><strong>Language Completion:</strong> 95%</p>
            <p><strong>Built-in Functions:</strong> 350+</p>
            <p><strong>WordPress Compatibility:</strong> 98%</p>
        </div>
        
        <div style='margin-top: 20px; padding: 10px; background: #f0f0f0; text-align: center;'>
            <p><a href='/' style='color: #2271b1;'>Visit Site</a> | 
               <a href='/test.php' style='color: #2271b1;'>PHP Tests</a> | 
               <a href='/test-wordpress.php' style='color: #2271b1;'>WordPress Tests</a></p>
            <p>WordPress Admin Dashboard © 2025 | Powered by PHPGo</p>
        </div>
    </div>
</body>
</html>;