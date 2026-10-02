# Environment Prod

The `prod` value of the `environment` key, by reference. Protected against
destroy because production guardrails test for it.

## What it configures

- `tagKey` — a reference to the `GcpTagKey`'s `name` output (`tagKeys/{id}`).
- `shortName: prod` — the word a condition tests:
  `resource.matchTag('{org}/environment', 'prod')`.

## Adjust before deploying

- **`tagKey.valueFrom.name`** — your key's manifest name, or a
  `tagKeys/{id}` literal.

## When to choose something else

Declare the sibling values with the **Environment Nonprod** preset.
