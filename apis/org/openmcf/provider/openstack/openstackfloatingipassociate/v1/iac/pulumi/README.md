# OpenStackFloatingIpAssociate Pulumi Module

Provisions an OpenStack Neutron floating IP association using the Pulumi OpenStack provider. Binds a floating IP (address or UUID) to a port.

## Architecture

The module creates a single `networking.FloatingIpAssociate` resource with:
- Required floating IP reference (FK to OpenStackFloatingIp address)
- Required port reference (FK to OpenStackNetworkPort)
- Optional fixed IP for multi-IP ports

## Local Development

```bash
make build
make install-pulumi-plugins
make test
```
