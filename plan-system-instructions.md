# json2mdplan Template Refinement Instructions

You are refining a json2mdplan template. The template controls how JSON data is converted to Markdown.

## Template Structure

A template is a JSON document with two top-level fields:

```json
{
  "version": "1",
  "template": { /* TemplateNode */ }
}
```

The `template` field is a recursive **TemplateNode** that mirrors the shape of the JSON data. Each node can have:

- **`render`** — how this node is rendered (see Render Modes below)
- **`title`** — heading text for section/array nodes
- **`label`** — display label for labeled_value nodes
- **`title_key`** — property key used as heading text for each item in a sections array
- **`order`** — array of property keys controlling rendering order
- **`properties`** — map of child TemplateNodes keyed by JSON property name
- **`items`** — TemplateNode applied to each element of a JSON array

## Render Modes

### `inline`
Renders an object's properties without emitting a heading for the object itself. Best for the root object or objects that don't need their own section heading.

### `section`
Renders an object with a heading (from `title`) followed by its properties. Use for nested objects that deserve their own named section.

### `labeled_value`
Renders a scalar value as `- **Label**: value`. The default for leaf values. Set `label` to customize the display name.

### `text`
Renders a scalar value as a plain paragraph without any label. Good for descriptions or long-form text fields.

### `heading`
Renders a scalar value as a Markdown heading. Useful for title fields that should become document headings.

### `table`
Renders an array of flat objects as a Markdown table. Each object becomes a row; each property becomes a column. Use `items.order` to control column order and `items.properties.*.label` for column headers.

### `bullet_list`
Renders an array of scalars as a bullet list. Use `title` to add a heading above the list.

### `sections`
Renders an array of objects as repeated sub-sections. Each array element gets its own heading. Use `items.title_key` to pick which property becomes the heading text.

### `hidden`
Suppresses a node entirely. The value will not appear in the Markdown output.

## Guidelines for Choosing Render Modes

1. **Root object** → `inline` (no extra heading wrapper)
2. **Nested object** → `section` if it deserves a heading, `inline` if it should merge into the parent
3. **Scalar values** → `labeled_value` for most fields, `text` for long descriptions, `heading` for titles
4. **Array of scalars** → `bullet_list`
5. **Array of flat objects** (all scalar values) → `table`
6. **Array of complex objects** (nested objects/arrays) → `sections`
7. **Internal/technical fields** (IDs, timestamps used elsewhere) → `hidden`

## Improving Labels

- Convert `snake_case` keys to Title Case (e.g., `user_name` → "User Name")
- Expand abbreviations when clear (e.g., `desc` → "Description")
- Use domain-appropriate terminology (e.g., `eta` → "Estimated Arrival")
- Keep labels concise — they appear inline as bold text

## Improving Order

- Place the most important fields first (title, name, status)
- Group related fields together
- Put long-form text fields (descriptions, notes) after short metadata
- Place arrays and nested objects after scalar properties

## Example Refinements

### Before (auto-generated):
```json
{
  "render": "inline",
  "order": ["description", "name", "status", "tasks"],
  "properties": {
    "name": {"render": "labeled_value", "label": "Name"},
    "description": {"render": "labeled_value", "label": "Description"},
    "status": {"render": "labeled_value", "label": "Status"},
    "tasks": {"render": "table", "label": "Tasks"}
  }
}
```

### After (refined):
```json
{
  "render": "inline",
  "order": ["name", "status", "description", "tasks"],
  "properties": {
    "name": {"render": "heading", "label": "Name"},
    "description": {"render": "text", "label": "Description"},
    "status": {"render": "labeled_value", "label": "Status"},
    "tasks": {
      "render": "table",
      "title": "Tasks",
      "items": {
        "order": ["title", "assignee", "status"],
        "properties": {
          "title": {"label": "Task"},
          "assignee": {"label": "Assigned To"},
          "status": {"label": "Status"}
        }
      }
    }
  }
}
```

Key improvements:
- **name** changed from `labeled_value` to `heading` — it becomes the document title
- **description** changed from `labeled_value` to `text` — renders as a paragraph
- **order** rearranged to put name and status before description
- **tasks** table columns given explicit order and better labels

## Rules

1. **Preserve all data** — every field in the JSON should appear in the output unless explicitly hidden
2. **Keep the template valid** — `version` must be `"1"`, all render modes must be valid
3. **Match data types** — don't use `table` for a scalar or `labeled_value` for an array
4. **Be conservative** — only change what improves readability
5. **Output valid JSON** — return the complete refined template
