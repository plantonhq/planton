# The Runner Starts an IaC Operation Only When Its Memory Holds It

**Date**: October 7, 2026
**Type**: Fix
**Components**: planton-runner Helm chart, Planton operator, KubernetesPlantonRunner and KubernetesPlantonOperator catalog kinds

## Summary

**A self-hosted runner admits IaC operations by the memory it actually holds, and its node reserves all of that memory.**
- **Before:** this morning's chart (0.10.0) described a runner that takes one operation per 1Gi beside ~768Mi. Both figures came from a laptop's resident memory. Measured in the runner image, an operation needs far less than 1Gi, and the runner holds about 45Mi idle.
- **What the runner does now:** from the next runner release, it starts an operation only when its own memory plus 768Mi for every operation started in the last three minutes fits in nine tenths of its limit. The rest wait in the queue, and `iac.maxConcurrency` is the most at once.
- **Why the node must reserve the memory:** the runner admits work against its memory limit, so a request below the limit lets a neighbour take memory the runner counted on. The chart and the operator now request the runner's memory equal to its 2Gi limit.

## What Changed

- **`helm/planton-runner` (0.11.0):**
  - `resources.requests.memory` is 2Gi, equal to the limit;
  - `iac.maxConcurrency` is documented as the most operations at once (0 means the runner's default, 10);
  - the README is updated to match.
- **Operator:** the runner's default sizing is 100m CPU and 2Gi memory, requested and limited alike. The CRD's description of `runner.resources` says so.
- **KubernetesPlantonRunner:**
  - the spec's `resources` comment, its catalog page, the preset and the cost note describe the admission and the new defaults;
  - its default chart is now 0.11.0.
- **KubernetesPlantonOperator:** its default chart is now 0.29.0, the operator release that carries the new runner sizing.
- Stubs, the reference pages and the proto-docs index are regenerated.

## How It Was Proven

**Measurement setup:**
- Each operation ran in the runner image under a 4g limit, alone in its container, so the container's cgroup was that operation's.
- The figure taken is the memory the kernel could not reclaim: usage less clean page cache. The plugins' program files stay out of it, because the kernel drops them under pressure instead of killing the pod.

**Results:**

| Engine | Cloud | Warm plugin cache | Cold cache (first operation of its kind) |
|---|---|---|---|
| OpenTofu | Google | ~150Mi | ~265Mi |
| OpenTofu | AWS | ~455Mi | ~665Mi, after ~70s |
| Pulumi | Google | ~305Mi | ~600Mi |
| Pulumi | AWS | ~285Mi | ~750Mi |

- **The charge:** 768Mi is the worst of these, rounded up.
- **The warm-up:** three minutes is the slowest cold peak, doubled and rounded up.
- **Finished operations:** an operation that finishes gives its charge back at once.
