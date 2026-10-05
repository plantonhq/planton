# GcpDataprocAutoscalingPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpDataprocAutoscalingPolicySpec defines a reusable Dataproc
autoscaling policy.

An autoscaling policy is a first-class, shareable resource: one
policy can govern many clusters (each cluster attaches it by
reference), so a platform team tunes scaling behavior in one place.
Policy contents are mutable — updating the policy re-tunes every
attached cluster — but a policy cannot be deleted while any cluster
references it.

The autoscaler evaluates YARN memory metrics every cooldown period
and resizes the cluster's primary and secondary worker groups within
the configured bounds.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDataprocAutoscalingPolicy
metadata:
  name: test-autoscaling-policy
spec:
  projectId:
    value: my-gcp-project-123 # replace with your project ID
  policyId: test-etl-autoscaling
  location: us-central1

  # Primary workers: the stable, HDFS-carrying base.
  workerConfig:
    maxInstances: 10
    minInstances: 2
    weight: 1

  # Secondary (spot) workers: the burst arm — weighted 3:1, so ~75% of
  # new capacity lands here; scales to zero when idle.
  secondaryWorkerConfig:
    maxInstances: 20
    minInstances: 0
    weight: 3

  basicAlgorithm:
    # Wait 2 minutes between scaling evaluations.
    cooldownPeriod: "120s"
    yarnConfig:
      # Running tasks get an hour to finish before a scale-down
      # forcefully removes their worker.
      gracefulDecommissionTimeout: "3600s"
      # Add ~50% of the suggested capacity per evaluation on the way up;
      # remove everything idle on the way down.
      scaleUpFactor: 0.5
      scaleDownFactor: 1.0
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.policyId` | `string` | yes |  |  |
| `spec.location` | `string` | yes |  |  |
| `spec.workerConfig` | `GcpDataprocAutoscalingPolicyWorkerConfig` | yes |  |  |
| `spec.workerConfig.maxInstances` | `int32` | yes |  |  |
| `spec.workerConfig.minInstances` | `int32` |  |  |  |
| `spec.workerConfig.weight` | `int32` |  |  |  |
| `spec.secondaryWorkerConfig` | `GcpDataprocAutoscalingPolicySecondaryWorkerConfig` |  |  |  |
| `spec.secondaryWorkerConfig.maxInstances` | `int32` |  |  |  |
| `spec.secondaryWorkerConfig.minInstances` | `int32` |  |  |  |
| `spec.secondaryWorkerConfig.weight` | `int32` |  |  |  |
| `spec.basicAlgorithm` | `GcpDataprocAutoscalingPolicyBasicAlgorithm` | yes |  |  |
| `spec.basicAlgorithm.cooldownPeriod` | `string` |  |  |  |
| `spec.basicAlgorithm.yarnConfig` | `GcpDataprocAutoscalingPolicyYarnConfig` | yes |  |  |
| `spec.basicAlgorithm.yarnConfig.gracefulDecommissionTimeout` | `string` | yes |  |  |
| `spec.basicAlgorithm.yarnConfig.scaleUpFactor` | `double` | yes |  |  |
| `spec.basicAlgorithm.yarnConfig.scaleDownFactor` | `double` | yes |  |  |
| `spec.basicAlgorithm.yarnConfig.scaleUpMinWorkerFraction` | `double` |  |  |  |
| `spec.basicAlgorithm.yarnConfig.scaleDownMinWorkerFraction` | `double` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the policy will be created.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.policyId

`string` · required

The policy ID (the GCP resource name). 3-50 characters: letters,
numbers, underscores, and hyphens; must start and end with a letter
or number. Immutable after creation.

- rule: {"required":true,"string":{"minLen":"3","maxLen":"50","pattern":"^[a-zA-Z0-9]([a-zA-Z0-9_-]{1,48}[a-zA-Z0-9])?$"}}

### spec.location

`string` · required

GCP region the policy lives in (e.g., "us-central1"). A cluster can
only attach policies in its own region. Immutable after creation.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.workerConfig

`GcpDataprocAutoscalingPolicyWorkerConfig` · required

Autoscaling bounds and weight for the primary worker group.

- rule: {"required":true}
- rule: max_instances must be greater than or equal to min_instances

### spec.workerConfig.maxInstances

`int32` · required

Maximum number of primary workers the autoscaler may create.

- rule: {"required":true,"int32":{"gte":1}}

### spec.workerConfig.minInstances

`int32`

Minimum number of primary workers. The Dataproc API requires at
least 2 primary workers on an autoscaled cluster; 0 leaves the
API's default (2).

- rule: min_instances must be at least 2 (or 0 to accept the API default of 2)

### spec.workerConfig.weight

`int32`

Relative weight of this group when the autoscaler distributes new
capacity between primary and secondary workers. Default: 1.
Example: primary weight 1 + secondary weight 3 sends ~75% of new
nodes to the secondary (spot) group.

- rule: {"int32":{"gte":0}}

### spec.secondaryWorkerConfig

`GcpDataprocAutoscalingPolicySecondaryWorkerConfig`

Autoscaling bounds and weight for the secondary (preemptible/spot)
worker group. Omit to keep the secondary group unscaled.

- rule: max_instances must be greater than or equal to min_instances when min_instances is set

### spec.secondaryWorkerConfig.maxInstances

`int32`

Maximum number of secondary workers the autoscaler may create.
Default: 0 (secondary group not scaled).

- rule: {"int32":{"gte":0}}

### spec.secondaryWorkerConfig.minInstances

`int32`

Minimum number of secondary workers. Default: 0 — the group can
scale to zero when idle.

- rule: {"int32":{"gte":0}}

### spec.secondaryWorkerConfig.weight

`int32`

Relative weight of this group when the autoscaler distributes new
capacity between primary and secondary workers. Default: 1.

- rule: {"int32":{"gte":0}}

### spec.basicAlgorithm

`GcpDataprocAutoscalingPolicyBasicAlgorithm` · required

The autoscaling algorithm: evaluation cadence + YARN scaling
behavior.

- rule: {"required":true}

### spec.basicAlgorithm.cooldownPeriod

`string`

How long the autoscaler waits between evaluations, letting the
cluster settle after a scaling event. Bounds: 2 minutes (120s) to
1 day (86400s) — the Dataproc API's own limits, enforced below so
an out-of-range value fails at spec time instead of at the
provider. Format: duration in seconds with "s" suffix.
Default: "120s".

- rule: cooldown_period must be a duration in seconds (e.g., '120s')
- rule: cooldown_period must be between 2 minutes (120s) and 1 day (86400s)

### spec.basicAlgorithm.yarnConfig

`GcpDataprocAutoscalingPolicyYarnConfig` · required

YARN memory-based scaling behavior.

- rule: {"required":true}

### spec.basicAlgorithm.yarnConfig.gracefulDecommissionTimeout

`string` · required

How long the autoscaler waits for a graceful YARN decommission
before forcefully removing a worker during scale-down. Running
tasks get this window to finish. Bounds: 0s to 1 day (86400s) —
the Dataproc API's own limit, enforced below so an out-of-range
value fails at spec time instead of at the provider.
Format: duration in seconds with "s" suffix (e.g., "3600s").

- rule: graceful_decommission_timeout must be at most 1 day (86400s)
- rule: {"required":true,"string":{"pattern":"^[0-9]+s$"}}

### spec.basicAlgorithm.yarnConfig.scaleUpFactor

`double` · required · optional (explicit presence)

Fraction of pending YARN memory the autoscaler adds capacity for
per evaluation (0.0-1.0). 1.0 scales up as fast as possible; 0.05
adds ~5% of the suggested capacity per cooldown period.
Declared optional (explicit presence) so 0.0 — a legitimate API
value — is expressible while the field itself stays required.

- rule: {"required":true,"double":{"lte":1,"gte":0}}

### spec.basicAlgorithm.yarnConfig.scaleDownFactor

`double` · required · optional (explicit presence)

Fraction of available (idle) YARN memory the autoscaler removes
capacity for per evaluation (0.0-1.0). 1.0 scales down as fast as
possible; 0.0 disables scale-down entirely (ever-growing cluster —
pair with idle_delete_ttl on the cluster for cost control).
Declared optional (explicit presence) so 0.0 — a legitimate API
value — is expressible while the field itself stays required.

- rule: {"required":true,"double":{"lte":1,"gte":0}}

### spec.basicAlgorithm.yarnConfig.scaleUpMinWorkerFraction

`double`

Minimum fractional change to the cluster size the autoscaler acts
on when scaling up (0.0-1.0). Example: 0.05 means a recommendation
must grow the cluster by at least 5% to trigger; 0.0 (default)
means any recommendation of at least one worker triggers.

- rule: {"double":{"lte":1,"gte":0}}

### spec.basicAlgorithm.yarnConfig.scaleDownMinWorkerFraction

`double`

Minimum fractional change to the cluster size the autoscaler acts
on when scaling down (0.0-1.0). Default: 0.0.

- rule: {"double":{"lte":1,"gte":0}}

### spec.deletionPolicy

`string`

Engine-side teardown behavior. "DELETE" (default) deletes the
policy; "PREVENT" fails any plan that would delete it; "ABANDON"
removes it from IaC management while leaving it in GCP. Note the API
refuses to delete a policy while any cluster still references it,
regardless of this setting.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpDataprocAutoscalingPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Fully qualified policy resource name as computed by the provider — projects/{project}/regions\|locations/{location}/autoscalingPolicies/{policy_id} (the Dataproc API treats the regions and locations path segments as equivalent, and both engines compute the identical form). The handle a Dataproc cluster's autoscaling_policy_uri reference resolves to. |
| `status.outputs.policy_id` | `string` | The policy ID (same as the spec's policy_id input). |
| `status.outputs.location` | `string` | The region the policy lives in (same as the spec's location input). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDataprocCluster | `spec.clusterConfig.autoscalingPolicyUri` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
