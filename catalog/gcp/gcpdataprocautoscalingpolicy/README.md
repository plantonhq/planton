# GCP Dataproc Autoscaling Policy

Deploys a Dataproc autoscaling policy (`google_dataproc_autoscaling_policy`) — the reusable, shareable resource that governs how Dataproc clusters scale their worker groups on YARN memory pressure. One policy can govern many clusters: each cluster attaches it by reference, so a platform team tunes scaling behavior in one place.

## Overview

Dataproc autoscaling is not inline cluster configuration. The policy is a first-class regional resource holding the scaling bounds and algorithm; a `GcpDataprocCluster` attaches it through `clusterConfig.autoscalingPolicyUri` (an in-place update on the cluster). The autoscaler then evaluates YARN memory metrics every cooldown period and resizes the cluster's primary and secondary worker groups within the policy's bounds.

Three properties define the operational model:

- **Shared** — many clusters, one policy; each cluster can only attach policies in its own region.
- **Mutable** — policy contents update in place, and every attached cluster re-tunes immediately. Only `policyId`, `location`, and `projectId` are immutable.
- **Delete-protected** — the API refuses to delete a policy while any cluster references it. Detach (or delete) the clusters first.

## Quick Start

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDataprocAutoscalingPolicy
metadata:
  name: etl-autoscaling
spec:
  policyId: etl-autoscaling
  location: us-central1
  workerConfig:
    maxInstances: 10
    minInstances: 2
  secondaryWorkerConfig:
    maxInstances: 20
    weight: 3
  basicAlgorithm:
    cooldownPeriod: "120s"
    yarnConfig:
      gracefulDecommissionTimeout: "3600s"
      scaleUpFactor: 0.5
      scaleDownFactor: 1.0
```

```shell
planton apply -f policy.yaml
```

Then attach it from a cluster:

```yaml
# In a GcpDataprocCluster spec:
clusterConfig:
  autoscalingPolicyUri:
    valueFrom:
      kind: GcpDataprocAutoscalingPolicy
      name: etl-autoscaling
      fieldPath: status.outputs.name
```

## Configuration Options

| Category | Options |
|----------|---------|
| **Identity** | `policyId` — 3-50 chars (letters, digits, underscores, hyphens; starts/ends alphanumeric); immutable. `location` — the policy's region (clusters attach same-region only); immutable. `projectId` — optional; omitted rides the provider default; reference `GcpProject` |
| **Primary bounds** | `workerConfig.maxInstances` (required, >=1), `minInstances` (0 accepts the API default of 2; otherwise >=2 — the Dataproc floor for autoscaled clusters), `weight` |
| **Secondary bounds** | `secondaryWorkerConfig` — optional; `maxInstances`/`minInstances` default 0 (scale-to-zero), `weight` |
| **Cadence** | `basicAlgorithm.cooldownPeriod` — seconds-suffixed duration (e.g. `"120s"`, the default); bounds 2 minutes to 1 day |
| **Scale-up** | `yarnConfig.scaleUpFactor` (required, 0.0-1.0) — fraction of pending YARN memory acted on per evaluation; `scaleUpMinWorkerFraction` — minimum fractional change worth acting on |
| **Scale-down** | `yarnConfig.scaleDownFactor` (required, 0.0-1.0) — fraction of idle YARN memory removed per evaluation; `scaleDownMinWorkerFraction` |
| **Drain window** | `yarnConfig.gracefulDecommissionTimeout` (required, e.g. `"3600s"`; bounds 0s to 1 day) — how long running tasks get to finish before a worker is forcefully removed |

### Weight semantics

`weight` steers where new capacity lands, proportionally across the two groups. Primary weight 1 + secondary weight 3 sends ~75% of new nodes to the secondary (spot) group. Primaries carry HDFS DataNodes, so they are the stable base; secondaries carry no data and absorb aggressive add/remove cycles.

### Factor semantics

The factors express how much of the metric-suggested change the autoscaler applies per evaluation: `1.0` is maximally aggressive (act on everything), `0.05` moves in ~5% steps. The min-worker fractions filter noise — e.g. `0.1` ignores any recommendation that would change the cluster by less than 10%.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | string | Fully qualified policy resource name (`projects/{p}/locations/{l}/autoscalingPolicies/{id}`) — the handle a cluster's `autoscalingPolicyUri` reference resolves to |
| `policy_id` | string | The policy ID |
| `location` | string | Region the policy lives in |

## Important Notes

- **One policy, many clusters**: attach the same policy from any number of same-region clusters; editing the policy re-tunes all of them in place.
- **Deletion ordering**: the API rejects deleting a policy while any cluster references it — detach or destroy the clusters first.
- **Region locality**: a cluster can only attach policies in its own region; a multi-region estate needs a policy per region.
- **`minInstances` of 1 is invalid** on the primary group: the Dataproc API requires at least 2 primary workers on an autoscaled cluster (0 accepts the default of 2).
- **`scaleDownFactor` pairs with cluster TTLs**: gentle scale-down keeps capacity around; use the cluster's `lifecycleConfig.idleDeleteTtl` as the backstop cost control.

### Deliberately not modeled (recorded reasons)

Nothing. Every configurable argument of `google_dataproc_autoscaling_policy` at the pinned provider is representable through this spec — including `deletionPolicy` (the API's own referenced-by-a-cluster guard still applies on top of it).

## Related Kinds

- **GcpDataprocCluster** — attaches this policy via `clusterConfig.autoscalingPolicyUri`
- **GcpProject** — provides the GCP project ID

## Additional Resources

- [Autoscaling clusters](https://cloud.google.com/dataproc/docs/concepts/configuring-clusters/autoscaling)
- [AutoscalingPolicies REST API](https://cloud.google.com/dataproc/docs/reference/rest/v1/projects.locations.autoscalingPolicies)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
