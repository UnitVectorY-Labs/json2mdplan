---
layout: default
title: json2mdplan
nav_order: 1
permalink: /
---

# json2mdplan

CLI tool that converts JSON files to Markdown using customizable, tree-based templates.

## What it does

- **Generate templates** automatically from any JSON file
- **Convert JSON to Markdown** deterministically using templates
- **Pipe-friendly** — reads from STDIN, writes to STDOUT
- **LLM-refinable** — templates are simple JSON that LLMs can improve

## Why json2mdplan?

Converting JSON to human-readable Markdown typically requires custom code for each data format. `json2mdplan` solves this with a template language that mirrors the shape of your JSON data.

The workflow is simple:
1. **Generate** a default template from a sample JSON file
2. **Refine** the template (manually or with an LLM) to improve labels, ordering, and render modes
3. **Convert** any JSON file with the same structure to Markdown

Templates are reusable — create one template and use it to convert any number of JSON files with the same structure.

## Two Modes

### Generate Mode (`--generate`)

Analyzes a JSON file and produces a template that describes how each field should be rendered. The auto-generated template uses sensible defaults:
- Objects become sections or inline groups
- Arrays of flat objects become tables
- Arrays of complex objects become repeated sub-sections
- Scalar values become labeled values

### Convert Mode (`--convert`)

Converts a JSON file to Markdown using a template. This step is completely deterministic — the same inputs always produce the same output. No network calls are made.
