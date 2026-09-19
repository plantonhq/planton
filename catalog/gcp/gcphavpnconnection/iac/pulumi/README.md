# GCP HA VPN Connection - Pulumi Module

## Overview

This directory contains the Pulumi implementation for connecting one HA VPN gateway to one peer — an on-premises site, another cloud, or another Google Cloud VPC — using Planton's `GcpHaVpnConnection` API. The module is written in Go and creates, through the bridged Google provider, an optional `compute.ExternalVpnGateway` and, per tunnel, `compute.VPNTunnel`, `compute.RouterInterface`, and `compute.RouterPeer` on the gateway's router.

## Prerequisites

1. **Pulumi CLI** installed (version 3.x or later)
2. **Go** installed (version 1.21 or later)
3. **GCP Credentials** configured:
   ```bash
   gcloud auth application-default login
   ```
4. **IAM permissions**: `roles/compute.networkAdmin` on the project
5. An existing `GcpHaVpnGateway` (the registry installs one as this kind's prerequisite in E2E)

## Directory Structure

```
iac/pulumi/
├── main.go                 # Pulumi program entry point
├── Pulumi.yaml             # Pulumi project configuration
├── README.md               # This file
└── module/
    ├── main.go             # Module coordinator
    ├── external_gateway.go # The peer device's addresses (external peer only)
    ├── tunnels.go          # Per tunnel: tunnel, router interface, BGP peer
    ├── locals.go           # Resolved gateway trio, names, label set
    └── outputs.go          # Stack output constants
```

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --stack org/project/stack
planton pulumi destroy --manifest ../../e2e/manifest.yaml --stack org/project/stack
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`.

## What the module does

- **External peer** (`peer.external_gateway`): creates the external VPN gateway with the device's interfaces (exactly one address each), the redundancy type, description and labels when set. Every tunnel then references it and names its interface. **Google peer** (`peer.gcp_gateway`): no external gateway; every tunnel names the peer gateway's self link and Google pairs interfaces itself.
- **Per tunnel**: the IPsec tunnel (secret marked as a Pulumi secret; `ike_version` always sent, 2 when unset; traffic selectors and cipher suite only when set; labels merged with the platform's); the router interface named from `bgp_session.name` or the tunnel's name, bound to the tunnel, with `ip_range` / `ip_version` when set; the BGP peer on that interface with `peer_asn`, `enable` always sent (true when unset), `enable_ipv6` always sent, and every optional numeric or address only when set (so 0 is a real priority and Google keeps assigning addresses it owns); BFD and the MD5 key when declared (the key marked secret; its name defaults to `<tunnel>-md5`).
- `deletion_policy` is fanned to every resource when set.
- Exports the index-aligned `tunnel_self_links`, `tunnel_names`, `router_interface_names`, `bgp_peer_names`, plus `external_gateway_self_link` (empty for a Google peer), `gateway_self_link`, `router_name`.

## Parity with Terraform

Both engines create the same resources with the same arguments and export the same seven keys. There is no `PARITY-EXCEPTION` in this module.

## Notes

- **Every tunnel field except labels is immutable.** Rotate a pre-shared key by adding a tunnel with the new key and removing the old one.
- **Two tunnels, one per gateway interface, to two peer addresses is the 99.99% shape.**
- **The MD5 key rides the peer**: the provider inserts it into the router's key table and Google requires each key to be used by exactly one session.
