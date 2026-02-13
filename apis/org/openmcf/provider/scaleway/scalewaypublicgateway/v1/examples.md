# Scaleway Public Gateway Examples

This document provides practical examples for deploying Scaleway Public Gateways using the OpenMCF API. Each example demonstrates different use cases from minimal NAT gateways to production-grade configurations with SSH bastion and port forwarding.

## Table of Contents

- [Example 1: Minimal NAT Gateway](#example-1-minimal-nat-gateway)
- [Example 2: Gateway with SSH Bastion](#example-2-gateway-with-ssh-bastion)
- [Example 3: Production Gateway with valueFrom Reference](#example-3-production-gateway-with-valuefrom-reference)
- [Example 4: Gateway with Port Forwarding](#example-4-gateway-with-port-forwarding)
- [Example 5: Full Kapsule Environment Stack](#example-5-full-kapsule-environment-stack)
- [Example 6: Email-Capable Gateway with Reverse DNS](#example-6-email-capable-gateway-with-reverse-dns)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Minimal NAT Gateway

**Use Case:** The simplest setup. A gateway that provides NAT masquerade for a Private Network, allowing resources to reach the internet.

**Configuration:**
- **Zone:** fr-par-1 (Paris zone 1)
- **Type:** VPC-GW-S (standard, sufficient for most workloads)
- **NAT:** Enabled (default)
- **Bastion:** Disabled
- **PAT rules:** None

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPublicGateway
metadata:
  name: dev-gateway
  org: my-org
  env: dev
spec:
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  zone: fr-par-1
  type: VPC-GW-S
```

**Deploy with OpenMCF CLI:**

```bash
openmcf apply -f dev-gateway.yaml
```

**What happens:**
- A Flexible IP is created and assigned to the gateway.
- A Public Gateway named `dev-gateway` is created in zone `fr-par-1`.
- The gateway is attached to the specified Private Network with NAT masquerade enabled (default).
- Resources in the Private Network can now reach the internet.

---

## Example 2: Gateway with SSH Bastion

**Use Case:** A gateway with SSH bastion enabled for secure access to instances in the Private Network. Access is restricted to specific IP ranges.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPublicGateway
metadata:
  name: bastion-gateway
  org: my-org
  env: staging
spec:
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  zone: fr-par-1
  type: VPC-GW-S
  bastion:
    enabled: true
    port: 22
    allowedIpRanges:
      - "203.0.113.0/24"
      - "198.51.100.10/32"
```

**What happens:**
- NAT masquerade is enabled (default).
- SSH bastion is activated on port 22.
- Only connections from `203.0.113.0/24` and `198.51.100.10/32` can reach the bastion.
- Developers SSH to the gateway's public IP, and the gateway proxies to the target instance's private IP.

**Connecting via bastion:**

```bash
ssh -J bastion@<GATEWAY_PUBLIC_IP> user@<INSTANCE_PRIVATE_IP>
```

---

## Example 3: Production Gateway with valueFrom Reference

**Use Case:** The most common production pattern. The gateway references a ScalewayPrivateNetwork using `valueFrom`, enabling the platform to build a dependency DAG and deploy in the correct order.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayVpc
metadata:
  name: prod-vpc
  org: my-org
  env: prod
spec:
  region: fr-par
  enableRouting: true
---
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPrivateNetwork
metadata:
  name: app-network
  org: my-org
  env: prod
spec:
  vpcId:
    valueFrom:
      kind: ScalewayVpc
      name: prod-vpc
      fieldPath: status.outputs.vpc_id
  region: fr-par
  ipv4Subnet: "10.0.1.0/24"
  enableDefaultRoutePropagation: true
---
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPublicGateway
metadata:
  name: prod-gateway
  org: my-org
  env: prod
spec:
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: app-network
      fieldPath: status.outputs.private_network_id
  zone: fr-par-1
  type: VPC-GW-S
  bastion:
    enabled: true
    allowedIpRanges:
      - "10.20.30.0/24"
```

**What happens:**
- The platform resolves the dependency chain: VPC -> Private Network -> Public Gateway.
- Each resource is deployed in topological order.
- The gateway's `privateNetworkId` is automatically populated from the Private Network's stack output.

---

## Example 4: Gateway with Port Forwarding

**Use Case:** Expose specific services running in the Private Network to the internet via port forwarding. Useful for small deployments that don't need a full Load Balancer.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPublicGateway
metadata:
  name: forwarding-gateway
  org: my-org
  env: staging
spec:
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  zone: fr-par-1
  type: VPC-GW-S
  patRules:
    - privateIp: "10.0.1.5"
      privatePort: 80
      publicPort: 8080
      protocol: tcp
    - privateIp: "10.0.1.10"
      privatePort: 5432
      publicPort: 15432
      protocol: tcp
```

**What happens:**
- NAT masquerade is enabled for outbound traffic.
- Public port 8080 on the gateway's IP forwards TCP traffic to `10.0.1.5:80` (a web server).
- Public port 15432 forwards TCP traffic to `10.0.1.10:5432` (a PostgreSQL instance).
- External clients access: `http://<GATEWAY_PUBLIC_IP>:8080` and `psql -h <GATEWAY_PUBLIC_IP> -p 15432`.

---

## Example 5: Full Kapsule Environment Stack

**Use Case:** A complete Kapsule (managed Kubernetes) environment with VPC, Private Network, and Public Gateway providing NAT and bastion access. This mirrors the `kapsule-environment` infra chart pattern.

```yaml
# Layer 0: VPC
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayVpc
metadata:
  name: k8s-vpc
  org: my-org
  env: prod
spec:
  region: fr-par
  enableRouting: true
---
# Layer 1: Private Network for the cluster
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPrivateNetwork
metadata:
  name: k8s-network
  org: my-org
  env: prod
spec:
  vpcId:
    valueFrom:
      kind: ScalewayVpc
      name: k8s-vpc
      fieldPath: status.outputs.vpc_id
  region: fr-par
  ipv4Subnet: "10.0.1.0/22"
  enableDefaultRoutePropagation: true
---
# Layer 2: Public Gateway for NAT + bastion
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPublicGateway
metadata:
  name: k8s-gateway
  org: my-org
  env: prod
spec:
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: k8s-network
      fieldPath: status.outputs.private_network_id
  zone: fr-par-1
  type: VPC-GW-S
  bastion:
    enabled: true
    allowedIpRanges:
      - "203.0.113.0/24"
```

**Dependency graph:**

```
ScalewayVpc (k8s-vpc)
    └── ScalewayPrivateNetwork (k8s-network)
            └── ScalewayPublicGateway (k8s-gateway)
```

**What happens:**
- VPC is created first (Layer 0).
- Private Network is created inside the VPC (Layer 1).
- Public Gateway is attached to the Private Network (Layer 2).
- Kapsule pods (deployed later) can pull container images and reach external APIs via NAT.
- Developers can SSH to nodes via the bastion.

---

## Example 6: Email-Capable Gateway with Reverse DNS

**Use Case:** A gateway for infrastructure that needs to send email directly. Enables SMTP and configures reverse DNS for deliverability compliance.

**Prerequisites:** A DNS A record pointing to the gateway's public IP must exist before setting `reverseDns`.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPublicGateway
metadata:
  name: mail-gateway
  org: my-org
  env: prod
spec:
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  zone: fr-par-1
  type: VPC-GW-S
  enableSmtp: true
  reverseDns: "mail.example.com"
```

**What happens:**
- Outbound SMTP (port 25) is unblocked for resources in the Private Network.
- The public IP's reverse DNS is set to `mail.example.com`.
- Email servers behind the gateway can send mail with proper PTR record compliance.

---

## Common Operations

### Get Gateway Outputs After Deployment

```bash
# Pulumi
pulumi stack output gateway_id
pulumi stack output public_ip_address

# Terraform
terraform output gateway_id
terraform output public_ip_address
```

### Check Public IP Address

```bash
# Quick connectivity test
curl -s http://$(pulumi stack output public_ip_address):8080
```

### Use Public IP in DNS Records

The `public_ip_address` output is useful for creating DNS records:

```yaml
# Example: ScalewayDnsRecord referencing the gateway's public IP
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayDnsRecord
metadata:
  name: gateway-dns
spec:
  # ... DNS zone reference ...
  records:
    - type: A
      value:
        valueFrom:
          kind: ScalewayPublicGateway
          name: k8s-gateway
          fieldPath: status.outputs.public_ip_address
```

### Destroy a Public Gateway

Destroying the gateway also destroys the attached GatewayNetwork, PAT rules, and Flexible IP.

```bash
# Pulumi
pulumi destroy

# Terraform
terraform destroy
```

---

## Best Practices

### Always Enable Masquerade for Kapsule Environments

Kapsule pods need internet access to pull container images and reach external services. Always deploy a Public Gateway with masquerade enabled alongside Kapsule clusters.

### Restrict Bastion Access

Never leave `allowedIpRanges` empty in production. Restrict to:
- Your office IP ranges
- VPN exit IPs
- CI/CD runner IPs

```yaml
bastion:
  enabled: true
  allowedIpRanges:
    - "203.0.113.0/24"   # Office network
    - "198.51.100.5/32"  # VPN exit IP
```

### Use valueFrom for Infra Chart Composability

Always use `valueFrom` references (not hardcoded IDs) when deploying as part of an infra chart. This enables:
- Automatic dependency ordering (VPC -> Private Network -> Gateway).
- Clean environment promotion (dev/staging/prod use the same template, different names).
- Impact analysis in the platform's dependency graph.

### Choose the Right Gateway Type

- **VPC-GW-S**: Standard gateway. Sufficient for development, staging, and most production workloads.
- **VPC-GW-XL**: High-bandwidth (up to 10 Gbps). Use only for high-throughput production workloads in Paris regions.

### Match Zone to Private Network Region

The gateway zone must be within the same region as the Private Network:

| Private Network Region | Valid Gateway Zones |
|---|---|
| `fr-par` | `fr-par-1`, `fr-par-2`, `fr-par-3` |
| `nl-ams` | `nl-ams-1`, `nl-ams-2`, `nl-ams-3` |
| `pl-waw` | `pl-waw-1`, `pl-waw-2`, `pl-waw-3` |

### Prefer Gateway NAT Over Individual Public IPs

Instead of assigning public IPs to each instance, use a single Public Gateway with NAT masquerade. This:
- Reduces cost (one IP instead of many).
- Simplifies firewall rules (one source IP for all outbound traffic).
- Improves security (instances are not directly reachable from the internet).
- Provides centralized control (bastion for SSH, PAT for selective inbound).
