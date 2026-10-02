# GcpCloudBuildTrigger — Pulumi Implementation

This directory contains the Pulumi implementation for a Cloud Build trigger from the Planton spec: `gcp.projects.Service` and one `gcp.cloudbuild.Trigger`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `trigger` |
| `module/locals.go` | Stack input |
| `module/trigger.go` | API enablement, the trigger, one builder per event source and build block, the send-when-set helpers, the outputs |
| `module/outputs.go` | Output key constants (`id`, `trigger_id`, `name`) |

## Send Posture (parity with Terraform)

- **Optional fields** -- set only when declared (`optionalString`, `optionalInt`, `optionalTrue`, `optionalStrings`, `optionalStringMap`); every block builder returns `nil` when its block is unset.
- **Name** -- set explicitly from `triggerName` (default `metadata.name`), never auto-named; an empty `location` leaves the provider's `global`.
- **SDK names** -- the spec's `steps`, `secrets`, `env`, `secretEnv`, `waitFor`, `secretManager`, and `sourceProvenanceHash` go to the SDK's `Steps`, `Secrets`, `Envs`, `SecretEnvs`, `WaitFors`, `SecretManagers`, and `SourceProvenanceHashes`.
- **Never sent** -- `DynamicSubstitutions`, `SubstitutionOption`, and a step's `Timing`.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
