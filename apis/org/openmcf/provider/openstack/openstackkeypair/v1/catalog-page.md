# OpenStack Keypair

Deploys an OpenStack Nova compute keypair for SSH authentication to instances. Supports both importing an existing public key and having OpenStack generate a new keypair.

## What Gets Created

When you deploy an OpenStackKeypair resource, OpenMCF provisions:

- **Compute Keypair** — an `openstack_compute_keypair_v2` resource that registers an SSH public key with Nova for injection into instances at launch time via cloud-init

When no `publicKey` is provided in the spec, OpenStack generates a new keypair. The generated private key is exported as a secret-level IaC output only (not stored in platform stack outputs). It must be retrieved immediately after creation via `pulumi stack output private_key --show-secrets`.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **Nova (Compute) API access** for the authenticated user or application credential
- **ssh-keygen** or equivalent tool if importing an externally generated public key (recommended for production)

## Quick Start

Create a file `keypair.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackKeypair
metadata:
  name: my-keypair
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackKeypair.my-keypair
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackkeypair/v1/iac/pulumi/module
spec:
  publicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ... user@host"
```

Deploy:

```shell
openmcf apply -f keypair.yaml
```

This imports the provided SSH public key into OpenStack as a keypair named `my-keypair`. The keypair can then be referenced by name when launching compute instances.

## Configuration Reference

### Required Fields

All spec fields are optional. The keypair name is derived from `metadata.name`.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `publicKey` | `string` | — | SSH public key to import, in OpenSSH authorized_keys format (e.g., `ssh-rsa AAAAB3... user@host`). If omitted, OpenStack generates a new keypair and the private key is available as a one-time secret IaC output. For production use, generating keys locally with `ssh-keygen` and importing the public key is recommended. |
| `region` | `string` | provider default | Overrides the region from the provider config for this keypair. |

## Examples

### Import an Existing Public Key

The most common and recommended approach. Generate a key pair locally with `ssh-keygen` and import the public key:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackKeypair
metadata:
  name: dev-keypair
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: dev.OpenstackKeypair.dev-keypair
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackkeypair/v1/iac/pulumi/module
spec:
  publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleKeyData developer@workstation"
```

### Generate a New Keypair

When no `publicKey` is provided, OpenStack generates a new keypair. The private key is stored encrypted in IaC state and must be retrieved immediately after creation. This approach is suitable for ephemeral environments where long-term key management is not required:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackKeypair
metadata:
  name: ephemeral-keypair
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: staging.OpenstackKeypair.ephemeral-keypair
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackkeypair/v1/iac/pulumi/module
spec: {}
```

After deployment, retrieve the generated private key:

```shell
pulumi stack output private_key --show-secrets
```

### Multi-Region Keypair with Region Override

Deploy a keypair to a specific region, overriding the default region from the provider config. Useful in multi-region deployments where keypairs need to exist in each region:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackKeypair
metadata:
  name: us-west-keypair
  labels:
    openmcf.org/provisioner: pulumi
    openmcf.org/stack.jobId: prod.OpenstackKeypair.us-west-keypair
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackkeypair/v1/iac/pulumi/module
spec:
  publicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7ExampleProdKeyData ops-team@corp"
  region: us-west-1
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | Name of the keypair, derived from `metadata.name` |
| `fingerprint` | `string` | MD5 fingerprint of the SSH public key (e.g., `d7:62:43:93:10:a8:7e:7c:01:c8:c5:67:ba:99:5c:25`) |
| `publicKey` | `string` | SSH public key in OpenSSH authorized_keys format. Either the imported key or the generated public key. |
| `region` | `string` | OpenStack region where the keypair was created |

The private key is intentionally excluded from stack outputs. When OpenStack generates a keypair (no `publicKey` in spec), the private key is available only as a secret IaC-level output, retrievable via `pulumi stack output private_key --show-secrets`.

## Related Components

- [OpenStackInstance](/docs/catalog/openstack/openstackinstance) — launches compute instances that reference keypairs for SSH access
- [OpenStackSecurityGroup](/docs/catalog/openstack/openstacksecuritygroup) — defines firewall rules, including SSH (port 22) ingress required for keypair-based access
- [OpenStackNetwork](/docs/catalog/openstack/openstacknetwork) — provides the network connectivity for SSH access to instances
