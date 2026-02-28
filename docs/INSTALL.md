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

## Requirements

V1 has no external service dependencies. The current implementation uses only
local JSON input, local plans, and local rendering.
