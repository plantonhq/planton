# Civo VPC

Deploys an isolated private network (VPC) on Civo Cloud within a specified region. The network can use a custom IPv4 CIDR range or let Civo auto-allocate one, and it exposes connection details that other components such as databases and Kubernetes clusters can reference.

## What Gets Created

When you deploy a CivoVpc resource, OpenMCF provisions:

- **Civo Network** -- a `civo_network` resource in the target region with the specified label and optional CIDR block
- **Resource Labels** -- standard OpenMCF labels applied to track the resource name, kind, organization, and environment

## Prerequisites

- **Civo credentials** configured via environment variables or OpenMCF provider config (the `civoCredentialId` field in the spec must reference a valid credential)
- **A target Civo region** -- the region must exist and be available on the Civo account (e.g., `lon1`, `fra1`, `nyc1`)

## Quick Start

Create a file `civo-vpc.yaml`:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoVpc
metadata:
  name: my-network
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoVpc.my-network
spec:
  civoCredentialId: my-civo-cred
  networkName: my-network
  region: lon1
```

Deploy:

```shell
openmcf apply -f civo-vpc.yaml
```

This creates a private network named `my-network` in the London region with an auto-allocated CIDR block.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `civoCredentialId` | `string` | ID of the Civo credential used to authenticate with the Civo API. | Required |
| `networkName` | `string` | DNS-friendly label for the network. Used as the `label` on the Civo network resource. | Required |
| `region` | `string` | Civo region where the network is created (e.g., `lon1`, `fra1`, `nyc1`, `phx1`). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ipRangeCidr` | `string` | auto-allocated | IPv4 CIDR range for the network (max `/24`). When omitted, Civo allocates an available range automatically. |
| `isDefaultForRegion` | `bool` | `false` | Whether the network should be the default for the region. Only one default network is allowed per region. Note: not currently supported by the Pulumi Civo provider; a warning is logged and the flag is skipped during provisioning. Use the Civo CLI (`civo network default <network-id>`) to set a network as default after creation. |
| `description` | `string` | `""` | Human-readable description for the network (max 100 characters). Recorded in OpenMCF metadata only; the Civo network provider does not expose a description field. |

## Examples

### Basic Network with Auto-Allocated CIDR

A minimal private network with Civo handling address allocation:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoVpc
metadata:
  name: dev-network
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoVpc.dev-network
spec:
  civoCredentialId: my-civo-cred
  networkName: dev-network
  region: fra1
```

### Custom CIDR Range

A network with an explicit address range for predictable IP planning:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoVpc
metadata:
  name: staging-network
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.CivoVpc.staging-network
spec:
  civoCredentialId: my-civo-cred
  networkName: staging-network
  region: nyc1
  ipRangeCidr: 10.0.0.0/24
  description: Staging environment private network
```

### Production Network with All Options

A fully specified network intended for production workloads:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoVpc
metadata:
  name: prod-network
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infra
    pulumi.openmcf.org/stack.name: prod.CivoVpc.prod-network
spec:
  civoCredentialId: prod-civo-cred
  networkName: prod-network
  region: lon1
  ipRangeCidr: 10.10.0.0/24
  isDefaultForRegion: true
  description: Production private network for lon1
```

Note: `isDefaultForRegion` is accepted in the manifest but is not applied during provisioning due to a limitation in the Pulumi Civo provider. After deployment, run `civo network default <network-id>` to set the network as default.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `networkId` | `string` | Unique identifier (UUID) of the created Civo network |
| `cidrBlock` | `string` | IPv4 CIDR block assigned to the network (either the specified `ipRangeCidr` or the auto-allocated range) |

Note: The `createdAtRfc3339` field is defined in the output schema but is not currently populated by the Pulumi Civo provider.

## Related Components

- [CivoFirewall](/docs/catalog/civo/civofirewall) -- defines firewall rules for controlling network traffic
- [CivoKubernetesCluster](/docs/catalog/civo/civokubernetescluster) -- deploys a Kubernetes cluster that can be attached to this network
- [CivoDatabase](/docs/catalog/civo/civodatabase) -- provisions a managed database instance connected to this network
- [CivoComputeInstance](/docs/catalog/civo/civocomputeinstance) -- launches compute instances within this network
