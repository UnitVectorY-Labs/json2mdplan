---
layout: default
title: Installation
nav_order: 2
permalink: /install
---

# Installation

## Download Binary

Download pre-built binaries from the [GitHub Releases](https://github.com/UnitVectorY-Labs/json2mdplan/releases) page.

Choose the appropriate binary for your platform and add it to your PATH.

## Install Using Go

Install directly from the Go toolchain:

```bash
go install github.com/UnitVectorY-Labs/json2mdplan@latest
```

## Build from Source

Build the application from source code:

```bash
git clone https://github.com/UnitVectorY-Labs/json2mdplan.git
cd json2mdplan
go build -o json2mdplan
```

## Authentication

`json2mdplan` requires Google Cloud credentials to access Gemini models for the `--plan` mode. The `--convert` mode does not require any authentication.

{: .important }
You will be charged for usage of Gemini models according to [Google Cloud's pricing](https://cloud.google.com/vertex-ai/pricing#generative_ai_models).

Authenticate locally:

```bash
gcloud auth application-default login
```

Or use a service account:

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json
```

Set your project (can be specified with `--project` flag as well):

```bash
export GOOGLE_CLOUD_PROJECT=your-project-id
```
