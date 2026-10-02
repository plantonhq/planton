# Folder Allow Internal Goto Next

A folder-scoped baseline for a landing zone: the policy lives on the
`production` folder declared by a `GcpFolder` in the same chart, allows the
traffic every workload needs (private address space to itself, Google's
health-check and IAP ranges), and delegates the rest with `goto_next`.

## What it configures

- `parent.folderId` and the association target — both reference the
  `GcpFolder` named `production`, so the chart builds the folder first and
  enforces the policy on it.
- Priority `1000` — `allow` all protocols from RFC 1918 space.
- Priority `1100` — `allow` tcp from `35.191.0.0/16` and `130.211.0.0/22`,
  Google's health-check probers, so every load-balanced backend passes.
- Priority `1200` — `allow` tcp:22 from `35.235.240.0/20`, Identity-Aware
  Proxy's TCP forwarding range, so SSH works without a public address.
- Priority `2000` — `goto_next` on `all`.

## Adjust before deploying

- **`valueFrom.name`** (both places) — the name of your `GcpFolder`.
- **The RFC 1918 allow** — narrow it to your own ranges if the folder's
  networks are peered with untrusted address space.

## When to choose something else

For the organization-wide denies, start from **Org Deny SSH From
Internet**; for blocking by threat list and country, from **Threat Intel And
Geo Block**.
