[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![Concept](https://img.shields.io/badge/Status-Concept-white)](https://guide.unitvectorylabs.com/bestpractices/status/#concept)

# json2mdplan

Unix-style CLI that generates a Markdown rendering plan from JSON and then uses
that plan to deterministically render Markdown from the same JSON input.

## Overview

`json2mdplan` is designed for converting JSON documents to human-readable Markdown:

- Generate a baseline `plan.json` from JSON input
- Validate that a plan covers all scalar content in the input JSON
- Render Markdown deterministically from JSON and a plan
- Support hand-edited plans and alternative valid plans
- Use data-driven fixtures under `tests/` to catch regressions

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

## Commands

- `json2mdplan plan` reads JSON and emits a baseline plan
- `json2mdplan render` reads JSON plus a plan and emits Markdown

See [docs/USAGE.md](docs/USAGE.md) for the planned CLI contract.
