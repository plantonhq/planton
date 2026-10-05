# GcpDeployPolicy — Pulumi Implementation

This directory contains the Pulumi implementation for a Cloud Deploy deploy policy from the Planton spec: `gcp.projects.Service` and one `gcp.clouddeploy.DeployPolicy`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `deployPolicy` |
| `module/locals.go` | IaC input, attribution labels, the label merge |
| `module/deploy_policy.go` | API enablement, the policy, one builder per nested block, the send-only-when-set helpers, the outputs |
| `module/outputs.go` | Output key constants (`name`, `deploy_policy_id`, `uid`) |

## Send Posture (parity with Terraform)

- **Optional fields** -- `optionalString`, `optionalInt`, and `optionalStrings` return nil for empty values; `Suspended` only when true.
- **Clock and date parts** -- each sent only when non-zero; a weekly window's times only when declared.
- **Labels** -- attribution labels on the policy's `Labels`, never on a selector's labels.
- **Name** -- set explicitly from `deployPolicyId` (default `metadata.name`), never auto-named.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
