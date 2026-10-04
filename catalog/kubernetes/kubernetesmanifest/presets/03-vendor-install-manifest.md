# Vendor Install Manifest

This preset is the "paste the vendor's install YAML" pattern: take the manifest a project publishes for `kubectl apply -f` — often hundreds of documents spanning CRDs, RBAC, Services, Deployments, and webhook configurations — and apply it exactly as downloaded, with Planton lifecycle management (declarative apply, update, destroy) wrapped around it. The manifest content is never mutated: no injected labels, no rewritten fields.

Before reaching for this preset, check the catalog: many popular installs have a first-class kind or ship as a Helm chart (use KubernetesHelmRelease for those). KubernetesManifest is the escape hatch for vendors that publish only raw YAML.

## When to Use

- A vendor or open-source project publishes its install as a single raw YAML file and offers no Helm chart worth using
- You want the install pinned in version control and managed declaratively instead of `kubectl apply`d from a URL
- The install mixes cluster-scoped and namespaced documents, CRDs and custom resources — all fine in one manifest

## Key Configuration Choices

- **`skip_await: true`** — the deploy returns as soon as the API server accepts every document, without waiting for readiness. Vendor installs frequently contain resources that are not ready until something else in the same bundle (or a later step) exists — the classic case is a webhook configuration waiting on its service. Awaiting readiness on such a bundle can deadlock the deploy; skipping it is the safe default for install manifests. Drop the field (or set `false`) if you want the deploy to block until rollouts complete
- **Namespace anchoring on a vendor bundle** — vendor manifests usually hard-code their namespaces; those documents keep them, and cluster-scoped documents (CRDs, ClusterRoles, ...) are unaffected. Only namespaced documents that declare no namespace land in the anchor (`vendor-system`). If the vendor's manifest includes its own `Namespace` document, set `create_namespace: false` and match `spec.namespace` to the vendor's namespace
- **Paste, don't edit** — the value of this pattern is that the applied objects are byte-for-byte what the vendor published, so upgrading means pasting the next version's file

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| the entire `manifest_yaml` body | The placeholder ServiceAccount stands in for the vendor's install manifest — replace the whole block with the downloaded file's contents | The vendor's release page or documentation (the URL they give for `kubectl apply -f`) |
| `vendor-system` | The anchor namespace for any namespaced documents the vendor left unanchored | The vendor's install docs — many bundles name their own namespace; match it |

## Related Presets

- **02-crd-and-custom-resource** — the CRD-ordering capability this pattern relies on, in miniature
- **01-config-bundle** — small first-party bundles anchored in one namespace
