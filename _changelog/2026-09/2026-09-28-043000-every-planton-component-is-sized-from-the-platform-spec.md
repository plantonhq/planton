# Every Planton Component Is Sized from the Platform's Spec

**Date**: September 28, 2026
**Type**: Feature
**Components**: Operator (`operator/api/v1`, `internal/resources`, `internal/component`, `internal/platformsizing`, `internal/controller`, `internal/status`), `hack/componentsizing`, the operator chart's CRDs

## Summary

Until now, a self-hosted platform's CPU and memory were Go constants in the operator. Only the store could be sized, and setting its `resources` replaced the whole block, so raising one limit dropped the requests beside it. A control plane OOM-killed by three parallel deploys could not be given more memory by any adopter.

Every per-platform workload can now be sized from the spec: the control plane, console, runner, gateway, identity server, PostgreSQL, Valkey, OpenBAO, OpenFGA, Neo4j, and Temporal's frontend, history, matching and worker. Each takes a `resources` field. The quantities you set win; the ones you leave out keep the operator's measured default. A size no workload can run with is refused before anything changes, and the status reports the size each workload actually runs with.

## What Changed

- **`ComponentResources`** (`requests`, `limits`) on `spec.controlPlane`, `spec.console`, `spec.runner`, `spec.gateway`, `spec.identity`, `spec.database.postgresql`, `spec.database.redis` (it replaces that field's `corev1.ResourceRequirements`), `spec.vault` and `spec.components.graph`. Two new blocks as well: `spec.temporal.{frontend,history,matching,worker}` and `spec.openfga`. It isn't the Kubernetes type because that type also carries `claims`, which the operator would accept and then ignore.
- **One registry** (`internal/resources/sizing.go`, `ComponentSizing`), keyed by spec path, holds each measured default with its reason, plus the one floor a program enforces (Neo4j's chart: 500m CPU and 2Gi memory). It replaces the per-renderer constants.
  - `Effective` merges quantity by quantity.
  - `mustBeSized` makes a builder refuse an empty size, so a component that forgets to resolve its sizing fails its first test instead of shipping an unsized pod.
- **`ResourcesValid`** is judged right after `VersionSupported` and in the same shape: refused whole, nothing created or changed, no requeue. It checks:
  - every effective request against its limit, saying which half was declared and which is the default;
  - Neo4j's floor;
  - Valkey's `maxMemory` against its effective memory limit, read in Valkey's own units.
- **`status.components.<component>.sizing`** reports each workload's effective requests and limits by path. A field the API server dropped (declared against an older definition) shows up here as its default.
- **Out-of-memory explanations** name the exact field to raise (`spec.<path>.limits.memory`), carried by `WorkloadRef.Sized`.
- **`api/v1/component_sizing.json`** is the registry published as data by `hack/componentsizing` through `make manifests`, and it's held fresh by `ensure_operator_manifests_fresh.sh`.
- **A CRD size guard** in `test/chart`: every definition stays under 230 KiB, below client-side `kubectl apply`'s 256 KiB annotation cap. The platform definition is now 211 KiB (it was 157 KiB).
- **Neo4j** is rendered with requests and limits in the house pattern (no CPU limit). The chart's own `cpu`/`memory` pair set both.

## Verification

- `go test ./...` passes in the operator module, envtest included (`KUBEBUILDER_ASSETS` pointed at `bin/k8s`).
- **Red proofs:**
  - Restoring the old "replace whole" rule in `Effective` fails the one-quantity and added-quantity tests.
  - Turning the sizing check off in the controller fails the envtest refusal case.
  - A builder handed no sizing panics with the path and the fix.
- `make manifests generate` is clean, and the guard is clean once staged.
