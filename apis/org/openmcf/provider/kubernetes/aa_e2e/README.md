# Kubernetes E2E Provider Harness

This package (`aa_e2e`) implements the E2E test harness for the Kubernetes
provider. It manages the test cluster lifecycle and delegates resource
verification to the `verify/` subpackage.

## Why `aa_e2e`?

Every cloud provider in OpenMCF can have an E2E harness colocated alongside
its component directories. The directory is named `aa_e2e` (not `_e2e` or
`e2e`) for two reasons:

- **Go ignores directories starting with `_`**. A directory named `_e2e`
  would be silently excluded from the Go build, meaning the package would
  never compile or be importable.
- **`aa_` sorts first alphabetically** across all providers (aws, azure, gcp,
  kubernetes, etc.), making the harness directory immediately visible in file
  explorers without needing to scroll past dozens of component directories.

This naming convention applies to all providers. When adding E2E support for
a new provider (e.g., AWS), create `apis/org/openmcf/provider/aws/aa_e2e/`.

## Directory Layout

```
aa_e2e/
  harness.go    -- Kind cluster lifecycle (Setup / Teardown / VerifyDeployed / VerifyDestroyed)
  README.md     -- This file
  verify/       -- Manifest-driven resource verification (separate package)
    manifest.go           -- ManifestInfo struct + ParseManifestInfo
    verifier.go           -- ResourceVerifier interface, GetVerifierFromManifest dispatch, kind maps
    kubectl.go            -- kubectl helper functions (retry, backoff, resource exist/absent)
    namespace.go          -- NamespaceVerifier (Tier 1)
    workload.go           -- WorkloadVerifier (Tier 1 deployments, statefulsets)
    resource_existence.go -- ResourceExistenceVerifier (Tier 1 secrets, services)
    operator.go           -- OperatorComponentVerifier (Tier 4 operators)
    crd_workload.go       -- CRDWorkloadVerifier (Tier 3 operator-dependent CRD workloads)
    helm.go               -- HelmComponentVerifier (Tier 2 Helm-based apps)
    generic.go            -- GenericVerifier (fallback, always passes)
```

## Harness (`harness.go`)

The `Harness` struct manages a kind (Kubernetes IN Docker) cluster. The
`e2e/e2e_test.go` TestMain creates one shared Harness for all Kubernetes
tests in a single run.

- `Setup` creates the kind cluster, writes a kubeconfig to a temp file, and
  sets the `KUBECONFIG` environment variable so Pulumi's Kubernetes provider
  finds it automatically.
- `Teardown` deletes the kind cluster and cleans up temp files.
- `VerifyDeployed` and `VerifyDestroyed` delegate to `verify.GetVerifierFromManifest`
  which parses the test manifest and returns the appropriate verifier.

The Harness implements the `provider.Harness` interface from
`e2e/framework/provider/provider.go`, keeping the framework provider-agnostic.

## Verification (`verify/`)

Verification is **manifest-driven**: the verifier reads the test manifest YAML
at runtime, extracts the `kind`, `metadata.name`, and `spec.namespace`, and
selects the appropriate verifier type. This means adding a new test scenario
(a YAML file in a component's `v1/e2e/` directory) never requires touching Go
code.

### Verifier Types

| Verifier | File | Tier | Checks |
|----------|------|------|--------|
| `NamespaceVerifier` | `namespace.go` | 1 | Namespace exists / absent |
| `WorkloadVerifier` | `workload.go` | 1 | Deployment or StatefulSet exists in namespace / absent |
| `ResourceExistenceVerifier` | `resource_existence.go` | 1 | Secret or Service exists in namespace / absent |
| `HelmComponentVerifier` | `helm.go` | 2 | Namespace + running pods + services |
| `OperatorComponentVerifier` | `operator.go` | 4 | Namespace + running pods (no service requirement) |
| `CRDWorkloadVerifier` | `crd_workload.go` | 3 | Namespace + running pods + services |
| `GenericVerifier` | `generic.go` | -- | Always passes (logs a skip message) |

### Dispatch (`verifier.go`)

`GetVerifierFromManifest` uses three kind classification maps plus a hardcoded
switch for Tier 1 native resources:

- **`operatorKinds`** -- Tier 4 operator/controller components (namespace + pods)
- **`crdWorkloadKinds`** -- Tier 3 CRD workloads (namespace + pods + services)
- **`helmTier2Kinds`** -- Tier 2 Helm-based applications (namespace + pods + services)

New component kinds are added to the appropriate map. Tier 1 native resources
(namespace, deployment, statefulset, secret, service) are routed via the switch.

### Retry Strategy (`kubectl.go`)

All kubectl operations use retry loops with progressive backoff to handle
Kubernetes eventual consistency:

- **Existence checks**: 5 attempts, 2-second base backoff
- **Absence checks**: 10 attempts, 2-second base backoff (resources take
  longer to finalize)
- **Pod readiness**: 15 attempts, 3-second base backoff (images must
  be pulled, init containers must complete, CRDs must reconcile)
- **Service existence**: 10 attempts, 2-second base backoff

## Adding a New Provider Harness

When extending E2E testing to a new provider (e.g., AWS):

1. Create `apis/org/openmcf/provider/aws/aa_e2e/`
2. Implement `harness.go` with provider-specific infrastructure lifecycle
3. Create a `verify/` subpackage with provider-specific verification logic
4. Implement the `provider.Harness` interface from `e2e/framework/provider/`
5. Register the new harness in `e2e/e2e_test.go` TestMain

## Test Manifests

Test manifests live colocated with their components at
`{component}/v1/e2e/*.yaml`, not in this directory. The test framework
discovers them automatically via `e2e/framework/discovery/`. Adding a new
test scenario means dropping a YAML file -- zero Go code changes.
