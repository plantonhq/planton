# Kubernetes catalog: served-chart pinning guidance, kept-CRD namespace semantics, and silent-failure caveats landed on the surfaces their readers use

**Date**: 2026-07-23
**Scope**: spec proto field comments + regenerated Go stubs across twelve Kubernetes kinds (`kubernetesbackendtlspolicy`, `kubernetescertmanager`, `kubernetescilium`, `kubernetescloudnativepgoperator`, `kubernetesclusterautoscaler`, `kubernetesexternaldns`, `kubernetesexternalsecretsoperator`, `kubernetesingressnginx`, `kuberneteskarpenter`, `kuberneteskeda`, `kubernetesmetricsserver`, `kubernetesvelero`); docs READMEs for cert-manager, external-secrets-operator, KEDA, and Velero. Comments and documentation only — zero behavior change; no engine module, CEL rule, or rendered artifact differs.

## What changed

### Chart-version pinning: the served chart is the contract (11 kinds)

Every Helm-backed kind's `chart_version` comment now teaches where valid
versions come from: the chart repository's index (`helm search repo`) — or,
for Karpenter's OCI-served charts, the registry's published tags. The
upstream source tree's `Chart.yaml` can claim a version at a tag that was
never served (or lag the published OCI tags), so a version picked from the
source tree can fail at install with "chart not found" despite looking
authoritative. Landed uniformly on cert-manager, Cilium, CloudNativePG
operator, Cluster Autoscaler, external-dns, external-secrets operator,
ingress-nginx, Karpenter, KEDA, metrics-server, and Velero.

Istio deliberately does not carry the sentence: its charts are served from
Istio's own release pipeline with chart version = app version, and its
`version` field instead carries the sequential single-minor upgrade
contract. KubernetesHelmRelease does not carry it either — the chart there
is user-supplied, so the guidance reduces to a tautology.

### Kept-CRD installs pin their namespace (cert-manager, external-secrets operator, KEDA)

These kinds keep their CRDs on uninstall by default so that removing the
kind never cascade-deletes cluster-wide data (certificates, synced
secrets, scaling declarations). The `namespace` field on each now teaches
the consequence: chart-templated CRDs retain the Helm release's namespace
in their ownership metadata, so a later install into a DIFFERENT namespace
fails with Helm's release-ownership error on the surviving CRDs. Treat the
install namespace as permanent; moving requires first deleting the kept
CRDs — with each kind's cascade spelled out. Each kind's docs keep-section
carries the same warning. (Karpenter's `namespace` already taught this,
live-verified.)

Velero deliberately does NOT carry the warning: its CRDs ship in the
chart's `crds/` directory, which Helm creates raw — no release adoption, no
ownership metadata (verified in the Helm source: `installCRDs` performs a
bare create and skips existing CRDs). A Velero re-install into a new
namespace simply proceeds, and backup records resync from the object
store.

### Velero CSI snapshots: the silent-no-snapshot trap

`volume_snapshots.enable_csi` (and the docs' CSI section) now teach the two
cluster prerequisites Velero does not install: the external snapshot
controller and a `VolumeSnapshotClass` labeled
`velero.io/csi-volumesnapshot-class: "true"` for each volume's CSI driver.
Without them a backup still reports Completed while the volumes were never
snapshotted — the docs now tell operators to confirm VolumeSnapshot entries
in the backup's own resource list before trusting it for recovery.

### Cilium WireGuard encryption: node-kernel dependency

`encryption.type` now states that WireGuard rides the NODE kernel's module
and names the failure mode: on a node without the module the agent fails
to start rather than silently sending plaintext (verified in the Cilium
source — the WireGuard agent's start hook aborts agent startup when the
device cannot be created).

### BackendTLSPolicy: implementation-dependent by design

The spec header now warns that the policy only takes effect when the
Gateway's controller implements BackendTLSPolicy — support is still uneven
across implementations. A policy behind a non-implementing gateway is
accepted by the API server and then silently ignored: the gateway keeps
sending plaintext (or its own default TLS posture) to the backend.

## Why

Live operation keeps proving that the costliest configuration mistakes are
the ones that deploy cleanly: a backup that reports Completed with
unprotected volumes, a TLS policy the gateway silently ignores, a chart
version that exists in the source tree but not in the repository, an
install namespace that quietly became permanent. The catalog's contract is
that the spec is enough to configure a kind correctly — so these
constraints belong on the exact fields and doc sections their authors read,
not in tribal memory.

## Validation

- `buf lint` + `buf format --diff` clean on all twelve edited proto
  directories; stubs regenerated per directory with coverage verified
  (every edited `spec.proto` has its regenerated `spec.pb.go` sibling).
- Spec test suites re-run green for all twelve kinds (proving the diffs
  are prose-only — no CEL or shape change).
- Per-kind Pulumi release-entrypoint builds green ×12; repo-wide
  `make build-go` green.
- No engine module, preset, scenario, or chart changed; existing E2E
  results stand by construction.
