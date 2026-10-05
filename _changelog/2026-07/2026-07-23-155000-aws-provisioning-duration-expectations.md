# AWS catalog: provisioning-duration expectations landed on MWAA and the FSx file systems

**Date**: 2026-07-23
**Scope**: `awsmwaaenvironment` (spec header + regenerated Go stub + docs), docs READMEs for `awsfsxlustrefilesystem`, `awsfsxwindowsfilesystem`, `awsfsxontapfilesystem`, `awsfsxopenzfsfilesystem`. Comments and documentation only — zero behavior change.

## What changed

### MWAA: the "is it hung?" expectation, set before the deploy

Amazon MWAA environment creation is exceptionally slow — the Terraform
provider's own default timeouts allow 120 minutes for create and 90 for
update/delete (verified in the provider source). The docs carried the
subnet-replacement and worker-scaling timings but never the headline
create/delete expectation, so a first deploy that sits "creating" for 40
minutes looks like a hang. The spec header now sets the expectation where
manifest authors read, and the docs gained a Provisioning Times section
covering create, update/replace, and delete.

### FSx file systems: create-duration notes

Each of the four FSx file-system kinds' technical references now states
its create-duration class up front, citing the provider's default create
timeout (Lustre 30 minutes, Windows File Server 45, ONTAP 60, OpenZFS 60)
so deploy windows are budgeted rather than discovered. The faster FSx
satellites (storage virtual machine, volume, data repository association —
30/30/10-minute timeouts) deliberately carry no note: they are below the
surprise threshold.

## Why

The costliest operational surprises are the ones that look like failures:
a slow managed-service create reads as a hung deploy, and the natural
reaction — aborting or re-running — is worse than waiting. The catalog's
contract is that the spec and its docs are enough to operate a kind
correctly, so duration expectations belong on the kind surfaces, not
in tribal memory.

## Validation

- `buf lint` + `buf format --diff` clean on the edited proto directory;
  stub regenerated and coverage verified.
- MWAA spec tests and Pulumi release-entrypoint build green; repo-wide
  `make build-go` green.
- No module, CEL, preset, or chart changed; existing E2E results stand by
  construction.
