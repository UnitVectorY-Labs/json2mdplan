---
layout: default
title: json2mdplan
nav_order: 1
permalink: /
---

# json2mdplan

Unix-style CLI that extracts structure-only JSON, uses Vertex AI (Gemini) structured outputs to generate a schema-validated Markdown plan, then renders Markdown locally from the original JSON without sending raw values to the model.

## What it does

- Generate a rendering plan from a JSON Schema using Gemini, keeping your raw data private
- Convert JSON instances to Markdown deterministically without any LLM calls
- Enforce output structure using a validated Plan JSON Schema
- Enable repeatable, inspectable document generation from the command line
- Support shell pipelines, scripts, and batch processing workflows
- Leverage Google Gemini models with Vertex AI for intelligent plan generation

## Why json2mdplan?

Converting JSON documents to human-readable Markdown typically requires custom code for each schema or manual template creation. Large language models can help automate this, but sending raw data to external APIs raises privacy concerns and introduces non-deterministic behavior.

`json2mdplan` takes a different approach. It uses an LLM only to analyze the JSON Schema structure and generate a rendering plan. Your actual JSON data never leaves your machine during the conversion step. The rendering plan defines how to order properties, which fields become headings, and how arrays should be displayed.

The key is separation of concerns:
- **Plan generation** uses an LLM to understand schema semantics and create intelligent rendering rules
- **Conversion** is completely deterministic and local—the same inputs always produce the same Markdown output

This hybrid approach gives you the intelligence of LLMs for understanding document structure while maintaining the predictability and privacy needed for production workflows.

## Two Modes

`json2mdplan` operates in two distinct modes:

### Plan Mode (`--plan`)

Generates a Plan JSON from a JSON Schema using Gemini. This is where the LLM analyzes your schema structure and decides:
- Which field should be the document title
- What order properties should appear
- How arrays of objects should be titled
- Which fields to suppress or render as JSON

### Convert Mode (`--convert`)

Converts a JSON instance to Markdown using the schema and plan. This step:
- Requires no LLM or API calls
- Is completely deterministic
- Validates the JSON against the schema
- Verifies the plan matches the schema via fingerprinting
