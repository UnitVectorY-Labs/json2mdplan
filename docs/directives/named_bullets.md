# named_bullets

`named_bullets` renders a JSON object as a Markdown bullet list where each item
uses a bold label followed by a scalar value.

## Shape

```json
{
  "op": "named_bullets",
  "path": ".",
  "fields": [
    {
      "path": "name",
      "label": "name"
    }
  ]
}
```

## Behavior

- `path` selects the object to render.
- `fields` lists the object members to output.
- Each field is rendered as `- **label:** value`.
- Field order is preserved exactly as written in the plan.

## Requirements

- `path` must resolve to a JSON object.
- `fields` must not be empty.
- Each `fields[].path` must resolve relative to the selected object.
- Each resolved field value must be a scalar JSON value.

## Validation

Validation fails when:

- the directive `path` does not resolve to an object
- a listed field does not exist
- a listed field resolves to an object or array

The current implementation also uses this directive for coverage checking. A
plan that uses `named_bullets` must still cover all scalar leaf values in the
input JSON.

## Example

Input JSON:

```json
{
  "name": "Alice",
  "role": "Engineer"
}
```

Plan:

```json
{
  "version": 1,
  "directives": [
    {
      "op": "named_bullets",
      "path": ".",
      "fields": [
        {
          "path": "name",
          "label": "name"
        },
        {
          "path": "role",
          "label": "role"
        }
      ]
    }
  ]
}
```

Output Markdown:

```md
- **name:** Alice
- **role:** Engineer
```
