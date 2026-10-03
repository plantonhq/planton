# Kubernetes Ingress

## Overview

**KubernetesIngress** is a catalog kind that creates and manages Kubernetes `networking/v1` Ingress objects as first-class, declaratively managed resources. An Ingress declares HTTP(S) exposure for in-cluster Services: host rules and path matches routing to Service backends, with optional TLS termination from certificate Secrets.

The kind covers the complete `networking/v1` IngressSpec surface — ingress class selection, a default backend, TLS blocks, and host/path rules with all three path types. The single deliberate omission is the `resource` backend variant (an ObjectRef to an arbitrary same-namespace object): it is controller-specific and rarely implemented, and Service backends cover the real exposure paths.

## Purpose

Exposure in Planton is **composed, never embedded**. A workload kind (KubernetesDeployment and friends) exports its Service name as a `service` output; this kind routes a hostname to that Service; a certificate comes from cert-manager or a KubernetesSecret that the `tls` block names. Every piece of the exposure path is a visible, independently managed node in the resource graph — no workload kind carries a hidden, half-featured ingress of its own.

**Key value over raw manifests:**

- **Schema-level validation**: DNS-subdomain names, precise-or-wildcard host syntax, "exactly one of port number/name" per backend, "path required for prefix/exact", and "at least one rule or a default backend" — all caught before anything reaches the cluster
- **Backends by value or reference**: `service_name` accepts a literal Service name or a reference to a KubernetesService resource, so a chart can deploy a workload and wire its exported `service` output into the Ingress without copying names by hand
- **TLS Secrets by value or reference**: `secret_name` accepts a literal name or a KubernetesSecret reference; with cert-manager the Secret need not exist yet
- **Dual IaC support**: Both Pulumi and Terraform implementations with feature parity
- **Non-blocking creation**: Deploys never hang waiting for an ingress controller (see below)

## The Controller Relationship

An Ingress object is **inert until an ingress controller** (ingress-nginx, AWS ALB, GCE, Traefik, ...) runs in the cluster and claims it — via `ingress_class_name` or the cluster's default IngressClass. Creating the Ingress before the controller exists is valid; its load-balancer status simply stays empty until a controller reconciles it.

Both IaC modules deliberately create the object **without waiting for a controller** (Terraform sets `wait_for_load_balancer = false`; Pulumi sets the `pulumi.com/skipAwait` annotation — the exact same choice). Infra charts routinely deploy the workload and its exposure before the ingress controller wave, and blocking every deploy until a controller populates the status would couple this kind to cluster addon ordering. The consequence: the `load_balancer_ip` / `load_balancer_hostname` outputs export empty on clusters where no controller has reconciled the object yet, and fill in once one has.

IngressClass objects ship with their controllers (`kubectl get ingressclass` lists what the cluster offers), which is why `ingress_class_name` is a plain name rather than a reference to a Planton kind.

## Relationship to Other Components

- **Workload components** (KubernetesDeployment and friends): Export a `service` output that backends here route to — deploy the app, then expose it, composed in one chart
- **KubernetesService**: The default reference kind for `service_name`; a backend can point at a managed Service directly
- **KubernetesSecret**: The default reference kind for `tls[].secret_name` — a `kubernetes.io/tls` Secret holding the certificate and key
- **cert-manager** (cluster addon): The alternative certificate path — add a `cert-manager.io/cluster-issuer` annotation and cert-manager creates the Secret named in the `tls` block
- **KubernetesNamespace**: Provides the target namespace; reference it from `spec.namespace` to deploy both in one chart

## Routing Model

### Rules and paths

A request is matched first on its Host header against `rules[].host`, then on the rule's HTTP paths. Hosts are either precise (`app.example.com`) or wildcard with a single leading `*` label (`*.example.com` matches `a.example.com` but not `b.a.example.com` or `example.com`). Omitting the host matches every host reaching the controller.

Each path routes to one Service backend with exactly one of `port_number` or `port_name`. Prefer port names when the Service defines them — the reference survives port-number changes.

### Path types

- **`prefix`** (the default): match on URL path prefix, split by `/` per element — `/api` matches `/api` and `/api/users` but not `/apiary`. The longest matching path wins. The one type every controller must implement identically
- **`exact`**: match the URL path exactly, case-sensitively
- **`implementation_specific`**: semantics delegated to the IngressClass (ingress-nginx treats these as regex candidates). Non-portable across controllers by definition

### Default backend

`default_backend` handles requests no rule matches — and is also the way to expose a single Service on all traffic the controller routes here (set only `default_backend`, omit `rules`). Either `rules` or `default_backend` must be present; an Ingress with neither routes nothing, so validation rejects it.

### Same-namespace constraint

Ingress backends can only reference Services in the Ingress's **own namespace** — a Kubernetes API constraint, not a Planton one. `spec.namespace` must be the namespace of the backend Services.

## TLS

Each `tls` entry names the hosts served under one certificate Secret; the controller multiplexes multiple entries on port 443 via SNI. The Secret is a `kubernetes.io/tls` Secret in the Ingress's namespace. With cert-manager, leave the Secret non-existent and add the issuer annotation — cert-manager creates it under exactly the name written in `secret_name`. Omitting `secret_name` asks the controller to serve those hosts with its default (or separately provisioned) certificate.

## Annotations

Controller-specific behavior goes through `annotations` — the upstream contract. Common ingress-nginx examples:

- `nginx.ingress.kubernetes.io/rewrite-target: /$2` — path rewriting
- `nginx.ingress.kubernetes.io/proxy-body-size: 50m` — upload size limit
- `nginx.ingress.kubernetes.io/ssl-redirect: "false"` — disable HTTPS redirect
- `cert-manager.io/cluster-issuer: letsencrypt-prod` — cert-manager issues the TLS certificate
- `external-dns.alpha.kubernetes.io/hostname: app.example.com` — DNS record automation

## Essential Configuration Fields

### Required

- **`spec.name`**: The Ingress name (DNS subdomain, 1–253 chars)
- **`spec.rules` or `spec.default_backend`**: At least one of the two

### Common

- **`spec.namespace`**: Literal name or KubernetesNamespace reference; must match the backend Services' namespace. Defaults to `default` when omitted
- **`spec.ingress_class_name`**: Selects the serving controller (`nginx`, `alb`, `gce`, ...). Omit to use the cluster's default class; on clusters without a default, a class-less Ingress is not served by any controller
- **`spec.tls`**: TLS termination blocks
- **`spec.labels`** / **`spec.annotations`**: Merged with standard Planton governance labels

## Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`ingress_name`**: The name of the Ingress object as created in the cluster
- **`namespace`**: The namespace the Ingress was created in
- **`load_balancer_ip`**: The IP the controller's load balancer exposes (IP-based controllers: GCE, ingress-nginx on most clouds); empty until a controller reconciles the Ingress
- **`load_balancer_hostname`**: The DNS hostname the controller's load balancer exposes (hostname-based controllers: AWS ALB/ELB); empty until a controller reconciles the Ingress
- **`first_host`**: The first host declared in the rules — the primary public FQDN, ready for downstream references (DNS records, dashboards, smoke tests). Empty for host-less catch-all rules or default-backend-only Ingresses

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Resolve the target namespace and backend/Secret references (literal values or resolved references)
2. Merge user labels and annotations with standard Planton tracking labels
3. Create the `networking/v1` Ingress with class, default backend, TLS blocks, and rules — without waiting for a controller
4. Export the name, namespace, load-balancer handles, and first host for downstream composition

Both IaC implementations have feature parity and follow identical logic.

## When to Use

Use **KubernetesIngress** when you need:

- HTTP(S) exposure of in-cluster Services under stable hostnames
- Path-based fan-out — one host, several backends (`/` to the frontend, `/api` to the backend)
- TLS termination at the edge, with cert-manager or pre-provisioned Secrets
- Exposure composed with workloads in one chart, wired through resource references

**Do NOT use** when:

- No ingress controller runs (or will run) in the cluster — the object would be valid but permanently unserved; use a `LoadBalancer`/`NodePort` Service instead
- You need non-HTTP protocols (raw TCP/UDP, gRPC-specific routing beyond what your controller's annotations offer) — that is Service or Gateway API territory
- You need a non-Service backend (the `resource` variant) — deliberately not modeled here

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster (any distribution: GKE, EKS, AKS, self-hosted)
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **Backend Services**: Must exist (or deploy in the same chart) in the same namespace as the Ingress
- **Ingress controller**: Required for the Ingress to serve traffic — but NOT required at creation time; the object waits inertly until one claims it

## Best Practices

1. **Pin the class explicitly**: Set `ingress_class_name` rather than relying on a cluster default — portable across clusters and explicit in review
2. **Prefer named ports**: `port_name: http` survives Service port-number changes; numbers do not
3. **Prefer `prefix` paths**: The one path type with identical semantics on every controller; reserve `implementation_specific` for controller features you consciously depend on
4. **Let cert-manager own TLS Secrets**: Name the Secret in `tls`, annotate with the issuer, and never handle certificate material by hand
5. **Keep backends in the Ingress's namespace**: The API constraint is absolute — plan namespace placement around it
6. **Compose DNS on the outputs**: Point records at `load_balancer_ip`/`load_balancer_hostname`, or let external-dns read the same status

## References

- [Kubernetes Ingress Documentation](https://kubernetes.io/docs/concepts/services-networking/ingress/)
- [Ingress Controllers](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/)
- [IngressClass](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class)
- [Ingress API Reference](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/ingress-v1/)
- [cert-manager: Securing Ingress Resources](https://cert-manager.io/docs/usage/ingress/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
