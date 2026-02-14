# Scaleway VPC

Deploys a Scaleway VPC (Virtual Private Cloud) — a regional, logical container that groups Private Networks. The VPC itself does not define IP ranges or CIDR blocks; IP planning happens at the Private Network level. OpenMCF provisions the VPC with optional inter-Private-Network routing and custom routes propagation, and tags it with standard resource metadata.

## What Gets Created

When you deploy a ScalewayVpc resource, OpenMCF provisions:

- **VPC** — a `network.Vpc` resource in the specified Scaleway region, named after `metadata.name`, with standard OpenMCF tags (`resource`, `resource-name`, `resource-kind`, plus optional `organization`, `environment`, and `resource-id` tags derived from metadata)

## Prerequisites

- **Scaleway credentials** configured via environment variables or OpenMCF provider config
- **A target region** — one of the Scaleway regions (e.g., `fr-par`, `nl-ams`, `pl-waw`)

## Quick Start

Create a file `vpc.yaml`:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayVpc
metadata:
  name: my-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayVpc.my-vpc
spec:
  region: fr-par
```

Deploy:

```shell
openmcf apply -f vpc.yaml
```

This creates a VPC in `fr-par` with routing disabled. The VPC ID is exported as a stack output for use by downstream resources such as ScalewayPrivateNetwork.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Scaleway region where the VPC will be created (e.g., `"fr-par"`, `"nl-ams"`, `"pl-waw"`). Cannot be changed after creation. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enableRouting` | `bool` | `false` | Enables routing between Private Networks attached to this VPC. Once enabled, routing cannot be disabled. Required for multi-tier architectures where resources in separate Private Networks need to communicate (e.g., a Kapsule cluster talking to an RDB instance). |
| `enableCustomRoutesPropagation` | `bool` | `false` | Enables custom routes propagation between Private Networks in this VPC. Once enabled, it cannot be disabled. Useful for advanced networking scenarios such as VPN gateways or network appliances. |

## Examples

### Minimal VPC

A VPC with no routing — suitable for isolating a single Private Network:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayVpc
metadata:
  name: isolated-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayVpc.isolated-vpc
spec:
  region: nl-ams
```

### VPC with Inter-Network Routing

Enable routing so that resources in different Private Networks within the VPC can reach each other. This is the typical setup for multi-tier applications:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayVpc
metadata:
  name: app-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayVpc.app-vpc
  env: prod
  org: acme
spec:
  region: fr-par
  enableRouting: true
```

### VPC with Routing and Custom Routes Propagation

Full networking capabilities enabled — routing between Private Networks plus custom routes advertised across them. Use this when deploying VPN gateways or network appliances that inject custom routes:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayVpc
metadata:
  name: network-vpc
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayVpc.network-vpc
  env: prod
  org: acme
spec:
  region: pl-waw
  enableRouting: true
  enableCustomRoutesPropagation: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `vpcId` | `string` | UUID of the created Scaleway VPC. Referenced by downstream resources (e.g., ScalewayPrivateNetwork) via `StringValueOrRef`. |

## Related Components

- [ScalewayPrivateNetwork](/docs/catalog/scaleway/scalewayprivatenetwork) — creates Private Networks attached to this VPC for workload-level IP planning and isolation
- [ScalewayKapsuleCluster](/docs/catalog/scaleway/scalewaykapsulecluster) — deploys a managed Kubernetes cluster that requires a Private Network (and therefore a VPC)
- [ScalewayRdbInstance](/docs/catalog/scaleway/scalewayrdbinstance) — provisions managed databases that can be attached to a Private Network within this VPC
