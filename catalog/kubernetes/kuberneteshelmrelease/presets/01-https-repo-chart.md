# HTTPS Repo Chart

This preset installs [podinfo](https://github.com/stefanprodan/podinfo) 6.9.2 from its HTTPS Helm repository, with a `values_yaml` block overriding the chart's defaults. It is the baseline shape for the component: a public HTTP(S) chart repository, a pinned version, and a values file — the equivalent of `helm install -f values.yaml` expressed declaratively.

**Before reaching for this component at all:** if the catalog has a first-class kind for what you're deploying, use it instead. Typed kinds validate their configuration before deploy and export composable outputs; KubernetesHelmRelease is the intentional passthrough for charts no kind covers.

## When to Use

- Installing any chart published to a plain HTTPS Helm repository (the most common publishing form)
- Charts whose configuration is naturally a small values file — nested maps, lists, numbers, booleans — rather than a handful of one-line overrides
- As the starting point to adapt for any other chart: swap `repo`, `chart`, `version`, and the values

## Key Configuration Choices

- **`repo: https://stefanprodan.github.io/podinfo`** — an HTTP(S) repository URL; the chart name stays bare and the repository is consulted for the index. For OCI registries, see preset 02
- **`version: 6.9.2`** — the exact chart version. `version` is required by the spec: an unpinned "latest" install is not reproducible, and reproducibility is the point of declaring the release here
- **`values_yaml`** — the Helm values file, inline. Full YAML expressiveness; applied first, before any `set*` overrides. Do not put secrets here — that is what `set_sensitive` is for (preset 03)
- **`create_namespace: true`** — the module creates the `podinfo` namespace (with the standard governance labels) before installing and deletes it with the resource. Set to `false` when the namespace is managed elsewhere
- **Release name** — defaults to `metadata.name` (`podinfo` here); set `spec.release_name` to override what `helm list` shows

## Placeholders to Replace

This preset has no placeholders — it deploys as-is on any cluster and serves as a working smoke test for the component. Adapt it to your chart by replacing `repo`, `chart`, `version`, and the `values_yaml` content.

## Related Presets

- **02-oci-registry-chart** — the same chart pulled from an OCI registry instead of an HTTPS repository
- **03-private-repo-with-secrets** — private repository credentials, secret chart values, and production lifecycle knobs
