<?php
/**
 * ArrayAccess and Iterator Example
 *
 * Demonstrates custom collection classes that behave like arrays.
 */

class Collection implements ArrayAccess, Iterator {
    private array $items = [];
    private int $position = 0;
    private array $keys = [];

    public function __construct(array $items = []) {
        $this->items = $items;
        $this->keys = array_keys($items);
    }

    // ArrayAccess methods
    public function offsetExists(mixed $offset): bool {
        return isset($this->items[$offset]);
    }

    public function offsetGet(mixed $offset): mixed {
        return $this->items[$offset] ?? null;
    }

    public function offsetSet(mixed $offset, mixed $value): void {
        if ($offset === null) {
            $this->items[] = $value;
        } else {
            $this->items[$offset] = $value;
        }
        $this->keys = array_keys($this->items);
    }

    public function offsetUnset(mixed $offset): void {
        unset($this->items[$offset]);
        $this->keys = array_keys($this->items);
    }

    // Iterator methods
    public function current(): mixed {
        return $this->items[$this->keys[$this->position]];
    }

    public function key(): mixed {
        return $this->keys[$this->position];
    }

    public function next(): void {
        $this->position++;
    }

    public function rewind(): void {
        $this->position = 0;
    }

    public function valid(): bool {
        return isset($this->keys[$this->position]);
    }

    // Additional utility methods
    public function count(): int {
        return count($this->items);
    }

    public function map(callable $callback): Collection {
        $result = [];
        foreach ($this->items as $key => $value) {
            $result[$key] = $callback($value, $key);
        }
        return new Collection($result);
    }

    public function filter(callable $callback): Collection {
        $result = [];
        foreach ($this->items as $key => $value) {
            if ($callback($value, $key)) {
                $result[$key] = $value;
            }
        }
        return new Collection($result);
    }

    public function toArray(): array {
        return $this->items;
    }
}

// Usage demonstration
echo "Creating collection with users:\n";
echo str_repeat("-", 40) . "\n";

$users = new Collection([
    'alice' => ['name' => 'Alice', 'age' => 25],
    'bob' => ['name' => 'Bob', 'age' => 30],
    'charlie' => ['name' => 'Charlie', 'age' => 35],
]);

// ArrayAccess - get item
echo "users['alice']: " . $users['alice']['name'] . "\n";

// ArrayAccess - check existence
echo "isset(users['bob']): " . (isset($users['bob']) ? 'true' : 'false') . "\n";
echo "isset(users['dave']): " . (isset($users['dave']) ? 'true' : 'false') . "\n";

// ArrayAccess - set item
$users['dave'] = ['name' => 'Dave', 'age' => 40];
echo "Added dave, count: " . $users->count() . "\n";

// Iterator - foreach loop
echo "\nIterating over users:\n";
echo str_repeat("-", 40) . "\n";
foreach ($users as $key => $user) {
    echo "$key: {$user['name']} (age {$user['age']})\n";
}

// Map operation
echo "\nMapped names:\n";
echo str_repeat("-", 40) . "\n";
$names = $users->map(function($user) {
    return $user['name'];
});
foreach ($names as $key => $name) {
    echo "$key: $name\n";
}

// Filter operation
echo "\nUsers over 30:\n";
echo str_repeat("-", 40) . "\n";
$over30 = $users->filter(function($user) {
    return $user['age'] > 30;
});
foreach ($over30 as $key => $user) {
    echo "$key: {$user['name']} (age {$user['age']})\n";
}
