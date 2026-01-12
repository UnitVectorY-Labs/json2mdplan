---
layout: default
title: heading
parent: Directive Operators
nav_order: 1
permalink: /operators/heading/
---

# heading Operator

Emits a Markdown heading line.

## Purpose

Create section headings at various depths (H1 through H6). Headings structure the document and provide navigation.

## Syntax

```json
{
  "op": "heading",
  "level": 1,
  "text": <text_union>
}
```

Or with level relative to base:

```json
{
  "op": "heading",
  "level_from_base": 0,
  "text": <text_union>
}
```

## Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `op` | string | Yes | Must be `"heading"` |
| `level` | integer | Conditional | Absolute heading level (1-6). Use `level` OR `level_from_base` |
| `level_from_base` | integer | Conditional | Level offset from `base_heading_level` setting. Use `level` OR `level_from_base` |
| `text` | text_union | Yes | The heading text content |
| `id` | string | No | Optional identifier for debugging |
| `when` | condition | No | Optional condition for execution |

## Text Union

The `text` field accepts any [text union](README.html#text-unions):
- Literal text
- Value from data
- Schema title
- Concatenated parts

## Examples

### Fixed Heading

```json
{
  "op": "heading",
  "level": 1,
  "text": {"literal": "Executive Summary"}
}
```

Output:
```markdown
# Executive Summary
```

### Heading from Data

```json
{
  "op": "heading",
  "level": 1,
  "text": {"value": {"path": "/title", "from": "root"}}
}
```

For JSON `{"title": "Q4 Report"}`, outputs:
```markdown
# Q4 Report
```

### Heading from Schema Title

```json
{
  "op": "heading",
  "level": 2,
  "text": {"schema_title": {"path": "/metadata", "fallback": "Metadata"}}
}
```

Uses the schema's title field for `/metadata`, or "Metadata" if not defined.

### Relative Level

```json
{
  "op": "heading",
  "level_from_base": 1,
  "text": {"literal": "Section"}
}
```

With `base_heading_level: 1`, produces H2. With `base_heading_level: 2`, produces H3.

### Concatenated Heading

```json
{
  "op": "heading",
  "level": 3,
  "text": {
    "concat": [
      {"literal": "Risk "},
      {"value": {"path": "/id", "from": "current"}}
    ]
  }
}
```

For `{"id": "R-001"}`, outputs:
```markdown
### Risk R\-001
```

### Conditional Heading

```json
{
  "op": "heading",
  "level": 2,
  "text": {"literal": "Appendix"},
  "when": {
    "value": {"path": "/appendix", "from": "root"},
    "test": "non_null"
  }
}
```

Only emits heading if `/appendix` exists and is not null.

## Behavior

- Heading level is clamped to 1-6 range
- Text is automatically escaped for Markdown special characters
- Emits heading followed by two newlines (blank line)
- If text resolves to empty string, nothing is emitted

## Best Practices

1. **Use schema titles** when available for semantic headings
2. **Use relative levels** (`level_from_base`) for reusable patterns
3. **Structure documents** with consistent heading hierarchy
4. **Conditional headings** prevent empty sections

## See Also

- [text_line](text_line.html) - For paragraph content
- [blank_line](blank_line.html) - For spacing
- [Text Unions](README.html#text-unions) - Text specification format
