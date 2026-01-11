---
layout: default
title: Directive Operators
nav_order: 5
has_children: true
permalink: /operators
---

# Directive Operators

json2mdplan uses a directive-based interpreter model to convert JSON to Markdown. The plan consists of a sequential list of directives that are executed top-to-bottom to produce deterministic Markdown output.

## Core Concepts

### Directives
Each directive is an operator that performs a specific action:
- Emit Markdown content (headings, text, lists)
- Control flow (loops, conditionals, scoping)
- Data transformation (formatting, styling)

### Sequential Execution
Directives execute in order from top to bottom. Each directive emits zero or more lines of Markdown to the output stream.

### Scope Stack
The interpreter maintains a scope stack for data access:
- **root scope**: The entire JSON instance
- **for_each** pushes item scope when looping through arrays
- **with_scope** pushes object scope for nested structures

### Data Addressing
Paths use JSON Pointer syntax:
- `/metadata/author` - absolute path from root
- `.` or empty string - current scope
- Use `from: "root"` (default) or `from: "current"` to control resolution

## Operator Categories

### Document Structure
- [heading](heading.html) - Emit Markdown headings
- [blank_line](blank_line.html) - Add whitespace
- [text_line](text_line.html) - Emit text paragraphs

### Value Display
- [labeled_value_line](labeled_value_line.html) - Display labeled values
- [bullet_list](bullet_list.html) - Render arrays as bullet lists

### Control Flow
- [for_each](for_each.html) - Loop through arrays
- [with_scope](with_scope.html) - Navigate into objects
- [if_present](if_present.html) - Conditional execution

### Metadata
- [suppress](suppress.html) - Mark paths as suppressed

## Text Unions

Many operators accept text via "text unions" which can be:

### Literal
Fixed text:
```json
{"literal": "My Heading"}
```

### Value
From data:
```json
{"value": {"path": "/title", "from": "root"}}
```

### Schema Title
From schema metadata:
```json
{"schema_title": {"path": "/metadata", "fallback": "Metadata"}}
```

### Concat
Combine multiple parts:
```json
{
  "concat": [
    {"literal": "ID: "},
    {"value": {"path": "/id", "from": "current"}}
  ]
}
```

## Value Formats

When reading values from JSON, specify the format:

- `text` - Default string representation
- `number` - Formatted number
- `boolean` - true/false
- `date` - ISO date format
- `datetime` - ISO datetime format
- `json_compact` - Single-line JSON

## Conditions

Most directives support optional `when` clauses:

```json
{
  "when": {
    "value": {"path": "/optional", "from": "root"},
    "test": "non_null"
  }
}
```

Test modes:
- `exists` - Path exists in data
- `non_null` - Exists and not null
- `non_empty` - Non-null and non-empty (arrays, strings, objects)
- `equals_literal` - Matches specific value

## Example Plan

```json
{
  "version": 1,
  "schema_fingerprint": {
    "sha256": "abc...",
    "canonicalization": "json-canonical-v1"
  },
  "settings": {
    "base_heading_level": 1
  },
  "directives": [
    {
      "op": "heading",
      "level": 1,
      "text": {"value": {"path": "/title", "from": "root"}}
    },
    {
      "op": "text_line",
      "text": {"value": {"path": "/summary", "from": "root"}}
    },
    {
      "op": "for_each",
      "array": {"path": "/items", "from": "root"},
      "do": [
        {
          "op": "heading",
          "level": 2,
          "text": {"value": {"path": "/name", "from": "current"}}
        }
      ]
    }
  ]
}
```

## See Also

- [Plan Schema](../plan-schema.html) - JSON Schema for plans
- [System Instructions](../system-instructions.html) - LLM generation guidelines
- [Examples](../examples.html) - Complete examples
