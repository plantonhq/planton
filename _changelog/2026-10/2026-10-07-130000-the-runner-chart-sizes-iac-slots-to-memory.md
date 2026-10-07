# The Runner Chart Sizes IaC Slots to Its Memory

**Date**: October 7, 2026
**Type**: Fix
**Components**: planton-runner Helm chart, KubernetesPlantonRunner catalog kind

## Summary

**A self-hosted runner no longer takes more IaC operations than its memory holds.**
- **The problem:** the chart gave the runner ten concurrent operations inside a 1Gi limit, and one OpenTofu run with the AWS provider alone peaks near 740Mi. A burst got the pod OOM-killed, and every operation in flight failed with it.
- **What changes:** from the next runner release, the runner reads its own memory limit and takes about one operation per 1Gi beside its own ~768Mi (at most ten). The rest wait in the queue.
- **Defaults:** the chart's default memory limit rises to 2Gi, the smallest that holds one operation.

## What Changed

- `helm/planton-runner`:
  - `temporal.maxConcurrency` is gone, and the runner no longer reads `TEMPORAL_MAX_CONCURRENCY`;
  - the new `iac.maxConcurrency` (default `0`) sizes the IaC worker to the memory limit, and any other value is passed as `IAC_MAX_CONCURRENCY`;
  - `resources.limits.memory` defaults to `2Gi`;
  - the README documents both.
- **KubernetesPlantonRunner:** the spec's `resources` comment, its catalog page, its preset and both modules' comments state the new default and how slots follow the memory limit. The reference page and the proto-docs index are regenerated.

## How It Was Proven

- **Measured peaks, per operation, for engine plus provider plugin:**

  | Engine | Provider | Peak |
  |---|---|---|
  | OpenTofu | Google | ~265Mi |
  | OpenTofu | AWS | ~740Mi |
  | Pulumi | AWS | ~770Mi |
  | Pulumi | Google | ~850Mi |

  The 1Gi budget is the largest of these, rounded up.
- `helm template` renders no `IAC_MAX_CONCURRENCY` by default, and renders it when `iac.maxConcurrency` is set.
