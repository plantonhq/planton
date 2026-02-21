# Import Existing Public Key

This preset registers an existing SSH public key in Hetzner Cloud so it can be injected into servers at creation time. Generate a keypair locally with `ssh-keygen` and import only the public half -- the private key never enters IaC state or the Hetzner Cloud API.

Hetzner Cloud supports ED25519, RSA (>= 1024-bit), and ECDSA key types. Changing the public key after creation forces replacement of the resource because Hetzner Cloud does not allow in-place updates to key material.

## When to Use

- Provisioning servers that require SSH key-based authentication
- Production environments where keys are generated and managed externally (ssh-keygen, HashiCorp Vault, 1Password SSH agent, etc.)
- Teams sharing an operational SSH key for server access

## Key Configuration Choices

- **Import mode** (`publicKey` set) -- brings an existing key into Hetzner Cloud rather than generating one server-side (Hetzner Cloud does not support server-side generation)
- **No private key in state** -- because only the public key is imported, the IaC state file contains no sensitive material

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<ssh-public-key>` | SSH public key in OpenSSH authorized_keys format | Output of `cat ~/.ssh/id_ed25519.pub` (ED25519), `cat ~/.ssh/id_rsa.pub` (RSA), or your key management system |
