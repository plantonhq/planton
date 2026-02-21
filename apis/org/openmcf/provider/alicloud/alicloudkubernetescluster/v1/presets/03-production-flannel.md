# Production ACK Cluster with Flannel Networking

This preset creates a production-grade ACK Managed Kubernetes cluster using Flannel overlay networking. It provides the same security and observability posture as the Terway production preset (ack.pro.small, RRSA, audit logging, deletion protection, maintenance windows) but uses a simpler overlay network that does not require dedicated pod VSwitches or per-pod ENI allocation.

## When to Use

- Production workloads where pod-level VPC security group isolation is not required
- Clusters with moderate pod counts where overlay networking overhead is acceptable
- Teams that want simpler VPC infrastructure (no dedicated pod VSwitches to manage)
- Environments where the pod CIDR is separate from the VPC CIDR and overlay routing is preferred

## Key Configuration Choices

- **Flannel overlay** (`flannel` addon + `podCidr: "172.20.0.0/16"`) -- Pods get cluster-internal IPs routed via VXLAN tunnels. Simpler VPC setup than Terway (no pod VSwitches needed) at the cost of losing VPC-native pod IPs and per-pod security group support. The /16 CIDR provides ~65,000 pod IPs.
- **ack.pro.small** (`clusterSpec: ack.pro.small`) -- Same professional tier as the Terway preset. Enhanced SLA, managed node pools, and topology-aware scheduling.
- **Three availability zones** -- Production-grade multi-AZ resilience for both the control plane and worker nodes.
- **RRSA enabled** (`enableRrsa: true`) -- Pod-level IAM remains available regardless of CNI choice. Service accounts can assume RAM roles via OIDC.
- **Full observability** (logtail-ds, arms-prometheus, metrics-server, ack-node-problem-detector) -- Same logging and monitoring stack as the Terway preset with 90-day control plane log retention and audit logging.
- **Patch-only auto-upgrade** (`channel: patch`) -- Automatic patch upgrades during the Wednesday maintenance window without minor version jumps.
- **External NAT** (`newNatGateway: false`) -- Same assumption as the Terway preset: NAT is managed via a dedicated AlicloudNatGateway component.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region code (e.g., `cn-hangzhou`, `ap-southeast-1`) | Your deployment region strategy |
| `<your-cluster-name>` | Cluster name (1-63 chars, alphanumeric) | Your naming convention |
| `<vswitch-id-zone-a>` | VSwitch in first availability zone | `AlicloudVswitch` stack outputs |
| `<vswitch-id-zone-b>` | VSwitch in second availability zone | `AlicloudVswitch` stack outputs |
| `<vswitch-id-zone-c>` | VSwitch in third availability zone | `AlicloudVswitch` stack outputs |
| `<your-log-project-name>` | SLS project for cluster logs and addon dashboards | `AlicloudLogProject` stack outputs |
| `<your-team>` | Team or business unit | Your organizational structure |
| `<your-cost-center>` | Cost center code | Your finance team |

## Related Presets

- **01-production-terway** -- Use instead when pods need VPC-native IP addresses, per-pod security group isolation, or you are running large-scale clusters where overlay overhead matters
- **02-development-flannel** -- Use for non-production environments where cost and simplicity are the priority
