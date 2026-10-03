# Istio Base CRDs

Installs the Istio base Custom Resource Definitions (the `istio/base` CRD bundle) on any Kubernetes cluster. These are the API types only -- no control plane, gateways, or sidecars. Installing them is the lightweight prerequisite that lets a cluster accept and server-side-validate Istio's networking, security, and telemetry resources: DestinationRule, ServiceEntry, PeerAuthentication, RequestAuthentication, AuthorizationPolicy, Telemetry, and EnvoyFilter. Install this once per cluster before creating any of those.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Istio CRDs** -- cluster-scoped Custom Resource Definitions applied from the official Istio release manifest. They register the Istio API types so the cluster can accept Istio resources; no pods, services, or workloads are created.
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Cluster admin access** required. CRDs are cluster-scoped resources that require elevated permissions to install.
- **No conflicting Istio installation** -- if a different Istio version's CRDs are already present, reconcile that first. The installed CRD set is pinned to a specific Istio release for schema compatibility (see below).

## Deploy

### Console

Open the deployment store, find **Istio Base CRDs**, and click **Deploy**. The creation wizard walks you through environment and connection selection and a short confirmation of what gets installed. There are no settings to configure -- the CRD set and its version are fixed.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesIstioBaseCrds
metadata:
  name: istio-base-crds
  org: acme-corp
  env: prod
spec: {}
```

```shell
planton apply -f istio-base-crds.yaml
```

This installs the Istio base CRDs cluster-wide. No namespace is required since CRDs are cluster-scoped, and there are no spec fields to set. An Infra Job tracks the provisioning in real time.

## Key Configuration

There is nothing to configure -- the spec is deliberately empty. The Istio CRD set is pinned to the Istio release that the typed Istio resources were built against, so the API schema on the cluster always matches the resources you create. A user-selectable version would risk a mismatch that silently drops or rejects fields, so it is intentionally fixed. The one real decision is WHETHER to use this kind at all: a cluster that will run the full mesh should deploy **Istio** directly (its module co-owns the same CRDs, so upgrading from CRDs-only to the full mesh later is a plain redeploy).

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components and operators can reference:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `installed_release` | The Istio release the CRDs were installed from (the pinned SDK tag, e.g. `1.30.3`) | Verifying the CRD schema version present on the cluster |
| `installed_manifest_url` | The exact CRD bundle URL that was applied | Auditing precisely what landed on the cluster |

## Common Patterns

Browse the [Presets](#presets) tab for a ready-to-deploy configuration.

**Standard installation** -- Installs the pinned Istio base CRDs with no additional configuration. Suitable for any cluster that will run typed Istio traffic, security, or telemetry resources. Start from the **Standard Istio CRD Installation** preset.

## Works With

- [**Istio Destination Rule**](/infra-catalog/kubernetes-destination-rule) and [**Istio Service Entry**](/infra-catalog/kubernetes-service-entry) -- the traffic-management resources these CRDs unlock.
- [**Istio Peer Authentication**](/infra-catalog/kubernetes-peer-authentication), [**Istio Request Authentication**](/infra-catalog/kubernetes-request-authentication), and [**Istio Authorization Policy**](/infra-catalog/kubernetes-authorization-policy) -- the security policy resources.
- [**Istio Telemetry**](/infra-catalog/kubernetes-telemetry) and [**Istio Envoy Filter**](/infra-catalog/kubernetes-envoy-filter) -- observability and extensibility.
- [**Istio**](/infra-catalog/kubernetes-istio) -- the full mesh (control plane and data plane); its module co-owns these CRDs, so a CRDs-only cluster upgrades with a plain redeploy.
