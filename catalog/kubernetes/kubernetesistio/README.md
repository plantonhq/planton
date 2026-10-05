# Kubernetes Istio

## When NOT to Use This

**If you need a gateway, do not look for it here.** This component deploys NO gateways and NO ingress. istiod implements the Kubernetes Gateway API natively, so north-south exposure composes from the Gateway API kinds: a KubernetesGateway with `gateway_class_name: istio` makes istiod provision and program the gateway deployment automatically (named `<gateway>-istio`), and route kinds (KubernetesHttpRoute, ...) attach to it.

**If your cluster only uses the typed Istio policy/config kinds without running a mesh**, you do not need a control plane at all — KubernetesIstioBaseCrds installs just the Istio CRD bundle. When a mesh is later wanted, deploying KubernetesIstio on that cluster is a plain redeploy: both kinds co-own the CRDs.

**One KubernetesIstio per cluster.** The CRDs and the validation-webhook plumbing are cluster singletons; a second installation would fight the first over them.

## Overview

**KubernetesIstio** installs the Istio service-mesh **control plane** from the official Helm charts (`base` + `istiod`, plus `cni` + `ztunnel` in ambient mode, all from `https://istio-release.storage.googleapis.com/charts`, at one pinned version — default `1.30.3`). istiod is the mesh's brain: it validates mesh configuration, issues workload certificates, and programs the data plane (sidecar proxies, ambient node proxies, and gateways).

**Key design points:**

- **The CRDs are module-owned**, applied via server-side apply OUTSIDE the Helm release (the `base` release excludes the whole bundle). Helm cannot adopt CRDs that already exist without its ownership metadata, so Helm-owned CRDs would make a CRDs-only cluster (KubernetesIstioBaseCrds) permanently unable to upgrade to the full mesh. Module-owned CRDs are co-ownable by both kinds — that upgrade is a plain redeploy. Destroying this component removes the CRDs with everything else, and mesh configuration objects cascade with them.
- **Data plane choice is first-class**: `dataplane_mode: sidecar` (the default — namespaces opt in to injection via the `istio-injection=enabled` label) or `ambient` (no sidecars — a per-node ztunnel DaemonSet carries mTLS/L4, waypoint proxies add L7 where needed; enrollment via the `istio.io/dataplane-mode=ambient` label). Choose at install time — switching modes later is a workload-by-workload migration, not a flag flip.
- **Upgrades are sequential and single-minor.** Istio supports stepping one minor at a time; skipping minors in place is unsupported. `revision` names the control plane (`istiod-<revision>`, workload selection via `istio.io/rev`) for organizations that always run revisioned control planes — side-by-side canary control planes are deliberately not modeled.
- **`helm_values` is a per-release escape hatch** (`base`/`istiod`/`cni`/`ztunnel` YAML documents, merged last with Helm `-f` semantics, identical on both engines) — a safety valve on top of the fully modeled spec, never the primary interface.

## Environment Injection (what differs per cluster type)

The control plane itself calls no cloud APIs — the same spec installs identically on any conformant cluster. What differs per environment is only how auto-provisioned gateways get exposed, and optionally where images come from:

| Environment | `gateway_defaults.service_type` | `images.hub` |
|---|---|---|
| EKS | `LoadBalancer` (upstream default) | optional ECR mirror |
| GKE | `LoadBalancer` (upstream default) | optional Artifact Registry mirror |
| AKS | `LoadBalancer` (upstream default) | optional ACR mirror |
| kind / k3s / bare metal | `NodePort` or `ClusterIP` (no cloud LB controller) | usually default (`docker.io/istio`) |
| Air-gapped | per cluster capability | required — private registry, plus `images.image_pull_secrets` |

A per-Gateway override of the service type also exists via the Gateway's `infrastructure` parameters.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: control-plane namespace (`istio-system` by convention) — literal or a KubernetesNamespace reference

### Common

- **`spec.create_namespace`**: create (and own) the namespace with the resource
- **`spec.version`**: Helm chart version for every release (Istio versions its charts in lockstep with the product; default `1.30.3`) — pin deliberately, upgrade one minor at a time
- **`spec.dataplane_mode`**: `sidecar` (default) or `ambient`
- **`spec.revision`**: control-plane revision name; empty runs the unnamed default revision (the standard posture)
- **`spec.istiod`**: control-plane sizing — replicas XOR autoscale (the chart's HPA defaults to min 1 / max 5 at 80% CPU), resources, PodDisruptionBudget, priority class, scheduling
- **`spec.mesh_config`**: trust domain (set a stable, organization-unique value BEFORE production — changing it later re-identifies every workload), outbound traffic policy (`REGISTRY_ONLY` is the egress-lockdown posture), access logging, multi-cluster identity fields
- **`spec.proxy`** / **`spec.sidecar_injector`**: sidecar-mode proxy defaults and injection behavior
- **`spec.cni`**: the node-level agent — always installed in ambient mode; opt-in in sidecar mode (replaces the injected privileged init-container; required on platforms that forbid NET_ADMIN init containers)
- **`spec.ztunnel`**: ambient node-proxy sizing (ambient mode only)
- **`spec.gateway_defaults.service_type`**: default Service type for auto-provisioned Gateway API gateways
- **`spec.images`**: hub/variant/pull-secrets for every Istio image — the air-gapped and hardening knobs
- **`spec.helm_values`**: per-release escape hatch — never the substitute for the typed fields

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Control-plane namespace (e.g. `istio-system`) |
| `istiod_service_name` | istiod Service name (`istiod`, or `istiod-<revision>` for a named revision) — the discovery address proxies connect to |
| `revision` | Installed control-plane revision (`default` when unnamed) — the `istio.io/rev` selection handle |
| `gateway_class_name` | GatewayClass istiod serves (`istio`) — the composition seam for KubernetesGateway |
| `trust_domain` | Identity root of every workload certificate — the prefix of principal strings in KubernetesAuthorizationPolicy rules |
| `dataplane_mode` | `sidecar` or `ambient` — tells composed resources how workloads enroll |

## Composing in Infra Charts

The mesh is the substrate; everything else composes against its outputs:

- **North-south exposure** (Gateway API seam): a KubernetesGateway with `gateway_class_name: istio` (or a `valueFrom` reference to this component's `status.outputs.gateway_class_name`) makes istiod auto-provision the gateway deployment; KubernetesHttpRoute resources attach to that Gateway. No gateway release ships with this component by design.
- **Mesh traffic policy** (typed Istio kinds): KubernetesDestinationRule, KubernetesPeerAuthentication, KubernetesAuthorizationPolicy, KubernetesRequestAuthentication, KubernetesServiceEntry, KubernetesTelemetry, KubernetesEnvoyFilter. These need only the CRDs this component installs.
- **Identity**: `status.outputs.trust_domain` prefixes the SPIFFE principals (`<trust_domain>/ns/<ns>/sa/<sa>`) that AuthorizationPolicy rules match on.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
