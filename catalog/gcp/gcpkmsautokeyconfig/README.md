# GCP KMS Autokey Config

Switches Cloud KMS Autokey on for a folder or a single project, so teams get customer-managed encryption keys on demand instead of designing key rings and grants themselves. With Autokey on, a `GcpKmsKeyHandle` asks for a key for one resource type and location, and Autokey creates an HSM key in the `autokey` key ring, grants the resource's service agent encrypt and decrypt on it, and hands the key back.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `cloudkms.googleapis.com` on a project configuration's project and on a folder's key project (never disabled on destroy)
- **Autokey configuration** -- `kms_autokey_config` on a folder, or `kms_project_autokey_config` on a project, applied over whatever configuration the scope already had

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Cloud KMS Autokey admin permissions (`roles/cloudkms.autokeyAdmin`) at the scope: the folder or the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Dedicated Key Project (folder configurations only)

Before the first key handle is requested under dedicated-project storage, the key project's Cloud KMS service agent must exist and hold `roles/cloudkms.admin` on the key project -- the one-time setup Google documents. See the [GUIDE](GUIDE.md). Same-project storage needs none of this.

### Optional Dependencies

- **`GcpFolder`** / **`GcpProject`** -- the scope, by reference (`scope.folderId`, `scope.projectId`).
- **`GcpProject`** -- the dedicated key project (`keyProject`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsAutokeyConfig
metadata:
  name: project-autokey
spec:
  scope:
    projectId:
      value: my-gcp-project
  keyProjectResolutionMode: RESOURCE_PROJECT
```

```shell
planton apply -f kms-autokey-config.yaml
```

## Configuration Reference

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `scope` | `object` | provider project | At most one of `folderId` (`GcpFolder` ref) or `projectId` (`GcpProject` ref). |
| `keyProjectResolutionMode` | `string` | not sent | `RESOURCE_PROJECT` (same-project storage), `DEDICATED_KEY_PROJECT` (folders only), or `DISABLED` (switch Autokey off under an enabled folder). |
| `keyProject` | `string` / ref | none | The dedicated key project for a folder (`GcpProject` ref). |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` clears the configuration (Autokey off for the scope), `PREVENT` fails destroy, `ABANDON` keeps it in force. |

### Validation Rules

- `scope` names at most one of folder or project.
- `keyProject` and `DEDICATED_KEY_PROJECT` are folder-only; `DEDICATED_KEY_PROJECT` needs a `keyProject`.
- `keyProjectResolutionMode` and `deletionPolicy` take only Google's values.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `folders/{id}/autokeyConfig` or `projects/{id}/autokeyConfig` |
| `parent` | `string` | `folders/{id}` or `projects/{id}` |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **One configuration per scope.** Applying takes over whatever configuration the folder or project had; two blocks for the same scope overwrite each other.
- **Destroy turns Autokey off.** Under the default `deletionPolicy`, destroy clears the configuration. Keys Autokey already created stay and keep protecting their resources.
- **A project overrides its folder.** Use a project configuration with `DISABLED` to opt one project out of a folder's Autokey.
- **Keys are HSM keys.** Autokey keys use the Cloud HSM protection level, rotate yearly, and bill as HSM key versions.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpKmsKeyHandle** -- requests a key from Autokey for one resource
- **GcpFolder** -- a folder configuration every project beneath it inherits
- **GcpProject** -- a project configuration, or the dedicated key project
- **GcpKmsKey** -- keys you design yourself, when Autokey's defaults do not fit

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
