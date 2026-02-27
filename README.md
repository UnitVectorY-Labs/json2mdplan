[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)

# json2mdplan

CLI tool that converts JSON files to Markdown using customizable, tree-based templates.

## Overview

`json2mdplan` converts arbitrary JSON data to well-structured Markdown documents. It uses a template language that mirrors the shape of your JSON data, giving you control over how each field is rendered.

- **Generate templates** automatically from any JSON file
- **Convert JSON to Markdown** deterministically using templates
- **Pipe-friendly** — reads from STDIN, writes to STDOUT
- **LLM-refinable** — templates are simple JSON that LLMs can improve

## Installation

```bash
go install github.com/UnitVectorY-Labs/json2mdplan@latest
```

Build from source:

```bash
git clone https://github.com/UnitVectorY-Labs/json2mdplan.git
cd json2mdplan
go build -o json2mdplan
```

## Usage

The tool has two mutually exclusive modes:

### Generate a Template

Create a template from a JSON file:

```bash
json2mdplan --generate --input data.json --pretty
```

Or from STDIN:

```bash
cat data.json | json2mdplan --generate --pretty
```

### Convert JSON to Markdown

Convert a JSON file to Markdown using a template:

```bash
json2mdplan --convert --input data.json --template template.json
```

Pipeline usage:

```bash
cat data.json | json2mdplan --convert --template template.json > output.md
```

### Options

| Flag | Description |
|---|---|
| `--generate` | Generate a template from JSON input |
| `--convert` | Convert JSON to Markdown using a template |
| `--input` | Input JSON file (default: STDIN) |
| `--template` | Template file (required for `--convert`) |
| `--output` | Output file (default: STDOUT) |
| `--pretty` | Pretty-print JSON output (`--generate` mode) |
| `--verbose` | Enable verbose logging to STDERR |
| `--version` | Show version and exit |

## Template Language

Templates are JSON documents that mirror the structure of your data. Each node specifies a **render mode** that controls how the corresponding JSON value appears in the Markdown output.

### Render Modes

| Mode | Data Type | Output |
|---|---|---|
| `inline` | Object | Properties rendered without a heading wrapper |
| `section` | Object | Heading + properties |
| `labeled_value` | Scalar | `- **Label**: value` |
| `text` | Scalar | Plain paragraph |
| `heading` | Scalar | Markdown heading |
| `table` | Array of flat objects | Markdown table |
| `bullet_list` | Array of scalars | Bullet list |
| `sections` | Array of objects | Repeated sub-sections |
| `hidden` | Any | Suppressed from output |

### Example

Given this JSON:

```json
{
  "title": "My Project",
  "status": "active",
  "tags": ["go", "cli"]
}
```

And this template:

```json
{
  "version": "1",
  "template": {
    "render": "inline",
    "order": ["title", "status", "tags"],
    "properties": {
      "title": {"render": "heading"},
      "status": {"render": "labeled_value", "label": "Status"},
      "tags": {"render": "bullet_list", "title": "Tags"}
    }
  }
}
```

Output:

```markdown
## My Project

- **Status**: active

## Tags

- go
- cli
```

## Design Philosophy

- **Data completeness** — every field in the JSON appears in the output unless explicitly hidden
- **Template reuse** — one template works for any JSON data with the same structure
- **Deterministic** — same inputs always produce the same Markdown output
- **LLM-friendly** — templates are simple enough for LLMs to generate and refine
