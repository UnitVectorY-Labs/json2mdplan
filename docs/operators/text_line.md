---
layout: default
title: text_line
parent: Directive Operators
nav_order: 3
---

# text_line Operator

Emits a line of text (paragraph line).

## Purpose

Output literal text or formatted values as paragraph content. This is the fundamental operator for rendering text.

## Syntax

```json
{
  "op": "text_line",
  "text": <text_union>
}
```

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `op` | string | Yes | - | Must be `"text_line"` |
| `text` | text_union | Yes | - | The text content |
| `escape` | boolean | No | true | Whether to escape Markdown characters |
| `style` | string | No | "plain" | Text style: "plain", "bold", "italic", "code_inline" |
| `prefix` | string | No | - | Literal prefix (e.g., "- " for bullets) |
| `suffix` | string | No | - | Literal suffix |
| `id` | string | No | - | Optional identifier for debugging |
| `when` | condition | No | - | Optional condition for execution |

## Examples

### Plain Text

```json
{
  "op": "text_line",
  "text": {"literal": "This is a paragraph."}
}
```

Output:
```markdown
This is a paragraph.
```

### Text from Data

```json
{
  "op": "text_line",
  "text": {"value": {"path": "/summary", "from": "root"}}
}
```

For `{"summary": "System overview"}`, outputs:
```markdown
System overview
```

### Bold Text

```json
{
  "op": "text_line",
  "text": {"literal": "Important notice"},
  "style": "bold"
}
```

Output:
```markdown
**Important notice**
```

### Italic Text

```json
{
  "op": "text_line",
  "text": {"literal": "Emphasis"},
  "style": "italic"
}
```

Output:
```markdown
*Emphasis*
```

### Inline Code

```json
{
  "op": "text_line",
  "text": {"literal": "variable_name"},
  "style": "code_inline"
}
```

Output:
```markdown
`variable_name`
```

### With Prefix (Manual Bullet)

```json
{
  "op": "text_line",
  "text": {"value": {"path": "/item", "from": "current"}},
  "prefix": "- "
}
```

Output:
```markdown
- Item content
```

### With Suffix

```json
{
  "op": "text_line",
  "text": {"literal": "Status"},
  "suffix": ":"
}
```

Output:
```markdown
Status:
```

### Unescaped HTML

```json
{
  "op": "text_line",
  "text": {"literal": "<em>HTML content</em>"},
  "escape": false
}
```

Output (unescaped):
```markdown
<em>HTML content</em>
```

### Concatenated Text

```json
{
  "op": "text_line",
  "text": {
    "concat": [
      {"literal": "User: "},
      {"value": {"path": "/username", "from": "root"}}
    ]
  }
}
```

For `{"username": "alice"}`, outputs:
```markdown
User: alice
```

## Behavior

- Text is escaped by default (Markdown special characters)
- Style is applied before prefix/suffix
- Emits text followed by two newlines (paragraph separator)
- Empty text emits nothing

## Style Application Order

1. Resolve text from text union
2. Apply style (bold, italic, code)
3. Add prefix
4. Add suffix
5. Apply escaping (if enabled)
6. Emit with paragraph break

## Best Practices

1. **Use bullet_list** instead of manual prefix for arrays
2. **Escape by default** unless you need raw HTML/Markdown
3. **Bold/italic sparingly** for true emphasis
4. **Inline code** for identifiers, variables, filenames

## Common Patterns

### Styled Value

```json
{
  "op": "text_line",
  "text": {"value": {"path": "/warning", "from": "root"}},
  "style": "bold"
}
```

### Manual List Item

```json
{
  "op": "for_each",
  "array": {"path": "/tasks", "from": "root"},
  "do": [
    {
      "op": "text_line",
      "text": {"value": {"path": "/description", "from": "current"}},
      "prefix": "• "
    }
  ]
}
```

### Formatted Quote

```json
{
  "op": "text_line",
  "text": {"value": {"path": "/quote", "from": "root"}},
  "prefix": "> ",
  "style": "italic"
}
```

## See Also

- [heading](heading.html) - Section headings
- [labeled_value_line](labeled_value_line.html) - Labeled values
- [bullet_list](bullet_list.html) - Automatic bullet lists
- [Text Unions](README.html#text-unions) - Text specification
