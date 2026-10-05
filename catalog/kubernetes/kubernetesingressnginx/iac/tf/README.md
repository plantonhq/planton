# KubernetesIngressNginx Terraform Module

Terraform/OpenTofu module for the KubernetesIngressNginx kind: installs
the ingress-nginx controller from the official Helm chart (`ingress-nginx`
at https://kubernetes.github.io/ingress-nginx) as the cluster's HTTP(S)
entry point, with multi-instance coexistence built in.

## Module Behavior

- **One Helm release named after `metadata.name`** — multiple controller
  instances per cluster (public + internal traffic splits, each owning its
  own IngressClass) are a first-class pattern, so nothing here is a fixed
  chart name. The chart's fullname is pinned to the release name too, so
  every chart object (Deployment/DaemonSet, Services, IngressClass, RBAC,
  admission webhook) carries a deterministic, manifest-derived name —
  including the leader-election identity (`<fullname>-leader`), which
  isolates election per instance.
- **The typed spec renders into chart values** (`locals.typed_values`), and
  the spec's `helm_values` escape hatch is passed as a SECOND values
  document that the provider merges over the first with Helm `-f`
  semantics — the exact semantic twin of the Pulumi module's
  `buildHelmValues` + `mergeMaps`.
- **Chart-name parity**: the repository serves this chart as
  `ingress-nginx` — the repo-prefixed spelling `kubernetes-ingress-nginx`
  does not exist in the index and fails at install time. The chart identity
  is kept byte-identical with the Pulumi module's vars; cross-engine drift
  would deploy two different products from one manifest.
- **The IngressClass controller identifier derives automatically**: the
  chart default `k8s.io/ingress-nginx` for class `nginx`, otherwise
  `k8s.io/<class-name>` — additional controllers isolate without the user
  inventing a vocabulary. The legacy annotation-vocabulary value
  (`controller.ingressClass`) tracks the class name too.
- **The install waits for the release to become ready**
  (`wait`/`wait_for_jobs`/`atomic`/`cleanup_on_fail`, 300s timeout) — a
  controller that never starts (bad image, unschedulable pod, webhook
  certgen failure) fails THIS deploy, not the first Ingress.
  `wait_for_jobs` covers the admission-webhook certgen hook Jobs.
- **LB-wait posture**: Helm's readiness check on a LoadBalancer-type
  Service also waits for the cloud LB address, so on clusters WITHOUT a
  cloud LB controller (kind, bare metal) a `load_balancer` service type
  times out loudly here — deliberate: use `node_port`/host access on such
  clusters, and the failure names the real problem instead of leaving a
  silently Pending entry point.
- **The controller Service is read back** (`data.kubernetes_service_v1`)
  for the load-balancer address outputs — gated on the `load_balancer`
  service type: for `node_port`/`cluster_ip` there is no LB status to read
  and the address outputs stay empty by design. For `load_balancer` the
  Helm wait guarantees the address exists by the time the read runs.

## Resources

| Resource | Condition |
|---|---|
| `kubernetes_namespace_v1.ingress_nginx` | `spec.create_namespace` |
| `helm_release.ingress_nginx` | always |
| `data.kubernetes_service_v1.controller` (read) | `spec.service.type` is `load_balancer` |

The chart owns ALL of the controller's Kubernetes objects; the module
itself creates only the optional anchor namespace (stamped with the
standard governance labels — never injected into chart resources) and reads
the controller Service back.

## Outputs

| Output | Meaning |
|---|---|
| `namespace` | Namespace the controller is installed in |
| `release_name` | Helm release name (= `metadata.name`; controller resources are named `<release>-controller`) |
| `ingress_class_name` | The IngressClass this controller owns — what KubernetesIngress resources reference |
| `controller_service_name` | The controller's external Service (`<name>-controller`) — the traffic entry point |
| `internal_service_name` | The internal Service — empty unless `spec.service.internal.enabled` |
| `load_balancer_ip` | External IP of the cloud LB (providers that populate an IP; empty otherwise) |
| `load_balancer_hostname` | External hostname of the cloud LB (providers that populate a DNS name; empty otherwise) |

## Parity

Kept in lockstep with the Pulumi module (`iac/pulumi/module/`): same chart
identity and pinned default version, same values rendering (ingress-class
derivation, replicas-vs-autoscaling exclusivity, host-network DNS policy,
default-TLS `extraArgs` flag, default-backend image split), same wait
posture, same outputs. Conditional objects use the null-prune idiom
throughout so numbers and booleans keep their types in the rendered values.
