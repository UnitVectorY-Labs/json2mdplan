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

## Generate a Template

Create a template from a JSON file:

```bash
json2mdplan --generate --input data.json --pretty > template.json
```

Given `data.json`:

```json
{
  "name": "Project Alpha",
  "status": "active",
  "tags": ["go", "cli"]
}
```

**Output (`template.json`):**

```json
{
  "version": "1",
  "template": {
    "render": "inline",
    "order": ["name", "status", "tags"],
    "properties": {
      "name": { "render": "labeled_value", "label": "Name" },
      "status": { "render": "labeled_value", "label": "Status" },
      "tags": { "render": "bullet_list", "label": "Tags" }
    }
  }
}
```

## Convert JSON to Markdown

Convert a JSON file using a template:

```bash
json2mdplan --convert --input data.json --template template.json
```

**Output:**

```markdown
- **Name**: Project Alpha
- **Status**: active

- go
- cli
```

## Refined Template

Improve the auto-generated template by changing render modes and adding titles:

```json
{
  "version": "1",
  "template": {
    "render": "inline",
    "order": ["name", "status", "tags"],
    "properties": {
      "name": { "render": "heading" },
      "status": { "render": "labeled_value", "label": "Status" },
      "tags": { "render": "bullet_list", "title": "Tags" }
    }
  }
}
```

**Output:**

```markdown
## Project Alpha

- **Status**: active

## Tags

- go
- cli
```

## Table Rendering

Arrays of flat objects render as tables:

Given `team.json`:

```json
{
  "team": [
    { "name": "Alice", "role": "Lead" },
    { "name": "Bob", "role": "Developer" }
  ]
}
```

With template:

```json
{
  "version": "1",
  "template": {
    "render": "inline",
    "properties": {
      "team": {
        "render": "table",
        "title": "Team",
        "items": {
          "order": ["name", "role"],
          "properties": {
            "name": { "label": "Name" },
            "role": { "label": "Role" }
          }
        }
      }
    }
  }
}
```

**Output:**

```markdown
## Team

| Name | Role |
| --- | --- |
| Alice | Lead |
| Bob | Developer |
```

## Sections Rendering

Arrays of complex objects render as sub-sections:

Given `phases.json`:

```json
{
  "phases": [
    { "name": "Phase 1", "status": "done", "tasks": [{"id": 1}] },
    { "name": "Phase 2", "status": "pending", "tasks": [{"id": 2}] }
  ]
}
```

With template using `title_key`:

```json
{
  "version": "1",
  "template": {
    "render": "inline",
    "properties": {
      "phases": {
        "render": "sections",
        "title": "Phases",
        "items": {
          "render": "inline",
          "title_key": "name",
          "order": ["status", "tasks"],
          "properties": {
            "status": { "render": "labeled_value", "label": "Status" },
            "tasks": { "render": "table", "title": "Tasks" }
          }
        }
      }
    }
  }
}
```

## Pipeline Processing

Use STDIN and STDOUT for pipeline integration:

```bash
cat data.json | json2mdplan --convert --template template.json > output.md
```

Generate and convert in one pipeline:

```bash
json2mdplan --generate --input data.json --pretty | \
  json2mdplan --convert --input data.json --template /dev/stdin

```

## Hidden Fields

Suppress fields from the output:

```json
{
  "version": "1",
  "template": {
    "render": "inline",
    "properties": {
      "name": { "render": "labeled_value", "label": "Name" },
      "internal_id": { "render": "hidden" }
    }
  }
}
```
