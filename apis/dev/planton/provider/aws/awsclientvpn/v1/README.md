# AwsClientVpn

The **AwsClientVpn** resource provides a standardized way to provision and manage AWS Client VPN endpoints through Planton. It enables secure remote access into AWS VPCs using OpenVPN, allowing developers and teams to connect from their laptops or remote networks to private resources within a VPC.

## Spec Fields (80/20)

### Essential Fields (80% Use Case)

- **vpc_id**: The target AWS VPC where the Client VPN endpoint will be created. Must be in the same region as your AWS credentials. Required.
- **subnets**: List of subnet IDs within the VPC to associate as target networks. Each subnet association enables VPN clients to access resources in that subnet's Availability Zone. Minimum of 1 required.
- **client_cidr_block**: IPv4 CIDR block (e.g., "10.0.0.0/22") from which to assign client IP addresses. Must not overlap with VPC CIDR or any VPC routes. CIDR block size must be between /22 and /12. Required.
- **server_certificate_arn**: ARN of the AWS Certificate Manager (ACM) certificate to use for the VPN server. This TLS certificate is presented to clients upon connection. Required.
- **cidr_authorization_rules**: List of IPv4 CIDR ranges that VPN clients are authorized to access (e.g., ["10.0.0.0/16"]). Typically corresponds to private subnet ranges within the VPC.

### Advanced Fields (20% Use Case)

- **description**: Human-friendly description for the Client VPN endpoint, visible in AWS Console.
- **authentication_type**: Authentication method for clients. Currently only "certificate" (mutual TLS) is supported in v1. This requires clients to present a certificate signed by a trusted CA.
- **disable_split_tunnel**: When false (default), only traffic to authorized CIDRs goes through VPN; other traffic stays local. When true, all client traffic is routed through the VPN (full-tunnel mode).
- **vpn_port**: Port number for VPN connections. Default is 443 (recommended for firewall traversal). Allowed values: 443 (TCP) or 1194 (UDP).
- **transport_protocol**: Protocol for VPN sessions - "tcp" (default, recommended for reliability) or "udp" (lower latency but may be blocked by firewalls).
- **log_group_name**: CloudWatch Logs group name for connection logging. If specified, VPN connection events are logged. If omitted, logging is disabled.
- **security_groups**: List of security group IDs to apply to the VPN endpoint's network associations. Controls traffic between VPN clients and VPC resources. If not provided, VPC's default security group is used.
- **dns_servers**: Custom DNS server IP addresses for VPN clients (maximum 2). If not set, clients use the VPC's DNS (AmazonProvidedDNS).

## Stack Outputs

After provisioning, the AwsClientVpn resource provides the following outputs:

- **client_vpn_endpoint_id**: The unique identifier of the Client VPN endpoint.
- **security_group_id**: The ID of the security group applied to the VPN endpoint associations.
- **subnet_association_ids**: Map of subnet IDs to their association IDs.
- **vpn_endpoint_dns_name**: DNS name of the VPN endpoint for client configuration.

## How It Works

When you define an AwsClientVpn resource, Planton:

1. **Creates VPN Endpoint**: Provisions an AWS Client VPN endpoint in the specified VPC with the configured authentication and network settings.
2. **Associates Subnets**: Attaches the specified subnets as target networks, enabling clients to access resources in those subnets.
3. **Configures Authorization**: Creates authorization rules for the specified CIDR ranges, allowing VPN clients to access those networks.
4. **Applies Security**: Attaches security groups to control traffic between VPN clients and VPC resources.
5. **Enables Logging (Optional)**: If a CloudWatch log group is specified, enables connection logging for audit and troubleshooting.
6. **Configures Client Settings**: Sets up split-tunnel routing, DNS servers, and transport protocol based on the specification.

The resource uses Pulumi or Terraform under the hood (depending on your stack configuration) to provision all necessary AWS resources consistently.

## Use Cases

### Remote Developer Access
Enable developers to securely connect to private databases, internal APIs, and other VPC resources from their local machines without exposing those resources to the internet.

### Secure Site-to-Site Access
Provide secure access for remote offices or contractors to AWS resources without setting up complex VPN hardware or dedicated connections.

### Split-Tunnel for Optimal Performance
Use split-tunnel mode (default) to route only internal traffic through the VPN, keeping internet traffic local for better performance and reduced VPN bandwidth costs.

### Full-Tunnel for Maximum Security
Enable full-tunnel mode for scenarios requiring all client traffic to be inspected or filtered through AWS security controls.

### Multi-AZ High Availability
Associate subnets across multiple Availability Zones to ensure VPN connectivity remains available even if one AZ experiences issues.

## Important Notes

### Certificate Requirements
- Certificate-based authentication (mutual TLS) requires both a server certificate and client certificates.
- The server certificate must be in AWS Certificate Manager in the same region as the VPN endpoint.
- Client certificates must be generated and distributed to VPN users separately.

### Network Planning
- The client CIDR block must not overlap with your VPC CIDR or any connected networks.
- Plan for enough client IPs - a /22 CIDR provides ~1,000 client IPs.
- Authorization rules must match the networks clients need to access.

### Port and Protocol Selection
- Port 443 with TCP is recommended as it traverses most corporate firewalls.
- Port 1194 with UDP offers lower latency but may be blocked by firewalls.
- The port and protocol must be paired correctly: TCP↔443, UDP↔1194.

### Split Tunnel vs Full Tunnel
- Split tunnel (default): Only internal traffic goes through VPN, internet traffic stays local.
- Full tunnel: All client traffic routes through VPN, providing more control but potentially slower internet access.

## References

- [AWS Client VPN Documentation](https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/what-is.html)
- [AWS Client VPN Getting Started](https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/cvpn-getting-started.html)
- [OpenVPN Client Configuration](https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/cvpn-working-endpoint-export.html)
- [AWS Client VPN Authentication Methods](https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/client-authentication.html)
