---
title: "Import Existing Public Key"
description: "This preset imports an existing SSH public key into OpenStack. This is the recommended approach for production -- generate a keypair locally with `ssh-keygen` and import only the public key. The..."
type: "preset"
rank: "01"
presetSlug: "01-import-public-key"
componentSlug: "keypair"
componentTitle: "Keypair"
provider: "openstack"
icon: "package"
order: 1
---

# Import Existing Public Key

This preset imports an existing SSH public key into OpenStack. This is the recommended approach for production -- generate a keypair locally with `ssh-keygen` and import only the public key. The private key never leaves your machine or IaC state.

## When to Use

- Any compute instance that needs SSH key-based authentication
- Production environments where key management is handled externally (ssh-keygen, Vault, etc.)
- Teams that maintain a shared SSH key for operational access

## Key Configuration Choices

- **Import mode** (`publicKey` set) -- brings an existing key into OpenStack rather than generating one
- **No private key in state** -- because only the public key is imported, the IaC state file contains no sensitive material

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<ssh-public-key>` | SSH public key in OpenSSH format (e.g., `ssh-rsa AAAAB3Nza... user@host`) | Output of `cat ~/.ssh/id_rsa.pub` or your key management system |
