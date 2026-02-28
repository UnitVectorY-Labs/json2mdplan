# bullet_list

`bullet_list` renders a JSON array of scalar values as a Markdown bullet list.

## Shape

```json
{
  "op": "bullet_list",
  "path": "."
}
```

## Behavior

- `path` selects the array to render.
- Each array item is rendered as a Markdown bullet.
- String values are emitted directly.
- Number, boolean, and null values are converted to their JSON text form.

## Requirements

- `path` must resolve to a JSON array.
- Every array item must be a scalar JSON value.
- `fields` is not supported for this directive.

## Validation

Validation fails when:

- the directive `path` does not resolve to an array
- any array item resolves to an object or array
- the directive contains unsupported `fields`

This directive also participates in coverage validation. Each scalar array item
is counted as consumed content.

## Example

Input JSON:

```json
[
  "red",
  "green",
  "blue"
]
```

Plan:

```json
{
  "version": 1,
  "directives": [
    {
      "op": "bullet_list",
      "path": "."
    }
  ]
}
```

Output Markdown:

```md
- red
- green
- blue
```
