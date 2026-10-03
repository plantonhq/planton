# GcpDataprocAutoscalingPolicy - Pulumi Module

This Pulumi (Go) module provisions a Dataproc autoscaling policy (`dataproc.AutoscalingPolicy`) — the reusable resource that governs how Dataproc clusters scale their worker groups on YARN memory pressure. It is the Pulumi-side implementation of the Planton `GcpDataprocAutoscalingPolicy` resource kind and has feature parity with the Terraform module.

## Overview

One policy can govern many clusters: each cluster attaches it by reference (`autoscalingPolicyUri`), so scaling behavior is tuned in one place. Policy contents are mutable — updating re-tunes every attached cluster — but the API refuses to delete a policy while any cluster references it.

The module enables the Dataproc API with `disable_on_destroy=false` so a fresh project works first try and teardown never disables the API project-wide. Zero-valued optional fields (weights, floors, fractions, cooldown) are withheld so GCP's defaults apply, identically to the Terraform module. The autoscaling-policy API has no labels surface, so no platform attribution labels are stamped.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpdataprocautoscalingpolicy/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/main.go` — provider setup and resource orchestration
- `module/locals.go` — metadata-derived values
- `module/autoscaling_policy.go` — API enablement + the policy resource + exports
- `module/outputs.go` — output keys (must match `outputs.proto`)

## Outputs

| Name | Description |
|------|-------------|
| `name` | Fully qualified policy resource name (`projects/{p}/locations/{l}/autoscalingPolicies/{id}`) — the handle a cluster's `autoscalingPolicyUri` reference resolves to |
| `policy_id` | The policy ID |
| `location` | Region the policy lives in (echoed from the spec, so callers address the policy without parsing paths) |

## Lifecycle Notes

`policyId`, `location`, and `projectId` are ForceNew; every bound, weight, and algorithm knob updates in place — re-tuning never recreates the policy or touches attached clusters. Destroy clusters (or detach the policy from them) before destroying the policy: the API rejects deleting a referenced policy.
