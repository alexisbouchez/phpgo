<?php
/**
 * Namespaces Example
 *
 * Demonstrates namespace declaration, use statements, and aliasing.
 */

namespace App\Models {
    class User {
        public function __construct(
            public int $id,
            public string $name,
            public string $email
        ) {}

        public function toArray(): array {
            return [
                'id' => $this->id,
                'name' => $this->name,
                'email' => $this->email
            ];
        }
    }

    class Product {
        public function __construct(
            public int $id,
            public string $name,
            public float $price
        ) {}
    }
}

namespace App\Services {
    use App\Models\User;

    class UserService {
        private array $users = [];

        public function create(int $id, string $name, string $email): User {
            $user = new User($id, $name, $email);
            $this->users[$id] = $user;
            return $user;
        }

        public function find(int $id): ?User {
            return $this->users[$id] ?? null;
        }

        public function all(): array {
            return $this->users;
        }
    }
}

namespace App\Controllers {
    use App\Services\UserService;
    use App\Models\User as UserModel;

    class UserController {
        private UserService $userService;

        public function __construct() {
            $this->userService = new UserService();
        }

        public function index(): void {
            echo "All users:\n";
            foreach ($this->userService->all() as $user) {
                echo "  - " . $user->name . "\n";
            }
        }

        public function store(int $id, string $name, string $email): UserModel {
            return $this->userService->create($id, $name, $email);
        }

        public function show(int $id): void {
            $user = $this->userService->find($id);
            if ($user) {
                echo "Found: " . $user->name . " <" . $user->email . ">\n";
            } else {
                echo "User not found\n";
            }
        }
    }
}

namespace {
    // Main application code in global namespace
    use App\Controllers\UserController;
    use App\Models\Product;

    echo "=== Namespaces Demo ===\n\n";

    // Using the controller
    $controller = new UserController();

    echo "Creating users...\n";
    $controller->store(1, "Alice", "alice@example.com");
    $controller->store(2, "Bob", "bob@example.com");
    $controller->store(3, "Charlie", "charlie@example.com");

    echo "\n";
    $controller->index();

    echo "\nLooking up user 2:\n";
    $controller->show(2);

    echo "\nLooking up user 99:\n";
    $controller->show(99);

    // Using Product from different namespace
    echo "\nCreating a product:\n";
    $product = new Product(1, "Widget", 29.99);
    echo "Product: {$product->name} - \${$product->price}\n";
}
