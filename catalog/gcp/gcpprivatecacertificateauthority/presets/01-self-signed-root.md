# Self-Signed Root

## Use Case

The trust anchor of a private PKI: a root authority with a P-384 Cloud HSM key and a twenty-year certificate, kept in its own pool to sign subordinates.

## When to Use

- Starting a private certificate hierarchy
- A single-tier lab or short-lived mesh where the root issues directly

## What This Creates

- An enabled root authority in `root-pool` (`us-central1`) whose certificate may sign certificates and CRLs, with deletion protection on

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `lifetime` | `630720000s` (20 years) | Shorter roots rotate sooner but force relying parties to re-trust. |
| `keySpec.algorithm` | `EC_P384_SHA384` | Use an RSA algorithm when relying parties cannot validate EC chains. |
| `config.subjectConfig.subject` | Example Root CA | Your organization's name. |
