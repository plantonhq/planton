# OpenStackContainerClusterTemplate Pulumi Module

Provisions an OpenStack Magnum cluster template using the Pulumi OpenStack provider. A cluster template defines the base image, network topology, node flavors, and container orchestration engine for Magnum-managed Kubernetes clusters.

## Architecture

The module creates a single `containerinfra.ClusterTemplate` resource with:
- Required COE and image configuration
- Optional SSH keypair reference (FK to OpenStackKeypair)
- Optional network references (FK to OpenStackNetwork, OpenStackSubnet)
- Optional node flavor, Docker volume, and HA settings

## Local Development

```bash
make build
make install-pulumi-plugins
make test
```
