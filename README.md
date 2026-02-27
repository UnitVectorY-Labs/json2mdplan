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
