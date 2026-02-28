---
layout: default
title: Usage
nav_order: 3
permalink: /usage
---

# Usage

`json2mdplan` is planned as a Unix-style CLI with subcommands.

## Command Summary

V1 has two subcommands:

- `plan` generates a baseline `plan.json` from input JSON
- `render` applies a plan to input JSON and emits Markdown.

## Unix Conventions

- Read primary input from STDIN by default when practical.
- Write primary output to STDOUT by default.
- All other output (e.g. logs, errors) go to STDERR.
- Allow explicit file-based input and output flags.

For any given input type, only one source should be allowed:

- STDIN
- inline content flag
- file flag

Supplying more than one source for the same input results in an error.

## `plan`

Generate a baseline plan from input JSON.

### Syntax

```bash
json2mdplan plan [--json <json> | --json-file <path>] [--out-file <path>]
```

### Arguments

| Argument | Required | Description |
| --- | --- | --- |
| `--json <json>` | No | Inline JSON input |
| `--json-file <path>` | No | Read JSON input from a file |
| `--out-file <path>` | No | Write the generated plan to a file instead of STDOUT |

### Input Rules

- If neither `--json` nor `--json-file` is provided, `plan` reads JSON from
  STDIN.
- `--json` and `--json-file` are mutually exclusive.

### Output Rules

- If `--out-file` is not provided, the generated plan is written to STDOUT.
- The emitted output should be JSON.

## `render`

Apply a plan to input JSON and emit Markdown.

### Syntax

```bash
json2mdplan render [--json <json>] [--json-file <path>] (--plan <plan-json> | --plan-file <path>) [--out-file <path>]
```

### Arguments

| Argument | Required | Description |
| --- | --- | --- |
| `--json <json>` | No | Inline JSON input |
| `--json-file <path>` | No | Read JSON input from a file |
| `--plan <plan-json>` | Yes | Inline plan JSON |
| `--plan-file <path>` | Yes | Read the plan JSON from a file |
| `--out-file <path>` | No | Write Markdown output to a file instead of STDOUT |

### Input Rules

- If neither `--json` nor `--json-file` is provided, `render` reads JSON from
  STDIN.
- `--json` and `--json-file` are mutually exclusive.
- For this subcommand, STDIN is reserved for JSON input.
- Exactly one of `--plan` or `--plan-file` must be provided.

### Output Rules

- If `--out-file` is not provided, rendered Markdown is written to STDOUT.
