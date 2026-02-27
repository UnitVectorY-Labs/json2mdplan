---
layout: default
title: for_each
parent: Directive Operators
nav_order: 5
permalink: /operators/for_each/
---

# for_each Operator

Iterates over an array and executes directives for each item.

## Purpose

Loop through arrays with controlled scope. Each iteration pushes the current array item onto the scope stack, allowing nested directives to access item properties.

## Syntax

```json
{
  "op": "for_each",
  "array": <value_reference>,
  "do": [<directives>]
}
```

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `op` | string | Yes | - | Must be `"for_each"` |
| `array` | value_reference | Yes | - | Path to array to iterate |
| `as` | string | No | - | Optional name for the loop scope |
| `index_as` | string | No | - | Optional name for index variable (0-based) |
| `do` | array | Yes | - | Directives to execute for each item |
| `between_items` | array | No | - | Directives to execute between items |
| `id` | string | No | - | Optional identifier for debugging |
| `when` | condition | No | - | Optional condition for execution |

## Examples

### Basic Loop

```json
{
  "op": "for_each",
  "array": {"path": "/items", "from": "root"},
  "do": [
    {
      "op": "text_line",
      "text": {"value": {"path": "/name", "from": "current"}}
    }
  ]
}
```

For `{"items": [{"name": "A"}, {"name": "B"}]}`, outputs:
```markdown
A

B
```

### With Headings

```json
{
  "op": "for_each",
  "array": {"path": "/sections", "from": "root"},
  "do": [
    {
      "op": "heading",
      "level": 2,
      "text": {"value": {"path": "/title", "from": "current"}}
    },
    {
      "op": "text_line",
      "text": {"value": {"path": "/content", "from": "current"}}
    }
  ]
}
```

### Between Items Spacing

```json
{
  "op": "for_each",
  "array": {"path": "/items", "from": "root"},
  "do": [
    {
      "op": "heading",
      "level": 3,
      "text": {"value": {"path": "/title", "from": "current"}}
    }
  ],
  "between_items": [
    {
      "op": "blank_line"
    }
  ]
}
```

Adds blank line between each item (but not after last item).

### Labeled Properties

```json
{
  "op": "for_each",
  "array": {"path": "/users", "from": "root"},
  "do": [
    {
      "op": "heading",
      "level": 2,
      "text": {"value": {"path": "/name", "from": "current"}}
    },
    {
      "op": "labeled_value_line",
      "label": {"literal": "Email"},
      "value": {"path": "/email", "from": "current"}
    },
    {
      "op": "labeled_value_line",
      "label": {"literal": "Role"},
      "value": {"path": "/role", "from": "current"}
    }
  ]
}
```

### Nested Loops

```json
{
  "op": "for_each",
  "array": {"path": "/departments", "from": "root"},
  "do": [
    {
      "op": "heading",
      "level": 2,
      "text": {"value": {"path": "/name", "from": "current"}}
    },
    {
      "op": "for_each",
      "array": {"path": "/employees", "from": "current"},
      "do": [
        {
          "op": "text_line",
          "text": {"value": {"path": "/name", "from": "current"}},
          "prefix": "- "
        }
      ]
    }
  ]
}
```

### Concatenated Title

```json
{
  "op": "for_each",
  "array": {"path": "/tasks", "from": "root"},
  "do": [
    {
      "op": "heading",
      "level": 3,
      "text": {
        "concat": [
          {"value": {"path": "/id", "from": "current"}},
          {"literal": ": "},
          {"value": {"path": "/title", "from": "current"}}
        ]
      }
    },
    {
      "op": "labeled_value_line",
      "label": {"literal": "Status"},
      "value": {"path": "/status", "from": "current"}
    }
  ]
}
```

For tasks with id and title, produces headings like "T-001: Implement feature".

### Named Scope (Future)

```json
{
  "op": "for_each",
  "array": {"path": "/items", "from": "root"},
  "as": "item",
  "index_as": "i",
  "do": [
    {
      "op": "text_line",
      "text": {"value": {"path": "/name", "from": "current"}}
    }
  ]
}
```

Note: `as` and `index_as` are reserved for future use.

## Behavior

- Skips execution if array is empty or null
- Each iteration pushes item onto scope stack
- Nested directives use `from: "current"` to access item properties
- `between_items` executes after each item except the last
- Scope is popped after each iteration

## Scope Stack

```
Before loop:
[root_data]

First iteration:
[root_data, item_0]

Second iteration:
[root_data, item_1]

After loop:
[root_data]
```

## Best Practices

1. **Use from: "current"** - Always use `from: "current"` in nested directives
2. **Between items** - Add spacing between complex items
3. **Headings per item** - Use heading + properties pattern
4. **Avoid deep nesting** - Limit nesting depth for readability

## Common Patterns

### Item Cards

```json
{
  "op": "for_each",
  "array": {"path": "/products", "from": "root"},
  "do": [
    {
      "op": "heading",
      "level": 3,
      "text": {"value": {"path": "/name", "from": "current"}}
    },
    {
      "op": "text_line",
      "text": {"value": {"path": "/description", "from": "current"}}
    },
    {
      "op": "labeled_value_line",
      "label": {"literal": "Price"},
      "value": {"path": "/price", "from": "current"},
      "value_format": "number"
    }
  ],
  "between_items": [
    {
      "op": "blank_line"
    }
  ]
}
```

### Numbered Items

```json
{
  "op": "for_each",
  "array": {"path": "/steps", "from": "root"},
  "do": [
    {
      "op": "text_line",
      "text": {"value": {"path": "/instruction", "from": "current"}},
      "prefix": "1. "
    }
  ]
}
```

Note: Markdown auto-numbers list items, so all can use "1. ".

### Sub-lists

```json
{
  "op": "for_each",
  "array": {"path": "/categories", "from": "root"},
  "do": [
    {
      "op": "heading",
      "level": 2,
      "text": {"value": {"path": "/name", "from": "current"}}
    },
    {
      "op": "bullet_list",
      "items": {"path": "/items", "from": "current"},
      "item_format": "text"
    }
  ]
}
```

## See Also

- [with_scope](with_scope.html) - Object scoping without loop
- [bullet_list](bullet_list.html) - Automatic bullet lists
- [if_present](if_present.html) - Conditional execution
- [blank_line](blank_line.html) - Spacing between items
