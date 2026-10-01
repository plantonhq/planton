# GcpBinaryAuthorizationPolicy — Pulumi Implementation

This directory contains the Pulumi implementation for a project's Binary Authorization policy from the Planton spec: `gcp.projects.Service` for the Binary Authorization API and one `gcp.binaryauthorization.Policy`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `policy` |
| `module/locals.go` | Stack input holder |
| `module/policy.go` | Project resolution, API enablement, the rules, the outputs |
| `module/outputs.go` | Output key constants (`name`, `project_id`) |

## Send Posture (parity with Terraform)

- **`Project`** -- the spec's project, or the provider's project from `organizations.GetClientConfig`.
- **`AdmissionWhitelistPatterns`** -- one entry per spec string.
- **`RequireAttestationsBies`** -- only when non-empty.
- **`Description`**, **`GlobalPolicyEvaluationMode`**, **`DeletionPolicy`** -- only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
