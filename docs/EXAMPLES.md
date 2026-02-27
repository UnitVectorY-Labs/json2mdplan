---
layout: default
title: Examples
nav_order: 4
permalink: /examples
---

# Examples
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

Each example demonstrates a different capability. The `--plan` mode requires authentication and project configuration as described in the [installation instructions](./INSTALL.md). The `--convert` mode requires no authentication.

## Generate a Plan from Schema

Create a rendering plan from a JSON Schema using Gemini.

Given a schema file `schema.json`:

```json
{
  "type": "object",
  "title": "Project",
  "properties": {
    "name": {
      "type": "string",
      "title": "Project Name"
    },
    "description": {
      "type": "string",
      "title": "Description"
    },
    "status": {
      "type": "string",
      "title": "Status"
    }
  },
  "required": ["name"]
}
```

Generate a plan:

```bash
json2mdplan --plan \
    --schema-file schema.json \
    --project my-project \
    --location us-central1 \
    --model gemini-2.5-flash \
    --pretty-print \
    --out plan.json
```

**Output (`plan.json`):**

```json
{
  "version": 1,
  "settings": {
    "base_heading_level": 1,
    "include_descriptions": false,
    "default_array_mode": "objects_as_subsections",
    "fallback_mode": "json_code_block"
  },
  "overrides": [
    {
      "path": "/name",
      "role": "document_title"
    },
    {
      "path": "/description",
      "role": "prominent_paragraph"
    }
  ]
}
```

## Convert JSON to Markdown

Convert a JSON instance to Markdown using a schema and plan.

Given a data file `data.json`:

```json
{
  "name": "Project Alpha",
  "description": "A revolutionary new project.",
  "status": "Active"
}
```

Convert to Markdown:

```bash
json2mdplan --convert \
    --json-file data.json \
    --schema-file schema.json \
    --plan-file plan.json
```

**Output:**

```markdown
# Project Alpha

A revolutionary new project.

## Status

Active
```

## Pipeline Processing

Use STDIN and STDOUT for pipeline integration.

```bash
cat data.json | json2mdplan --convert \
    --schema-file schema.json \
    --plan-file plan.json \
    > output.md
```

## Array of Objects

Handle arrays with custom item titles.

Given a schema with an array of team members:

```json
{
  "type": "object",
  "properties": {
    "team_name": { "type": "string", "title": "Team Name" },
    "members": {
      "type": "array",
      "title": "Members",
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string", "title": "Name" },
          "role": { "type": "string", "title": "Role" }
        }
      }
    }
  }
}
```

With a plan that specifies `array_section` with `item_title_from`:

```json
{
  "version": 1,
  "settings": {
    "base_heading_level": 1,
    "include_descriptions": false,
    "default_array_mode": "objects_as_subsections",
    "fallback_mode": "json_code_block"
  },
  "overrides": [
    { "path": "/team_name", "role": "document_title" },
    { "path": "/members", "role": "array_section", "item_title_from": "/name", "item_title_fallback": "Member {{index}}" }
  ]
}
```

And data:

```json
{
  "team_name": "Engineering",
  "members": [
    { "name": "Alice", "role": "Lead Developer" },
    { "name": "Bob", "role": "Designer" }
  ]
}
```

**Output:**

```markdown
# Engineering

## Members

### Alice

#### Role

Lead Developer

### Bob

#### Role

Designer
```

## Suppress Fields

Hide internal or sensitive fields from the output.

Use the `suppress` role in the plan:

```json
{
  "overrides": [
    { "path": "/internal_id", "role": "suppress" },
    { "path": "/debug_info", "role": "suppress" }
  ]
}
```

## Render Complex Data as JSON

For complex nested structures, render them as JSON code blocks.

Use the `render_as_json` role:

```json
{
  "overrides": [
    { "path": "/config", "role": "render_as_json" }
  ]
}
```

**Output:**

````markdown
## Config

```json
{
  "setting1": "value1",
  "setting2": 42,
  "nested": {
    "deep": true
  }
}
```
````

## Custom Property Order

Override the default property ordering for an object.

Use the `object_order` role:

```json
{
  "overrides": [
    {
      "path": "",
      "role": "object_order",
      "order": ["summary", "details", "metadata"]
    }
  ]
}
```

Properties will be rendered in the specified order, followed by any remaining properties in their default order (required first, then alphabetically).

## Include Schema Descriptions

Enable schema descriptions as paragraphs under headings.

In the plan settings:

```json
{
  "settings": {
    "base_heading_level": 1,
    "include_descriptions": true,
    "default_array_mode": "objects_as_subsections",
    "fallback_mode": "json_code_block"
  }
}
```

If your schema has:

```json
{
  "properties": {
    "name": {
      "type": "string",
      "title": "Name",
      "description": "The full name of the person"
    }
  }
}
```

**Output:**

```markdown
## Name

The full name of the person

John Doe
```
