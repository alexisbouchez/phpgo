<?php
/**
 * Attributes Example (PHP 8+ Style)
 *
 * Demonstrates PHP 8 attributes for metadata annotation.
 */

// Define custom attributes
#[Attribute]
class Route {
    public function __construct(
        public string $path,
        public string $method = 'GET'
    ) {}
}

#[Attribute]
class Validate {
    public function __construct(
        public string $rule
    ) {}
}

#[Attribute(Attribute::TARGET_PROPERTY)]
class Column {
    public function __construct(
        public string $type,
        public bool $nullable = false
    ) {}
}

// Use attributes on a controller class
class UserController {
    #[Route('/users', 'GET')]
    public function index() {
        return "List all users";
    }

    #[Route('/users/{id}', 'GET')]
    public function show(int $id) {
        return "Show user $id";
    }

    #[Route('/users', 'POST')]
    public function create(#[Validate('required|email')] string $email) {
        return "Create user with email: $email";
    }
}

// Entity with column attributes
class Product {
    #[Column('int', nullable: false)]
    public int $id;

    #[Column('varchar(255)')]
    public string $name;

    #[Column('decimal(10,2)')]
    public float $price;

    #[Column('text', nullable: true)]
    public ?string $description = null;
}

// Using reflection to read attributes
$reflection = new ReflectionClass(UserController::class);

echo "UserController Routes:\n";
echo str_repeat("-", 40) . "\n";

foreach ($reflection->getMethods() as $method) {
    $attributes = $method->getAttributes(Route::class);
    foreach ($attributes as $attribute) {
        $route = $attribute->newInstance();
        echo sprintf("%-6s %s -> %s()\n",
            $route->method,
            $route->path,
            $method->getName()
        );
    }
}

echo "\nProduct Columns:\n";
echo str_repeat("-", 40) . "\n";

$productReflection = new ReflectionClass(Product::class);
foreach ($productReflection->getProperties() as $property) {
    $attributes = $property->getAttributes(Column::class);
    foreach ($attributes as $attribute) {
        $column = $attribute->newInstance();
        $nullable = $column->nullable ? "NULL" : "NOT NULL";
        echo sprintf("%-15s %-20s %s\n",
            $property->getName(),
            $column->type,
            $nullable
        );
    }
}
