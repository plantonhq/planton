# GcpBinaryAuthorizationPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpBinaryAuthorizationPolicySpec is a project's Binary Authorization
policy (`google_binary_authorization_policy`): the rule GKE applies to
every pod creation -- allow, deny, or require that each container image
was signed by trusted attestors (GcpBinaryAuthorizationAttestor) -- with
per-cluster overrides and exempt image patterns.

GKE clusters enforce it when they set
binary_authorization_evaluation_mode to PROJECT_SINGLETON_POLICY_ENFORCE
(GcpGkeCluster). Roll a new rule out with enforcement_mode
DRYRUN_AUDIT_LOG_ONLY first: denials are logged, nothing is blocked.

Google keeps exactly one policy per project. Applying this block replaces
whatever policy the project had (every apply sends the whole policy).
Destroying it (deletion_policy DELETE, the default) does not leave the
project without a policy: Google's default is written back -- allow
every image, enforced, with gcr.io/google_containers/* exempt.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBinaryAuthorizationPolicy
metadata:
  name: prod-policy
spec:
  projectId:
    value: my-gcp-project
  globalPolicyEvaluationMode: ENABLE
  defaultAdmissionRule:
    evaluationMode: REQUIRE_ATTESTATION
    enforcementMode: DRYRUN_AUDIT_LOG_ONLY
    requireAttestationsBy:
      - value: projects/my-gcp-project/attestors/built-by-ci
  clusterAdmissionRules:
    - cluster: us-central1.sandbox
      evaluationMode: ALWAYS_ALLOW
      enforcementMode: DRYRUN_AUDIT_LOG_ONLY
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.description` | `string` |  |  |  |
| `spec.globalPolicyEvaluationMode` | `string` |  |  |  |
| `spec.admissionWhitelistPatterns` | `[]string` |  |  |  |
| `spec.defaultAdmissionRule` | `GcpBinaryAuthorizationPolicyAdmissionRule` | yes |  |  |
| `spec.defaultAdmissionRule.evaluationMode` | `string` | yes |  |  |
| `spec.defaultAdmissionRule.enforcementMode` | `string` | yes |  |  |
| `spec.defaultAdmissionRule.requireAttestationsBy` | `[]string \| valueFrom` |  |  | GcpBinaryAuthorizationAttestor (`status.outputs.attestor_id`) |
| `spec.clusterAdmissionRules` | `[]GcpBinaryAuthorizationPolicyClusterAdmissionRule` |  |  |  |
| `spec.clusterAdmissionRules[].cluster` | `string` | yes |  |  |
| `spec.clusterAdmissionRules[].evaluationMode` | `string` | yes |  |  |
| `spec.clusterAdmissionRules[].enforcementMode` | `string` | yes |  |  |
| `spec.clusterAdmissionRules[].requireAttestationsBy` | `[]string \| valueFrom` |  |  | GcpBinaryAuthorizationAttestor (`status.outputs.attestor_id`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The project the policy governs: a literal project ID or a GcpProject
reference. Empty means the provider's default project. The module
enables binaryauthorization.googleapis.com there.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.description

`string`

What the policy is for.

### spec.globalPolicyEvaluationMode

`string`

Google's global policy for its own system images (GKE add-ons and
similar), evaluated before this policy:
  "ENABLE"  -- Google-maintained system images are always admitted, so
               a strict default rule cannot break the cluster's own
               components (recommended with REQUIRE_ATTESTATION)
  "DISABLE" -- system images face this policy like any other image
Empty sends nothing and keeps Google's current value.

- rule: global_policy_evaluation_mode must be ENABLE or DISABLE

### spec.admissionWhitelistPatterns

`[]string`

Image name patterns admitted regardless of every rule, in the form
registry/path/to/image; a trailing * is a wildcard and may appear only
after the registry/ part, e.g. "us-docker.pkg.dev/my-project/base/*".
Use them for images no build pipeline signs.

- rule: {"repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.defaultAdmissionRule

`GcpBinaryAuthorizationPolicyAdmissionRule` · required

The rule for every cluster without its own rule below. Required.

- rule: {"required":true}
- rule: require_attestations_by is required for REQUIRE_ATTESTATION and must be empty otherwise

### spec.defaultAdmissionRule.evaluationMode

`string` · required

How images are judged:
  "ALWAYS_ALLOW"        -- every image is admitted
  "REQUIRE_ATTESTATION" -- an image is admitted only when every attestor
                           in require_attestations_by has signed it
  "ALWAYS_DENY"         -- every image is denied (lock a cluster down)

- rule: {"required":true,"string":{"in":["ALWAYS_ALLOW","REQUIRE_ATTESTATION","ALWAYS_DENY"]}}

### spec.defaultAdmissionRule.enforcementMode

`string` · required

What a denial does:
  "ENFORCED_BLOCK_AND_AUDIT_LOG" -- the pod is blocked and the denial
                                    logged
  "DRYRUN_AUDIT_LOG_ONLY"        -- the pod runs and the would-be denial
                                    is logged (the safe rollout mode)

- rule: {"required":true,"string":{"in":["ENFORCED_BLOCK_AND_AUDIT_LOG","DRYRUN_AUDIT_LOG_ONLY"]}}

### spec.defaultAdmissionRule.requireAttestationsBy

`[]string | valueFrom`

The attestors that must all have signed an image: GcpBinaryAuthorizationAttestor
references or literals. A bare attestor name resolves to the policy's
project; an attestor in another project must be the full
projects/{project}/attestors/{attestor}. Each attestor must exist
before the policy names it.

- references: GcpBinaryAuthorizationAttestor (`status.outputs.attestor_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBinaryAuthorizationAttestor, name: <that resource's name>, fieldPath: status.outputs.attestor_id}} -- a bare string does not parse

### spec.clusterAdmissionRules

`[]GcpBinaryAuthorizationPolicyClusterAdmissionRule`

Per-cluster rules, one per cluster, overriding the default rule there.

- rule: each cluster may have only one rule
- rule: require_attestations_by is required for REQUIRE_ATTESTATION and must be empty otherwise

### spec.clusterAdmissionRules[].cluster

`string` · required

The cluster, as Google keys it: {location}.{cluster_name}, where the
location is the cluster's zone (e.g. "us-central1-a.prod") or region
(e.g. "us-central1.prod").

- rule: cluster must be {location}.{cluster_name}, e.g. us-central1.prod or us-central1-a.prod
- rule: {"required":true}

### spec.clusterAdmissionRules[].evaluationMode

`string` · required

How images are judged on this cluster; see
GcpBinaryAuthorizationPolicyAdmissionRule.evaluation_mode.

- rule: {"required":true,"string":{"in":["ALWAYS_ALLOW","REQUIRE_ATTESTATION","ALWAYS_DENY"]}}

### spec.clusterAdmissionRules[].enforcementMode

`string` · required

What a denial does on this cluster; see
GcpBinaryAuthorizationPolicyAdmissionRule.enforcement_mode.

- rule: {"required":true,"string":{"in":["ENFORCED_BLOCK_AND_AUDIT_LOG","DRYRUN_AUDIT_LOG_ONLY"]}}

### spec.clusterAdmissionRules[].requireAttestationsBy

`[]string | valueFrom`

The attestors that must all have signed an image on this cluster; see
GcpBinaryAuthorizationPolicyAdmissionRule.require_attestations_by.

- references: GcpBinaryAuthorizationAttestor (`status.outputs.attestor_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBinaryAuthorizationAttestor, name: <that resource's name>, fieldPath: status.outputs.attestor_id}} -- a bare string does not parse

### spec.deletionPolicy

`string`

What destroying this block does:
  "" / "DELETE" -- Google's default policy (allow every image) is
                   written back
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the block leaves management and this policy stays in
                   force

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpBinaryAuthorizationPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | The policy's resource name: projects/{project}/policy. |
| `status.outputs.project_id` | `string` | The project the policy governs. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.defaultAdmissionRule.requireAttestationsBy` | GcpBinaryAuthorizationAttestor | `status.outputs.attestor_id` |
| `spec.clusterAdmissionRules[].requireAttestationsBy` | GcpBinaryAuthorizationAttestor | `status.outputs.attestor_id` |

## See Also

- [Overview](../README.md)
