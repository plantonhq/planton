# Kubernetes Ingress Nginx

## When NOT to Use This

**This kind is the controller only — the machinery that answers
traffic, not the routing rules.** Routing rules are separate first-class
resources: create KubernetesIngress objects that reference this controller's
`ingress_class_name` output. TLS certificates come from cert-manager
(KubernetesCertManager plus the issuer and certificate kinds); this
component only consumes their Secrets. And if your traffic model is Gateway
API (Gateway/HTTPRoute) rather than Ingress, use the Gateway API kinds
instead — this controller reconciles Ingress resources.

## Overview

**KubernetesIngressNginx** installs the ingress-nginx controller — the
cluster's HTTP(S) entry point — from the official Helm chart
(`ingress-nginx` at `https://kubernetes.github.io/ingress-nginx`, default
chart 4.15.1 = controller v1.15.1). The controller watches Ingress
resources of its IngressClass and programs an embedded NGINX to route
external traffic to Services. The typed spec covers the chart's meaningful
configuration surface — ingress class identity, replicas and autoscaling,
the controller Service and its cloud annotations, workload shape,
NGINX tuning, TLS defaults, webhooks, metrics, TCP/UDP exposure,
scheduling — with a `helm_values` escape hatch (merged last, Helm `-f`
semantics, identical on both engines) for anything beyond it.

**Key design points:**

- **Multiple controllers per cluster are first-class**: the upstream
  pattern for public + internal traffic splits. Each instance gets its own
  resource with a DISTINCT `ingress_class.name` — the Helm release name,
  chart fullname, controller resource names, and leader-election identity
  all derive from `metadata.name`, so instances never collide
- **Each instance owns one IngressClass**: default `"nginx"`; the
  IngressClass's controller identifier derives automatically
  (`k8s.io/ingress-nginx` for class `"nginx"`, otherwise
  `k8s.io/<class-name>`) so additional controllers isolate without you
  inventing a vocabulary
- **The Service annotations map is the cloud integration surface**: the
  controller itself never calls cloud APIs — what the host cloud provisions
  for the LoadBalancer Service (NLB vs. classic ELB on AWS, internal vs.
  external on GCP/Azure) is driven entirely by `service.annotations`. There
  is no provider oneof and no workload-identity block by design
- **Replicas vs. autoscaling**: `replicas` pins a count; `autoscaling`
  hands the count to a chart-managed HPA (requires metrics-server for the
  utilization targets to receive values). With effective replicas above one
  the chart automatically guards the controller with a PodDisruptionBudget
  (minAvailable 1) — there is no separate PDB toggle
- **Snippet annotations stay off by default**: `allow_snippet_annotations`
  follows upstream's post-CVE-2021-25742 default — snippets let any Ingress
  author run arbitrary NGINX directives in the shared controller

## Environment Injection (how the cloud shapes the load balancer)

The controller runs identically everywhere; only the Service annotations
change per cloud. These are the standard recipes:

| Cloud / posture | `service.annotations` |
|---|---|
| AWS NLB (the production default on EKS) | `service.beta.kubernetes.io/aws-load-balancer-type: "external"` + `service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: "ip"` |
| AWS internal | add `service.beta.kubernetes.io/aws-load-balancer-scheme: "internal"` |
| GCP internal | `networking.gke.io/load-balancer-type: "Internal"` |
| GCP static IP | `networking.gke.io/load-balancer-ip-addresses: "<address-name>"` (or `controller.service.loadBalancerIP` via `helm_values`) |
| Azure internal | `service.beta.kubernetes.io/azure-load-balancer-internal: "true"` |

Two ways to serve public and private traffic:

- **Two controller instances** (recommended for real separation): each with
  its own namespace, ingress class, and Service annotations — fully
  independent stacks
- **One instance, dual LB** (`service.internal.enabled`): the chart's
  second, internal Service in front of the SAME controller pods. The
  internal Service REQUIRES at least one annotation — the cloud's
  internal-LB annotation is what makes it internal

On clusters WITHOUT a cloud LB controller (kind, bare metal), a
`load_balancer` service type FAILS LOUDLY at install time — Helm's
readiness wait includes the LB address, and an address that can never
arrive times out. This is deliberate: use `node_port` or host access
(`host_ports` / `host_network`, typically with `controller_kind:
daemon_set`) on such clusters.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: installation namespace (`ingress-nginx` by
  convention; a per-instance namespace like `ingress-nginx-internal` for
  additional controllers) — literal or a KubernetesNamespace reference

### Common

- **`spec.create_namespace`**: create (and own) the namespace with the
  release
- **`spec.chart_version`**: pinned chart version (default `4.15.1`, which
  ships controller v1.15.1 — chart and controller are released together but
  numbered independently)
- **`spec.ingress_class.name`**: the class this instance owns (default
  `nginx`); `is_default_class` marks it the cluster default for Ingresses
  that specify no class
- **`spec.replicas`** / **`spec.autoscaling`**: production entry points run
  2+ replicas; autoscaling hands the count to the chart's HPA
- **`spec.service`**: type (`load_balancer` / `node_port` / `cluster_ip`),
  the per-cloud `annotations`, `external_traffic_policy` (`local` preserves
  client source IPs — the usual production choice), source ranges, and the
  optional `internal` dual-LB Service
- **`spec.nginx_config`**: global NGINX tuning in upstream's own ConfigMap
  key/value vocabulary (`proxy-body-size`, `use-forwarded-headers`, ...)
- **`spec.default_tls_certificate`**: cluster-wide default certificate —
  the cert-manager seam: point it at a KubernetesCertificate's secret
  output for an auto-renewed wildcard default
- **`spec.metrics`**: controller metrics (port 10254) and an opt-in
  ServiceMonitor (requires the Prometheus operator CRDs — the release fails
  to install without them)
- **`spec.tcp_services`** / **`spec.udp_services`**: raw L4 exposure —
  maps of external port to `"namespace/service:port"`
- **`spec.helm_values`**: escape hatch for chart values beyond the typed
  fields — never the primary interface

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Installation namespace |
| `release_name` | Helm release name (equals `metadata.name`; controller resources are named `<release>-controller`) |
| `ingress_class_name` | The IngressClass this controller owns — what KubernetesIngress resources reference |
| `controller_service_name` | The controller's external Service (`<name>-controller`) — the traffic entry point |
| `internal_service_name` | The internal Service — populated only when `service.internal.enabled` |
| `load_balancer_ip` | External IP of the cloud LB (GCP/Azure populate an IP; empty otherwise) |
| `load_balancer_hostname` | External hostname of the cloud LB (AWS populates a DNS name; empty otherwise) |

`load_balancer_ip` and `load_balancer_hostname` are empty on clusters
without a cloud LB controller and for `node_port` / `cluster_ip` service
types.

## Composing in Infra Charts

The standard wiring: this controller deploys first; KubernetesIngress
resources reference its `ingress_class_name` output to route through it;
KubernetesExternalDns publishes DNS records for the LB address the cloud
assigns; cert-manager kinds (KubernetesCertManager, issuers,
KubernetesCertificate) mint the certificates that Ingresses — or the
`default_tls_certificate` field — consume. A cluster with a public +
internal split runs two instances of this component, each with its own
ingress class and Service annotations.

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesIngressNginx
metadata:
  name: ingress-nginx
spec:
  namespace:
    value: ingress-nginx
  createNamespace: true
  ingressClass:
    name: nginx
    isDefaultClass: true
  replicas: 2
  service:
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-type: external
      service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: ip
    externalTrafficPolicy: local
  defaultTlsCertificate:
    secretName:
      valueFrom:
        kind: KubernetesCertificate
        name: wildcard-cert
        fieldPath: status.outputs.secret_name
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
