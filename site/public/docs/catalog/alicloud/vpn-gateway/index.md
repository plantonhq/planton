---
title: "VPN Gateway"
description: "VPN Gateway deployment documentation"
icon: "package"
order: 100
componentName: "alicloudvpngateway"
---

# AliCloud VPN Gateway

Deploys an Alibaba Cloud VPN Gateway with bundled customer gateways and IPsec VPN connections. The component provisions all resources as a single atomic unit, establishing encrypted site-to-site tunnels between your VPC and remote networks.

## What Gets Created

When you deploy an AliCloudVpnGateway resource, OpenMCF provisions:

- **VPN Gateway** -- an `alicloud_vpn_gateway` resource in the specified VPC and VSwitch, with configurable bandwidth and optional SSL VPN
- **Customer Gateways** -- one `alicloud_vpn_customer_gateway` per connection, representing the remote device's public IP and optional BGP ASN
- **VPN Connections** -- one `alicloud_vpn_connection` per connection, with IKE/IPsec tunnel configuration, DPD, NAT traversal, and optional health checks

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or OpenMCF provider config
- **An Alibaba Cloud VPC** -- the VPN Gateway must belong to a VPC (create one with AliCloudVpc)
- **A VSwitch** -- the VPN Gateway requires placement in a VSwitch (create with AliCloudVswitch)
- **Remote device public IP** -- the on-premises router, firewall, or peer cloud gateway's public IP address
- **Network CIDR planning** -- VPC-side and remote-site CIDR blocks that should be reachable through the tunnels

## Quick Start

Create a file `vpn-gateway.yaml`:

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudVpnGateway
metadata:
  name: my-vpn
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-abc123
  vswitchId:
    value: vsw-abc123
  vpnGatewayName: my-vpn
  bandwidth: 10
  connections:
    - name: office-hq
      customerGatewayIp: "203.0.113.1"
      localSubnets:
        - "10.0.0.0/8"
      remoteSubnets:
        - "192.168.0.0/16"
```

Deploy:

```shell
openmcf apply -f vpn-gateway.yaml
```

This creates a 10 Mbps VPN Gateway with a single IPsec tunnel to a remote network at `203.0.113.1`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | string | Alibaba Cloud region (e.g., `cn-hangzhou`, `us-west-1`) | Required; non-empty |
| `vpcId` | StringValueOrRef | VPC ID for the VPN Gateway. Can reference AliCloudVpc via `valueFrom`. | Required |
| `vswitchId` | StringValueOrRef | VSwitch ID for gateway placement. Can reference AliCloudVswitch via `valueFrom`. | Required |
| `vpnGatewayName` | string | Gateway name (2-128 characters) | Required; 2-128 chars |
| `bandwidth` | int | Maximum bandwidth in Mbps | Must be one of: 5, 10, 20, 50, 100, 200, 500, 1000 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | | Human-readable description |
| `paymentType` | string | `PayAsYouGo` | Billing method: `PayAsYouGo` or `Subscription` |
| `enableSsl` | bool | `false` | Enable SSL VPN for remote client access |
| `sslConnections` | int | | Max concurrent SSL VPN clients (when `enableSsl` is `true`) |
| `tags` | map | | Key-value tags for the VPN Gateway |
| `resourceGroupId` | string | | Resource group for organizational grouping |
| `connections` | list | | IPsec VPN connections (see below) |

### Connection Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | string | *required* | Connection name (2-128 chars). Used for both the customer gateway and VPN connection. |
| `customerGatewayIp` | string | *required* | Public IP of the remote VPN device |
| `customerGatewayAsn` | string | | BGP ASN of the remote device (e.g., `65001`) |
| `localSubnets` | list | *required* | VPC-side CIDRs reachable through the tunnel (1-10 entries) |
| `remoteSubnets` | list | *required* | Remote-site CIDRs reachable through the tunnel (1-10 entries) |
| `enableDpd` | bool | `true` | Dead Peer Detection to verify remote peer is alive |
| `enableNatTraversal` | bool | `true` | NAT traversal (UDP encapsulation) for peers behind NAT |
| `effectImmediately` | bool | `true` | Start IPsec negotiation immediately instead of waiting for traffic |
| `ikeConfig` | object | | IKE Phase 1 parameters (see below) |
| `ipsecConfig` | object | | IPsec Phase 2 parameters (see below) |
| `healthCheckConfig` | object | | Tunnel health monitoring (see below) |

### IKE Config Fields

| Field | Type | Default | Valid Values |
|-------|------|---------|--------------|
| `psk` | string | auto-generated | Pre-shared key (1-100 characters) |
| `ikeVersion` | string | `ikev2` | `ikev1`, `ikev2` |
| `ikeMode` | string | `main` | `main`, `aggressive` |
| `ikeEncAlg` | string | `aes` | `aes`, `aes192`, `aes256`, `des`, `3des` |
| `ikeAuthAlg` | string | `sha1` | `md5`, `sha1`, `sha256`, `sha384`, `sha512` |
| `ikePfs` | string | `group2` | `group1`, `group2`, `group5`, `group14` |
| `ikeLifetime` | int | `86400` | 0-86400 seconds |

### IPsec Config Fields

| Field | Type | Default | Valid Values |
|-------|------|---------|--------------|
| `ipsecEncAlg` | string | `aes` | `aes`, `aes192`, `aes256`, `des`, `3des` |
| `ipsecAuthAlg` | string | `md5` | `md5`, `sha1`, `sha256`, `sha384`, `sha512` |
| `ipsecPfs` | string | `group2` | `disabled`, `group1`, `group2`, `group5`, `group14` |
| `ipsecLifetime` | int | `86400` | 0-86400 seconds |

### Health Check Config Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enable` | bool | `false` | Enable health probes for this tunnel |
| `sip` | string | | Source IP for probes (VPC-side, routable through the tunnel) |
| `dip` | string | | Destination IP for probes (remote-side) |
| `interval` | int | `3` | Seconds between probes |
| `retry` | int | `3` | Consecutive failures before the tunnel is declared unhealthy |

## Examples

### Single Site-to-Site Connection

The simplest VPN setup: one gateway with a single IPsec connection to a remote office.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudVpnGateway
metadata:
  name: office-vpn
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-abc123
  vswitchId:
    value: vsw-abc123
  vpnGatewayName: office-vpn
  bandwidth: 10
  connections:
    - name: office-hq
      customerGatewayIp: "203.0.113.1"
      localSubnets:
        - "10.0.0.0/8"
      remoteSubnets:
        - "192.168.0.0/16"
```

### Multi-Site Production with Custom Encryption

VPN Gateway connecting to two remote sites with AES-256, SHA-256 authentication, DH group14, and health checks for failover.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudVpnGateway
metadata:
  name: prod-vpn
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  vpcId:
    valueFrom:
      name: prod-vpc
  vswitchId:
    valueFrom:
      name: vpn-vswitch
  vpnGatewayName: prod-vpn-gateway
  description: Production VPN for datacenter connectivity
  bandwidth: 100
  tags:
    team: network
    costCenter: shared-infra
  connections:
    - name: datacenter-primary
      customerGatewayIp: "198.51.100.1"
      localSubnets:
        - "10.0.0.0/8"
        - "172.16.0.0/12"
      remoteSubnets:
        - "192.168.1.0/24"
        - "192.168.2.0/24"
      ikeConfig:
        psk: "strong-secret-key-dc1"
        ikeVersion: ikev2
        ikeEncAlg: aes256
        ikeAuthAlg: sha256
        ikePfs: group14
      ipsecConfig:
        ipsecEncAlg: aes256
        ipsecAuthAlg: sha256
        ipsecPfs: group14
      healthCheckConfig:
        enable: true
        sip: "10.0.0.1"
        dip: "192.168.1.1"
        interval: 5
        retry: 3
    - name: datacenter-dr
      customerGatewayIp: "198.51.100.2"
      localSubnets:
        - "10.0.0.0/8"
      remoteSubnets:
        - "192.168.10.0/24"
      ikeConfig:
        psk: "strong-secret-key-dc2"
        ikeVersion: ikev2
        ikeEncAlg: aes256
        ikeAuthAlg: sha256
        ikePfs: group14
      ipsecConfig:
        ipsecEncAlg: aes256
        ipsecAuthAlg: sha256
        ipsecPfs: group14
      healthCheckConfig:
        enable: true
        sip: "10.0.0.1"
        dip: "192.168.10.1"
```

### SSL VPN with Site-to-Site

VPN Gateway with SSL VPN enabled for remote client access alongside a site-to-site connection.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudVpnGateway
metadata:
  name: hybrid-vpn
spec:
  region: ap-southeast-1
  vpcId:
    value: vpc-sea1
  vswitchId:
    value: vsw-sea1a
  vpnGatewayName: hybrid-vpn
  bandwidth: 50
  enableSsl: true
  sslConnections: 50
  connections:
    - name: singapore-office
      customerGatewayIp: "203.0.113.10"
      localSubnets:
        - "10.0.0.0/8"
      remoteSubnets:
        - "172.20.0.0/16"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `vpn_gateway_id` | string | VPN Gateway resource ID (e.g., `vpn-xxxxx`) |
| `internet_ip` | string | VPN Gateway's public IP address, used as the local endpoint for IPsec tunnels |
| `ssl_vpn_internet_ip` | string | SSL VPN IP address (populated only when `enableSsl` is `true`) |
| `connection_ids` | map&lt;string, string&gt; | Map of connection names to VPN connection IDs |

## Related Components

- **AliCloudVpc** -- VPC that the VPN Gateway belongs to
- **AliCloudVswitch** -- VSwitch for gateway placement
- **AliCloudEipAddress** -- EIPs if dedicated public IPs are needed (the VPN Gateway gets its own)
- **AliCloudCenInstance** -- For multi-region VPC connectivity (alternative to VPN for Alibaba Cloud-to-Alibaba Cloud)
