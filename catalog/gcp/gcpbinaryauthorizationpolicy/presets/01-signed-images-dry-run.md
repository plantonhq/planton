# Signed Images Dry Run

## Use Case

See which pods would be blocked if only CI-signed images were allowed, before blocking anything.

## When to Use

- The first policy in a project whose pipelines are starting to sign images
- Measuring unsigned images across every cluster

## What This Creates

- The Binary Authorization API on the project
- A policy requiring the `built-by-ci` attestor in dry-run mode, with Google's system images always admitted

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `defaultAdmissionRule.enforcementMode` | `DRYRUN_AUDIT_LOG_ONLY` | `ENFORCED_BLOCK_AND_AUDIT_LOG` once the logs are clean. |
| `defaultAdmissionRule.requireAttestationsBy` | `built-by-ci` | Add a QA or vulnerability-scan attestor. |
