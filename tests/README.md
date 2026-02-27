# Tests

This directory contains data-driven test cases for `json2mdplan`.

The purpose of these fixtures is to define the desired behavior before the
implementation is complete. These test cases are the source of truth for how
JSON should be translated into Markdown.

## Status

The exact `plan.json` format is still being designed.

For now, the test cases focus on:

- the shape of the input JSON
- the desired Markdown output
- alternative valid and invalid plans that can be used later for regression
  testing

## Test Case Layout

Each test case lives in its own directory under `tests/`.

Example:

```text
tests/
  basic-single-attribute/
    input.json
    output.md
    schema.json
    plan.json
    plan-enhanced.json
    output-enhanced.md
    valid-plans/
      alternate-plan.json
      alternate-plan.md
    invalid-plans/
      missing-field.json
      missing-field.error
```

## Standard Files

These files may appear in a test case directory:

- `input.json`
  The source JSON document for the test case.
- `output.md`
  The expected Markdown output for the default or baseline plan.
- `schema.json`
  Optional JSON Schema for the input document.
- `plan.json`
  Optional baseline plan for converting `input.json` into `output.md`.
- `plan-enhanced.json`
  Optional LLM-enhanced plan. This may produce a different Markdown structure
  from `plan.json` as long as it remains valid for the same input.
- `output-enhanced.md`
  Optional expected Markdown output produced by `plan-enhanced.json`.

## Valid Alternative Plans

If a test case has multiple valid plans, they live under `valid-plans/`.

Files in `valid-plans/` are paired by basename:

- `valid-plans/foo.json`
- `valid-plans/foo.md`

The `.json` file is an alternative valid plan for the test case input.
The matching `.md` file is the expected Markdown output for that plan.

This is intended to support cases where multiple plans are valid for the same
JSON, and where those plans may intentionally render different Markdown.

## Invalid Plans

If a test case includes plans that should fail validation or execution, they
live under `invalid-plans/`.

Files in `invalid-plans/` are paired by basename:

- `invalid-plans/foo.json`
- `invalid-plans/foo.error`

The `.json` file is an invalid plan.
The matching `.error` file records the expected failure outcome.

The exact structure of `.error` files is still to be defined, but they should
be stable enough to support regression testing.

## Current Output Conventions

The first agreed Markdown conventions are:

- A flat JSON object with scalar fields renders as a bullet list.
- Each field renders as `- **field-name:** value`.
- A top-level array of scalar values renders as a plain bullet list.

Examples:

```json
{
  "name": "Alice",
  "role": "Engineer"
}
```

renders as:

```md
- **name:** Alice
- **role:** Engineer
```

And:

```json
[
  "red",
  "green",
  "blue"
]
```

renders as:

```md
- red
- green
- blue
```

## Design Intent

These tests should start simple and grow in complexity slowly.

The main design challenge is the plan format, not the Markdown itself. The test
suite should therefore make it easy to compare:

- baseline plans
- enhanced plans
- alternative valid plans
- invalid plans

The long-term goal is to ensure that all meaningful content in the input JSON
is preserved in the Markdown output, even when the rendering is lossy with
respect to structure or attribute names.
