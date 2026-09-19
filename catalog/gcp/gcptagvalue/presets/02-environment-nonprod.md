# Environment Nonprod

A sibling value under the same key -- one key, several values, each its own
resource. Bind it to the nonprod folder and every relaxed guardrail can
test for it.

## What it configures

- `tagKey` — the same `GcpTagKey` reference as the prod value.
- `shortName: nonprod`.

## Adjust before deploying

- **`shortName`** — your vocabulary (`staging`, `dev`, `sandbox`); one
  `GcpTagValue` each.

## When to choose something else

The production value takes the **Environment Prod** preset.
