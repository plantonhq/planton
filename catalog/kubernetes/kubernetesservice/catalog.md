# Kubernetes Service

Deploys a standalone Kubernetes Service — the stable network identity in front of a set of pods, giving its backends one durable name and address while the pods behind it come and go. Workload kinds create a Service for their own pods already; this kind covers everything else: exposing pods managed outside Planton, LoadBalancer and NodePort exposure with cloud annotations, ExternalName aliases to hosts outside the cluster, headless services for custom discovery, dual-stack addressing, and selectorless services fronting manually-managed endpoints. It is also the unit every other networking construct composes on — Ingress backends, Gateway API route backends, DestinationRule hosts, and BackendTLSPolicy targets all point at a Service.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Service** — one core/v1 Service in the specified namespace, with the declared type (ClusterIP, NodePort, LoadBalancer, or ExternalName), selector, ports, traffic policies, and annotations. The spec covers the complete core/v1 ServiceSpec surface with one deliberate omission: the deprecated `loadBalancerIP` field (upstream deprecated it as under-specified; every cloud pins an address through its own annotation instead).

For LoadBalancer services, the cloud's controller then provisions the external load balancer and writes its address back to the Service — surfaced through the `load_balancer_ip` / `load_balancer_hostname` outputs.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **The target namespace exists** — the module does not create it; reference a **Kubernetes Namespace** or type the name directly. Omitted means the cluster's `default` namespace.
- **The right load-balancer controller, for LoadBalancer types** — which controller answers the annotations matters (verified live on EKS): `aws-load-balancer-type: "nlb"` is handled by EKS's built-in cloud controller, while the `aws-load-balancer-type: "external"` annotation family is handled ONLY by the AWS Load Balancer Controller. With those set and that controller absent, the Service never receives an address and no error surfaces anywhere.
- **Pods carrying the selector labels** — the Service routes to pods matching ALL selector pairs; nothing checks that such pods exist.

## Deploy

### Console

Open the deployment store, find **Kubernetes Service**, and click **Deploy**. The creation wizard walks you through the namespace and name, the type, the selector, the ports, traffic policies, and exposure tuning. Start from the **ClusterIP for an Application** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesService
metadata:
  name: orders-api
  org: acme-corp
  env: prod
spec:
  name: orders-api
  namespace:
    value: backend-services
  selector:
    app: orders-api
  ports:
    - name: http
      port: 80
      targetPort: "8080"
```

```shell
planton apply -f service.yaml
```

This creates a ClusterIP Service in `backend-services` routing port 80 to port 8080 on every pod labeled `app: orders-api`, reachable in-cluster at `orders-api.backend-services.svc.cluster.local`. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the Service to its Planton-managed namespace:

```yaml
spec:
  name: orders-api
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: backend-namespace
      fieldPath: spec.name
  selector:
    app: orders-api
  ports:
    - name: http
      port: 80
```

The InfraPipeline deploys the namespace first, then creates the Service inside it.

## Key Configuration

These are the most important decisions when configuring a Service. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The type ladder — and the one that isn't on it** — ClusterIP (the default) is in-cluster only; NodePort builds on it with a static port on every node; LoadBalancer builds on NodePort with a cloud-provisioned external address. ExternalName is a different mechanism entirely, not a fourth exposure level: cluster DNS returns a CNAME to `externalDnsName` and no traffic passes through Kubernetes — selectors, cluster IPs, external IPs, and ports are all rejected on it.

**The selector is the contract, and its failure mode is silent** — traffic reaches a pod only if that pod carries EVERY label in the selector. A selector matching nothing produces a healthy-looking Service with zero endpoints: DNS resolves and connections hang. Planton workloads stamp `app: <workload-name>` on their pods and export the full set as their `selector_labels` output, so that single pair is enough to select one. An empty selector is a real posture, not an omission — a selectorless Service whose EndpointSlices are managed by hand or by a controller.

**Headless is a modifier, not a type** — `headless: true` removes the virtual IP so DNS returns pod addresses directly: the tool for StatefulSet peer discovery and clients that must reach each pod individually. It cannot combine with NodePort or LoadBalancer (they need a virtual IP to build on) or with a static `clusterIpAddress`. Pair it with `publishNotReadyAddresses: true` only for a StatefulSet's governing Service, where peers must find each other DURING startup — on an ordinary traffic-serving Service that flag sends real traffic to pods that cannot handle it.

**Name target ports, don't number them** — a `targetPort` naming a container port ("http") survives the workload changing its listening port; a numbered one does not. Once a Service exposes more than one port, every port needs a unique name.

**Cloud behavior comes from annotations, and the reader matters** — the annotation set is the portable way to tune a provisioned balancer: AWS NLB selection, internal placement on GCP and Azure, external-dns hostnames. An annotation with no controller reading it is silently inert — the EKS `external` family without the AWS Load Balancer Controller is the classic case: the Service just never gets an address.

**`externalTrafficPolicy: local` trades balance for the real client IP** — Cluster (the default) spreads load evenly but masquerades the source IP and can add a hop; Local preserves the client IP and removes the hop, but nodes without endpoints drop traffic — external balancers must probe the health-check NodePort to learn which nodes hold endpoints. `healthCheckNodePort` is settable only in this combination, and immutable once set.

**Immutable-after-create fields** — `clusterIpAddress` (a static in-CIDR virtual IP) and `loadBalancerClass` (which LB implementation provisions the Service when the cluster runs more than one, e.g. MetalLB beside the cloud default) both lock at creation. Choose them deliberately or leave them empty.

**`loadBalancerSourceRanges` is a courtesy, not a firewall** — it restricts client CIDRs only on providers that enforce it. Treat it as defense-in-depth over a NetworkPolicy, never as the only gate.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

The selector targets pods by label, not by resource reference — workloads and the Services selecting them carry no deploy-ordering edge.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `service_name` | The Service object's name in the cluster | Ingress backends and Gateway API route `backendRefs` |
| `kube_endpoint` | The in-cluster DNS endpoint (`<name>.<namespace>.svc.cluster.local`) | Sibling workloads' connection strings |
| `cluster_ip` | The assigned virtual IP — empty for headless and ExternalName | Diagnostics, static egress rules |
| `load_balancer_ip` | The provisioned address, on providers exposing an IP (GCP, Azure, MetalLB) | DNS A records |
| `load_balancer_hostname` | The provisioned address, on providers exposing a hostname (AWS ELB/NLB) | DNS CNAME records |
| `port_forward_command` | A ready-to-run `kubectl port-forward` command | Reaching the Service from a laptop with no exposure |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Front an application with a stable name** — a ClusterIP Service selecting the workload's `app` label, named ports mapping 80 to the container port. Start from the **ClusterIP for an Application** preset.

**Public exposure without an ingress layer** — a LoadBalancer Service with the cloud's annotations (NLB mode, source-range restrictions, an external-dns hostname); the right choice for non-HTTP protocols where an L7 layer adds nothing. Start from the **Public Load Balancer** preset.

**StatefulSet peer discovery** — a headless Service with `publishNotReadyAddresses: true` so database peers can bootstrap a quorum before reporting Ready. Start from the **Headless Service for StatefulSet Peers** preset.

**Alias an external dependency** — an ExternalName Service giving an out-of-cluster host (a managed database, a legacy API) an in-cluster DNS name, so consumers switch backends by editing one resource. Start from the **ExternalName Alias** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — placement (optional — omitted means the `default` namespace).
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) and [**Kubernetes StatefulSet**](/infra-catalog/kubernetes-stateful-set) — the workloads whose pods it selects.
- [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) — routes HTTP traffic to this Service.
- [**Kubernetes HTTPRoute**](/infra-catalog/kubernetes-http-route) and [**Kubernetes GRPCRoute**](/infra-catalog/kubernetes-grpc-route) — Gateway API backends pointing here.
- [**Kubernetes BackendTLSPolicy**](/infra-catalog/kubernetes-backend-tls-policy) — secures the gateway-to-Service hop.
- [**Istio Destination Rule**](/infra-catalog/kubernetes-destination-rule) — applies mesh traffic policy to this host.
- [**Kubernetes NetworkPolicy**](/infra-catalog/kubernetes-network-policy) — governs which pods may reach the selected pods.
