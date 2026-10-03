# Ingress NGINX

Installs the ingress-nginx controller — the cluster's HTTP and HTTPS entry point — from the official `ingress-nginx` Helm chart. The controller watches Ingress resources of its IngressClass and programs an embedded NGINX to route external traffic to Services. This component installs the CONTROLLER only: routing rules are separate first-class KubernetesIngress resources that name this controller's published class, and TLS certificates come from cert-manager through each Ingress's TLS secret or cluster-wide through the default certificate this spec can reference.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** — created only when `createNamespace` is true; otherwise deploys into an existing namespace
- **Helm Release** — the official `ingress-nginx` chart (release named `metadata.name`; controller resources are named `<release>-controller`), creating:
  - The controller workload — a Deployment by default, or a DaemonSet when `controllerKind: daemon_set` (the bare-metal/edge pattern), with the declared replicas, resources, node selector, and tolerations
  - The **IngressClass** this controller owns — the name KubernetesIngress resources reference to be served by this instance
  - The controller **Service** — the traffic entry point (LoadBalancer by default; NodePort or ClusterIP per `service.type`), shaped entirely by `service.annotations`
  - A second **internal Service** — only when `service.internal.enabled` is true (the single-controller dual-LB pattern)
  - The **admission webhook** with its certgen bootstrap Job — enabled by chart default; rejects broken Ingress objects at apply time
  - A **PodDisruptionBudget** (minAvailable 1) — added by the chart automatically whenever more than one replica (or autoscaling with `minReplicas` above 1) is declared; there is no separate toggle
  - The optional **default backend** Deployment, and the metrics Service plus ServiceMonitor when `metrics` is declared

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A cloud load-balancer controller that reads your Service annotations** — only for `service.type: load_balancer`. On EKS, the `service.beta.kubernetes.io/aws-load-balancer-type: "external"` annotation family belongs to the AWS Load Balancer Controller: set it without that controller installed and the Service simply never receives an address — no error surfaces anywhere, and this module's readiness wait times out. EKS's built-in cloud controller answers `"nlb"` instead. Both paths are verified live.
- **metrics-server** — only when `autoscaling` is enabled; the utilization targets receive no values without it.
- **The Prometheus Operator CRDs** — only when `metrics.serviceMonitor` is true; the release FAILS to install without them.

## Deploy

### Console

Open the deployment store, find **Ingress NGINX**, and click **Deploy**. The creation wizard walks you through placement, the chart pin, the IngressClass identity, replicas and autoscaling, the controller Service and its cloud annotations, NGINX tuning, the default TLS certificate, the admission webhook, metrics, and TCP/UDP passthrough. Start from the **AWS NLB Public Entry (EKS)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesIngressNginx
metadata:
  name: ingress-nginx
  org: acme-corp
  env: prod
spec:
  namespace:
    value: ingress-nginx
  createNamespace: true
  replicas: 2
  service:
    externalTrafficPolicy: local
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
```

```shell
planton apply -f ingress-nginx.yaml
```

This installs a two-replica controller (the chart adds a PodDisruptionBudget automatically) owning the standard `nginx` IngressClass, behind an AWS NLB provisioned by EKS's built-in cloud controller, with client source IPs preserved. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the placement and the cluster-wide default certificate to resources managed by other Infra Components:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: ingress-namespace
      fieldPath: spec.name
  defaultTlsCertificate:
    secretName:
      valueFrom:
        kind: KubernetesCertificate
        name: wildcard-cert
        fieldPath: status.outputs.secret_name
```

The InfraPipeline deploys the namespace and the certificate first, then installs the controller against them.

## Key Configuration

These are the most important decisions when configuring an ingress-nginx controller. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Several controllers per cluster is the normal pattern** — a public entry point and an internal one is the common split. Each instance gets its own resource and its own DISTINCT `ingressClass.name` — release naming, controller resource names, and leader-election identity all derive from `metadata.name`, so instances never collide. Two controllers sharing a class both program the same Ingresses and fight over them. IngressClasses are immutable: changing the name replaces the class.

**The Service annotations ARE the cloud integration** — the controller itself never calls a cloud API; what the host cloud provisions is driven entirely by `service.annotations`. Which controller READS those annotations decides whether anything happens at all: on EKS the `external` annotation family belongs to the AWS Load Balancer Controller, and without it the Service silently never gets an address. GCP internal placement uses `networking.gke.io/load-balancer-type`, Azure uses `service.beta.kubernetes.io/azure-load-balancer-internal`, AWS uses the scheme annotation.

**Pick the deployment shape for the terrain** — managed cloud runs the default Deployment behind a LoadBalancer Service. Bare metal and edge run `controllerKind: daemon_set` with `hostPorts` or `hostNetwork` and a NodePort Service. `hostNetwork` and `hostPorts` are alternatives — host networking already binds every listener on the node, and the spec rejects the pair.

**An entry point is a single point of failure** — two or more replicas gives zero-drop rollouts and node-failure tolerance, and above one replica the chart adds the PodDisruptionBudget automatically. When `autoscaling` is enabled it OWNS the replica count between min and max, and `replicas` is ignored. Upstream recommends leaving the CPU limit unset: throttling an entry point turns a traffic spike into cluster-wide latency.

**`externalTrafficPolicy: local` is the usual production choice** — it preserves client source IPs and avoids an extra node hop; the default `cluster` policy SNATs sources. Health-check semantics differ between the two — nodes without a local controller pod fail the LB health check by design under `local`.

**Snippet annotations stay off** — `allowSnippetAnnotations` lets any Ingress author inject raw NGINX directives into the shared controller. Upstream disabled them by default after CVE-2021-25742; enable only where every Ingress author is trusted at controller level.

**Keep the admission webhook on** — it rejects a broken Ingress at apply time instead of at NGINX reload time, which is the difference between one developer seeing an error and every service behind the controller going down. The certgen Job is how the webhook bootstraps its certificate, so the only real reason to disable it is a cluster policy forbidding hook Jobs. `failurePolicy: fail` (the default) keeps a broken Ingress from slipping in during webhook downtime.

**Global NGINX tuning rides `nginxConfig` verbatim** — upstream's own ConfigMap key/value vocabulary (proxy-body-size, use-forwarded-headers, ssl-protocols, ...) passed through as-is; the authoritative key list is the upstream ConfigMap reference. `helmValues` merges last for chart surface beyond the typed fields — never the substitute for them, and never a place for secrets.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesCertificate** | `defaultTlsCertificate.secretName` | `status.outputs.secret_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the controller runs in | Locating the controller for diagnostics |
| `release_name` | Helm release name (= metadata.name; controller resources are `<release>-controller`) | Helm management and debugging |
| `ingress_class_name` | The IngressClass this controller owns | A KubernetesIngress's `ingressClassName` |
| `controller_service_name` | The controller's external Service — the traffic entry point | DNS records and exposure composition |
| `internal_service_name` | The second internal Service — populated only when `service.internal.enabled` is true | Private DNS for the internal entry |
| `load_balancer_ip` | The provisioned LB address, on providers exposing an IP (GCP/Azure) | DNS A records |
| `load_balancer_hostname` | The provisioned LB address, on providers exposing a hostname (AWS ELB/NLB) | DNS CNAME records |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public cloud entry** — two replicas behind an AWS NLB with source IPs preserved — the managed-cloud workhorse. Start from the **AWS NLB Public Entry (EKS)** preset.

**Second internal controller** — a separate instance with its own class (e.g. `nginx-internal`) behind a cloud-internal load balancer, so internal APIs never share an entry point with the public site. Start from the **Internal-Only Controller (Second Instance)** preset.

**Bare metal / edge** — a DaemonSet binding host ports 80/443 on every node, no cloud LB anywhere. Start from the **Bare Metal / Edge (DaemonSet + Host Ports)** preset.

**Cluster-wide default TLS** — a cert-manager-issued wildcard certificate served on every HTTPS request that matches no Ingress TLS block. Start from the **Cluster-Wide Default TLS Certificate (cert-manager Composition)** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — where the controller runs.
- [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) — the routing rules this controller serves, selected by class name.
- [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) — supplies the cluster-wide default TLS certificate by reference.
- [**Cert Manager**](/infra-catalog/kubernetes-cert-manager) — issues and renews the certificates behind each Ingress's TLS secret.
- [**Kubernetes Service**](/infra-catalog/kubernetes-service) — the backends Ingresses route to, and the raw TCP/UDP passthrough targets.
- [**Kubernetes PriorityClass**](/infra-catalog/kubernetes-priority-class) — ranks controller pods above ordinary workloads on production clusters.
- [**Metrics Server**](/infra-catalog/kubernetes-metrics-server) — required for the autoscaling utilization targets.
