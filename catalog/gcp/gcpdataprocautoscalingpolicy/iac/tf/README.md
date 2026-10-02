# GcpDataprocAutoscalingPolicy - Terraform Module

This Terraform module provisions a Dataproc autoscaling policy (`google_dataproc_autoscaling_policy`) — the reusable resource that governs how Dataproc clusters scale their worker groups on YARN memory pressure. It is the Terraform-side implementation of the Planton `GcpDataprocAutoscalingPolicy` resource kind and has feature parity with the Pulumi module.

## Overview

One policy can govern many clusters: each cluster attaches it by reference (`autoscaling_policy_uri`), so scaling behavior is tuned in one place. Policy contents are mutable — updating re-tunes every attached cluster — but the API refuses to delete a policy while any cluster references it.

The module enables the Dataproc API with `disable_on_destroy=false` so a fresh project works first try and teardown never disables the API project-wide. Zero-valued optional fields (weights, floors, fractions, cooldown) are withheld so GCP's defaults apply, identically to the Pulumi module. The autoscaling-policy API has no labels surface, so no platform attribution labels are stamped. The module runs on the plain `google` provider — every modeled field is GA on the released 8.x line.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpdataprocautoscalingpolicy/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpDataprocAutoscalingPolicy spec | — |

The `spec` object includes: identity (`policy_id`, `location` — both ForceNew; `project_id` empty rides the provider default), `worker_config` (`max_instances` required, `min_instances` 0-or->=2, `weight`), optional `secondary_worker_config` (bounds default 0 — scale-to-zero), and `basic_algorithm` (`cooldown_period` optional; `yarn_config` with the required `graceful_decommission_timeout`, `scale_up_factor`, `scale_down_factor`, plus optional min-worker fractions).

## Outputs

| Name | Description |
|------|-------------|
| `name` | Fully qualified policy resource name (`projects/{p}/locations/{l}/autoscalingPolicies/{id}`) — the handle a cluster's `autoscaling_policy_uri` reference resolves to |
| `policy_id` | The policy ID |
| `location` | Region the policy lives in (echoed from the spec, so callers address the policy without parsing paths) |

## Lifecycle Notes

`policy_id`, `location`, and `project_id` are ForceNew; every bound, weight, and algorithm knob updates in place — re-tuning never recreates the policy or touches attached clusters. Destroy clusters (or detach the policy from them) before destroying the policy: the API rejects deleting a referenced policy.
