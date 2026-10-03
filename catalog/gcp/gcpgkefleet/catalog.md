# GCP GKE Fleet

Runs many GKE clusters as one. A fleet is the home of your clusters' shared configuration: the security posture and Binary Authorization defaults every cluster inherits, the team scopes that give teams namespaces and access across clusters, and the fleet features that turn on Config Sync, Policy Controller, Service Mesh, and multi-cluster ingress everywhere at once.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- the Fleet API on the host project
- **Fleet** -- the project's one fleet, with its cluster defaults

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with fleet admin permissions on the host project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP GKE Fleet**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Platform Fleet** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeFleet
metadata:
  name: platform-fleet
  org: acme-corp
  env: prod
spec:
  projectId:
    value: platform-host
  displayName: Platform fleet
  defaultClusterConfig:
    securityPostureConfig:
      mode: BASIC
      vulnerabilityMode: VULNERABILITY_BASIC
```

```shell
planton apply -f gke-fleet.yaml
```

This creates the fleet before any cluster joins it, ready for team scopes and features. An Infra Job tracks the provisioning in real time.

### InfraChart

Declare the fleet first and reference its `status.outputs.project_id` from every scope, feature, and membership's `projectId`; the chart then creates them inside this fleet, after it.

## Key Configuration

These are the most important decisions when configuring a fleet. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Host project** -- the project that holds the fleet; clusters in other projects can still join it.

**Security posture defaults** -- `BASIC` configuration auditing and vulnerability scanning on every cluster, included with GKE.

**Binary Authorization defaults** -- audit every cluster's workloads against GKE platform policies.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `project_id` | The fleet host project | `projectId` of every fleet scope, feature, and membership |
| `name` | The fleet's resource name | Audits |
| `uid` | The fleet's unique identifier | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Platform fleet** -- a fleet with Google's standard posture checks. Start from the **Platform Fleet** preset.

**Audited fleet** -- every workload audited against a platform policy. Start from the **Binary Authorization Audit** preset.

## Works With

- [**GCP GKE Fleet Scope**](/infra-catalog/gcp-gke-fleet-scope) -- team scopes inside the fleet
- [**GCP GKE Fleet Feature**](/infra-catalog/gcp-gke-fleet-feature) -- fleet-wide features
- [**GCP GKE Fleet Membership**](/infra-catalog/gcp-gke-fleet-membership) -- explicit cluster registration
- [**GCP GKE Cluster**](/infra-catalog/gcp-gke-cluster) -- clusters that join through `fleetProject`
