# phpgo

A PHP interpreter written in Go.

## Features

### Language Support

- **Namespaces** with `use` statements and aliasing
- **Classes** with inheritance, interfaces, traits, and abstract classes
- **Magic methods** (`__construct`, `__destruct`, `__get`, `__set`, `__call`, `__toString`, `__clone`, `__debugInfo`, `__sleep`, `__wakeup`)
- **Visibility modifiers** (public, protected, private)
- **Static properties and methods**
- **Constructor property promotion**
- **Generators** with `yield` and `yield from`
- **Named arguments**
- **Attributes** (PHP 8+ style)
- **Strict types** with type checking via `declare(strict_types=1)`
- **ArrayAccess** and **Iterator** interfaces
- **Reflection API**
- **Output buffering** (`ob_start`, `ob_get_contents`, etc.)
- **Serialization** with `__sleep` and `__wakeup` support

### Built-in Functions

**String functions:** `strlen`, `substr`, `strpos`, `str_replace`, `strtoupper`, `strtolower`, `trim`, `ltrim`, `rtrim`, `explode`, `implode`, `sprintf`, `str_repeat`, `ucfirst`, `lcfirst`, `ucwords`, `str_pad`, `str_split`, `chunk_split`, `wordwrap`, `nl2br`, `ord`, `chr`, `str_contains`, `str_starts_with`, `str_ends_with`, `htmlspecialchars`, `htmlentities`, `strip_tags`, `addslashes`, `stripslashes`, `number_format`

**Array functions:** `count`, `array_push`, `array_pop`, `array_shift`, `array_unshift`, `array_merge`, `array_keys`, `array_values`, `array_reverse`, `array_slice`, `array_search`, `in_array`, `array_key_exists`, `array_map`, `array_filter`, `array_reduce`, `array_unique`, `array_flip`, `array_sum`, `array_product`, `range`, `sort`, `rsort`, `array_combine`, `array_fill`, `array_chunk`, `array_column`, `array_count_values`, `array_diff`, `array_intersect`

**Math functions:** `abs`, `ceil`, `floor`, `round`, `max`, `min`, `pow`, `sqrt`, `rand`, `mt_rand`

**Type functions:** `gettype`, `is_null`, `is_bool`, `is_int`, `is_float`, `is_string`, `is_array`, `is_object`, `is_numeric`, `intval`, `floatval`, `strval`, `boolval`

**File functions:** `file_get_contents`, `file_put_contents`, `file_exists`, `is_file`, `is_dir`, `is_readable`, `is_writable`, `file`, `dirname`, `basename`, `pathinfo`, `realpath`, `glob`

**Date/time functions:** `time`, `date`, `strtotime`, `mktime`, `microtime`

**Regex functions:** `preg_match`, `preg_match_all`, `preg_replace`, `preg_split`

**Hash functions:** `md5`, `sha1`, `hash`, `base64_encode`, `base64_decode`

**JSON functions:** `json_encode`, `json_decode`, `serialize`, `unserialize`

**Output functions:** `var_dump`, `print_r`, `ob_start`, `ob_end_clean`, `ob_end_flush`, `ob_get_contents`, `ob_get_clean`, `ob_get_flush`, `ob_get_level`, `ob_flush`, `ob_clean`

**Misc functions:** `defined`, `function_exists`, `class_exists`, `call_user_func`, `call_user_func_array`, `func_get_args`, `func_num_args`, `exit`, `die`

## Installation

```bash
go install github.com/alexisbouchez/phpgo@latest
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/alexisbouchez/phpgo/interpreter"
)

func main() {
    i := interpreter.New()
    i.Eval(`<?php
        echo "Hello, World!";
    ?>`)
    fmt.Print(i.Output())
}
```

## Examples

The `examples/` directory contains practical examples demonstrating phpgo's features:

| File | Description |
|------|-------------|
| `01_hello_world.php` | Basic syntax, variables, arrays, and string operations |
| `02_classes_oop.php` | Classes, inheritance, interfaces, traits, and magic methods |
| `03_generators.php` | Generators with `yield` and `yield from` |
| `04_attributes.php` | PHP 8+ style attributes for metadata annotation |
| `05_reflection.php` | Reflection API for runtime introspection |
| `06_strict_types.php` | Type checking with `declare(strict_types=1)` |
| `07_array_access_iterator.php` | Custom collections with ArrayAccess and Iterator |
| `08_output_buffering.php` | Output buffering for template rendering |
| `09_namespaces.php` | Namespaces, use statements, and aliasing |
| `10_practical_data_structures.php` | Stack, Queue, LinkedList, and BST implementations |
| `11_serialization.php` | Serialize/unserialize with `__sleep` and `__wakeup` |

### Running Examples

```bash
# Build phpgo
go build -o phpgo ./cmd/phpgo

# Run an example
./phpgo examples/01_hello_world.php
```

## License

This project is licensed under the GNU Affero General Public License v3.0 - see the [LICENSE](LICENSE) file for details.
