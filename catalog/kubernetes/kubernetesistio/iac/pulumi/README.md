# KubernetesIstio Pulumi Module

Installs the Istio control plane from the official Helm charts (`base`,
`istiod`, and in ambient mode `cni` + `ztunnel`, all at one pinned version
from `https://istio-release.storage.googleapis.com/charts`) as real Helm
releases, in upstream's own order.

## Resources Created

1. **Namespace** (optional) — created and owned when `create_namespace` is
   set, with the standard resource-identity labels
2. **Istio CRDs** (`yaml.ConfigFile`, server-side apply) — the upstream
   `crd-all.gen.yaml` bundle at the pinned version, applied by the module
   itself
3. **Helm release `istio-base`** — validation-webhook plumbing, with
   `base.excludedCRDs` covering the entire bundle so the release templates NO
   CRDs
4. **Helm release `istiod`** (or `istiod-<revision>` for a named revision) —
   the control plane; the module waits for real readiness (a control plane
   whose webhooks are not serving rejects every mesh-config apply)
5. **Helm release `istio-cni`** — the node agent, in ambient mode or when
   `cni.enabled` in sidecar mode
6. **Helm release `ztunnel`** — the ambient per-node L4 proxy

All releases are atomic with cleanup-on-fail: a failed install never leaves a
half-deployed mesh.

## Why the CRDs Are Module-Owned

Helm refuses to adopt CRDs that already exist without its ownership metadata.
If the `base` release owned the CRDs, a cluster running the CRDs-only
KubernetesIstioBaseCrds kind could never upgrade to the full mesh.
Server-side-applied CRDs are co-ownable by both kinds, making that migration
a plain redeploy. Destroying the module removes the CRDs with everything
else — mesh configuration objects cascade.

## Values Rendering (per-release documents)

The typed spec renders into per-chart values in `module/values.go`
(`buildBaseValues`, `buildIstiodValues`, `buildCniValues`,
`buildZtunnelValues`). Each release's `helm_values` escape hatch merges LAST
over the typed values with Helm `-f` semantics (nested maps deep-merge with
the override winning; scalars and lists replace).

**Parity:** the Terraform module reaches byte-identical values natively (each
`helm_release` passes the typed document plus the escape hatch as a second
values document) — chart identity (repo/name/version) and every typed mapping
are kept in lockstep between the two engines, so both deploy the same product
from one manifest.

## Usage

```bash
planton pulumi up --manifest istio.yaml --module-dir <path-to-this-module>
```

## Module Structure

- `main.go` — entrypoint that calls the module
- `module/main.go` — resource orchestration (CRDs, then the releases in order)
- `module/values.go` — typed-spec-to-chart-values rendering and the escape-hatch merge
- `module/locals.go` — resolved names (revision, release/Service names, trust domain)
- `module/namespace.go` — optional namespace creation
- `module/vars.go` — chart identity, pinned default version, CRD bundle URL and exclusion list
- `module/outputs.go` — output constants

## Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Control-plane namespace |
| `istiod_service_name` | istiod Service name (`istiod`, or `istiod-<revision>`) |
| `revision` | Installed control-plane revision (`default` when unnamed) |
| `gateway_class_name` | GatewayClass istiod serves (`istio`) |
| `trust_domain` | The mesh's trust domain |
| `dataplane_mode` | `sidecar` or `ambient` |
