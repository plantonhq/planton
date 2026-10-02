# Subordinate By Reference

## Use Case

An issuing authority chained to a root in another pool: the root signs it on create, and it issues the day-to-day certificates while the root stays idle.

## When to Use

- The second tier of a two-tier hierarchy
- One issuing authority per environment or team under a shared root

## What This Creates

- An enabled subordinate in `internal-servers` signed by `root-ca`, with a five-year P-256 certificate that may sign leaf certificates but no further authorities

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `lifetime` | `157680000s` (5 years) | Keep it well inside the root's lifetime; issued certificates stop at this expiry. |
| `config.x509Config.caOptions.maxIssuerPathLength` | `0` | Raise it only for a three-tier hierarchy. |
