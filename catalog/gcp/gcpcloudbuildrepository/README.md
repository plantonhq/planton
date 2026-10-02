# GCP Cloud Build Repository

Declares one repository linked into Cloud Build through an existing connection. Triggers (`repositoryEventConfig`, `sourceToBuild`, `gitFileSource`) and Cloud Deploy custom target types reference the link's `name` output. The link lives in its connection's project and region; destroy removes the link, never the repository on the code host.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Repository link** -- one `cloudbuildv2_repository` under the parent connection

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/cloudbuild.connectionAdmin` (or the permissions in `iac/permissions.yaml`) on the connection's project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpCloudBuildConnection`** -- with its installation complete (`installation_stage` `COMPLETE`); reference its `name` from `parentConnection`.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudBuildRepository
metadata:
  name: orders
spec:
  parentConnection:
    valueFrom:
      kind: GcpCloudBuildConnection
      name: acme-github
  remoteUri: https://github.com/acme/orders.git
```

```shell
planton apply -f cloud-build-repository.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `parentConnection` | `string` / ref | The connection's full name (`GcpCloudBuildConnection` ref or `projects/{p}/locations/{l}/connections/{c}`). Immutable. |
| `remoteUri` | `string` | The repository's `https://` clone URI. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `repositoryId` | `string` | `metadata.name` | The repository's ID in Cloud Build (letters, digits, `-._~%!$&'()*+,;=@`). Immutable. |
| `annotations` | `map` | -- | AIP-128 annotations. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- A literal `parentConnection` is a full connection name; `remoteUri` starts with `https://`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/connections/{connection}/repositories/{repository_id}` |
| `repository_id` | `string` | The repository's ID in Cloud Build |
| `remote_uri` | `string` | The repository's clone URI |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Every field is immutable.** Changing any of them replaces the link; the repository on the host is untouched.
- **Project and region come from the connection.** Both modules parse them from `parentConnection`, which is why a short connection name is refused.
- **The connection must be complete.** Google links a repository only through a connection whose installation is finished.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpCloudBuildConnection** -- the connection the repository is linked through
- **GcpCloudBuildTrigger** -- builds started by the repository's events
- **GcpDeployCustomTargetType** -- Skaffold modules read from the repository

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
