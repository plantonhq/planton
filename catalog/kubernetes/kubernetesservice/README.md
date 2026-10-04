# Kubernetes Service

## Overview

**KubernetesService** is a catalog kind that creates and manages a standalone Kubernetes Service — the stable network identity in front of a set of pods. A Service gives its backends one durable virtual IP and DNS name while the pods behind it come and go, and it is the unit every other networking construct composes on: Ingress backends point at a Service, NetworkPolicies allow traffic to the pods a Service selects, and sibling workloads connect to its in-cluster DNS name.

The kind covers the complete core/v1 ServiceSpec surface: all four service types, headless services, static cluster IPs, external IPs, traffic policies, topology-aware routing, session affinity, LoadBalancer tuning knobs, dual-stack addressing, and per-port protocol/appProtocol detail. The single deliberate omission is the deprecated `loadBalancerIP` field — upstream deprecated it as under-specified and non-portable; every cloud expresses a pinned LB address through provider-specific annotations instead (set them in `spec.annotations`).

## Purpose

Planton workload kinds (KubernetesDeployment, KubernetesStatefulSet) already create a Service for their own pods, so most applications never need this kind. The standalone Service covers everything else:

- **Exposing pods managed outside Planton** — Helm releases, operators, raw manifests — under a stable, declaratively managed name
- **LoadBalancer / NodePort exposure** with cloud-provider annotations, source-range restrictions, and traffic-policy control
- **ExternalName aliases** — one stable in-cluster DNS name for an endpoint outside the cluster
- **Headless services** for StatefulSet peer discovery and per-pod addressing
- **Dual-stack addressing** — explicit IPv4/IPv6 family selection and ordering
- **Selectorless services** fronting manually-managed endpoints or external controllers

**Key value over raw manifests:**

- **API-server rules at validation time**: Eleven cross-field rules (headless vs. NodePort/LoadBalancer, ExternalName vs. selector/ports, LoadBalancer-only knobs on other types, affinity timeout without ClientIP affinity, dual-stack consistency, and more) each mirror a live kube-apiserver rejection — caught before anything reaches the cluster
- **Namespace by value or reference**: `spec.namespace` accepts a literal name or a reference to a `KubernetesNamespace` resource, so an Infra Chart creates the namespace and the Service in one run
- **Dual IaC support**: Both Pulumi and Terraform implementations with feature parity (one documented exception: `traffic_distribution` deploys only through the Pulumi engine — the Terraform kubernetes provider does not expose the field, and the Terraform module fails the plan loudly rather than silently dropping it)
- **Composable outputs**: The load-balancer address, in-cluster endpoint, and a ready-to-run port-forward command are exported for downstream automation

## Relationship to Other Components

- **Workload components** (KubernetesDeployment, KubernetesStatefulSet): Already ship with their own Service — use this kind only for additional or differently-shaped exposure. Planton workloads stamp `app: <workload-metadata-name>` on their pods, so selecting one is a single selector entry: `selector: {app: <workload-metadata-name>}`
- **KubernetesNamespace**: Provides the target namespace. Reference it from `spec.namespace` to deploy both in one chart
- **Ingress / Gateway components**: Consume the Service by name as their backend; the created name and namespace are exported as outputs for exactly this composition

## Service Types

- **`cluster_ip`** (default) — a cluster-internal virtual IP, reachable only from inside the cluster. The most common type
- **`node_port`** — builds on ClusterIP and opens a static port (30000–32767) on every node; reachable at `<NodeIP>:<NodePort>`
- **`load_balancer`** — builds on NodePort and asks the cloud provider to provision an external load balancer. Tuned through `spec.annotations` (NLB vs. ELB, internal vs. external, pinned addresses, external-dns records) plus the LoadBalancer-only spec knobs (`load_balancer_source_ranges`, `load_balancer_class`, `allocate_load_balancer_node_ports`, `health_check_node_port`)
- **`external_name`** — a pure DNS alias: cluster DNS returns a CNAME to `spec.external_dns_name`. No proxying, no selectors, no ports

Orthogonal to the type, **`headless: true`** creates a service with no virtual IP (`clusterIP: None`): DNS returns the pod IPs directly, and each StatefulSet pod gets its own stable name. Incompatible with `node_port`/`load_balancer` and with a static `cluster_ip_address`.

## Traffic Shaping

- **`external_traffic_policy`** (`cluster` | `local`) — for NodePort/LoadBalancer traffic. `local` preserves the client source IP and skips a hop, but nodes without endpoints drop traffic — pair with the load balancer's health check (`health_check_node_port`)
- **`internal_traffic_policy`** (`internal_cluster` | `internal_local`) — for ClusterIP traffic; `internal_local` serves node-local agent patterns (DNS caches, log collectors)
- **`traffic_distribution`** (`prefer_same_zone` | `prefer_same_node`) — a topology routing hint that cuts cross-zone transfer cost and latency. **Pulumi engine only** (see above)
- **`session_affinity: client_ip`** plus `session_affinity_timeout_seconds` — pins a client IP to one backend pod for the timeout window (Kubernetes default 10800s)

## Essential Configuration Fields

### Required

- **`spec.name`**: The Service name (DNS-1035 label: lowercase alphanumeric and hyphens, must start with a letter, max 63 chars). This becomes the service's DNS name
- **`spec.ports`**: At least one port for every type except `external_name`. Each port carries `port` (what clients connect to), optional `target_port` (number or named container port; defaults to `port`), `protocol` (TCP/UDP/SCTP), optional `app_protocol` (L7 hint), and optional `node_port`

### Common

- **`spec.namespace`**: Literal name or KubernetesNamespace reference; defaults to `default`
- **`spec.type`**: One of the four types above; defaults to `cluster_ip`
- **`spec.selector`**: Label selector for the backend pods; leave empty for ExternalName or selectorless services
- **`spec.headless`**: Headless service toggle
- **`spec.external_dns_name`**: CNAME target — required for, and only allowed on, `external_name`
- **`spec.publish_not_ready_addresses`**: Publish pod addresses before Ready — the StatefulSet-bootstrap pattern
- **`spec.ip_families`** / **`spec.ip_family_policy`**: Dual-stack family selection and requirement level
- **`spec.labels`** / **`spec.annotations`**: Merged with standard Planton labels; annotations carry the cloud LB configuration

## Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`service_name`**: The name of the Service object as created in the cluster
- **`namespace`**: The namespace the Service was created in
- **`type`**: The service type as deployed (ClusterIP, NodePort, LoadBalancer, ExternalName)
- **`cluster_ip`**: The cluster-internal virtual IP — empty for headless services (no virtual IP) and ExternalName services (DNS aliases)
- **`load_balancer_ip`**: The provisioned load balancer's IP — populated only on providers that expose an IP (GCP, Azure, MetalLB)
- **`load_balancer_hostname`**: The provisioned load balancer's DNS hostname — populated only on providers that expose a hostname (AWS ELB/NLB)
- **`kube_endpoint`**: In-cluster DNS endpoint, e.g. `my-app.my-ns.svc.cluster.local`
- **`port_forward_command`**: Ready-to-run `kubectl port-forward` command for reaching the Service from a developer machine; empty for ExternalName services

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Resolve the target namespace (literal value or resolved reference)
2. Merge user labels and annotations with standard Planton tracking labels
3. Translate spec enums to Kubernetes API strings (`load_balancer` → `LoadBalancer`, `client_ip` → `ClientIP`, ...)
4. Create the Service, sending each optional field only when the user set it — several fields (clusterIP, healthCheckNodePort, loadBalancerClass) are immutable or type-gated, and an empty value is not the same as an omitted one
5. Export the outputs above for downstream composition

Both implementations follow identical logic and export the identical output set, with the single `traffic_distribution` exception noted above.

## When to Use

Use **KubernetesService** when you need:

- A stable name in front of pods that Planton does not manage
- Internet or node-level exposure (LoadBalancer/NodePort) with cloud annotations and traffic policies
- A DNS alias to an endpoint outside the cluster (ExternalName)
- Headless per-pod discovery for StatefulSets and quorum systems
- Dual-stack or selectorless services

**Do NOT use** when:

- You are exposing a Planton-managed workload's pods in the ordinary way — the workload kind already creates that Service
- You need L7 routing (paths, hostnames, TLS termination) — that is an Ingress/Gateway concern; the Service is its backend

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster (any distribution: GKE, EKS, AKS, self-hosted)
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **Namespace**: The target namespace must exist before creating the Service (unless deploying to `default`, or creating the namespace in the same chart via a reference)
- **For LoadBalancer**: a cloud provider (or MetalLB-style implementation) able to provision load balancers

## Best Practices

1. **Default to `cluster_ip`**: Expose externally only when required, and prefer one Ingress over many LoadBalancers for HTTP workloads (each LoadBalancer service is a billed provider resource)
2. **Name every port on multi-port services**: Required by the API, and named ports (`http`, `grpc`) keep consumers readable
3. **Use `external_traffic_policy: local` when the client IP matters**: And run enough replicas that every schedulable node is likely to hold one
4. **Reserve `publish_not_ready_addresses` for bootstrap discovery**: On ordinary services it routes real traffic to pods that cannot handle it
5. **Pin LB behavior in annotations, not imperative tweaks**: The annotation set is the reproducible record of how the load balancer is configured

## References

- [Kubernetes Service Documentation](https://kubernetes.io/docs/concepts/services-networking/service/)
- [Service API Reference](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/service-v1/)
- [DNS for Services and Pods](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/)
- [Dual-stack Services](https://kubernetes.io/docs/concepts/services-networking/dual-stack/)
- [Traffic Distribution](https://kubernetes.io/docs/concepts/services-networking/service/#traffic-distribution)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
