# GCP GKE Fleet Feature

Turns on a fleet-wide capability for every cluster in a GKE fleet: Config Sync to keep clusters in sync with Git, Policy Controller to enforce Kubernetes policy, Cloud Service Mesh, one ingress across clusters, fleet logging, and upgrade sequencing that lets production wait for dev. Set it once as a fleet default and every cluster that joins later gets it too; override it for a named cluster when one needs something different.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- the Fleet API and the feature's own API
- **Feature** -- the fleet feature with its fleet-wide settings
- **Per-cluster settings** -- one entry per declared cluster override

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with fleet admin permissions on the fleet host project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP GKE Fleet Feature**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Config Sync from Git** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeFleetFeature
metadata:
  name: policy-controller
  org: acme-corp
  env: prod
spec:
  projectId:
    valueFrom:
      kind: GcpGkeFleet
      name: platform-fleet
      fieldPath: status.outputs.project_id
  feature: policycontroller
  fleetDefaultMemberConfig:
    policycontroller:
      policyControllerHubConfig:
        installSpec: INSTALL_SPEC_ENABLED
        policyContent:
          templateLibrary:
            installation: ALL
          bundles:
            - bundle: pss-baseline-v2022
```

```shell
planton apply -f gke-fleet-feature.yaml
```

This installs Policy Controller with the Pod Security Standards baseline on every cluster in the fleet. An Infra Job tracks the provisioning in real time.

### InfraChart

Reference the fleet's `status.outputs.project_id` from `projectId`; per-cluster overrides reference a cluster's `status.outputs.fleet_membership` or a membership's `status.outputs.name`.

## Key Configuration

These are the most important decisions when configuring a fleet feature. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Which feature** -- `feature` names it; only that feature's settings are accepted.

**Fleet default or per cluster** -- `fleetDefaultMemberConfig` reaches every cluster; `membershipConfigs` overrides named ones.

**Upgrade sequencing** -- `clusterupgrade` makes this fleet wait for upgrades to soak in an upstream fleet.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpGkeFleet** | `projectId`, `clusterupgrade.upstreamFleets` | `status.outputs.project_id` |
| **GcpGkeCluster** | `membershipConfigs[].membership`, `multiclusteringress.configMembership` | `status.outputs.fleet_membership` |
| **GcpGkeFleetMembership** | `membershipConfigs[].membership`, `multiclusteringress.configMembership` | `status.outputs.name` |
| **GcpServiceAccount** | Config Sync service-account emails | `status.outputs.email` |
| **GcpWorkloadIdentityPool** | `workloadidentity.scopeTenancyPool` | `status.outputs.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The feature's resource name | Audits, tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**GitOps everywhere** -- start from the **Config Sync from Git** preset.

**Policy baseline** -- start from the **Policy Controller Baseline** preset.

**Prod behind dev** -- start from the **Upgrade Behind the Dev Fleet** preset.

## Works With

- [**GCP GKE Fleet**](/infra-catalog/gcp-gke-fleet) -- the fleet the feature configures
- [**GCP GKE Fleet Scope**](/infra-catalog/gcp-gke-fleet-scope) -- team scopes that use allowlisted custom roles
- [**GCP GKE Cluster**](/infra-catalog/gcp-gke-cluster) -- clusters the features reach
