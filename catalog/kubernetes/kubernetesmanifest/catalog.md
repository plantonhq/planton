# Kubernetes Manifest

Deploys raw Kubernetes YAML manifests to any Kubernetes cluster, acting as a generic escape hatch for resources that do not have a dedicated component. Supports single-document and multi-document manifests (separated by `---`), including Deployments, Services, ConfigMaps, CRDs, Custom Resources, and any other valid Kubernetes resource types. The manifest content is applied exactly as written -- no injected labels, no rewritten fields -- with automatic CRD-before-custom-resource ordering inside one manifest.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created (with the standard governance labels) only when `createNamespace` is `true`, and deleted with the resource; otherwise deploys into an existing namespace
- **Kubernetes Resources from Manifest** -- all resources defined in the `manifestYaml` field are applied to the cluster exactly as written. Multi-document YAML (resources separated by `---`) is supported with automatic CRD ordering: a CRD and its custom resources can ship in the same manifest. The manifest content itself is never mutated -- no injected labels, no rewritten fields.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Appropriate RBAC permissions** for the resource types in your manifest -- see `iac/permissions.yaml` for the full permission statement.
- **CRDs pre-installed** if your manifest references custom resources. The manifest component does not handle CRD installation ordering across separate Infra Components.

## Deploy

### Console

Open the deployment store, find **Kubernetes Manifest**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Config Bundle** preset in the [Presets](#presets) tab, then replace the YAML with your actual manifest.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesManifest
metadata:
  name: app-config
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "my-app"
  createNamespace: false
  manifestYaml: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: app-settings
    data:
      config.yaml: |
        database_pool_size: 10
        log_level: info
```

```shell
planton apply -f manifest.yaml
```

This deploys a ConfigMap into the `my-app` namespace. The `manifestYaml` field accepts any valid Kubernetes YAML, including multi-document manifests with `---` separators. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the manifest deployment to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: app-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then applies the manifest resources into it.

## Key Configuration

These are the most important decisions when configuring a Kubernetes Manifest deployment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Manifest YAML content** -- The `manifestYaml` field is the core of this component. Paste any valid Kubernetes YAML, including multi-document manifests separated by `---`. The IaC module handles resource ordering automatically, applying CRDs before custom resources that depend on them.

**Namespace behavior** -- The `namespace` field is an ANCHOR, not a rewrite: namespaced documents that do not declare their own `metadata.namespace` are applied there, documents with an explicit namespace keep it, and cluster-scoped documents (CRDs, ClusterRoles) are never distorted by it. Set `createNamespace: true` to create the anchor namespace if it does not exist.

**Readiness semantics** -- By default the deploy blocks until every applied resource becomes ready: Deployments, DaemonSets, and StatefulSets complete their rollout, and other kinds pass readiness checks. Set `skipAwait: true` to return as soon as the API server accepts every document -- for manifests whose readiness depends on something deployed later (a webhook configuration waiting on its Service) or that intentionally stay not-ready at install time.

**Use case boundaries** -- A first-class catalog kind always wins: typed components validate configuration before deploy, export composable outputs, and document their trade-offs field by field -- raw YAML does none of that. Reach for a manifest only when the catalog has no component for what you need to apply (a vendor's install manifest, a CRD bundle, an exotic custom resource).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | The anchor namespace: where namespaced documents without an explicit `metadata.namespace` were applied | Application deployment manifests and service discovery |
| `applied_resources` | One entry per applied document, in manifest order, formatted as `<apiVersion>/<Kind>/<name>` (e.g. `apps/v1/Deployment/app`) | Auditing what a manifest shipped without re-parsing its YAML |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Config bundle** -- A multi-document manifest (a ConfigMap and a Secret) anchored in one namespace: write plain documents once, point the whole bundle at a namespace from the outside. Check the catalog first -- a single ConfigMap belongs in the typed KubernetesConfigMap component. Start from the **Config Bundle** preset.

**CRD and custom resource in one pass** -- A CustomResourceDefinition and a custom resource of that new type ship in the same manifest; the module orders the CRD install first, avoiding the two-apply dance that breaks a naive `kubectl apply -f`. Start from the **CRD and Custom Resource** preset.

**Vendor install manifest** -- Paste the YAML a project publishes for `kubectl apply -f` -- often hundreds of documents spanning CRDs, RBAC, Services, and webhooks -- and get declarative apply, update, and destroy wrapped around it. Charts belong in KubernetesHelmRelease instead. Start from the **Vendor Install Manifest** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the default namespace for manifest resources
- [**Helm Release**](/infra-catalog/kubernetes-helm-release) -- the right escape hatch when the vendor publishes a chart rather than raw YAML
- [**Kubernetes ConfigMap**](/infra-catalog/kubernetes-config-map) -- the typed component a single configuration document belongs in
