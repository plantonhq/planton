# Enforced With Sandbox Cluster

## Use Case

Block every unsigned image in the project, except on one sandbox cluster where anything may run.

## When to Use

- A production project after a dry-run rollout
- Keeping a sandbox for experiments next to enforced clusters

## What This Creates

- A policy enforcing the CI attestor by default, admitting a base-image pattern, allowing everything on `us-central1.sandbox`, with destroy blocked

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `clusterAdmissionRules[].cluster` | `us-central1.sandbox` | Your sandbox cluster's `{location}.{name}`. |
| `admissionWhitelistPatterns` | base images | Images nobody signs that must still run. |
| `deletionPolicy` | `PREVENT` | Keep: destroy would allow every image. |
