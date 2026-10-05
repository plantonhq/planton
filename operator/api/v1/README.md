# api/v1 -- PlantonPlatform CRD Types

This package defines the `PlantonPlatform` custom resource API at `planton.ai/v1`.

## Why This Package Exists

The CRD types are the user contract. Everything a user can configure and everything the operator reports back is defined here. Changes to this package change the external API surface.

## API Group

`apiVersion: planton.ai/v1` -- the domain `planton.ai` serves as the API group. This is a deliberate choice for a single-CRD operator: clean, no group prefix clutter. If the operator grows to manage multiple CRDs, we would introduce a group prefix (e.g., `install.planton.ai/v1`).

## Design Decisions

### Opinionated Defaults

A user should be able to apply a CR with only `spec.version` set and get a working deployment. All storage sizes, component toggles, and ingress settings have sensible defaults:

- PostgreSQL: 10Gi per instance
- Store (Valkey, redis-protocol): persisted on 1Gi, a 768mb dataset ceiling with `allkeys-lru` eviction under a 1Gi memory limit -- the store carries the live build-log stream beside the cache, so it persists, and the ceiling is what makes what it persists always reload
- Every workload the operator renders carries a sizing the operator chose (requests for CPU and memory, a memory limit, never a CPU limit), measured and recorded with its reason in one registry (`internal/resources/sizing.go`) -- never a chart's smallest preset by omission. See Sizing below
- Ingress: disabled (use port-forward)
- Runner and builds (Tekton): enabled -- deploying and building are the product; opting out is the explicit act
- Optional components (Graph): disabled
- Observability: metrics always served, inside the cluster, on the control plane's and runner's Service port `metrics` (9464); traces off until `spec.observability.otlpHttpEndpoint` names a trace store -- the address is the switch, so there is no separate flag to disagree with it

### Sizing

Every workload's size is one entry in the operator's registry (`internal/resources/sizing.go`, `ComponentSizing`), keyed by the spec field that overrides it: `spec.controlPlane.resources`, `spec.temporal.history.resources`, and so on. The registry is the one home of each measured default; the builders, the sizing check, the status readback, and the published `component_sizing.json` all read it.

- **Per-quantity merge.** A quantity set in a `resources` field wins; one left unset keeps the default. Raising one memory limit never means restating the numbers beside it, and a re-measured default still reaches every quantity nobody chose.
- **`ComponentResources`, not `corev1.ResourceRequirements`.** The Kubernetes type also carries `claims`, which the operator would accept and ignore; the two lists alone keep each repeated block small.
- **Judged before anything runs.** A size no workload can run with -- a request above its limit, a request under a floor a chart enforces (Neo4j's 500m and 2Gi), the store's `maxMemory` at or above its memory limit -- is refused whole: `ResourcesValid` False naming each field and its fix, nothing created or changed, no requeue. The same shape as `VersionSupported`.
- **Reported as it runs.** `status.components.<component>.sizing` lists each workload's effective requests and limits by path. A field declared against an older operator's definition is dropped by the API server without a word; this list is where that shows.
- **Published as data.** `make manifests` writes the registry to `component_sizing.json` for readers outside this module (the catalog kind that installs a platform states each default as its own field's option, and a test holds the two equal).
- **Not sized, deliberately:** the cluster-shared operators (CloudNativePG, the Barman plugin, Tekton), one-shot init containers and Jobs, and the operator's own pod (its chart's `resources`).

Resizing rolls the component's pods. A single-instance PostgreSQL restarts the only database (CloudNativePG switches over when there are replicas); OpenBAO comes back sealed and the operator unseals it again.

### Status Structure

Status uses two complementary mechanisms:

1. **ComponentStatuses**: A structured object with one field per component. This provides clear `kubectl get -o yaml` output and type-safe access from the reconciler. We chose named fields over a map to avoid stringly-typed component names.

2. **Conditions**: Standard `metav1.Condition` slice following Kubernetes API conventions. `Ready` aggregates every enabled component and its message is the `MESSAGE` column of `kubectl get plantonplatform`; `VersionSupported` says whether `spec.version` names a platform release this operator runs, and when it does not, its message names the oldest release the operator supports and the two ways forward; `ResourcesValid` says whether every declared size is one its workload can run with, and when one is not, names the field and the fix.

### Phase Enums

`PlantonPhase` and `ComponentPhase` are string-typed enums validated by kubebuilder markers. They are intentionally separate types because the overall deployment has lifecycle states (Upgrading) that individual components do not.

## Code Generation

After modifying types in this package, run:

```bash
make generate   # Regenerate DeepCopy methods
make manifests  # Regenerate CRD YAML and RBAC
```
