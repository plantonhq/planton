# Root Pool

## Use Case

The top of a two-tier hierarchy: a pool whose root authority signs only subordinate authorities, each of which issues the leaf certificates from its own pool.

## When to Use

- Keeping the root offline from day-to-day issuance while subordinates rotate beneath it
- Separate subordinate pools per environment or team, all chaining to one root

## What This Creates

- An Enterprise-tier pool in `us-central1` that accepts only CSRs (subordinate activation) and stamps every certificate it signs as a CA that may sign leaves but no further authorities (`maxIssuerPathLength: 0`)

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `issuancePolicy.baselineValues.caOptions.maxIssuerPathLength` | `0` | Raise it to allow a three-tier hierarchy. |
| `publishingOptions.publishCrl` | `true` | Keep it on so a revoked subordinate is known to relying parties. |
