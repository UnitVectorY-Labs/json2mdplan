---
layout: default
title: Template Language
nav_order: 5
has_children: true
permalink: /template/
---

# Template Language

json2mdplan uses a tree-based template language to describe how JSON data is converted to Markdown. Templates mirror the shape of your JSON data — each node in the template corresponds to a node in the JSON.

## Template Structure

A template is a JSON document with two fields:

```json
{
  "version": "1",
  "template": { /* TemplateNode */ }
}
```

## TemplateNode

Each node in the template can have the following fields:

| Field | Type | Description |
|---|---|---|
| `render` | string | How to render this node (see [Render Modes](render-modes.md)) |
| `title` | string | Heading text for sections or arrays |
| `label` | string | Display label for labeled values and table columns |
| `title_key` | string | Property key used as heading for array items in `sections` mode |
| `order` | string array | Property rendering order |
| `properties` | object | Map of child TemplateNodes by property name |
| `items` | TemplateNode | Template for array elements |

All fields are optional. When `render` is not specified, a default is chosen based on the data type.

## Defaults

When no template is provided for a node, json2mdplan uses these defaults:

| Data Type | Default Render |
|---|---|
| Object (root) | `inline` |
| Object (nested) | `section` |
| Array of scalars | `bullet_list` |
| Array of flat objects | `table` |
| Array of complex objects | `sections` |
| Scalar | `labeled_value` |

## Template Generation

Use `--generate` to create a template from a JSON file:

```bash
json2mdplan --generate --input data.json --pretty > template.json
```

The generated template captures the structure of the input JSON and assigns default render modes. You can then refine it manually or with an LLM.

## Template Refinement

Common refinements include:

- Changing a field from `labeled_value` to `heading` to make it a document title
- Changing a field from `labeled_value` to `text` for long descriptions
- Reordering properties with `order`
- Improving labels for better readability
- Hiding internal fields with `hidden`
- Adding `title` to arrays for section headings
