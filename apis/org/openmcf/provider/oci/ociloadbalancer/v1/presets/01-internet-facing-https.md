# Internet-Facing HTTPS Load Balancer

This preset creates a public OCI Application Load Balancer with HTTPS termination on port 443 and an automatic HTTP-to-HTTPS redirect on port 80. It uses the flexible shape with 10-100 Mbps bandwidth, deploys across two subnets for regional high availability, and applies a Network Security Group for network-level access control. This is the standard production configuration for the vast majority of internet-facing web applications and APIs on OCI.

## When to Use

- Production web applications or APIs that need HTTPS termination at the load balancer
- Any internet-facing workload where all HTTP traffic should be redirected to HTTPS
- Deployments requiring regional high availability across multiple availability domains
- Environments where network-level access control via NSGs is required (e.g., restricting source IPs)

## Key Configuration Choices

- **Public load balancer** (`isPrivate: false`) -- Receives a public IP address accessible from the internet. Use preset 02-internal-http instead for VCN-internal services.
- **Flexible shape with 10-100 Mbps** (`shape: flexible`, `shapeDetails`) -- Starts at 10 Mbps minimum (cost-efficient baseline) and bursts to 100 Mbps under load. Adjust `maximumBandwidthInMbps` upward for high-traffic workloads; OCI supports up to 8000 Mbps.
- **Two subnets across ADs** (`subnetIds`) -- Provides regional high availability. If one availability domain experiences issues, the load balancer continues serving traffic from the other. Both subnets must be public subnets for a public load balancer.
- **HTTPS listener with OCI Certificate Service** (`listeners[0].sslConfiguration.certificateIds`) -- References a certificate managed by the OCI Certificate Service rather than embedding PEM content. This enables automatic certificate rotation and lifecycle management. TLS 1.2 and 1.3 are both enabled for broad client compatibility with modern security.
- **HTTP-to-HTTPS redirect** (`listeners[1]` + `ruleSets[0]`) -- Port 80 listener with a rule set that returns a 301 permanent redirect to the HTTPS equivalent URL. This ensures no plaintext traffic reaches backends and is the standard operational pattern for production load balancers.
- **Round-robin policy** (`backendSets[0].policy: round_robin`) -- Distributes requests evenly across all healthy backends. Appropriate for stateless web servers with similar capacity. Switch to `least_connections` if backends have variable request processing times.
- **HTTP health check on /health** (`backendSets[0].healthChecker`) -- Checks the `/health` endpoint every 30 seconds, expecting HTTP 200, with a 3-second timeout and 3 retries before marking a backend unhealthy. Adjust `urlPath` to match your application's health endpoint.
- **Delete protection enabled** (`isDeleteProtectionEnabled: true`) -- Prevents accidental deletion of a production load balancer. Must be explicitly disabled before the load balancer can be destroyed.
- **NSG applied** (`networkSecurityGroupIds`) -- Controls which source IPs and ports can reach the load balancer at the network level. Configure the referenced NSG to allow inbound TCP on ports 80 and 443.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the load balancer will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<public-subnet-ocid-ad1>` | OCID of a public subnet in the first availability domain | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<public-subnet-ocid-ad2>` | OCID of a public subnet in the second availability domain | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<lb-nsg-ocid>` | OCID of the NSG controlling access to the load balancer (allow TCP 80, 443 inbound) | OCI Console > Networking > NSGs, or `OciNetworkSecurityGroup` status outputs |
| `<oci-certificate-service-certificate-ocid>` | OCID of a certificate managed by OCI Certificate Service | OCI Console > Identity & Security > Certificates |
| `<backend-ip-1>` | Private IP address of the first backend server | OCI Console > Compute > Instances, or `OciComputeInstance` status outputs |
| `<backend-ip-2>` | Private IP address of the second backend server | OCI Console > Compute > Instances, or `OciComputeInstance` status outputs |

## Related Presets

- **02-internal-http** -- Use instead for private load balancers that serve traffic only within the VCN (e.g., internal APIs, microservice communication)
- **03-development** -- Use instead for dev/test environments where HTTPS, HA, and delete protection are unnecessary
