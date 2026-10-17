# KubernetesAgentSandboxController Terraform Module

Installs agent-sandbox v1.0.7 from upstream's own release: the release asset the extensions setting selects (`sandbox.yaml`, or `sandbox-with-extensions.yaml`), and, when `router` is set, the router's deployment files from upstream's repository at the same tag. The typed spec is applied onto upstream's own objects in `locals.tf` and `documents.tf`; nothing is rebuilt from scratch, so every setting left unset keeps upstream's value. The exact semantic twin of the Pulumi module (`../pulumi/module/`).

## How the Documents Are Built

1. `data "http" "release_asset"` downloads the pinned asset; `data "http" "router_file"` downloads each router file the settings need (the TokenReview RBAC only in that mode, the cache RBAC only with the cache on).
2. The text is split on YAML separator lines and every document keyed by its composed identity, `apiVersion//kind//name[//namespace]`, which is also what `../import-map.yaml` reads.
3. The CustomResourceDefinitions leave this document set and go through the catalog's kept-definition block (`helm_crds.tf`, bundle branch), which downloads the same asset.
4. Typed settings patch the controller and router Deployments, Services, the PodDisruptionBudget, the NetworkPolicy and the label-domain ConfigMap. A flag upstream already passes keeps its position and takes the new value; any other flag is appended in a fixed order, and only when it differs from upstream's default.

## Order and Refusals

Every refusal fires at plan time, before anything is written. The preconditions sit on `terraform_data.preflight`, whose values are all known at plan, and conditions are not evaluated on destroy, so no refusal stands between a user and deleting their stack. No data source carries `depends_on`: a data source that depends on a resource is read at apply time.

- **A namespace other than upstream's** — a `namespace` reference that resolved to another name (the schema refuses another literal).
- **A controller or router someone else created** — `data "kubernetes_resources" "owned_deployment"` reads each Deployment by its fixed name; one without the `planton.ai/kind=KubernetesAgentSandboxController` label belongs to another installation (the GKE add-on, a `kubectl apply` of upstream's release).
- **The namespace's owner** — `data "kubernetes_resources" "namespace"` reads `agent-sandbox-system`: missing while `create_namespace` is false; created by someone else while it is true; or created by this resource while it is now false, which would delete it.
- **A definition someone else owns, or a newer one** — `helm_crds.tf` reads the definitions already on the cluster and refuses foreign owners and downgrades.
- **Upstream changed shape** — preconditions check that the release still carries the controller container and Service the module patches, that the router still runs `sandbox-router-go`, and that the router NetworkPolicy still has the rules the module rewrites. Any of them failing means the pin needs a catalog fix.

Each refusal says what was observed, what it means and the next step. Then the apply order: the Namespace when `create_namespace` is true (its delete blocks until it is fully gone, because the name is fixed), the kept definitions, and everything else after both, with `wait_for_rollout`. The controller holds no finalizers and registers no webhook at v1.0.7, so nothing deadlocks and a destroy drains nothing.

## Keep Mechanism

The definitions are applied with `apply_only`, so `tofu destroy` removes the controller and the router and leaves every Sandbox, SandboxTemplate, SandboxWarmPool and SandboxClaim on the cluster. `crds.keep_on_uninstall: false` turns it off. Server-side apply re-adopts kept definitions on reinstall.

## Resources

| Resource | Condition |
|---|---|
| `data.http.release_asset` | always |
| `data.http.router_file` (one per router file) | `router` set and enabled |
| `data.kubernetes_resources.owned_deployment` (one per fixed Deployment name) | always |
| `data.kubernetes_resources.namespace` | always |
| `terraform_data.preflight` (the refusals; writes nothing to the cluster) | always |
| `data.http.helm_crds_bundle`, `data.kubernetes_resources.helm_crds_existing` (`helm_crds.tf`) | `crds.install` (default true) |
| `kubectl_manifest.helm_crds` (one per definition, keyed by its name) | `crds.install` (default true) |
| `kubectl_manifest.namespace` (upstream's Namespace with the identity labels) | `create_namespace` |
| `kubectl_manifest.workloads` (one per document, keyed by composed identity) | always |

## Usage

```bash
planton tofu apply --manifest agent-sandbox.yaml
```

## Local Development

```bash
tofu init -backend=false
tofu validate
```

## Inputs

See `variables.tf`, generated from the spec proto. The namespace, Secret and certificate references arrive resolved to names before Terraform runs. OpenTofu or Terraform 1.4 or later (`terraform_data`).

## Outputs

| Output | Description |
|--------|-------------|
| `namespace` | `agent-sandbox-system` |
| `metrics_service_name` | The controller's metrics Service |
| `metrics_port` | The controller's metrics port |
| `router_service_name` | `sandbox-router-svc`, empty without a router |
| `router_endpoint` | The router's in-cluster URL (HTTPS when the plain listener is off), empty without a router |

## Parity

Kept in lockstep with the Pulumi module: the same pinned release and file addresses, the same composed identities, the same patches and flag order, the same refusals, the same outputs. `compose_test.go` there pins the release the two modules name, and its gated parity test compares this module's planned objects with the Pulumi module's composition across several specs.
