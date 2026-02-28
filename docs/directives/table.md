# table

`table` renders a JSON array of objects as a Markdown table. Each object becomes
a row, and the fields listed in the plan become the columns.

## Shape

```json
{
  "op": "table",
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

- `path` selects the array to render.
- `fields` lists the object members to output as columns.
- The first row is a header row using the `label` values.
- The second row is the Markdown table separator.
- Each subsequent row corresponds to one array item.
- Field order in the plan determines column order.

## Requirements

- `path` must resolve to a JSON array.
- Every array item must be a JSON object.
- `fields` must not be empty.
- Each `fields[].path` must resolve relative to each array item object.
- Each resolved field value must be a scalar JSON value.

## Validation

Validation fails when:

- the directive `path` does not resolve to an array
- any array item is not an object
- a listed field does not exist in an array item
- a listed field resolves to an object or array

This directive also participates in coverage validation. Each scalar field in
each array item referenced by the fields list is counted as consumed content.

## Example

Input JSON:

```json
[
  {
    "name": "Alice",
    "role": "Engineer",
    "city": "Boston"
  },
  {
    "name": "Bob",
    "role": "Designer",
    "city": "Seattle"
  },
  {
    "name": "Charlie",
    "role": "Manager",
    "city": "Denver"
  }
]
```

Plan:

```json
{
  "version": 1,
  "directives": [
    {
      "op": "table",
      "path": ".",
      "fields": [
        {
          "path": "name",
          "label": "name"
        },
        {
          "path": "role",
          "label": "role"
        },
        {
          "path": "city",
          "label": "city"
        }
      ]
    }
  ]
}
```

Output Markdown:

```md
| name | role | city |
| --- | --- | --- |
| Alice | Engineer | Boston |
| Bob | Designer | Seattle |
| Charlie | Manager | Denver |
```
