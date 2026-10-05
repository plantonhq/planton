# GCP GKE Fleet Membership

Brings an existing GKE cluster into a fleet so team scopes and fleet features can use it -- the explicit registration for clusters that did not join the fleet when they were created.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- the Fleet API on the fleet host project
- **Membership** -- the cluster's registration with the fleet

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with fleet admin permissions, and read access to the cluster, on the fleet host project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP GKE Fleet Membership**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Register a GKE Cluster** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeFleetMembership
metadata:
  name: legacy-cluster
  org: acme-corp
  env: prod
spec:
  projectId:
    valueFrom:
      kind: GcpGkeFleet
      name: platform-fleet
      fieldPath: status.outputs.project_id
  gkeCluster:
    value: projects/legacy-prod/locations/us-central1/clusters/legacy-cluster
```

```shell
planton apply -f gke-fleet-membership.yaml
```

This registers the legacy cluster with the platform fleet. An Infra Job tracks the provisioning in real time.

### InfraChart

Reference the fleet's `status.outputs.project_id` from `projectId` and the cluster's `status.outputs.cluster_id` from `gkeCluster`; scopes and features then reference this membership's `status.outputs.name`.

## Key Configuration

These are the most important decisions when configuring a membership. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Which path** -- use this block only for clusters that did not set `fleetProject` at creation.

**Fleet Workload Identity** -- set `issuer` so the cluster's workloads use fleet identities.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpGkeFleet** | `projectId` | `status.outputs.project_id` |
| **GcpGkeCluster** | `gkeCluster` | `status.outputs.cluster_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The membership's full name | `GcpGkeFleetScope.membershipBindings[].membership`, `GcpGkeFleetFeature.membershipConfigs[].membership` |
| `membership_id` | The membership's ID | Tooling |
| `location` | The membership's location | Tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Register a cluster** -- start from the **Register a GKE Cluster** preset.

**Cross-project with Workload Identity** -- start from the **With Fleet Workload Identity** preset.

## Works With

- [**GCP GKE Fleet**](/infra-catalog/gcp-gke-fleet) -- the fleet the cluster joins
- [**GCP GKE Cluster**](/infra-catalog/gcp-gke-cluster) -- the cluster to register
- [**GCP GKE Fleet Scope**](/infra-catalog/gcp-gke-fleet-scope) -- give the cluster to a team
