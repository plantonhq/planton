# GCP Binary Authorization Policy

A project's Binary Authorization policy: the rule GKE applies to every pod creation -- allow, deny, or admit only container images signed by trusted attestors -- with per-cluster overrides and image patterns that are always admitted. It is how a platform team makes sure only images its build pipeline produced, scanned, and signed can run in production.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `binaryauthorization.googleapis.com` on the project (never disabled on destroy)
- **Policy** -- the project's `binary_authorization_policy`, replacing whatever policy it had

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Binary Authorization policy admin permissions (`roles/binaryauthorization.policyEditor`) on the project, and read access to every attestor the policy names.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Enforcing Clusters

GKE clusters enforce the policy when they set `binaryAuthorizationEvaluationMode: PROJECT_SINGLETON_POLICY_ENFORCE` (`GcpGkeCluster`).

### Optional Dependencies

- **`GcpBinaryAuthorizationAttestor`** -- attestors a rule requires (`requireAttestationsBy`). Each must exist before the policy names it.
- **`GcpProject`** -- the project, by reference (`projectId`).

## Deploy

### CLI

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
```

```shell
planton apply -f binary-authorization-policy.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `defaultAdmissionRule` | `object` | `evaluationMode` (`ALWAYS_ALLOW` / `REQUIRE_ATTESTATION` / `ALWAYS_DENY`), `enforcementMode` (`ENFORCED_BLOCK_AND_AUDIT_LOG` / `DRYRUN_AUDIT_LOG_ONLY`), `requireAttestationsBy` (attestor refs, with `REQUIRE_ATTESTATION` only). |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The project the policy governs (`GcpProject` ref). |
| `description` | `string` | none | What the policy is for. |
| `globalPolicyEvaluationMode` | `string` | not sent | `ENABLE` always admits Google-maintained system images; `DISABLE` makes them face the policy. |
| `admissionWhitelistPatterns` | `string[]` | none | Image patterns always admitted, e.g. `us-docker.pkg.dev/my-project/base/*`. |
| `clusterAdmissionRules` | `object[]` | none | One rule per cluster (`cluster` as `{location}.{cluster_name}`), overriding the default rule. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` writes Google's default policy (allow every image) back, `PREVENT` fails destroy, `ABANDON` keeps this policy in force. |

### Validation Rules

- `requireAttestationsBy` is required for `REQUIRE_ATTESTATION` and empty otherwise.
- Each cluster has at most one rule; `cluster` is `{location}.{cluster_name}`.
- Modes take only Google's values.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/policy` |
| `project_id` | `string` | The project the policy governs |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Destroy allows everything.** Destroying the block does not leave the project without a policy: Google's default is written back -- allow every image, enforced, with `gcr.io/google_containers/*` exempt. Use `PREVENT` on production policies.
- **One policy per project.** Applying replaces whatever policy the project had; every apply sends the whole policy.
- **Roll out in dry run.** `DRYRUN_AUDIT_LOG_ONLY` logs what would be denied without blocking anything.
- **Admit Google's system images.** With a strict default rule, `globalPolicyEvaluationMode: ENABLE` keeps GKE's own components running.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpBinaryAuthorizationAttestor** -- the signers a rule requires
- **GcpGkeCluster** -- enforces the policy with `PROJECT_SINGLETON_POLICY_ENFORCE`
- **GcpKmsKey** -- signing keys held in Cloud KMS

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
