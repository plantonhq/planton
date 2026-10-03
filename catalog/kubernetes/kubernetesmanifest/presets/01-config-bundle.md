# Config Bundle

This preset applies a multi-document manifest — a ConfigMap and a Secret — anchored in a single namespace. Neither document declares its own `metadata.namespace`, so both land in the anchor namespace from `spec.namespace`, which the component creates first (`create_namespace: true`). It is the smallest useful demonstration of the component's namespace anchoring: write plain documents once, point the whole bundle at a namespace from the outside.

Before reaching for this preset, check the catalog: a single ConfigMap belongs in the first-class **KubernetesConfigMap** component, which validates configuration before deploy and exports composable outputs. KubernetesManifest is the escape hatch — this preset earns its place only when the documents must travel together as one raw bundle and no typed kind covers the set.

## When to Use

- A handful of plain configuration objects (ConfigMaps, Secrets, ServiceAccounts) that support some other workload and do not warrant separate resources
- Keeping a group of related documents in one manifest so they apply, update, and delete together
- As a template for any "several namespaced documents, one namespace" bundle — replace the documents with your own

## Key Configuration Choices

- **No `metadata.namespace` in the documents** — this is deliberate. Documents that declare no namespace land in the anchor namespace (`spec.namespace`); documents that declare one keep it. Leaving the documents unanchored makes the bundle reusable: retarget it by changing one spec field, not every document
- **`create_namespace: true`** — the anchor namespace is created (with standard Planton governance labels) before the documents apply, and deleted with the resource. Set it to `false` to apply into a namespace that already exists — the manifest documents themselves never receive injected labels either way
- **`stringData` on the Secret** — plain-text values that the API server encodes at rest; more readable in a manifest than pre-encoded `data`
- **Await is on by default** — `skip_await` is unset, so the deploy blocks until the applied resources pass readiness checks. For ConfigMaps and Secrets this is instantaneous

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-api-key>` | The secret value to store in the `app-credentials` Secret | Your credential store; never commit real values to version control |

The namespace (`app-config`), object names, and ConfigMap keys are example values — rename them to fit your application.

## Related Presets

- **02-crd-and-custom-resource** — a CRD and its custom resource in one manifest
- **03-vendor-install-manifest** — applying a vendor's published install YAML with `skip_await`
