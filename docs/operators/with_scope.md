---
layout: default
title: with_scope
parent: Directive Operators
nav_order: 6
permalink: /operators/with_scope/
---

# with_scope Operator

Sets the current scope to a specific object for nested directives.

## Purpose

Navigate into a subtree without looping. Useful for rendering object properties when you want all nested directives to reference fields within that object.

## Syntax

```json
{
  "op": "with_scope",
  "value": <value_reference>,
  "do": [<directives>]
}
```

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `op` | string | Yes | - | Must be `"with_scope"` |
| `value` | value_reference | Yes | - | Path to object to scope into |
| `as` | string | No | - | Optional name for the scope |
| `do` | array | Yes | - | Directives to execute in this scope |
| `id` | string | No | - | Optional identifier for debugging |
| `when` | condition | No | - | Optional condition for execution |

## Examples

### Basic Scope

```json
{
  "op": "with_scope",
  "value": {"path": "/metadata", "from": "root"},
  "do": [
    {
      "op": "labeled_value_line",
      "label": {"literal": "Author"},
      "value": {"path": "/author", "from": "current"}
    },
    {
      "op": "labeled_value_line",
      "label": {"literal": "Date"},
      "value": {"path": "/date", "from": "current"}
    }
  ]
}
```

For `{"metadata": {"author": "Alice", "date": "2024-01-15"}}`, outputs:
```markdown
**Author**: Alice

**Date**: 2024\-01\-15
```

### With Heading

```json
{
  "op": "heading",
  "level": 2,
  "text": {"literal": "Configuration"}
},
{
  "op": "with_scope",
  "value": {"path": "/config", "from": "root"},
  "do": [
    {
      "op": "labeled_value_line",
      "label": {"literal": "Host"},
      "value": {"path": "/host", "from": "current"}
    },
    {
      "op": "labeled_value_line",
      "label": {"literal": "Port"},
      "value": {"path": "/port", "from": "current"},
      "value_format": "number"
    }
  ]
}
```

### Nested Scopes

```json
{
  "op": "with_scope",
  "value": {"path": "/system", "from": "root"},
  "do": [
    {
      "op": "heading",
      "level": 2,
      "text": {"literal": "System Settings"}
    },
    {
      "op": "with_scope",
      "value": {"path": "/database", "from": "current"},
      "do": [
        {
          "op": "heading",
          "level": 3,
          "text": {"literal": "Database"}
        },
        {
          "op": "labeled_value_line",
          "label": {"literal": "Connection String"},
          "value": {"path": "/connection_string", "from": "current"}
        }
      ]
    }
  ]
}
```

### Combined with for_each

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
      "op": "with_scope",
      "value": {"path": "/address", "from": "current"},
      "do": [
        {
          "op": "heading",
          "level": 3,
          "text": {"literal": "Address"}
        },
        {
          "op": "labeled_value_line",
          "label": {"literal": "Street"},
          "value": {"path": "/street", "from": "current"}
        },
        {
          "op": "labeled_value_line",
          "label": {"literal": "City"},
          "value": {"path": "/city", "from": "current"}
        }
      ]
    }
  ]
}
```

### Conditional Scope

```json
{
  "op": "with_scope",
  "value": {"path": "/optional_section", "from": "root"},
  "do": [
    {
      "op": "heading",
      "level": 2,
      "text": {"literal": "Optional Section"}
    },
    {
      "op": "text_line",
      "text": {"value": {"path": "/content", "from": "current"}}
    }
  ],
  "when": {
    "value": {"path": "/optional_section", "from": "root"},
    "test": "non_null"
  }
}
```

Only enters scope if the path exists.

## Behavior

- Pushes value onto scope stack
- All nested directives execute with new scope
- Scope is popped after `do` directives complete
- If value is null, nothing is emitted

## Scope Stack

```
Before with_scope:
[root_data]

Inside with_scope for /metadata:
[root_data, metadata_object]

After with_scope:
[root_data]
```

## Best Practices

1. **Group related fields** - Use for object properties that belong together
2. **Add heading** - Include a heading before with_scope
3. **Use from: "current"** - All nested directives should use current scope
4. **Conditional scoping** - Use `when` clause to avoid null objects

## Common Patterns

### Metadata Block

```json
[
  {
    "op": "heading",
    "level": 2,
    "text": {"literal": "Metadata"}
  },
  {
    "op": "with_scope",
    "value": {"path": "/metadata", "from": "root"},
    "do": [
      {
        "op": "labeled_value_line",
        "label": {"literal": "Version"},
        "value": {"path": "/version", "from": "current"}
      },
      {
        "op": "labeled_value_line",
        "label": {"literal": "Created"},
        "value": {"path": "/created_at", "from": "current"},
        "value_format": "datetime"
      }
    ]
  }
]
```

### Nested Objects

```json
{
  "op": "with_scope",
  "value": {"path": "/server", "from": "root"},
  "do": [
    {
      "op": "heading",
      "level": 2,
      "text": {"literal": "Server Configuration"}
    },
    {
      "op": "labeled_value_line",
      "label": {"literal": "Hostname"},
      "value": {"path": "/hostname", "from": "current"}
    },
    {
      "op": "with_scope",
      "value": {"path": "/tls", "from": "current"},
      "do": [
        {
          "op": "heading",
          "level": 3,
          "text": {"literal": "TLS Settings"}
        },
        {
          "op": "labeled_value_line",
          "label": {"literal": "Enabled"},
          "value": {"path": "/enabled", "from": "current"},
          "value_format": "boolean"
        }
      ]
    }
  ]
}
```

## Comparison with for_each

| Feature | with_scope | for_each |
|---------|------------|----------|
| Purpose | Navigate into single object | Loop through array items |
| Scope change | Once (for the object) | Per array item |
| Use case | Grouped properties | Repeated structures |

## See Also

- [for_each](for_each.html) - Array iteration
- [if_present](if_present.html) - Conditional execution
- [labeled_value_line](labeled_value_line.html) - Display properties
- [Value References](README.html#data-addressing) - Path syntax
