# Kubernetes Manifest

## When NOT to Use This

Before anything else: **a first-class catalog kind always wins.** Typed kinds validate configuration before deploy, export composable outputs other resources can reference, and document their trade-offs field by field — raw YAML does none of that. If the catalog has a kind for what you're deploying (a Deployment, a Helm chart, a StorageClass, ...), use it.

**KubernetesManifest** is the catalog's bring-your-own-manifest escape hatch, for the YAML no kind covers:

- A vendor's install manifest, published as one raw file for `kubectl apply -f`
- A CRD bundle — CRDs plus the custom resources that configure an operator
- An exotic custom resource with no typed kind

## Overview

Hand the component any valid Kubernetes manifest — a single document or many separated by `---`, core kinds or custom resources, even a CRD and its custom resources together — and both IaC engines apply it to the cluster **exactly as written**: no injected labels, no rewritten fields, no interpretation. Around that, the component provides what `kubectl apply` cannot: full lifecycle (create, update, destroy with nothing orphaned), one-pass CRD ordering, readiness awaits, and outputs.

## Namespace Semantics

The one piece of defaulting the kind performs, identical on both engines:

- Documents that declare their own `metadata.namespace` **keep it**
- Namespaced documents that declare none **land in `spec.namespace`** (the anchor)
- Cluster-scoped documents (CRDs, ClusterRoles, ...) are **applied as-is** — the anchor never distorts them

`spec.namespace` accepts a literal namespace name or a reference to a KubernetesNamespace resource. With `create_namespace: true`, the anchor namespace is created (with the standard Planton governance labels — on the namespace object only, never inside your manifest) before the manifest applies, and deleted with the resource. With `false`, the namespace must already exist.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: The anchor namespace (value or ref). Where namespaced documents without an explicit `metadata.namespace` are applied
- **`spec.manifest_yaml`**: The raw Kubernetes YAML to apply — single or multi-document, must contain at least one non-blank document. Content is never validated beyond that: the manifest's kinds are the API server's business, by design

### Optional

- **`spec.create_namespace`** (default `false`): Whether this resource creates and owns the anchor namespace
- **`spec.skip_await`** (default `false`): When `true`, the deploy returns as soon as the API server accepts every document. When `false`, both engines block until readiness: Deployments/DaemonSets/StatefulSets complete their rollout, and other kinds pass their engine's readiness checks. Skip the await for manifests whose readiness depends on something deployed later (e.g. a webhook configuration waiting on its service) or that intentionally stay not-ready at install time — vendor install bundles are the usual case

## Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`namespace`**: The anchor namespace — where namespaced documents without an explicit `metadata.namespace` were applied
- **`applied_resources`**: One entry per applied document, in manifest order, formatted as `<apiVersion>/<Kind>/<name>` (e.g. `apps/v1/Deployment/app`). Derived by parsing the input YAML — identically on both engines — so downstream tooling sees what was applied without re-parsing the manifest

## Quick Start

Create a file `manifest.yaml` — a ConfigMap and Secret anchored in one namespace:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesManifest
metadata:
  name: app-config
spec:
  namespace:
    value: my-app
  create_namespace: true
  manifest_yaml: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: app-settings
    data:
      LOG_LEVEL: info
    ---
    apiVersion: v1
    kind: Secret
    metadata:
      name: app-credentials
    type: Opaque
    stringData:
      api-key: replace-me
```

Deploy:

```shell
planton apply -f manifest.yaml
```

Neither document declares a namespace, so both land in `my-app`, which the component creates first.

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that uphold one parity contract:

1. **Anchoring**: Pulumi constructs its Kubernetes provider with `spec.namespace` as the default namespace (scope-aware — cluster-scoped kinds are skipped); Terraform sets `override_namespace` per document, only on documents that declare no namespace. Same outcome
2. **Apply**: both engines apply server-side. Pulumi feeds the whole manifest to a `yaml/v2` ConfigGroup; Terraform creates one `kubectl_manifest` per document, keyed by the document's identity (`apiVersion/Kind/namespace/name`) so reordering documents never churns state
3. **CRD ordering**: a CRD and its custom resources apply in one pass on both engines — the ConfigGroup orders CRDs first, and `kubectl_manifest` needs no cluster connection at plan time
4. **Await**: `skip_await` maps to `SkipAwait` on Pulumi and inverted `wait`/`wait_for_rollout` on Terraform; both default to awaiting readiness. (One benign breadth difference: when awaiting, Pulumi also readiness-checks non-workload kinds like Services; kubectl awaits workload rollouts only. The applied objects are identical)
5. **Outputs**: both engines export the anchor namespace and the applied-resource inventory, derived by parsing the input YAML with an identical document-split rule

## When to Use

Use **KubernetesManifest** when:

- The catalog has no kind for the resource you need to apply
- A vendor publishes raw install YAML and you want it lifecycle-managed, byte-for-byte as published
- You are installing CRDs together with the custom resources that use them
- A small bundle of plain objects (ConfigMaps, RBAC, quotas) should apply, update, and delete together

**Do NOT use** when:

- A first-class kind covers the resource — it always wins on validation, outputs, and documentation
- The vendor ships a Helm chart — use KubernetesHelmRelease and keep the chart's upgrade path
- You need templating or per-environment value substitution — this component applies YAML exactly as written; parameterize upstream of it

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **The anchor namespace**: must already exist unless `create_namespace: true`

## Best Practices

1. **Exhaust the catalog first** — treat every KubernetesManifest resource as a pointer at a missing typed kind
2. **Leave your own documents unanchored**: omit `metadata.namespace` and anchor through `spec.namespace`, so retargeting the bundle is a one-field change
3. **Paste vendor manifests verbatim**: the pattern's value is byte-for-byte fidelity; upgrades are a paste of the next release's file
4. **Skip the await on install bundles**: manifests with internal ordering (webhooks waiting on services) can deadlock a readiness await; `skip_await: true` is the safe setting for vendor installs
5. **One concern per resource**: don't couple a vendor install and your own configuration in one manifest — split them so each changes independently

## References

- [Kubernetes Objects and Manifests](https://kubernetes.io/docs/concepts/overview/working-with-objects/)
- [Namespaces and Object Scope](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)
- [Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
- [Server-Side Apply](https://kubernetes.io/docs/reference/using-api/server-side-apply/)
- [Pulumi Kubernetes yaml/v2 ConfigGroup](https://www.pulumi.com/registry/packages/kubernetes/api-docs/yaml/v2/configgroup/)
- [alekc/kubectl Terraform Provider](https://registry.terraform.io/providers/alekc/kubectl/latest/docs)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
