<?php
/**
 * Classes and Object-Oriented Programming Example
 *
 * Demonstrates classes, inheritance, interfaces, traits, and magic methods.
 */

// Interface definition
interface Describable {
    public function describe(): string;
}

// Trait for common functionality
trait Timestampable {
    private string $createdAt;

    public function setCreatedAt(string $timestamp): void {
        $this->createdAt = $timestamp;
    }

    public function getCreatedAt(): string {
        return $this->createdAt;
    }
}

// Abstract base class
abstract class Entity implements Describable {
    protected int $id;

    public function __construct(int $id) {
        $this->id = $id;
    }

    public function getId(): int {
        return $this->id;
    }

    abstract public function getType(): string;
}

// Concrete class with constructor property promotion
class User extends Entity {
    use Timestampable;

    public function __construct(
        int $id,
        private string $name,
        private string $email
    ) {
        parent::__construct($id);
        $this->setCreatedAt(date("Y-m-d H:i:s"));
    }

    public function getName(): string {
        return $this->name;
    }

    public function getEmail(): string {
        return $this->email;
    }

    public function getType(): string {
        return "user";
    }

    public function describe(): string {
        return "User #{$this->id}: {$this->name} <{$this->email}>";
    }

    // Magic method for string representation
    public function __toString(): string {
        return $this->describe();
    }

    // Magic method for debug info
    public function __debugInfo(): array {
        return [
            'id' => $this->id,
            'name' => $this->name,
            'email' => '[REDACTED]',
            'created' => $this->getCreatedAt()
        ];
    }
}

// Create and use objects
$user = new User(1, "Alice", "alice@example.com");
echo $user . "\n";
echo "Type: " . $user->getType() . "\n";
echo "Created: " . $user->getCreatedAt() . "\n";

// Var dump shows __debugInfo output
echo "\nDebug info:\n";
var_dump($user);
