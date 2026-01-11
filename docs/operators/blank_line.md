---
layout: default
title: blank_line
parent: Directive Operators
nav_order: 2
---

# blank_line Operator

Emits one or more blank lines for spacing.

## Purpose

Add explicit whitespace between sections for improved readability. Blank lines are never inserted automatically - you must use this operator.

## Syntax

```json
{
  "op": "blank_line"
}
```

Or for multiple blank lines:

```json
{
  "op": "blank_line",
  "count": 2
}
```

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `op` | string | Yes | - | Must be `"blank_line"` |
| `count` | integer | No | 1 | Number of blank lines to emit |
| `id` | string | No | - | Optional identifier for debugging |
| `when` | condition | No | - | Optional condition for execution |

## Examples

### Single Blank Line

```json
{
  "op": "blank_line"
}
```

Output:
```markdown

```

### Multiple Blank Lines

```json
{
  "op": "blank_line",
  "count": 3
}
```

Output:
```markdown



```

### Conditional Spacing

```json
{
  "op": "blank_line",
  "when": {
    "value": {"path": "/has_separator", "from": "root"},
    "test": "equals_literal",
    "literal": true
  }
}
```

Only emits blank line if `has_separator` is true.

### Between Loop Items

```json
{
  "op": "for_each",
  "array": {"path": "/items", "from": "root"},
  "do": [
    {
      "op": "text_line",
      "text": {"value": {"path": "/description", "from": "current"}}
    }
  ],
  "between_items": [
    {
      "op": "blank_line"
    }
  ]
}
```

Adds spacing between array items.

## Behavior

- Emits the specified number of newline characters
- No Markdown formatting is applied
- Useful for separating major sections

## Best Practices

1. **Use sparingly** - Too many blank lines reduce readability
2. **Consistent spacing** - Use same count throughout document
3. **Between sections** - Add after major headings or before new topics
4. **Loop separation** - Use `between_items` in `for_each` for clarity

## Common Patterns

### Section Separation

```json
[
  {
    "op": "heading",
    "level": 2,
    "text": {"literal": "Introduction"}
  },
  {
    "op": "text_line",
    "text": {"literal": "Intro content"}
  },
  {
    "op": "blank_line"
  },
  {
    "op": "heading",
    "level": 2,
    "text": {"literal": "Details"}
  }
]
```

### List Spacing

```json
{
  "op": "bullet_list",
  "items": {"path": "/highlights", "from": "root"},
  "item_format": "text"
},
{
  "op": "blank_line"
},
{
  "op": "text_line",
  "text": {"literal": "End of highlights."}
}
```

## See Also

- [heading](heading.html) - Section headings
- [text_line](text_line.html) - Text content
- [for_each](for_each.html) - Loop with between_items
