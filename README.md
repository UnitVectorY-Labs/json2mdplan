[![GitHub release](https://img.shields.io/github/release/UnitVectorY-Labs/json2mdplan.svg)](https://github.com/UnitVectorY-Labs/json2mdplan/releases/latest) [![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![Active](https://img.shields.io/badge/Status-Active-green)](https://guide.unitvectorylabs.com/bestpractices/status/#active)

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

## Limitations

- The `--plan` mode requires Gemini API access via Vertex AI
- Schema complexity may affect plan generation quality
- Heading levels are clamped at H6 for deeply nested structures
- v1 supports headings and paragraphs only (no tables, lists, or advanced formatting)
