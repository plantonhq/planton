# CRD and Custom Resource

This preset applies a CustomResourceDefinition and a custom resource of that new type in one manifest — the ordering problem that breaks a naive `kubectl apply -f` (the custom resource is rejected because its type does not exist yet, forcing a second apply). Both engines handle it in a single pass: the Pulumi engine's `yaml/v2` ConfigGroup orders the CRD install before the custom resources that use it, and the Terraform engine's `kubectl_manifest` applies server-side without needing the type registered at plan time.

Before reaching for this preset, check the catalog: if a first-class kind covers what you need, use it — typed components validate configuration before deploy and export composable outputs. KubernetesManifest is the escape hatch for resources no kind covers, and a CRD bundle is one of its canonical uses.

## When to Use

- Installing an operator's or vendor's CRDs together with the initial custom resources that configure it
- Deploying an exotic custom resource for which no first-class catalog kind exists
- Any manifest where a document's type is defined by another document in the same manifest

## Key Configuration Choices

- **CRD first, custom resource second** — the manifest lists the CRD before the resource that uses it. Keep this order: it reads correctly, and it is the order the applied-resource inventory reports
- **The CRD is cluster-scoped** — it is applied exactly as written; the anchor namespace never touches cluster-scoped documents
- **The custom resource declares no `metadata.namespace`** — its CRD sets `scope: Namespaced`, so it lands in the anchor namespace (`crontab-demo`). Documents with an explicit namespace would keep it
- **`create_namespace: true`** — the anchor namespace is created before the manifest applies, so the custom resource has somewhere to land
- **A deliberately minimal example CRD** — the `CronTab` type (from the Kubernetes documentation's canonical example) stands in for whatever real CRD you are installing. Replace both documents wholesale

## Placeholders to Replace

This preset deploys as-is on any cluster — the example CRD and custom resource are self-contained and conflict with nothing. In real use, replace the entire `manifest_yaml` with your CRD bundle and rename the anchor namespace. Note that nothing reconciles the example `CronTab`: applying a CRD defines a type, it does not install the controller that acts on it.

## Related Presets

- **01-config-bundle** — plain namespaced documents anchored in one namespace
- **03-vendor-install-manifest** — a vendor's full install YAML, where CRDs, RBAC, and workloads arrive in one file
