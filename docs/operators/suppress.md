---
layout: default
title: suppress
parent: Directive Operators
nav_order: 9
permalink: /operators/suppress/
---

# suppress Operator

Declares a subtree should not be emitted.

## Purpose

Mark paths as suppressed for documentation purposes. While the interpreter doesn't enforce suppression, this operator documents which paths should not have directives that reference them.

## Syntax

```json
{
  "op": "suppress",
  "path": "/internal"
}
```

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `op` | string | Yes | - | Must be `"suppress"` |
| `path` | string | Yes | - | JSON Pointer to suppress |
| `reason` | string | No | - | Optional explanation |
| `id` | string | No | - | Optional identifier for debugging |

## Examples

### Basic Suppression

```json
{
  "op": "suppress",
  "path": "/internal"
}
```

Marks `/internal` field as suppressed.

### With Reason

```json
{
  "op": "suppress",
  "path": "/debug_info",
  "reason": "Internal debugging data not for end users"
}
```

Documents why the field is suppressed.

### Multiple Suppressions

```json
[
  {
    "op": "suppress",
    "path": "/internal_id",
    "reason": "Database implementation detail"
  },
  {
    "op": "suppress",
    "path": "/raw_data",
    "reason": "Binary data blob"
  },
  {
    "op": "suppress",
    "path": "/_metadata",
    "reason": "System metadata"
  }
]
```

Suppress multiple paths at start of plan.

### Nested Path

```json
{
  "op": "suppress",
  "path": "/user/credentials",
  "reason": "Sensitive authentication data"
}
```

Suppress nested object.

## Behavior

- Marks path in interpreter's suppressed map
- Does not prevent directives from accessing the path
- Primarily serves as documentation
- Useful for plan validation tools

## When to Suppress

Suppress paths that contain:
1. **Internal identifiers** - Database IDs, UUIDs not meaningful to users
2. **Debug information** - Debugging blobs, stack traces
3. **Sensitive data** - Credentials, tokens (though these shouldn't be in instance data)
4. **Binary data** - Non-textual content
5. **Duplicated data** - Fields that duplicate information rendered elsewhere
6. **Implementation details** - Internal state not relevant to documentation

## Best Practices

1. **Document reasons** - Always include `reason` field
2. **At plan start** - Place suppress directives at beginning of plan
3. **Don't over-suppress** - Only suppress truly irrelevant fields
4. **Validate suppression** - Ensure no other directives reference suppressed paths

## Validation Usage

Plan validation tools can use suppress directives to:
- Warn if directives reference suppressed paths
- Check schema coverage (all non-suppressed fields are rendered)
- Generate documentation about excluded fields

## Common Patterns

### System Fields

```json
{
  "op": "suppress",
  "path": "/_id",
  "reason": "MongoDB internal identifier"
},
{
  "op": "suppress",
  "path": "/_rev",
  "reason": "Document revision number"
},
{
  "op": "suppress",
  "path": "/created_by_system",
  "reason": "System audit field"
}
```

### Sensitive Fields

```json
{
  "op": "suppress",
  "path": "/api_key",
  "reason": "Sensitive credential"
},
{
  "op": "suppress",
  "path": "/password_hash",
  "reason": "Credential hash"
}
```

Note: These fields should ideally not be in the JSON instance at all.

### Debug Data

```json
{
  "op": "suppress",
  "path": "/debug",
  "reason": "Development debugging information"
},
{
  "op": "suppress",
  "path": "/performance_metrics",
  "reason": "Internal performance data"
}
```

### Raw Data

```json
{
  "op": "suppress",
  "path": "/image_binary",
  "reason": "Binary image data, not displayable as text"
},
{
  "op": "suppress",
  "path": "/encrypted_payload",
  "reason": "Encrypted binary data"
}
```

## Alternative: Just Don't Reference

In most cases, simply not creating directives that reference unwanted paths is sufficient. Use `suppress` when:
- You want to document excluded fields
- Plan validation tools need the information
- Schema coverage analysis is important
- You're generating plans automatically and want explicit exclusions

## Complete Example

```json
{
  "version": 1,
  "settings": {
    "base_heading_level": 1
  },
  "directives": [
    {
      "op": "suppress",
      "path": "/_id",
      "reason": "Database identifier"
    },
    {
      "op": "suppress",
      "path": "/internal_state",
      "reason": "Internal processing state"
    },
    {
      "op": "heading",
      "level": 1,
      "text": {"value": {"path": "/title", "from": "root"}}
    },
    {
      "op": "text_line",
      "text": {"value": {"path": "/content", "from": "root"}}
    }
  ]
}
```

## See Also

- [if_present](if_present.html) - Conditional rendering
- [labeled_value_line](labeled_value_line.html) - Has skip_if_missing for optional fields
- [Plan Schema](../plan-schema.html) - JSON Schema for plans
