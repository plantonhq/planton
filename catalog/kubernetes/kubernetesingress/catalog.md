# Kubernetes Ingress

Creates a namespaced Kubernetes `networking/v1` Ingress -- the object that declares HTTP(S) exposure for in-cluster Services: host rules and path matches routing to Service backends, with optional TLS termination from certificate Secrets. Exposure in Planton is composed, never embedded: a workload exports its Service, this kind routes a hostname to it, and a certificate Secret (often issued by cert-manager) terminates TLS -- every piece a visible, independently managed node in the resource graph. This kind covers the complete `networking/v1` IngressSpec surface with proto validation, typed SDKs, and InfraChart composability.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A namespaced Ingress** named `spec.name` in `spec.namespace`, carrying the host/path rules, the optional default backend, the TLS blocks, and your annotations (merged with the standard governance labels).
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking.

Both IaC modules deliberately create the object **without waiting for an ingress controller** (Terraform sets `wait_for_load_balancer = false`; Pulumi sets the `pulumi.com/skipAwait` annotation). An Ingress is inert until a controller claims it -- creating it first is valid, and its load-balancer address fills in once a controller reconciles it.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **An ingress controller** -- ingress-nginx (deployable as `KubernetesIngressNginx`), a cloud L7 controller (ALB on EKS, GCE on GKE), or any other. Without one the Ingress is accepted but never served; `kubectl get ingressclass` lists what the cluster offers.
- **The namespace and backend Services exist** -- the Ingress can only route to Services in its own namespace (a Kubernetes API constraint). When `spec.namespace` is omitted, the Ingress lands in the cluster's `default` namespace.
- **For TLS with cert-manager** -- cert-manager installed (deployable as `KubernetesCertManager`) with a `KubernetesClusterIssuer`; the certificate Secret named in the `tls` block need not exist yet.

## Deploy

### Console

Open the deployment store, find **Kubernetes Ingress**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and six spec steps: **Namespace** (optional -- must be the backend Services' namespace), **Name & Controller** (the in-cluster object name and the IngressClass), **Routing Rules** (hosts, paths, and backends), **Default Backend** (the catch-all, and the single-Service mode), **TLS** (certificate Secrets per host set), and **Labels & Annotations** (the controller-behavior contract). Start from the **Single Host** or **TLS with cert-manager** preset in the [Presets](#presets) tab for a directly deployable configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesIngress
metadata:
  name: web-ingress
  org: acme-corp
  env: prod
spec:
  namespace:
    value: payments
  name: web-app
  ingressClassName: nginx
  rules:
    - host: app.example.com
      paths:
        - path: /
          backend:
            serviceName:
              value: web-svc
            portNumber: 80
  tls:
    - hosts:
        - app.example.com
      secretName:
        value: app-tls
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
```

```shell
planton apply -f ingress.yaml
```

This creates an Ingress in `payments` served by the `nginx` controller: requests for `app.example.com` route to the `web-svc` Service on port 80 over HTTPS, with cert-manager issuing the certificate into the `app-tls` Secret. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the backend Service by reference — the classic "deploy the app, then expose it" composition without copying names by hand:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: payments-namespace
      fieldPath: spec.name
  name: web-app
  ingressClassName: nginx
  rules:
    - host: app.example.com
      paths:
        - path: /
          backend:
            serviceName:
              valueFrom:
                kind: KubernetesService
                name: web-service
                fieldPath: status.outputs.service_name
            portNumber: 80
```

The InfraPipeline deploys the namespace and the Service first, then creates the Ingress against them.

## Key Configuration

These are the most important decisions when configuring an Ingress. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Namespace** -- Where the Ingress lives, and therefore where ALL its backends must live: Ingress backends cannot cross namespaces. Optional -- omitted means the cluster's `default` namespace. Reference an existing `KubernetesNamespace` or type the name directly.

**Ingress class** -- Which controller serves this Ingress (`nginx`, `alb`, `gce`, ...). A plain name, not a reference: IngressClass objects ship with their controllers. Omitted, the cluster's default class applies -- and on clusters without a default, an unclassed Ingress is served by nothing.

**Rules** -- The routing table: a request is matched first on its Host header (precise or single-label wildcard like `*.example.com`), then on the rule's HTTP paths (`prefix` matching per path element by default; `exact` byte-for-byte; `implementationSpecific` delegated to the controller). Each path forwards to one backend Service, by port number or -- surviving port renumbering -- by port name. Validation enforces exactly one of the two, and at least one rule or a default backend.

**Default backend** -- The catch-all for unmatched requests, and the whole configuration for single-Service exposure (default backend set, zero rules). Left unset, most controllers fall back to their own global default.

**TLS** -- Each entry terminates HTTPS for a set of hosts under one `kubernetes.io/tls` certificate Secret, multiplexed via SNI. The Secret is a name reference; with cert-manager it is created for you, named exactly as written.

**Annotations** -- The upstream contract for controller-specific behavior: rewrites (`nginx.ingress.kubernetes.io/rewrite-target`), body-size limits, SSL redirects -- plus the composition hooks `cert-manager.io/cluster-issuer` and `external-dns.alpha.kubernetes.io/hostname`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesService** | `rules[].paths[].backend.serviceName` (and `defaultBackend.serviceName`) | `status.outputs.service_name` |
| **KubernetesSecret** | `tls[].secretName` | `status.outputs.secret_name` |

All three also accept plain literals for resources managed outside Planton.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `ingress_name` | Name of the created Ingress object | Ordering and reference in InfraCharts |
| `namespace` | The namespace the Ingress was created in | Co-locating companion resources |
| `load_balancer_ip` | The controller's load-balancer IP (IP-based controllers) | DNS A records; empty until a controller reconciles |
| `load_balancer_hostname` | The controller's load-balancer hostname (AWS ALB/ELB-class) | DNS CNAME records; empty until a controller reconciles |
| `first_host` | The first host declared in the rules -- the primary public FQDN | Dashboards, smoke tests, DNS automation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single host** -- One hostname, one Service, everything under `/`. Start from the **Single Host** preset.

**HTTPS with cert-manager** -- One host served over TLS with the certificate issued and renewed automatically. Start from the **TLS with cert-manager** preset.

**Path fan-out** -- One domain, several Services: `/` to the frontend, `/api` to the API. Start from the **Path Fan-Out** preset.

**Catch-all** -- No rules, just a default backend: every request the controller routes here reaches one Service. Start from the **Default Backend Only** preset.

## Works With

- [**Ingress NGINX**](/infra-catalog/kubernetes-ingress-nginx) -- installs the ingress-nginx controller (and its `nginx` IngressClass) that serves this Ingress.
- [**Kubernetes Service**](/infra-catalog/kubernetes-service) -- the backend Services (`rules` / `defaultBackend`) receiving the routed traffic; workload kinds export the Service this Ingress references.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the namespace (`spec.namespace`) the Ingress and its backends run in.
- [**Cert Manager**](/infra-catalog/kubernetes-cert-manager) and [**Cert Manager Cluster Issuer**](/infra-catalog/kubernetes-cluster-issuer) -- issue and renew the TLS certificates into the Secrets the `tls` block names.
- [**ExternalDNS**](/infra-catalog/kubernetes-external-dns) -- creates DNS records pointing at this Ingress's load-balancer address automatically.
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) -- pre-existing certificate Secrets referenced by `tls[].secretName`.
