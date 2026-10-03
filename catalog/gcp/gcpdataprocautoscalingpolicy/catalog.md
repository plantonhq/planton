# GCP Dataproc Autoscaling Policy

Deploys a reusable Dataproc autoscaling policy: the scaling contract that governs how attached clusters add primary and secondary (spot) workers when YARN memory is pending and remove them when it is idle. One policy can govern many clusters — a platform team tunes scaling behavior in one place, and editing the policy re-tunes every attached cluster in place. The API refuses to delete a policy while any cluster still references it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Dataproc API enablement** -- `dataproc.googleapis.com` is enabled in the target project (never disabled on destroy, so tearing down one policy cannot break the rest of the project)
- **Dataproc Autoscaling Policy** -- a regional policy resource with the configured worker bounds, capacity weights, and YARN scaling algorithm
- **Primary Worker Bounds** -- the min/max envelope and capacity weight for the on-demand worker group (the stable HDFS base)
- **Secondary Worker Bounds** -- created only when `secondaryWorkerConfig` is set; the envelope and weight for the preemptible/spot group (the cost-optimized burst arm)
- **YARN Scaling Algorithm** -- the scale-up/scale-down factors (what fraction of suggested capacity change is applied per evaluation), the graceful decommission window running tasks get before scale-down removes their worker, deadband fractions, and the evaluation cooldown

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the policy will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef; the module enables the Dataproc API itself.
- **Region co-location** -- a Dataproc cluster can only attach policies in its own region; create the policy in the region your clusters run in (`location` is immutable after creation).

## Deploy

### Console

Open the deployment store, find **GCP Dataproc Autoscaling Policy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Balanced Autoscaling** preset in the [Presets](#presets) tab for a moderate general-purpose policy.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDataprocAutoscalingPolicy
metadata:
  name: balanced-autoscaling
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  policyId: balanced-autoscaling
  location: us-central1
  workerConfig:
    maxInstances: 10
    minInstances: 2
  basicAlgorithm:
    yarnConfig:
      gracefulDecommissionTimeout: "1800s"
      scaleUpFactor: 0.5
      scaleDownFactor: 0.5
```

```shell
planton apply -f autoscaling-policy.yaml
```

This creates a policy that lets attached clusters grow their primary group to 10 workers and shrink back to 2, moving at half the suggested capacity change per evaluation, with a 30-minute drain window before any worker is removed. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying the project and policy together, wire the project reference with ValueFromRef:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: data-platform
      fieldPath: status.outputs.project_id
  policyId: balanced-autoscaling
  location: us-central1
```

The InfraPipeline resolves the dependency graph, creates the project first, then the policy inside it. A GcpDataprocCluster in the same chart attaches the policy by referencing this component's `status.outputs.name`, so the policy always deploys before the clusters that use it.

## Key Configuration

These are the most important decisions when configuring an autoscaling policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scale factors** -- `scaleUpFactor` and `scaleDownFactor` (both required, 0.0-1.0) set the policy's personality: the fraction of YARN's suggested capacity change applied per evaluation. 1.0 reacts at maximum speed; 0.05 creeps smoothly. An explicit `scaleDownFactor: 0.0` disables scale-down entirely — an ever-growing cluster; pair it with the cluster's idle-delete TTL for cost control.

**Worker bounds and weights** -- `workerConfig.maxInstances` (required) is the primary group's hard ceiling; `minInstances: 0` accepts the API's default floor of 2 (Dataproc requires at least 2 primary workers on an autoscaled cluster). The weight pair splits new capacity between on-demand primaries and spot secondaries — weight 1:3 sends ~75% of new nodes to the spot group. Leave `secondaryWorkerConfig` unset to keep the secondary group unscaled.

**Graceful decommission** -- `gracefulDecommissionTimeout` (required, e.g. `"1800s"`, bounded at 1 day) is the window running tasks get to finish before scale-down removes their worker. Short windows kill long-running tasks mid-flight; long windows delay cost savings.

**Cooldown and deadbands** -- `basicAlgorithm.cooldownPeriod` sets the evaluation cadence (blank keeps GCP's `"120s"` default; longer smooths scaling but slows every reaction). `scaleUpMinWorkerFraction` / `scaleDownMinWorkerFraction` filter out recommendations smaller than that fraction of the cluster — the noise gate for SLA-bound clusters where oscillation hurts.

**Identity is immutable** -- `policyId` and `location` cannot change after creation; everything else is mutable, and an edit re-tunes every attached cluster on the next evaluation.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | Fully qualified policy resource name (`projects/{p}/locations/{l}/autoscalingPolicies/{id}`) | The exact value a GcpDataprocCluster's `autoscalingPolicyUri` attaches |
| `policy_id` | The short policy ID | Addressing the policy in gcloud and Dataproc API calls |
| `location` | The policy's region | Region co-location checks when placing clusters |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Balanced autoscaling** -- Half-strength factors, an even primary/secondary split, and a 30-minute drain window. A sensible default for mixed interactive and batch clusters. Start from the **Balanced Autoscaling** preset.

**Aggressive batch** -- Full-strength factors on a spot-heavy weight mix with a scale-to-zero secondary group, for retryable batch pipelines where reaction speed wins. Start from the **Aggressive Batch** preset.

**Conservative production** -- Small factors, minimum-change fractions that filter scaling noise, a long cooldown, and a 2-hour drain window for SLA-bound shared clusters where oscillation hurts interactive users. Start from the **Conservative Production** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the policy is created
- [**GCP Dataproc Cluster**](/infra-catalog/gcp-dataproc-cluster) -- attaches this policy via `autoscalingPolicyUri`; one policy can govern many clusters
