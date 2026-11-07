# KubernetesAgentSubstrate Pulumi Module

Installs the Agent Substrate control plane of the release `spec.version` names (v1.0.0 when unset) straight from what upstream publishes: the tag's own files and definitions, and every image pinned by the digest upstream's registry serves (`module/release.go`), composed the way upstream's installer composes them, with the typed spec applied onto upstream's own objects. A new upstream release needs nothing from Planton; `release.go` says what does. The exact semantic twin of the Terraform module (`../tf/`).

## Layout

| File | Role |
|---|---|
| `module/vars.go` | The default release, upstream's fixed names, the files the composition reads |
| `module/release.go` | What upstream publishes and how the module reads it: the tag's files, the image placeholders, the digest lookup |
| `module/settings.go` | The spec resolved to settings with upstream's defaults filled in |
| `module/documents.go` | Download, split on separator lines, key by composed identity |
| `module/main.go` (`fetchReleaseFrom`) | Every file the composition reads, at the release's tag |
| `module/patch.go` | Upstream's kustomize components and overlays, applied as written |
| `module/compose.go` | The document set, in four phases, with the passes every document gets |
| `module/workloads.go` | The typed settings on each workload |
| `module/splice.go` | Upstream's text markers, the agentgateway and Cloud SQL components, the image mirror |
| `module/config.go` | The spec-derived ConfigMaps and Secret, and the rollout digests |
| `module/pools.go`, `module/secrets.go` | The generated certificate authorities, signing key and its public key set, as typed Secrets |
| `module/upgrade.go`, `module/preflight.go` | Upstream's upgrade policy and every refusal |
| `module/main.go` | Refusals, the kept definitions, the Secrets, then the phases |

## Order and Refusals

Everything that can refuse runs before anything registers: a namespace other than upstream's; a workload or namespace another installation holds (`workloadowner`); a minor-release move upstream cannot survive, read from the stamps on the API server and the bundled database's volume; a file or image the release does not publish, or an image placeholder the module does not know; and, in `keptcrds`, a definition someone else owns (the upgrade rule, not the never-downgrade check, judges a move to an older release's definitions). Then the namespaces, the definitions, the Secrets, and the prerequisites, database, system and runtime phases, each awaited before the next.

## Tests

- `compose_test.go` — hermetic, against `testdata/<release>` (upstream's own files, verbatim) with a stand-in registry: an empty section installs upstream as its installer writes it; typed settings land where upstream's installer puts them; upstream's components and markers apply as upstream applies them; every image placeholder is pinned at the release, an unknown one refused; the rollout digests cover what each workload reads; the version picks the release and a missing file or image is refused; the upgrade rule follows upstream's policy; the pools and key set have upstream's wire shape; the Terraform twin reads the same files. Gated by `AGENT_SUBSTRATE_UPSTREAM_PIN=1`: the fixtures equal upstream's tag, and every image resolves at upstream's registry to an amd64 and arm64 index.
- `pools_upstream_test.go` — gated by `AGENT_SUBSTRATE_UPSTREAM_PARSE=1`: both engines' generated pools parse with upstream's own `localca` and `localjwtauthority` code, and their key sets equal upstream's `PublicJWKS`.
- `parity_test.go` — gated by `AGENT_SUBSTRATE_TF_PARITY=1` with a kubeconfig: plans the Terraform twin and compares its documents and outputs with this module's composition, across five specs.

## Usage

```bash
planton pulumi up --manifest agent-substrate.yaml
```

## Outputs

See `module/outputs.go`; every output is described in `../../v1alpha1/outputs.proto`.
