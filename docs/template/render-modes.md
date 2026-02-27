---
layout: default
title: Render Modes
parent: Template Language
nav_order: 1
permalink: /template/render-modes
---

# Render Modes
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

Each template node has a `render` field that controls how the corresponding JSON value is converted to Markdown.

## inline

Renders an object's properties without emitting a heading for the object itself.

**Data type:** Object

**Use for:** The root object, or nested objects that should merge into their parent section.

```json
{ "render": "inline", "order": ["name", "status"] }
```

**Output:** Properties are rendered sequentially at the current heading level.

---

## section

Renders an object with a heading followed by its properties.

**Data type:** Object

**Fields:** `title` (required) — the heading text.

```json
{ "render": "section", "title": "Metadata" }
```

**Output:**

```markdown
## Metadata

- **Key**: value
```

---

## labeled_value

Renders a scalar value as a labeled bullet item.

**Data type:** Scalar (string, number, boolean)

**Fields:** `label` — the display label (defaults to a human-friendly version of the key).

```json
{ "render": "labeled_value", "label": "Status" }
```

**Output:**

```markdown
- **Status**: active
```

Consecutive labeled values within an object are grouped together.

---

## text

Renders a scalar value as a plain paragraph without any label.

**Data type:** Scalar

```json
{ "render": "text" }
```

**Output:**

```markdown
A great project description that renders as a paragraph.
```

---

## heading

Renders a scalar value as a Markdown heading.

**Data type:** Scalar

```json
{ "render": "heading" }
```

**Output:**

```markdown
## My Document Title
```

The heading level is determined by the nesting depth in the template.

---

## table

Renders an array of flat objects as a Markdown table.

**Data type:** Array of objects (all scalar properties)

**Fields:**
- `title` — optional heading above the table
- `items.order` — column order
- `items.properties.*.label` — column headers

```json
{
  "render": "table",
  "title": "Team Members",
  "items": {
    "order": ["name", "role"],
    "properties": {
      "name": { "label": "Name" },
      "role": { "label": "Role" }
    }
  }
}
```

**Output:**

```markdown
## Team Members

| Name | Role |
| --- | --- |
| Alice | Lead |
| Bob | Developer |
```

---

## bullet_list

Renders an array of scalars as a bullet list.

**Data type:** Array of scalars

**Fields:** `title` — optional heading above the list.

```json
{ "render": "bullet_list", "title": "Tags" }
```

**Output:**

```markdown
## Tags

- go
- cli
- tool
```

---

## sections

Renders an array of objects as repeated sub-sections, each with its own heading.

**Data type:** Array of objects

**Fields:**
- `title` — optional heading above all sections
- `items.title_key` — property key whose value becomes each section's heading
- `items.order` — property order within each section
- `items.properties` — child templates for each section's properties

```json
{
  "render": "sections",
  "title": "Phases",
  "items": {
    "render": "inline",
    "title_key": "name",
    "order": ["status"],
    "properties": {
      "status": { "render": "labeled_value", "label": "Status" }
    }
  }
}
```

**Output:**

```markdown
## Phases

### Phase 1

- **Status**: done

### Phase 2

- **Status**: pending
```

If `title_key` is not set, items are titled "Item 1", "Item 2", etc.

---

## hidden

Suppresses a node entirely from the Markdown output.

**Data type:** Any

```json
{ "render": "hidden" }
```

Use for internal IDs, timestamps, or other fields that should not appear in the document.
