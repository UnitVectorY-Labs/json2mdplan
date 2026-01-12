---
layout: default
title: bullet_list
parent: Directive Operators
nav_order: 8
permalink: /operators/bullet_list/
---

# bullet_list Operator

Emits a Markdown bullet list from an array.

## Purpose

Render arrays as bulleted lists. Supports both arrays of scalars (strings, numbers) and arrays of objects with customizable item text.

## Syntax

For array of scalars:
```json
{
  "op": "bullet_list",
  "items": <value_reference>,
  "item_format": "text"
}
```

For array of objects:
```json
{
  "op": "bullet_list",
  "items": <value_reference>,
  "item_text": <text_union>
}
```

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `op` | string | Yes | - | Must be `"bullet_list"` |
| `items` | value_reference | Yes | - | Path to array |
| `item_format` | string | Conditional | - | Format for scalar items (text, number, boolean, etc.) |
| `item_text` | text_union | Conditional | - | Text expression for object items |
| `bullet` | string | No | "- " | Bullet character(s) |
| `skip_empty` | boolean | No | true | Skip empty/null items |
| `id` | string | No | - | Optional identifier for debugging |
| `when` | condition | No | - | Optional condition for execution |

**Note:** Use either `item_format` OR `item_text`, not both.

## Examples

### Array of Strings

```json
{
  "op": "bullet_list",
  "items": {"path": "/highlights", "from": "root"},
  "item_format": "text"
}
```

For `{"highlights": ["Point 1", "Point 2", "Point 3"]}`, outputs:
```markdown
- Point 1
- Point 2
- Point 3
```

### Array of Numbers

```json
{
  "op": "bullet_list",
  "items": {"path": "/scores", "from": "root"},
  "item_format": "number"
}
```

For `{"scores": [95, 87, 92]}`, outputs:
```markdown
- 95
- 87
- 92
```

### Array of Objects - Simple

```json
{
  "op": "bullet_list",
  "items": {"path": "/tasks", "from": "root"},
  "item_text": {"value": {"path": "/title", "from": "current"}}
}
```

For `{"tasks": [{"title": "Task A"}, {"title": "Task B"}]}`, outputs:
```markdown
- Task A
- Task B
```

### Array of Objects - Concatenated

```json
{
  "op": "bullet_list",
  "items": {"path": "/items", "from": "root"},
  "item_text": {
    "concat": [
      {"value": {"path": "/id", "from": "current"}},
      {"literal": ": "},
      {"value": {"path": "/name", "from": "current"}}
    ]
  }
}
```

For `{"items": [{"id": "A1", "name": "First"}, {"id": "A2", "name": "Second"}]}`, outputs:
```markdown
- A1: First
- A2: Second
```

### Custom Bullet

```json
{
  "op": "bullet_list",
  "items": {"path": "/items", "from": "root"},
  "item_format": "text",
  "bullet": "* "
}
```

Outputs:
```markdown
* Item 1
* Item 2
```

### Don't Skip Empty

```json
{
  "op": "bullet_list",
  "items": {"path": "/values", "from": "root"},
  "item_format": "text",
  "skip_empty": false
}
```

Will include empty strings in the output.

### Nested in for_each

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
      "op": "bullet_list",
      "items": {"path": "/items", "from": "current"},
      "item_format": "text"
    }
  ]
}
```

Each section gets its own bullet list.

### With Heading

```json
[
  {
    "op": "heading",
    "level": 2,
    "text": {"literal": "Key Features"}
  },
  {
    "op": "bullet_list",
    "items": {"path": "/features", "from": "root"},
    "item_format": "text"
  }
]
```

## Behavior

- Skips execution if array is empty or null
- Each item emits bullet + text + newline
- Ends with single blank line after all items
- Values are automatically escaped
- Item scope is available in `item_text` via `from: "current"`

## Item Formats

| Format | Description | Example |
|--------|-------------|---------|
| `text` | String representation | "hello" |
| `number` | Formatted number | 42 or 3.14 |
| `boolean` | true/false | true |
| `json_compact` | Single-line JSON | {"a":1} |

## Best Practices

1. **Use for simple arrays** - Arrays of strings or simple values
2. **Concat for objects** - Build meaningful item text
3. **Consistent bullets** - Use same bullet style throughout document
4. **Heading before list** - Add context with heading
5. **Consider for_each** - For complex items with multiple properties, use `for_each` with headings instead

## Common Patterns

### Simple List with Heading

```json
[
  {
    "op": "heading",
    "level": 2,
    "text": {"literal": "Requirements"}
  },
  {
    "op": "bullet_list",
    "items": {"path": "/requirements", "from": "root"},
    "item_format": "text"
  }
]
```

### Formatted Object List

```json
{
  "op": "heading",
  "level": 2,
  "text": {"literal": "Team Members"}
},
{
  "op": "bullet_list",
  "items": {"path": "/team", "from": "root"},
  "item_text": {
    "concat": [
      {"value": {"path": "/name", "from": "current"}},
      {"literal": " ("},
      {"value": {"path": "/role", "from": "current"}},
      {"literal": ")"}
    ]
  }
}
```

Outputs:
```markdown
## Team Members

- Alice (Developer)
- Bob (Designer)
```

### Multiple Lists

```json
[
  {
    "op": "heading",
    "level": 2,
    "text": {"literal": "Pros"}
  },
  {
    "op": "bullet_list",
    "items": {"path": "/pros", "from": "root"},
    "item_format": "text"
  },
  {
    "op": "heading",
    "level": 2,
    "text": {"literal": "Cons"}
  },
  {
    "op": "bullet_list",
    "items": {"path": "/cons", "from": "root"},
    "item_format": "text"
  }
]
```

## When to Use for_each Instead

Use `for_each` instead of `bullet_list` when:
- Items need multiple lines (heading + details)
- Items have complex structure
- You need sub-bullets or nested content
- Items should be numbered sections

Example where `for_each` is better:
```json
{
  "op": "for_each",
  "array": {"path": "/risks", "from": "root"},
  "do": [
    {
      "op": "heading",
      "level": 3,
      "text": {"value": {"path": "/title", "from": "current"}}
    },
    {
      "op": "labeled_value_line",
      "label": {"literal": "Severity"},
      "value": {"path": "/severity", "from": "current"}
    },
    {
      "op": "text_line",
      "text": {"value": {"path": "/mitigation", "from": "current"}}
    }
  ]
}
```

## See Also

- [for_each](for_each.html) - Loop with complex item rendering
- [text_line](text_line.html) - Manual bullets with prefix
- [heading](heading.html) - Section headings
- [Text Unions](README.html#text-unions) - Text specification
