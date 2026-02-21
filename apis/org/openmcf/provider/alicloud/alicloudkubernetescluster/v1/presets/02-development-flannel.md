# Development ACK Cluster with Flannel Networking

This preset creates a minimal ACK Managed Kubernetes cluster for development and testing. It uses the free ack.standard tier with Flannel overlay networking, two availability zones, and only the essential addons (CNI, CSI). No logging, RRSA, deletion protection, or maintenance windows are configured, keeping the cluster simple and cost-free on the control plane side.

## When to Use

- Development, testing, and learning environments
- Proof-of-concept deployments where cost and simplicity are priorities
- Clusters that do not need per-pod VPC-native IP addresses or ENI isolation
- Environments where audit logging and RRSA are unnecessary

## Key Configuration Choices

- **ack.standard** (`clusterSpec: ack.standard`) -- Free managed control plane with basic SLA. No management fee; only worker node costs apply. Suitable for non-production workloads.
- **Flannel overlay** (`flannel` addon + `podCidr`) -- Simple overlay networking that does not require dedicated pod VSwitches. Pods get cluster-internal IPs from the pod CIDR, routed via VXLAN tunnels between nodes. Simpler to set up than Terway but without VPC-native pod networking.
- **Two availability zones** -- Minimum for basic resilience. A single AZ would work for development but two prevents zone-level outages from blocking work entirely.
- **Public API server** (`slbInternetEnabled: true`) -- Enables kubectl access from outside the VPC without a VPN, convenient for development workflows.
- **Minimal addons** (flannel, csi-plugin, csi-provisioner) -- Only the CNI and storage drivers. No monitoring, logging, or node problem detection to minimize resource overhead on small dev clusters.
- **No deletion protection** -- Development clusters are disposable; deletion protection would add friction to teardown.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region code (e.g., `cn-hangzhou`, `ap-southeast-1`) | Your deployment region |
| `<vswitch-id-zone-a>` | VSwitch in first availability zone | `AlicloudVswitch` stack outputs |
| `<vswitch-id-zone-b>` | VSwitch in second availability zone | `AlicloudVswitch` stack outputs |

## Related Presets

- **01-production-terway** -- Use for production workloads requiring Terway ENI networking, RRSA, logging, and professional SLA
- **03-production-flannel** -- Use for production workloads that prefer Flannel overlay networking with full security and observability features
