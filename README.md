[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![Concept](https://img.shields.io/badge/Status-Concept-white)](https://guide.unitvectorylabs.com/bestpractices/status/#concept)

# json2mdplan

Unix-style CLI that extracts structure-only JSON, uses Vertex AI (Gemini) structured outputs to generate a schema-validated Markdown plan, then renders Markdown locally from the original JSON without sending raw values to the model.

## Overview

`json2mdplan` is designed for converting JSON documents to human-readable Markdown:

- Generate a rendering plan from a JSON Schema using Gemini, keeping your raw data private
- Convert JSON instances to Markdown deterministically without any LLM calls
- Enforce output structure using a validated Plan JSON Schema
- Enable repeatable, inspectable document generation from the command line
- Support shell pipelines, scripts, and batch processing workflows

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

## Examples

### Generate a Plan from Schema

Generate a rendering plan from a JSON Schema using Gemini:

```bash
json2mdplan --plan \
    --schema-file schema.json \
    --project my-project \
    --location us-central1 \
    --model gemini-2.5-flash \
    --out plan.json
```

### Convert JSON to Markdown

Convert a JSON instance to Markdown using the schema and plan (no LLM required):

```bash
json2mdplan --convert \
    --json-file data.json \
    --schema-file schema.json \
    --plan-file plan.json \
    --out output.md
```

## Usage

The `json2mdplan` application has two mutually exclusive modes:

- `--plan`: Generate a Plan JSON from a JSON Schema using Gemini on Vertex AI
- `--convert`: Convert a JSON instance to Markdown using a schema and plan (no LLM required)

### Authentication

`json2mdplan` uses Google Application Default Credentials for the `--plan` mode.

Authenticate locally with:

```bash
gcloud auth application-default login
```

Or via service account:

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json
```

For complete usage documentation including all options, environment variables, and command line conventions, see the [Usage documentation](https://unitvectory-labs.github.io/json2mdplan/usage).

## Plan Model

json2mdplan uses a **directive-based interpreter** model. The plan contains a sequential list of directives that are executed top-to-bottom to produce deterministic Markdown output.

### Directive Operators

Plans consist of operators that:
- **Emit content**: headings, text, bullet lists
- **Control flow**: loops (`for_each`), conditionals (`if_present`), scoping (`with_scope`)
- **Format data**: labeled values, text styling, value formatting

Key operators include:
- `heading` - Emit Markdown headings
- `text_line` - Emit paragraphs
- `labeled_value_line` - Display **Label**: value pairs
- `for_each` - Loop through arrays
- `bullet_list` - Render arrays as bullet lists
- `if_present` - Conditional execution
- `with_scope` - Navigate into objects

For complete operator documentation, see the [Operators documentation](https://unitvectory-labs.github.io/json2mdplan/operators/).

### Example Plan

```json
{
  "version": 1,
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
      "op": "for_each",
      "array": {"path": "/items", "from": "root"},
      "do": [
        {
          "op": "heading",
          "level": 2,
          "text": {"value": {"path": "/name", "from": "current"}}
        },
        {
          "op": "text_line",
          "text": {"value": {"path": "/description", "from": "current"}}
        }
      ]
    }
  ]
}
```

## Supported Features

- Headings (H1-H6) with absolute or relative levels
- Text paragraphs with styling (bold, italic, inline code)
- Labeled values with customizable separators
- Bullet lists from arrays
- Loops with scope management
- Conditional rendering based on data presence
- Object scoping for nested structures
- Text concatenation and formatting
- Value formatting (text, number, boolean, date, json_compact)
- Schema title lookups

## Limitations

- The `--plan` mode requires Gemini API access via Vertex AI
- Schema complexity may affect plan generation quality
- Heading levels are clamped at H6 for deeply nested structures
