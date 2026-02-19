# Internal HTTP Load Balancer

This preset creates a private OCI Application Load Balancer for distributing traffic across backend services within the VCN. It uses the flexible shape with 10-50 Mbps bandwidth, deploys in a single private subnet, and applies a least-connections load balancing policy optimized for internal API workloads with variable request cost. No SSL, rule sets, or hostnames are configured -- internal traffic stays within the VCN where encryption overhead is unnecessary.

## When to Use

- Internal microservices or APIs that receive traffic only from within the VCN or peered networks
- Backend service tiers in a multi-tier architecture where the load balancer must not be internet-accessible
- Internal API gateways that aggregate traffic from multiple frontend services
- Any workload where a stable private IP is needed for service discovery or DNS-based routing within the VCN

## Key Configuration Choices

- **Private load balancer** (`isPrivate: true`) -- Receives only private IP addresses from the assigned subnet. Not accessible from the public internet. All traffic must originate from within the VCN or connected networks (peering, FastConnect, VPN).
- **Flexible shape with 10-50 Mbps** (`shape: flexible`, `shapeDetails`) -- Internal traffic volumes are typically more predictable than internet-facing traffic, so a lower maximum bandwidth ceiling reduces cost. Increase `maximumBandwidthInMbps` if the service handles high internal throughput.
- **Single subnet** (`subnetIds`) -- Simpler than multi-AD deployment. For internal services, the application layer typically handles redundancy (multiple backend instances). Add a second subnet in a different AD if the internal LB itself needs HA.
- **Least-connections policy** (`backendSets[0].policy: least_connections`) -- Routes each new request to the backend with the fewest active connections. Better than round-robin for internal APIs where some endpoints are computationally expensive and others are cheap, preventing slow backends from accumulating a request backlog.
- **HTTP health check on port 8080** (`backendSets[0].healthChecker`) -- Probes the `/health` endpoint on the backend's application port (8080) rather than the listener port (80). This verifies the application process is responsive, not just that the port is open. Adjust `port` and `urlPath` to match your service.
- **Listener on port 80 forwarding to backends on 8080** -- A common internal pattern where the LB accepts traffic on a standard port (80) and forwards to the application's native port (8080). Adjust both ports to match your service topology.
- **No SSL** -- Traffic between internal services stays within the VCN's private network. Adding TLS for internal traffic increases latency and operational complexity (certificate management) with minimal security benefit when the network is already isolated. If your compliance requirements mandate encryption in transit even within the VCN, add an `sslConfiguration` to the listener and backend set.
- **No delete protection** -- Internal services are redeployed and replaced more frequently than internet-facing infrastructure. Omitting delete protection simplifies teardown during service migrations.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the load balancer will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<private-subnet-ocid>` | OCID of the private subnet hosting the load balancer | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<internal-lb-nsg-ocid>` | OCID of the NSG controlling access to the internal load balancer (allow traffic from VCN CIDR) | OCI Console > Networking > NSGs, or `OciNetworkSecurityGroup` status outputs |
| `<backend-ip-1>` | Private IP address of the first backend server | OCI Console > Compute > Instances, or `OciComputeInstance` status outputs |
| `<backend-ip-2>` | Private IP address of the second backend server | OCI Console > Compute > Instances, or `OciComputeInstance` status outputs |

## Related Presets

- **01-internet-facing-https** -- Use instead for public-facing web applications that need HTTPS termination and HTTP-to-HTTPS redirect
- **03-development** -- Use instead for dev/test environments where a minimal configuration is sufficient
