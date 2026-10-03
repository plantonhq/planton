# GCP Cloud Build Repository

Links one repository from your code host into Cloud Build through an existing connection, so triggers can build on its pushes and pull requests and Cloud Deploy can read Skaffold configuration from it. The repository itself stays on GitHub, GitLab, or Bitbucket; this is the link.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Repository link** -- the repository, linked under its connection in the connection's project and region

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with permission to manage Cloud Build repositories in the connection's project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Code Host Account

- **A finished connection** -- a `GcpCloudBuildConnection` whose installation is complete, with access to the repository.

## Deploy

### Console

Open the deployment store, find **GCP Cloud Build Repository**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Linked Repository** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudBuildRepository
metadata:
  name: orders
  org: acme-corp
  env: prod
spec:
  parentConnection:
    value: projects/acme-ci/locations/us-central1/connections/acme-github
  remoteUri: https://github.com/acme/orders.git
```

```shell
planton apply -f cloud-build-repository.yaml
```

This links the orders repository so a trigger can build it. An Infra Job tracks the provisioning in real time.

### InfraChart

Reference the connection's `status.outputs.name` from `parentConnection`.

## Key Configuration

These are the most important decisions when configuring a repository link. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Connection** -- the connection's full name; the link lives in its project and region.

**Remote URI** -- the repository's HTTPS clone URI.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpCloudBuildConnection** | `parentConnection` | `status.outputs.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The repository link's full resource name | A trigger's repository fields, a custom target type's Cloud Build repository |
| `repository_id` | The repository's ID in Cloud Build | Tooling |
| `remote_uri` | The repository's clone URI | Tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Linked repository** -- one repository through a GitHub connection. Start from the **Linked Repository** preset.

## Works With

- [**GCP Cloud Build Connection**](/infra-catalog/gcp-cloud-build-connection) -- the connection the repository is linked through
- [**GCP Cloud Build Trigger**](/infra-catalog/gcp-cloud-build-trigger) -- builds started by the repository's events
- [**GCP Deploy Custom Target Type**](/infra-catalog/gcp-deploy-custom-target-type) -- Skaffold modules read from the repository
