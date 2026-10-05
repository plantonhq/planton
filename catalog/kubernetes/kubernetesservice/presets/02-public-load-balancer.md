# Public Load Balancer

This preset asks the cloud provider to provision an external load balancer in front of the selected pods. Once deployed, the provider's address lands in the outputs — `load_balancer_ip` on IP-based providers (GCP, Azure, MetalLB) or `load_balancer_hostname` on hostname-based ones (AWS ELB/NLB).

## When to Use

- Exposing a TCP/UDP workload directly to the internet without an Ingress layer
- Non-HTTP protocols (databases, game servers, MQTT) where an L7 Ingress does not apply
- When you need the real client source IP to reach the application

## Key Configuration Choices

- **`annotations`** — the portable way to tune the provisioned load balancer; each cloud reads its own keys. The example pair requests an AWS NLB and an external-dns DNS record. Swap for your provider's annotations (e.g. `cloud.google.com/load-balancer-type: "Internal"` on GCP, `service.beta.kubernetes.io/azure-load-balancer-internal: "true"` on Azure)
- **`external_traffic_policy: local`** — routes external traffic only to pods on the receiving node, which preserves the client source IP and skips a second hop. The trade-off: nodes without a pod fail the load balancer's health check and receive no traffic, so run enough replicas to spread across nodes. Drop this line for the default `cluster` policy (even spreading, masqueraded source IP)
- **Two named ports** — with multiple ports, every port needs a unique name; clients hit 443/80 while containers listen on 8443/8080
- **Locking down sources** — add `load_balancer_source_ranges` with CIDRs to restrict who can reach the load balancer (enforced by providers that support it)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Namespace the selected pods run in | Your namespace management |
| `<your-app-name>` | Value of the `app` label on the pods to expose | `kubectl get pods --show-labels`, or the workload manifest |
| `app.example.com` | Public hostname for the external-dns record (remove the annotation if you don't run external-dns) | Your DNS zone |

## Related Presets

- **01-cluster-ip-app** — internal-only exposure of the same pods
- **04-external-name** — the inverse direction: aliasing an external endpoint into the cluster
