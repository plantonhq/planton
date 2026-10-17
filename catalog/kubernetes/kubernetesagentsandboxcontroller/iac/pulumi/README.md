# KubernetesAgentSandboxController Pulumi Module

Installs agent-sandbox v1.0.7 from upstream's own release: the release asset the extensions setting selects (`sandbox.yaml`, or `sandbox-with-extensions.yaml`), and, when `router` is set, the router's deployment files from upstream's repository at the same tag. The typed spec is applied onto upstream's own objects; nothing is rebuilt from scratch, so every setting left unset keeps upstream's value. The exact semantic twin of the Terraform module (`../tf/`).

## Layout

| File | Role |
|---|---|
| `module/vars.go` | The pinned release, upstream's fixed names, the file addresses |
| `module/settings.go` | The spec resolved to settings with upstream's defaults filled in |
| `module/documents.go` | Download, split on separator lines, key by composed identity |
| `module/flags.go` | Typed settings written onto upstream's container arguments |
| `module/workloads.go` | The patches on the Deployments, Services, PodDisruptionBudget, NetworkPolicy and ConfigMap |
| `module/compose.go` | The document set: upstream minus the definitions, patched |
| `module/locals.go` | The identity labels on the Namespace this module creates |
| `module/main.go` | Refusals, the kept definitions, then the namespace and the workloads |

## Order and Refusals

Everything that can refuse runs before anything registers:

1. **A namespace other than upstream's** — a `namespace` reference that resolved to another name (the schema refuses another literal).
2. **Upstream changed shape** — composition fails when the release no longer carries the objects the module patches, naming the pin as what needs a catalog fix.
3. **A controller or router someone else created** — `workloadowner` reads each Deployment by its fixed name; one without the `planton.ai/kind=KubernetesAgentSandboxController` label belongs to another installation (the GKE add-on, a `kubectl apply` of upstream's release) and stops the preview with what was observed, what it means and the next step.
4. **The namespace's owner** — `workloadowner` then reads `agent-sandbox-system`: missing while `create_namespace` is false; created by someone else while it is true; or created by this resource while it is now false, which would delete it.
5. **A definition someone else owns, or a newer one** — `keptcrds` derives the definitions from the same release asset (its bundle source, named `agent-sandbox`, pinned to v1.0.7) and refuses foreign owners and downgrades.

Then the namespace when `create_namespace` is true (upstream's, with the identity labels), the kept definitions, and the workloads, which depend on both. Pulumi awaits each Deployment's rollout. The controller holds no finalizers and registers no webhook at v1.0.7, so nothing deadlocks and a destroy drains nothing.

## Keep Mechanism

Every definition carries `retainOnDelete`, so `pulumi destroy` removes the controller and the router and leaves every Sandbox, SandboxTemplate, SandboxWarmPool and SandboxClaim on the cluster. `crds.keep_on_uninstall: false` turns it off. The definitions ride a dedicated upsert provider (server-side apply with force), so a reinstall adopts them again.

## Tests

- `compose_test.go` — hermetic: every typed setting reaches the objects, flags keep upstream's positions, the router's optional files follow their modes, and the Terraform twin pins the same release. Gated by `AGENT_SANDBOX_UPSTREAM_PIN=1`: the testdata fixtures still equal what upstream serves at the tag.
- `parity_test.go` — gated by `AGENT_SANDBOX_TF_PARITY=1` with a kubeconfig: runs `tofu plan` on the Terraform twin and compares its planned objects with this module's composition, across several specs.

## Usage

```bash
planton pulumi up --manifest agent-sandbox.yaml
```

## Outputs

| Output | Description |
|--------|-------------|
| `namespace` | `agent-sandbox-system` |
| `metrics_service_name` | The controller's metrics Service |
| `metrics_port` | The controller's metrics port |
| `router_service_name` | `sandbox-router-svc`, empty without a router |
| `router_endpoint` | The router's in-cluster URL (HTTPS when the plain listener is off), empty without a router |
