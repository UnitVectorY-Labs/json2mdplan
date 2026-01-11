---
layout: default
title: labeled_value_line
parent: Directive Operators
nav_order: 4
---

# labeled_value_line Operator

Emits a single line displaying a label and value, typically formatted as **Label**: value.

## Purpose

Display structured data as labeled attribute-value pairs. Common for metadata, properties, and key-value information.

## Syntax

```json
{
  "op": "labeled_value_line",
  "label": <text_union>,
  "value": <value_reference>
}
```

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `op` | string | Yes | - | Must be `"labeled_value_line"` |
| `label` | text_union | Yes | - | The label text |
| `value` | value_reference | Yes | - | Path to the value |
| `label_style` | string | No | "bold" | "bold" or "plain" |
| `separator` | string | No | ": " | Separator between label and value |
| `value_format` | string | No | "text" | Value format (text, number, boolean, date, datetime, json_compact) |
| `skip_if_missing` | boolean | No | true | Skip if value is null |
| `id` | string | No | - | Optional identifier for debugging |
| `when` | condition | No | - | Optional condition for execution |

## Examples

### Basic Labeled Value

```json
{
  "op": "labeled_value_line",
  "label": {"literal": "Status"},
  "value": {"path": "/status", "from": "root"}
}
```

For `{"status": "active"}`, outputs:
```markdown
**Status**: active
```

### Plain Label Style

```json
{
  "op": "labeled_value_line",
  "label": {"literal": "Description"},
  "value": {"path": "/description", "from": "root"},
  "label_style": "plain"
}
```

Output:
```markdown
Description: System description
```

### Custom Separator

```json
{
  "op": "labeled_value_line",
  "label": {"literal": "Version"},
  "value": {"path": "/version", "from": "root"},
  "separator": " = "
}
```

Output:
```markdown
**Version** = 1.2.3
```

### Number Formatting

```json
{
  "op": "labeled_value_line",
  "label": {"literal": "Count"},
  "value": {"path": "/count", "from": "root"},
  "value_format": "number"
}
```

For `{"count": 42}`, outputs:
```markdown
**Count**: 42
```

### Date Formatting

```json
{
  "op": "labeled_value_line",
  "label": {"literal": "Created"},
  "value": {"path": "/created_at", "from": "root"},
  "value_format": "datetime"
}
```

### JSON Compact

```json
{
  "op": "labeled_value_line",
  "label": {"literal": "Config"},
  "value": {"path": "/config", "from": "root"},
  "value_format": "json_compact"
}
```

For `{"config": {"key": "value"}}`, outputs:
```markdown
**Config**: {"key":"value"}
```

### From Current Scope

```json
{
  "op": "for_each",
  "array": {"path": "/users", "from": "root"},
  "do": [
    {
      "op": "labeled_value_line",
      "label": {"literal": "Name"},
      "value": {"path": "/name", "from": "current"}
    }
  ]
}
```

### Don't Skip Missing

```json
{
  "op": "labeled_value_line",
  "label": {"literal": "Optional"},
  "value": {"path": "/optional_field", "from": "root"},
  "skip_if_missing": false
}
```

Will emit "**Optional**: null" even if field is null.

### Schema Title as Label

```json
{
  "op": "labeled_value_line",
  "label": {"schema_title": {"path": "/status", "fallback": "Status"}},
  "value": {"path": "/status", "from": "root"}
}
```

Uses the schema's title for `/status` field as the label.

## Behavior

- Label is escaped before style is applied
- Value is always escaped
- Bold labels wrap text in `**...**`
- If value is null and `skip_if_missing` is true, nothing is emitted
- Emits label + separator + value followed by two newlines

## Value Formats

| Format | Description | Example Output |
|--------|-------------|----------------|
| `text` | String representation | "hello" |
| `number` | Formatted number | 42 or 3.14 |
| `boolean` | true/false | true |
| `date` | ISO date | 2024-01-15 |
| `datetime` | ISO datetime | 2024-01-15T10:30:00Z |
| `json_compact` | Single-line JSON | {"a":1} |

## Best Practices

1. **Consistent separators** - Use same separator throughout document
2. **Bold for emphasis** - Use bold labels for important attributes
3. **Skip missing** - Default behavior prevents null clutter
4. **Appropriate formatting** - Use correct value_format for data type

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
    "op": "labeled_value_line",
    "label": {"literal": "Author"},
    "value": {"path": "/author", "from": "root"}
  },
  {
    "op": "labeled_value_line",
    "label": {"literal": "Date"},
    "value": {"path": "/date", "from": "root"},
    "value_format": "date"
  },
  {
    "op": "labeled_value_line",
    "label": {"literal": "Version"},
    "value": {"path": "/version", "from": "root"}
  }
]
```

### Object Properties

```json
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

## See Also

- [text_line](text_line.html) - Plain text paragraphs
- [with_scope](with_scope.html) - Object scoping
- [for_each](for_each.html) - Array iteration
- [Value Formats](README.html#value-formats) - Format specifications
