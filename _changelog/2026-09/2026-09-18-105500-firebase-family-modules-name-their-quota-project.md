# The Firebase-family modules name their quota project

**Date**: September 18, 2026
**Type**: Bug fix
**Components**: `catalog/gcp/gcpfirebaseproject`, `gcpfirebaseandroidapp`, `gcpfirebaseappleapp`, `gcpfirebasewebapp`, `gcpapikey` (both engines); `pkg/iac/pulumi/pulumimodule/provider/gcp/pulumigoogleprovider`

## Summary

The five Firebase-family modules armed `user_project_override` so Google attributes quota to the resource's own project, and that is enough for every resource call. It is not enough for a data-source read: the Firebase app-config and Admin SDK reads carry no project the header can borrow, so under a user's gcloud sign-in the read was attributed to Google's shared ADC project, where the Firebase API is disabled, and failed with `403 ... requires a quota project` after the create had already succeeded. The live proof never met it because it ran under a service-account identity, which needs no quota project. Both engines now name the resource's project as `billing_project` beside the override, the shape Google's own Firebase provider documentation shows.

## What Changed

- **Terraform**: every provider block of the five modules gains `billing_project = local.project_id` beside `user_project_override = true`, with the why written once above the blocks.
- **Pulumi**: `pulumigoogleprovider.GetWithQuotaProject(ctx, config, quotaProject, ...)` arms the override and sets `billingProject`, on both the typed and the keyless (raw-property) arms; an empty project degrades to the override alone. The five modules call it with their spec's project. `GetWithUserProjectOverride` is unchanged for the Identity Platform kinds.
- **Docs**: the kinds' READMEs and module READMEs that described the quota posture say so in one sentence.

## Verification

`go build` and `go vet` on the helper and the five Pulumi modules; `tofu fmt -check` and `tofu validate` on the five Terraform modules; `go test ./pkg/providerparity`; the beta-admission and provider-pin guards. Found live on 2026-09-18 enabling Firebase on a project from a runner-mode user credential: the enablement, the API services, and the three API keys were created; the read-back failed as described. No live re-run in this change; the first apply after the release is the proof.
