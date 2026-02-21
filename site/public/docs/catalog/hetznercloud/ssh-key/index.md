---
title: "SSH Key"
description: "SSH Key deployment documentation"
icon: "package"
order: 100
componentName: "hetznercloudsshkey"
---

# Hetzner Cloud SSH Key

Registers an SSH public key in a Hetzner Cloud account for injection into servers at creation time. Supports RSA (>= 1024 bits), ED25519, and ECDSA key types. The key name and labels are derived from resource metadata, leaving only the public key content as the user-specified field.

## What Gets Created

When you deploy a HetznerCloudSshKey resource, OpenMCF provisions:

- **SSH Key** — an `hcloud_ssh_key` resource containing the public key material, a display name derived from `metadata.name`, and standard labels computed from resource metadata. The key is registered at the account level and referenced by servers via its numeric ID.

## Prerequisites

- **Hetzner Cloud API token** configured via environment variable (`HCLOUD_TOKEN`) or OpenMCF provider config
- **An SSH key pair** generated locally (e.g., `ssh-keygen -t ed25519`). Only the public key is needed.

## Quick Start

Create a file `ssh-key.yaml`:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSshKey
metadata:
  name: deploy-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudSshKey.deploy-key
spec:
  publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleKeyData deploy@ci"
```

Deploy:

```shell
openmcf apply -f ssh-key.yaml
```

This registers an ED25519 SSH public key named `deploy-key` in your Hetzner Cloud account.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `publicKey` | `string` | SSH public key in OpenSSH `authorized_keys` format. Supports ED25519, RSA (>= 1024 bits), and ECDSA. Changing this value forces replacement of the resource. | Required, non-empty (`min_len = 1`) |

### Optional Fields

This component has no optional spec fields. The SSH key name is derived from `metadata.name` and labels are computed from resource metadata.

## Examples

### Minimal ED25519 Key

The simplest deployment: a single ED25519 key with no organizational context.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSshKey
metadata:
  name: my-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudSshKey.my-key
spec:
  publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleKeyData user@host"
```

### Production Key with Org and Environment

A key scoped to a specific organization and environment. The metadata drives label generation for resource tracking.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSshKey
metadata:
  name: prod-deploy-key
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudSshKey.prod-deploy-key
    team: platform
spec:
  publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIProdDeployKeyData deploy@acme-ci"
```

### Server Composition via valueFrom

An SSH key referenced by a HetznerCloudServer using `valueFrom`. The server receives the key's numeric ID from the SSH key's stack outputs, establishing a dependency edge in the deployment DAG.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSshKey
metadata:
  name: web-key
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudSshKey.web-key
spec:
  publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIWebKeyData web-deploy@acme"
```

The server references this key:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: web-01
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudServer.web-01
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
  sshKeyIds:
    - valueFrom:
        kind: HetznerCloudSshKey
        name: web-key
        fieldPath: status.outputs.ssh_key_id
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `ssh_key_id` | `string` | Hetzner Cloud numeric ID of the created SSH key. Referenced by HetznerCloudServer via `sshKeyIds`. |
| `fingerprint` | `string` | MD5 fingerprint of the SSH public key (e.g., `"aa:bb:cc:dd:..."`). Computed by Hetzner Cloud from the uploaded key material. |

## Related Components

- [HetznerCloudServer](/docs/catalog/hetznercloud/server) — References SSH key IDs for password-less access at server boot
- [HetznerCloudFirewall](/docs/catalog/hetznercloud/firewall) — Commonly deployed alongside SSH keys to restrict SSH port access
