---
layout: default
title: Usage
nav_order: 3
permalink: /usage
---

# Usage

The `json2mdplan` application has two mutually exclusive modes and follows Unix-style CLI conventions.

```
json2mdplan --plan [OPTIONS]     # Generate a plan from JSON Schema
json2mdplan --convert [OPTIONS]  # Convert JSON to Markdown
```

## Plan Mode Options

Generate a Plan JSON from a JSON Schema using Gemini.

| Option          | Arg    | Required | Notes                                           |
|-----------------|--------|----------|-------------------------------------------------|
| `--plan`        |        | yes      | Enable plan generation mode                     |
| `--schema`      | json   | yes*     | Exactly one* of this or `--schema-file`         |
| `--schema-file` | path   | yes*     | Exactly one* of this or `--schema`              |
| `--project`     | id     | yes      | Environment variable fallback supported         |
| `--location`    | region | yes      | Environment variable fallback supported         |
| `--model`       | name   | yes      | Gemini model id                                 |
| `--timeout`     | int    | no       | HTTP request timeout in seconds; default is 60  |
| `--out`         | path   | no       | Output file path; defaults to STDOUT if not set |
| `--pretty-print`|        | no       | Pretty-print JSON output; default is minified   |
| `--verbose`     |        | no       | Logs additional information to STDERR           |

If neither `--schema` nor `--schema-file` is provided, the schema is read from STDIN.

## Convert Mode Options

Convert a JSON instance to Markdown using a schema and plan (no LLM required).

| Option          | Arg    | Required | Notes                                           |
|-----------------|--------|----------|-------------------------------------------------|
| `--convert`     |        | yes      | Enable convert mode                             |
| `--json`        | json   | no*      | JSON instance inline                            |
| `--json-file`   | path   | no*      | JSON instance from file                         |
| `--schema`      | json   | yes*     | Exactly one* of this or `--schema-file`         |
| `--schema-file` | path   | yes*     | Exactly one* of this or `--schema`              |
| `--plan-json`   | json   | yes*     | Exactly one* of this or `--plan-file`           |
| `--plan-file`   | path   | yes*     | Exactly one* of this or `--plan-json`           |
| `--out`         | path   | no       | Output file path; defaults to STDOUT if not set |
| `--verbose`     |        | no       | Logs additional information to STDERR           |

If neither `--json` nor `--json-file` is provided, the JSON instance is read from STDIN.

## Common Options

| Option      | Notes                    |
|-------------|--------------------------|
| `--version` | Print version and exit   |
| `--help`    | Print help and exit      |

## Environment Variables

Options always take precedence over environment variables.

| Option      | Environment Variables                                                     |
|-------------|---------------------------------------------------------------------------|
| `--project` | `GOOGLE_CLOUD_PROJECT`, `CLOUDSDK_CORE_PROJECT`                           |
| `--location`| `GOOGLE_CLOUD_LOCATION`, `GOOGLE_CLOUD_REGION`, `CLOUDSDK_COMPUTE_REGION` |

## Command Line Conventions

The `json2mdplan` CLI follows standard UNIX conventions for input and output.

- STDIN is used for schema input (plan mode) or JSON instance (convert mode) when flags are not provided
- STDOUT emits the result when `--out` is not specified
- STDERR is reserved for logs, errors, and verbose output

### Exit Codes

| Code | Meaning                     |
|------|-----------------------------|
| 0    | Success                     |
| 2    | CLI usage error             |
| 3    | Input read/parse error      |
| 4    | Validation or response error|
| 5    | API/auth error              |

## Plan Schema

The generated plan follows a strict JSON Schema that defines:

- `version`: Plan format version (must be 1)
- `schema_fingerprint`: SHA-256 hash of the canonicalized schema for compatibility verification
- `settings`: Global rendering settings (base heading level, include descriptions, array mode, fallback mode)
- `overrides`: Path-based rendering overrides

### Override Roles

| Role                  | Purpose                                           |
|-----------------------|---------------------------------------------------|
| `document_title`      | Designate a scalar field as the document title    |
| `prominent_paragraph` | Render a field as a paragraph without a heading   |
| `section`             | Force a heading boundary for a node               |
| `object_order`        | Override property output order                    |
| `array_section`       | Specify how to title array items                  |
| `suppress`            | Omit a field or subtree from output               |
| `render_as_json`      | Render a subtree as a JSON code block             |

## Validation

The convert mode performs several validation steps:

1. Parse all inputs as JSON
2. Validate JSON instance against JSON Schema
3. Validate Plan against Plan Schema
4. Verify plan-schema compatibility using fingerprint matching
