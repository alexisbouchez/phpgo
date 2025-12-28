<?php
/**
 * Serialization Example
 *
 * Demonstrates serialize/unserialize with __sleep and __wakeup magic methods.
 */

class DatabaseConnection {
    private string $host;
    private string $database;
    private string $username;
    private bool $connected = false;

    public function __construct(string $host, string $database, string $username) {
        $this->host = $host;
        $this->database = $database;
        $this->username = $username;
        $this->connect();
    }

    public function connect(): void {
        // Simulate connection
        echo "Connecting to {$this->database}@{$this->host}...\n";
        $this->connected = true;
        echo "Connected!\n";
    }

    public function isConnected(): bool {
        return $this->connected;
    }

    public function query(string $sql): string {
        if (!$this->connected) {
            return "Error: Not connected";
        }
        return "Executed: $sql";
    }

    // Called before serialization - return properties to serialize
    public function __sleep(): array {
        echo "__sleep called - preparing for serialization\n";
        // Don't serialize the connection state
        return ['host', 'database', 'username'];
    }

    // Called after unserialization - reinitialize the object
    public function __wakeup(): void {
        echo "__wakeup called - restoring from serialization\n";
        $this->connect();
    }

    public function getInfo(): string {
        return "{$this->username}@{$this->host}/{$this->database}";
    }
}

class Session {
    public string $id;
    public array $data = [];
    public int $createdAt;
    public int $lastAccess;

    public function __construct(string $id) {
        $this->id = $id;
        $this->createdAt = time();
        $this->lastAccess = time();
    }

    public function set(string $key, mixed $value): void {
        $this->data[$key] = $value;
        $this->lastAccess = time();
    }

    public function get(string $key): mixed {
        return $this->data[$key] ?? null;
    }

    public function __sleep(): array {
        echo "Serializing session {$this->id}\n";
        return ['id', 'data', 'createdAt', 'lastAccess'];
    }

    public function __wakeup(): void {
        echo "Restoring session {$this->id}\n";
        $this->lastAccess = time();
    }
}

// === Database Connection Demo ===
echo "=== Database Connection Serialization ===\n\n";

echo "Creating connection...\n";
$db = new DatabaseConnection("localhost", "myapp", "admin");
echo "Connection info: " . $db->getInfo() . "\n";
echo "Is connected: " . ($db->isConnected() ? "yes" : "no") . "\n\n";

echo "Serializing...\n";
$serialized = serialize($db);
echo "Serialized length: " . strlen($serialized) . " bytes\n\n";

echo "Unserializing...\n";
$restored = unserialize($serialized);
echo "Connection info: " . $restored->getInfo() . "\n";
echo "Is connected: " . ($restored->isConnected() ? "yes" : "no") . "\n";
echo "Query result: " . $restored->query("SELECT * FROM users") . "\n\n";

// === Session Demo ===
echo "=== Session Serialization ===\n\n";

$session = new Session("abc123");
$session->set("user_id", 42);
$session->set("username", "alice");
$session->set("cart", ["item1", "item2", "item3"]);

echo "Original session:\n";
echo "  ID: " . $session->id . "\n";
echo "  User: " . $session->get("username") . "\n";
echo "  Cart items: " . count($session->get("cart")) . "\n\n";

echo "Serializing session...\n";
$sessionData = serialize($session);
echo "Stored " . strlen($sessionData) . " bytes\n\n";

echo "Restoring session...\n";
$restoredSession = unserialize($sessionData);
echo "Restored session:\n";
echo "  ID: " . $restoredSession->id . "\n";
echo "  User: " . $restoredSession->get("username") . "\n";
echo "  Cart items: " . count($restoredSession->get("cart")) . "\n";

// === JSON vs Serialize comparison ===
echo "\n=== JSON vs Serialize ===\n\n";

$data = [
    "name" => "Test",
    "values" => [1, 2, 3, 4, 5],
    "nested" => ["a" => 1, "b" => 2]
];

$jsonEncoded = json_encode($data);
$serialized = serialize($data);

echo "Original data: ";
print_r($data);
echo "\n";
echo "JSON encoded (" . strlen($jsonEncoded) . " bytes): $jsonEncoded\n";
echo "Serialized (" . strlen($serialized) . " bytes): $serialized\n\n";

$jsonDecoded = json_decode($jsonEncoded, true);
$unserialized = unserialize($serialized);

echo "Both methods restore data correctly: ";
echo ($jsonDecoded === $unserialized) ? "true" : "false";
echo "\n";
