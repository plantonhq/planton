# Kubernetes registry family sub-bands, AKS + self-managed credential seam, persistent E2E cluster lanes

**Date**: 2026-07-21
**Scope**: `apis/dev/planton/shared/catalogkind`, `apis/dev/planton/provider/kubernetes`, `pkg/kubernetes/{kubetoken,execcredential,kubeconfig}`, `e2e`, site catalog

## What changed

### Kubernetes kind registry reorganized into family sub-bands

The Kubernetes enum band (800–999) is renumbered into self-describing family
sub-bands, so the registry itself teaches where a kind belongs and where the
next one goes:

| Sub-band | Family |
|---|---|
| 800–829 | Building blocks (core API primitives) |
| 830–859 | Foundation addons (certs, DNS, secrets, ingress, Gateway API, mesh) |
| 860–879 | Observability |
| 880–899 | Security, policy, identity |
| 900–929 | Data platforms |
| 930–949 | Reserved: analytics & ML |
| 950–969 | GitOps & CI/CD |
| 970–989 | App platforms |
| 990–999 | Reserved for growth |

The catalog has zero external adoption, which is what makes a one-time
renumber safe; kind names and id_prefixes are unchanged. Generated kind maps,
stubs, and the E2E matrix were regenerated.

### KubernetesGitlab removed

Running GitLab on a cluster is not a plausible destination for this catalog's
audience; the kind, its modules, site catalog pages, and registry entry are
removed.

### AKS credential seam completed

`KubernetesProviderConfigAzureAks` was an empty placeholder and the kubeconfig
builder rejected AKS connections. AKS now rides the same ExecCredential seam
as EKS and GKE:

- The provider config carries cluster identity (endpoint, CA) plus an Entra
  service-principal credential; an empty client secret selects the ambient
  Azure credential chain (env, managed identity, Azure CLI).
- `pkg/kubernetes/kubetoken` gains an AKS arm minting Entra access tokens for
  the AKS AAD server application (the kubelogin protocol) via the client
  credentials flow — offline-unit-tested, credentials never in argv.
- The single kubeconfig builder serves both deploy engines, so cross-engine
  drift on AKS wiring is structurally impossible.

Entra-less AKS clusters (local accounts) connect through the new self-managed
arm instead.

### self_managed provider arm

`KubernetesProvider` gains `self_managed = 5` with a kubeconfig-passthrough
config, covering kind/minikube/k3s, on-prem, bare metal, vSphere, and
kubeconfig-only clouds (Civo, Scaleway). Membership principle documented on
the enum: a platform earns its own value only when it brings its own
credential shape — mirroring the platform connection API.

### E2E harness: persistent kind cluster + external-cluster lane

- The kind cluster now PERSISTS across runs: setup reuses a running cluster
  (stable default name `planton-e2e`) and creates one only when absent;
  teardown leaves it running unless `PLANTON_E2E_DESTROY_CLUSTER=1` (set in
  the CI workflow for ephemeral runners). Cluster create/destroy dominated
  local run time; a warm rerun of a full scenario now completes in seconds.
- `PLANTON_E2E_KUBECONFIG` adds an external-cluster lane for batch-provisioned
  real clusters (EKS/GKE/AKS): the harness adopts the kubeconfig and never
  touches cluster lifecycle. Kind verify semantics are identical in both
  lanes.

## Validation

- `make protos`, `make generate-catalog-kind-map`, `make e2e-matrix`,
  `make build-go` — all green.
- `go build ./pkg/kubernetes/...`, `go test ./pkg/kubernetes/...` — green,
  including new AKS token/kubeconfig/dispatch unit tests.
- Live kind lane: `TestKubernetesNamespace_{Pulumi,Terraform}` — all scenarios,
  both engines, six-phase runner green; persistent-reuse path verified on a
  second run.
