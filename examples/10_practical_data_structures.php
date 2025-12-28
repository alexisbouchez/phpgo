<?php
/**
 * Practical Data Structures Example
 *
 * Demonstrates building useful data structures with PHP classes.
 */

// Stack implementation
class Stack {
    private array $items = [];

    public function push(mixed $item): void {
        $this->items[] = $item;
    }

    public function pop(): mixed {
        if ($this->isEmpty()) {
            return null;
        }
        return array_pop($this->items);
    }

    public function peek(): mixed {
        if ($this->isEmpty()) {
            return null;
        }
        return $this->items[count($this->items) - 1];
    }

    public function isEmpty(): bool {
        return count($this->items) === 0;
    }

    public function size(): int {
        return count($this->items);
    }
}

// Queue implementation
class Queue {
    private array $items = [];

    public function enqueue(mixed $item): void {
        $this->items[] = $item;
    }

    public function dequeue(): mixed {
        if ($this->isEmpty()) {
            return null;
        }
        return array_shift($this->items);
    }

    public function front(): mixed {
        return $this->items[0] ?? null;
    }

    public function isEmpty(): bool {
        return count($this->items) === 0;
    }

    public function size(): int {
        return count($this->items);
    }
}

// Linked List Node
class ListNode {
    public mixed $value;
    public ?ListNode $next = null;

    public function __construct(mixed $value) {
        $this->value = $value;
    }
}

// Linked List implementation
class LinkedList {
    private ?ListNode $head = null;
    private int $size = 0;

    public function append(mixed $value): void {
        $node = new ListNode($value);

        if ($this->head === null) {
            $this->head = $node;
        } else {
            $current = $this->head;
            while ($current->next !== null) {
                $current = $current->next;
            }
            $current->next = $node;
        }
        $this->size++;
    }

    public function prepend(mixed $value): void {
        $node = new ListNode($value);
        $node->next = $this->head;
        $this->head = $node;
        $this->size++;
    }

    public function toArray(): array {
        $result = [];
        $current = $this->head;
        while ($current !== null) {
            $result[] = $current->value;
            $current = $current->next;
        }
        return $result;
    }

    public function size(): int {
        return $this->size;
    }
}

// Binary Tree Node
class TreeNode {
    public mixed $value;
    public ?TreeNode $left = null;
    public ?TreeNode $right = null;

    public function __construct(mixed $value) {
        $this->value = $value;
    }
}

// Binary Search Tree
class BinarySearchTree {
    private ?TreeNode $root = null;

    public function insert(int $value): void {
        $this->root = $this->insertNode($this->root, $value);
    }

    private function insertNode(?TreeNode $node, int $value): TreeNode {
        if ($node === null) {
            return new TreeNode($value);
        }

        if ($value < $node->value) {
            $node->left = $this->insertNode($node->left, $value);
        } else {
            $node->right = $this->insertNode($node->right, $value);
        }

        return $node;
    }

    public function inOrder(): array {
        $result = [];
        $this->inOrderTraversal($this->root, $result);
        return $result;
    }

    private function inOrderTraversal(?TreeNode $node, array &$result): void {
        if ($node === null) {
            return;
        }
        $this->inOrderTraversal($node->left, $result);
        $result[] = $node->value;
        $this->inOrderTraversal($node->right, $result);
    }
}

// === Demonstrations ===

echo "=== Stack Demo ===\n";
$stack = new Stack();
$stack->push(1);
$stack->push(2);
$stack->push(3);
echo "Stack size: " . $stack->size() . "\n";
echo "Peek: " . $stack->peek() . "\n";
echo "Pop: " . $stack->pop() . "\n";
echo "Pop: " . $stack->pop() . "\n";
echo "Size after pops: " . $stack->size() . "\n\n";

echo "=== Queue Demo ===\n";
$queue = new Queue();
$queue->enqueue("First");
$queue->enqueue("Second");
$queue->enqueue("Third");
echo "Queue size: " . $queue->size() . "\n";
echo "Front: " . $queue->front() . "\n";
echo "Dequeue: " . $queue->dequeue() . "\n";
echo "Dequeue: " . $queue->dequeue() . "\n";
echo "Size after dequeues: " . $queue->size() . "\n\n";

echo "=== Linked List Demo ===\n";
$list = new LinkedList();
$list->append(1);
$list->append(2);
$list->append(3);
$list->prepend(0);
echo "List: [" . implode(" -> ", $list->toArray()) . "]\n";
echo "Size: " . $list->size() . "\n\n";

echo "=== Binary Search Tree Demo ===\n";
$bst = new BinarySearchTree();
$values = [5, 3, 7, 1, 4, 6, 8];
echo "Inserting: " . implode(", ", $values) . "\n";
foreach ($values as $v) {
    $bst->insert($v);
}
echo "In-order traversal: " . implode(", ", $bst->inOrder()) . "\n";
