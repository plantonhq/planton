# Multi-Region Target

## Use Case

Roll a release out to several regional targets in one step, with one approval for all of them.

## When to Use

- A production stage that spans regions
- Regional Cloud Run or GKE targets that should always run the same release

## What This Creates

- The Cloud Deploy API on the delivery project
- A multi-target that deploys to two regional targets in parallel, after approval

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `multiTarget.targetIds` | two `GcpDeployTarget` references | The regional targets; each must be in this target's project and region. |
| `requireApproval` | `true` | One approval gates every child rollout. |
