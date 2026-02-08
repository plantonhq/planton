# OpenStack Compute Keypair

Provision and manage SSH keypairs in OpenStack Nova using OpenMCF's unified API.

## Overview

OpenStack compute keypairs are SSH key pairs used for authenticating access to compute instances. They are a fundamental security primitive—when you launch an instance, you specify a keypair, and OpenStack injects the public key into the instance so you can SSH in with the corresponding private key.

This component supports two workflows:
1. **Import an existing public key**: Bring your own key generated with `ssh-keygen` (recommended for production)
2. **Generate a new keypair**: Let OpenStack create the key pair (the private key is available once via IaC state)

The keypair name is derived from `metadata.name`.

## Prerequisites

1. **OpenStack Cloud**: Access to an OpenStack deployment with Nova (Compute service)
2. **Credentials**: OpenStack credentials configured via the credential management system
3. **OpenMCF CLI**: Install from [openmcf.org](https://openmcf.org)

## Quick Start

### Import an Existing Public Key (Recommended)

Generate a key locally and import the public half:

```bash
# Generate a keypair locally
ssh-keygen -t ed25519 -C "deploy@myapp" -f ~/.ssh/myapp-key -N ""
```

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackComputeKeypair
metadata:
  name: myapp-deploy-key
spec:
  public_key: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA... deploy@myapp"
```

### Generate a Keypair via OpenStack

Let OpenStack create both halves (private key available once via IaC state):

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackComputeKeypair
metadata:
  name: auto-generated-key
spec: {}
```

> **Security Note**: When generating keypairs, the private key is stored in IaC state (Terraform state file or Pulumi state). For production use, generate keys locally with `ssh-keygen` and import the public key.

### Deploy

```bash
openmcf apply --manifest keypair.yaml \
  -p openstack-creds.yaml
```

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `public_key` | string | No | SSH public key in OpenSSH format. If omitted, OpenStack generates a keypair. |
| `region` | string | No | Override the region from the provider config. |

## Outputs

| Field | Description |
|-------|-------------|
| `name` | The keypair name (from `metadata.name`) |
| `fingerprint` | MD5 fingerprint of the SSH public key |
| `public_key` | The SSH public key |
| `region` | Region where the keypair was created |

## IaC Implementations

This component is implemented with both Pulumi (Go) and Terraform (HCL) with full feature parity.

- **Terraform resource**: `openstack_compute_keypair_v2`
- **Pulumi resource**: `openstack.compute.Keypair`
