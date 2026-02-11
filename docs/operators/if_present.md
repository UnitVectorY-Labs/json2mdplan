---
layout: default
title: if_present
parent: Directive Operators
nav_order: 7
permalink: /operators/if_present/
---

# if_present Operator

Conditional execution based on value presence.

## Purpose

Avoid emitting content when data is missing, null, or empty. Essential for handling optional fields and preventing empty sections in output.

## Syntax

```json
{
  "op": "if_present",
  "value": <value_reference>,
  "mode": "non_null",
  "then": [<directives>]
}
```

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `op` | string | Yes | - | Must be `"if_present"` |
| `value` | value_reference | Yes | - | Path to test |
| `mode` | string | No | "non_null" | Test mode: "exists", "non_null", "non_empty" |
| `then` | array | Yes | - | Directives to execute if test passes |
| `else` | array | No | - | Directives to execute if test fails |
| `id` | string | No | - | Optional identifier for debugging |

## Test Modes

| Mode | Description | Passes When |
|------|-------------|-------------|
| `exists` | Path exists in data | Path is defined (even if null) |
| `non_null` | Exists and not null | Path exists and value is not null |
| `non_empty` | Non-null and non-empty | Value is not null AND not empty (arrays length > 0, strings not "", objects not {}) |

## Examples

### Basic Presence Check

```json
{
  "op": "if_present",
  "value": {"path": "/appendix", "from": "root"},
  "mode": "non_null",
  "then": [
    {
      "op": "heading",
      "level": 2,
      "text": {"literal": "Appendix"}
    },
    {
      "op": "text_line",
      "text": {"value": {"path": "/appendix", "from": "root"}}
    }
  ]
}
```

Only renders appendix section if the field exists and is not null.

### With Else Clause

```json
{
  "op": "if_present",
  "value": {"path": "/status", "from": "root"},
  "mode": "non_empty",
  "then": [
    {
      "op": "labeled_value_line",
      "label": {"literal": "Status"},
      "value": {"path": "/status", "from": "root"}
    }
  ],
  "else": [
    {
      "op": "text_line",
      "text": {"literal": "Status: Unknown"}
    }
  ]
}
```

Shows status if available, otherwise shows "Unknown".

### Non-Empty Array

```json
{
  "op": "if_present",
  "value": {"path": "/items", "from": "root"},
  "mode": "non_empty",
  "then": [
    {
      "op": "heading",
      "level": 2,
      "text": {"literal": "Items"}
    },
    {
      "op": "bullet_list",
      "items": {"path": "/items", "from": "root"},
      "item_format": "text"
    }
  ]
}
```

Only shows items section if array is non-empty.

### Nested Conditionals

```json
{
  "op": "if_present",
  "value": {"path": "/config", "from": "root"},
  "mode": "non_null",
  "then": [
    {
      "op": "heading",
      "level": 2,
      "text": {"literal": "Configuration"}
    },
    {
      "op": "if_present",
      "value": {"path": "/config/advanced", "from": "root"},
      "mode": "non_null",
      "then": [
        {
          "op": "heading",
          "level": 3,
          "text": {"literal": "Advanced Settings"}
        },
        {
          "op": "text_line",
          "text": {"value": {"path": "/config/advanced", "from": "root"}}
        }
      ]
    }
  ]
}
```

### Exists vs Non-Null

```json
{
  "op": "if_present",
  "value": {"path": "/optional", "from": "root"},
  "mode": "exists",
  "then": [
    {
      "op": "text_line",
      "text": {"literal": "Field is defined (may be null)"}
    }
  ]
}
```

Passes even if value is null, as long as the key exists.

### With for_each

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
      "op": "if_present",
      "value": {"path": "/bio", "from": "current"},
      "mode": "non_empty",
      "then": [
        {
          "op": "heading",
          "level": 3,
          "text": {"literal": "Biography"}
        },
        {
          "op": "text_line",
          "text": {"value": {"path": "/bio", "from": "current"}}
        }
      ]
    }
  ]
}
```

Shows bio section only for users who have one.

## Behavior

- Tests value according to specified mode
- If test passes, executes `then` directives
- If test fails and `else` exists, executes `else` directives
- If test fails and no `else`, emits nothing

## Mode Behavior Details

### exists
```json
{"field": null}        // Passes
{"field": ""}          // Passes
{"field": "value"}     // Passes
{}                     // Fails
```

### non_null
```json
{"field": null}        // Fails
{"field": ""}          // Passes
{"field": "value"}     // Passes
{}                     // Fails
```

### non_empty
```json
{"field": null}        // Fails
{"field": ""}          // Fails (string)
{"field": []}          // Fails (array)
{"field": {}}          // Fails (object)
{"field": "value"}     // Passes
{"field": ["item"]}    // Passes
{"field": {"key":"v"}} // Passes
```

## Best Practices

1. **Prevent empty sections** - Use before heading + content blocks
2. **Use non_empty** - For arrays and strings to avoid rendering empty content
3. **Use non_null** - For optional object fields
4. **Provide fallbacks** - Use `else` for required information
5. **Simplify with skip_if_missing** - For single values, use `skip_if_missing` in `labeled_value_line`

## Common Patterns

### Optional Section

```json
{
  "op": "if_present",
  "value": {"path": "/recommendations", "from": "root"},
  "mode": "non_empty",
  "then": [
    {
      "op": "heading",
      "level": 2,
      "text": {"literal": "Recommendations"}
    },
    {
      "op": "for_each",
      "array": {"path": "/recommendations", "from": "root"},
      "do": [
        {
          "op": "text_line",
          "text": {"value": {"path": "/text", "from": "current"}},
          "prefix": "- "
        }
      ]
    }
  ]
}
```

### Fallback Value

```json
{
  "op": "if_present",
  "value": {"path": "/description", "from": "root"},
  "mode": "non_empty",
  "then": [
    {
      "op": "text_line",
      "text": {"value": {"path": "/description", "from": "root"}}
    }
  ],
  "else": [
    {
      "op": "text_line",
      "text": {"literal": "No description available."},
      "style": "italic"
    }
  ]
}
```

### Multi-Field Check

To check if ANY of multiple fields exist, use nested conditions:

```json
{
  "op": "if_present",
  "value": {"path": "/field1", "from": "root"},
  "mode": "non_null",
  "then": [
    {
      "op": "text_line",
      "text": {"value": {"path": "/field1", "from": "root"}}
    }
  ],
  "else": [
    {
      "op": "if_present",
      "value": {"path": "/field2", "from": "root"},
      "mode": "non_null",
      "then": [
        {
          "op": "text_line",
          "text": {"value": {"path": "/field2", "from": "root"}}
        }
      ]
    }
  ]
}
```

## See Also

- [labeled_value_line](labeled_value_line.html) - Has built-in `skip_if_missing`
- [for_each](for_each.html) - Automatically skips empty arrays
- [with_scope](with_scope.html) - Scope navigation
- [Conditions](README.html#conditions) - When clauses
