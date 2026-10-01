# GCP KMS Autokey Config

Turns on Cloud KMS Autokey for a folder or a project, so every team gets customer-managed encryption keys on demand without designing keys, key rings, or IAM grants. Developers ask for a key with a key handle when they create a bucket, disk, dataset, or topic; Autokey creates an HSM key that follows Google's recommended settings and wires up the access, while security keeps full control of the keys.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `cloudkms.googleapis.com` on the configured project and on a folder's key project
- **Autokey configuration** -- the folder's or project's `kms.AutokeyConfig` / `kms.ProjectAutokeyConfig`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Cloud KMS Autokey admin permissions at the folder or project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP KMS Autokey Config**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Project Same-Project Storage** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsAutokeyConfig
metadata:
  name: orders-autokey
  org: acme-corp
  env: prod
spec:
  scope:
    projectId:
      value: orders-prod
  keyProjectResolutionMode: RESOURCE_PROJECT
```

```shell
planton apply -f kms-autokey-config.yaml
```

This lets anyone who can create resources in the project request customer-managed keys for them. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a `GcpFolder` from `scope.folderId` and a `GcpProject` from `keyProject` to give a whole folder dedicated-project key storage; `GcpKmsKeyHandle` resources in the folder's projects then request their keys.

## Key Configuration

These are the most important decisions when configuring Autokey. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Where keys live** -- `RESOURCE_PROJECT` keeps each key beside the resource it protects (works on a folder or a project); `DEDICATED_KEY_PROJECT` puts every key for a folder in one key project, the separation-of-duties model.

**Folder or project** -- a folder configuration covers every project beneath it; a project configuration overrides its folder, including switching Autokey off with `DISABLED`.

**What destroy means** -- by default destroy clears the configuration and turns Autokey off; existing keys stay.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `scope.projectId` | `status.outputs.project_id` |
| **GcpFolder** | `scope.folderId` | `status.outputs.folder_id` |
| **GcpProject** | `keyProject` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The Autokey configuration's resource name | Audits and console links |
| `parent` | The folder or project Autokey is configured on | Reporting |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Project same-project storage** -- Autokey for one project, keys beside the resources. Start from the **Project Same-Project Storage** preset.

**Folder dedicated key project** -- one key project for a whole folder of workloads. Start from the **Folder Dedicated Key Project** preset.

## Works With

- [**GCP KMS Key Handle**](/cloud-catalog/gcp-kms-key-handle) -- requests a key from Autokey
- [**GCP Folder**](/cloud-catalog/gcp-folder) -- a configuration inherited by every project in the folder
- [**GCP Project**](/cloud-catalog/gcp-project) -- a project configuration or the key project
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- hand-designed keys
