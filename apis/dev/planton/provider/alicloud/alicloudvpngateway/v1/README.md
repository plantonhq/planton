# AliCloudVpnGateway

Manages an Alibaba Cloud VPN Gateway with bundled customer gateways and IPsec VPN connections.

## Overview

A VPN Gateway provides encrypted site-to-site connectivity between an Alibaba Cloud VPC and remote networks (on-premises data centers, branch offices, or other cloud environments) over the public internet using IPsec.

### What Gets Created

- **VPN Gateway** -- an IPsec/SSL VPN gateway placed in a VPC/VSwitch
- **Customer Gateways** -- one per connection, representing the remote VPN device's public IP
- **VPN Connections** -- one per connection, establishing an IPsec tunnel with IKE/IPsec negotiation parameters

### How IPsec VPN Works

Each connection defines a tunnel between the Alibaba Cloud VPN Gateway and a remote device. The tunnel is negotiated in two phases:

1. **IKE (Phase 1)** -- Establishes a secure channel between peers using pre-shared keys and Diffie-Hellman key exchange.
2. **IPsec (Phase 2)** -- Negotiates the encryption and authentication for actual data traffic through the tunnel.

Multiple connections can share a single VPN Gateway, each connecting to a different remote site.

### SSL VPN

The gateway optionally supports SSL VPN for remote client access when `enableSsl` is true. SSL VPN servers and client certificates are managed separately.

## Build and Test (Localized)

All build and test commands are scoped to this component directory. Never run project-wide `make build`.

```bash
# Proto compilation (from planton repo root, once after proto changes)
make protos

# Go build (Pulumi module)
go build ./apis/dev/planton/provider/alicloud/alicloudvpngateway/v1/iac/pulumi/...

# Go vet
go vet ./apis/dev/planton/provider/alicloud/alicloudvpngateway/v1/iac/pulumi/...

# Spec tests
go test ./apis/dev/planton/provider/alicloud/alicloudvpngateway/v1/...

# Terraform validation
cd apis/dev/planton/provider/alicloud/alicloudvpngateway/v1/iac/tf
terraform init -backend=false
terraform validate
```

## Configuration Reference

See [catalog-page.md](catalog-page.md) for complete field documentation, examples, and presets.
