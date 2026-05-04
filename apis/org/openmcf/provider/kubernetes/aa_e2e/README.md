# Kubernetes E2E Provider Harness

This package (`aa_e2e`) implements the E2E test harness for the Kubernetes
provider. It manages the test cluster lifecycle and provides verification
logic that confirms deployed resources exist and destroyed resources are gone.

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

## Architecture

The harness has two responsibilities: **cluster lifecycle** and **resource
verification**.

```
aa_e2e/
  harness.go   -- Cluster lifecycle (Setup / Teardown / VerifyDeployed / VerifyDestroyed)
  verify.go    -- Verifier dispatch, verifier types, and kubectl helpers
```

### Harness (`harness.go`)

The `Harness` struct manages a kind (Kubernetes IN Docker) cluster. The
`e2e/e2e_test.go` TestMain creates one shared Harness for all Kubernetes
tests in a single run.

- `Setup` creates the kind cluster, writes a kubeconfig to a temp file, and
  sets the `KUBECONFIG` environment variable so Pulumi's Kubernetes provider
  finds it automatically.
- `Teardown` deletes the kind cluster and cleans up temp files.
- `VerifyDeployed` and `VerifyDestroyed` delegate to the appropriate
  `ResourceVerifier` by parsing the test manifest to determine what was
  deployed and what should be checked.

The Harness implements the `provider.Harness` interface from
`e2e/framework/provider/provider.go`, keeping the framework provider-agnostic.

### Verification (`verify.go`)

Verification is **manifest-driven**: the verifier reads the test manifest YAML
at runtime, extracts the `kind`, `metadata.name`, and `spec.namespace`, and
selects the appropriate verifier type. This means adding a new test scenario
(a YAML file in a component's `v1/e2e/` directory) never requires touching Go
code.

#### Verifier Types

| Verifier | Scope | Checks |
|----------|-------|--------|
| `NamespaceVerifier` | Tier 1 namespace | Namespace exists / absent |
| `WorkloadVerifier` | Tier 1 deployments, statefulsets | Resource exists in namespace / absent |
| `ResourceExistenceVerifier` | Tier 1 secrets, services | Resource exists in namespace / absent |
| `HelmComponentVerifier` | Tier 2 Helm-based apps | Namespace + at least one Running pod + at least one Service |
| `OperatorComponentVerifier` | Operator/controller installs | Namespace + at least one Running pod (no Service required) |
| `GenericVerifier` | Fallback | Always passes (logs a skip message) |

The dispatch logic in `GetVerifierFromManifest` uses two kind maps
(`operatorKinds` and `helmTier2Kinds`) plus a hardcoded switch for Tier 1
native resources. New component kinds are added to the appropriate map.

#### Retry Strategy

All kubectl operations use retry loops with increasing backoff to handle
Kubernetes eventual consistency:

- **Existence checks**: 5 attempts, 2-second base backoff
- **Absence checks**: 10 attempts, 2-second base backoff (resources take
  longer to finalize)
- **Helm pod readiness**: 15 attempts, 3-second base backoff (images must
  be pulled, init containers must complete)
- **Service existence**: 10 attempts, 2-second base backoff

## Adding a New Provider Harness

When extending E2E testing to a new provider (e.g., AWS):

1. Create `apis/org/openmcf/provider/aws/aa_e2e/`
2. Implement `harness.go` with provider-specific infrastructure lifecycle
   (e.g., creating a test VPC, configuring credentials from env vars)
3. Implement `verify.go` with provider-specific verification (e.g., AWS SDK
   calls to confirm resources exist)
4. Implement the `provider.Harness` interface from `e2e/framework/provider/`
5. Register the new harness in `e2e/e2e_test.go` TestMain

The test framework (`e2e/framework/`) is provider-agnostic. The 6-phase
lifecycle (VALIDATE, DEPLOY, VERIFY-OUT, VERIFY-RES, DESTROY, VERIFY-CLN)
and the runner work identically across all providers. Only the harness
implementation changes.

## Test Manifests

Test manifests live colocated with their components at
`{component}/v1/e2e/*.yaml`, not in this directory. The test framework
discovers them automatically via `e2e/framework/discovery/`. Adding a new
test scenario means dropping a YAML file -- zero Go code changes.
