# pulumi-kubernetes SDK v4.18.4 → v4.33.0; SeaweedFS fully live-proven

## What changed

- **The `pulumi-kubernetes` SDK dependency moved from v4.18.4 to
  v4.33.0** (with `go mod tidy` + Bazel dependency sync). The driver:
  the provider's embedded Helm SDK. v4.18.4 embeds Helm 3.15.4, which
  cannot parse charts using Helm ≥ 3.16 template functions — the
  official SeaweedFS chart uses `fromToml`, and Helm parses every
  template file before evaluating gates, so no values could avoid it;
  every Pulumi `helm.sh/v3:Release` install of the chart failed at
  render. v4.33.0 embeds Helm 3.20.2. Because a Release renders
  templates only at APPLY, previews pass on an incompatible engine —
  the kind update workflow now teaches checking a chart's
  template-function floor against the provider's embedded Helm version
  when adopting a chart pin.

- **Verification:** the repo-wide Bazel build gate compiles every
  Pulumi module on the new line; the SeaweedFS Pulumi E2E lanes ran
  green end to end (S3 put/get round-trips with chart-generated
  credentials, the declared-bucket install-hook proof, and the
  volume-loss durability proof — an object read back byte-identical
  after the volume server pod was deleted); and the Qdrant Pulumi lanes
  were re-run as a proven-kind canary on the new provider — green
  including the vector-persistence proof, no behavioral drift.

- **KubernetesSeaweedFs is now fully live-proven on both engines** and
  entered the green E2E CI matrix.

## Validation

`make build-go` green on v4.33.0; SeaweedFS Pulumi 3/3 scenario lanes
green; Qdrant Pulumi canary 3/3 green; zero orphaned resources; the
E2E CI matrix regenerated.
