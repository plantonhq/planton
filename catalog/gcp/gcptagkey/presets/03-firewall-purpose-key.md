# Firewall Purpose Key

A `GCE_FIREWALL` key: its values are secure tags usable as sources and
targets in network firewall policy rules, scoped to one VPC.

## What it configures

- `purpose: GCE_FIREWALL` — immutable once set.
- `purposeData.network` — `{project}/{vpc}` the secure tags are scoped to.

## Adjust before deploying

- **`organizationId`** and **`purposeData.network`** — yours.
- Declare the roles (`web`, `db`) as `GcpTagValue`s and bind them to VM
  instances with `GcpTagBinding` (zonal: set `location`).

## When to choose something else

An ordinary governance tag takes the **Environment Key (Organization)** preset.
