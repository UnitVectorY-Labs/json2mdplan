---
layout: default
title: Usage
nav_order: 3
permalink: /usage
---

# Usage

The `json2mdplan` application has two mutually exclusive modes and follows Unix-style CLI conventions.

```
json2mdplan --generate [OPTIONS]  # Generate a template from JSON
json2mdplan --convert [OPTIONS]   # Convert JSON to Markdown
```

## Generate Mode

Generate a template from a JSON file.

```bash
json2mdplan --generate --input data.json --pretty
```

Or from STDIN:

```bash
cat data.json | json2mdplan --generate --pretty > template.json
```

## Convert Mode

Convert a JSON file to Markdown using a template.

```bash
json2mdplan --convert --input data.json --template template.json
```

Or from STDIN:

```bash
cat data.json | json2mdplan --convert --template template.json > output.md
```

## All Options

| Flag | Arg | Mode | Description |
|---|---|---|---|
| `--generate` | | | Generate a template from JSON input |
| `--convert` | | | Convert JSON to Markdown using a template |
| `--input` | path | both | Input JSON file (default: STDIN) |
| `--template` | path | convert | Template file (required for `--convert`) |
| `--output` | path | both | Output file (default: STDOUT) |
| `--pretty` | | generate | Pretty-print JSON output |
| `--verbose` | | both | Enable verbose logging to STDERR |
| `--version` | | | Show version and exit |

## Command Line Conventions

- STDIN is used for JSON input when `--input` is not provided
- STDOUT emits the result when `--output` is not specified
- STDERR is reserved for logs, errors, and verbose output

### Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 2 | CLI usage error |
| 3 | Input read/parse error |
| 4 | Template validation error |
