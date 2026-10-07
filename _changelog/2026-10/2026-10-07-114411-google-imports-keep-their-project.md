# Google Imports Keep Their Project

**Date**: October 7, 2026
**Type**: Fix
**Components**: every Google OpenTofu module's provider block, the Pulumi Google provider builder and every Google Pulumi module, the provider-block guard, the GCP provider-parity manifest

## Summary

**Taking over an existing Google resource no longer plans to replace it.**
- **The bug:** a bucket imported by its name landed in state with `project = ""`. The next preview read `project = "" -> "<project>"`, a change that forces replacement, so the preview wanted to destroy and recreate a live bucket.
- **The cause:** no Google module told the provider which project it works in. Google's provider fills an imported resource's project from the provider's default project when the import ID carries none, and the default was empty. The same holds on Pulumi, because pulumi-gcp is bridged from the same provider.
- **The fix:** every Google module whose resources take a project from the spec now hands that project to its provider, on both engines. The project always comes from the component's own configuration, never from the connection. A spec that names no project keeps Google's ambient default, as before.

Resources created by a deploy are unaffected. They recorded their project when they were created, so their previews show no difference. Only imports change, and now every import form records the project.

## What Changed

- **OpenTofu (166 modules):**
  - every `provider "google"` and `provider "google-beta"` block carries the same lines, `project = local.project_id`, under one comment;
  - `local.project_id` is the spec's project or null.
  - Twelve modules resolve their project through a `google_client_config` fallback, which the provider cannot read without depending on its own data source. Each gains a provider-safe `local.project_id`: `gcpbinaryauthorization*`, `gcpgkefleet`, `gcpkmskeyhandle`, the three `gcpscc*` kinds, `gcpkmsautokeyconfig`, `gcpmodelarmorfloorsetting`, `gcpprojectiammember`, `gcpsharedvpchost` and `gcplogbucket`. In the last three, the resolved value their resources use is renamed `local.project`, as the others already call it.
  - `gcpplantonrunner`'s provider comment had said the runtime injects `GOOGLE_PROJECT`. It does not, and the block now carries the project like every other module.
- **The 20 kinds without a project of their own** keep empty blocks, each listed with its reason. They are the billing, identity, folder, organization and tag kinds, the grants that name their target by path, the KMS key, `gcpproject`, and the two-project kinds.
- **Pulumi:**
  - `pulumigoogleprovider.Get`, `GetWithUserProjectOverride` and `GetWithQuotaProject` take the component's project as an explicit argument, so the compiler makes every Google module decide;
  - a `projects/<id>` spelling is trimmed to the bare ID;
  - every Google Pulumi module passes its spec's project, or its scope's, or (Cloud Build repository) its connection's;
  - the 20 kinds above pass `""`, with the reason beside the call.
- **The guard:** `pkg/iac/iacinput/providerenvvars/gcpprovidertf_guard_test.go`, the twin of the AWS provider-block guard.
  - Every Google module carries the project lines in each provider block and passes a project to the Pulumi builder, unless its kind is on the list.
  - A listed kind that names a project fails too, so the list cannot go stale.
  - Credentials and the provider config may never be wired into the HCL.
- **Parity:** `project` moves from the GCP provider-config manifest's exclusions to `moduleOwned`, and `terraform-parity.md` is regenerated.

## How It Was Proven

**One live resource of each of eight kinds** in a real Google project (bucket, image repository, network, subnetwork, router NAT, service account, Cloud SQL, Cloud Run) was imported into throwaway local state and previewed. Import and plan only read the cloud.

| Bucket import | Provider | State's project | Preview |
|---|---|---|---|
| by name | before the change | `""` | 2 replacements |
| by name | after the change | the project | no replacement |

The other kinds' import IDs carry the project, so they recorded it either way, and the change leaves them alone.

**On Pulumi,** an import of the same bucket:
- **before the change:** planned `project: <nil> => <project>` and a replace;
- **after it:** a clean import.

**Checks:**
- the guard passes, and fails when one module's project is removed, in either engine;
- `tofu validate` passes on every changed module;
- every Google Pulumi module builds;
- `go test ./pkg/providerparity/...` and the builder's unit tests pass.
