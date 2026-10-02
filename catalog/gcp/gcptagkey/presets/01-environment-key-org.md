# Environment Key (Organization)

The `environment` key every landing zone starts with, owned by the
organization so its values can be bound anywhere. Protected against
destroy: every environment guardrail keys on it.

## What it configures

- `parent.organizationId` — organization-owned; values bindable estate-wide.
- `shortName: environment` — the name written in `resource.matchTag('{org}/environment', ...)`.
- `deletionPolicy: PREVENT`.

## Adjust before deploying

- **`organizationId`** — yours.
- Declare the values (`prod`, `staging`, `dev`) with `GcpTagValue`
  resources referencing this key's `name` output.

## When to choose something else

A key one application manages inside its project takes the
**Project-Scoped Key** preset; a key for firewall rules takes **Firewall
Purpose Key**.
