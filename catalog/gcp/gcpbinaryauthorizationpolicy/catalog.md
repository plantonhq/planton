# GCP Binary Authorization Policy

Decides which container images GKE may run in a project. Allow everything, deny everything, or admit only images your trusted pipelines have signed -- per cluster if needed, with a dry-run mode that logs what would be blocked before anything is.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `binaryauthorization.googleapis.com` on the project
- **Policy** -- the project's `binaryauthorization.Policy`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Binary Authorization policy admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP Binary Authorization Policy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Signed Images Dry Run** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBinaryAuthorizationPolicy
metadata:
  name: prod-policy
  org: acme-corp
  env: prod
spec:
  projectId:
    value: shop-prod
  globalPolicyEvaluationMode: ENABLE
  defaultAdmissionRule:
    evaluationMode: REQUIRE_ATTESTATION
    enforcementMode: DRYRUN_AUDIT_LOG_ONLY
    requireAttestationsBy:
      - value: projects/shop-security/attestors/built-by-ci
```

```shell
planton apply -f binary-authorization-policy.yaml
```

This logs every pod whose images the CI attestor has not signed, without blocking any yet. An Infra Job tracks the provisioning in real time.

### InfraChart

Reference `GcpBinaryAuthorizationAttestor` resources from `requireAttestationsBy` (their `status.outputs.attestor_id`), and set `binaryAuthorizationEvaluationMode: PROJECT_SINGLETON_POLICY_ENFORCE` on the project's `GcpGkeCluster` resources so they enforce it.

## Key Configuration

These are the most important decisions when configuring a policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The default rule** -- `ALWAYS_ALLOW`, `REQUIRE_ATTESTATION` (with the attestors that must all have signed), or `ALWAYS_DENY`.

**Enforce or dry run** -- `ENFORCED_BLOCK_AND_AUDIT_LOG` blocks; `DRYRUN_AUDIT_LOG_ONLY` only logs. Start with dry run.

**Exceptions** -- per-cluster rules, always-admitted image patterns, and Google's global policy for system images.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpBinaryAuthorizationAttestor** | `defaultAdmissionRule.requireAttestationsBy`, `clusterAdmissionRules[].requireAttestationsBy` | `status.outputs.attestor_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The policy's resource name | Audits |
| `project_id` | The project the policy governs | Reporting |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Signed images dry run** -- require the CI attestor everywhere, logging only. Start from the **Signed Images Dry Run** preset.

**Enforced with sandbox exception** -- enforce signatures and let one sandbox cluster run anything. Start from the **Enforced With Sandbox Cluster** preset.

## Works With

- [**GCP Binary Authorization Attestor**](/infra-catalog/gcp-binary-authorization-attestor) -- the signers a rule requires
- [**GCP GKE Cluster**](/infra-catalog/gcp-gke-cluster) -- enforces the policy
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- signing keys
