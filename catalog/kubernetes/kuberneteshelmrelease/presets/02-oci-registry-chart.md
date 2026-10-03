# OCI Registry Chart

This preset installs [podinfo](https://github.com/stefanprodan/podinfo) 6.9.2 from an OCI registry (`oci://ghcr.io/stefanprodan/charts`) — the same chart as preset 01, pulled the other way charts are published. More and more projects ship charts only to OCI registries (GHCR, ECR, ACR, Artifact Registry); the manifest shape is identical to the HTTPS form except for the `repo` scheme.

**Before reaching for this component at all:** if the catalog has a first-class component for what you're deploying, use it instead. KubernetesHelmRelease is the intentional passthrough for charts no kind covers.

## When to Use

- Charts published to an OCI registry rather than a classic HTTPS Helm repository — the `repo` field accepts both; only the scheme differs
- Registries like `oci://ghcr.io/...`, `oci://public.ecr.aws/...`, or a private OCI registry (add credentials as in preset 03)

## Key Configuration Choices

- **`repo: oci://ghcr.io/stefanprodan/charts`** — an OCI registry reference. The chart is pulled as `<repo>/<chart>:<version>` (here `oci://ghcr.io/stefanprodan/charts/podinfo:6.9.2`); both engines perform that join identically, so the manifest keeps `repo` and `chart` separate just like the HTTPS form
- **`set` vs `set_string`** — this preset demonstrates the two override layers and why both exist:
  - `set` uses Helm's `--set` coercion: `replicaCount: "2"` arrives at the chart as the **number** 2 (also: `"true"`/`"false"` become booleans, `"null"` deletes a key)
  - `set_string` keeps values as **literal strings**: `image.tag: "6.9.2"` stays the string `"6.9.2"`. This matters for version-like tags — under `set`, a tag like `"1.30"` would be coerced to the number 1.3 and break the image reference
- **`version: 6.9.2`** — required and pinned; for OCI charts this becomes the reference tag

## Placeholders to Replace

This preset has no placeholders — it deploys as-is. For a private OCI registry, add `repository_username` and `repository_password` (see preset 03); the credentials are used for the registry login.

## Related Presets

- **01-https-repo-chart** — the same chart from its HTTPS repository, configured with a `values_yaml` block
- **03-private-repo-with-secrets** — private registry/repository credentials and secret chart values
