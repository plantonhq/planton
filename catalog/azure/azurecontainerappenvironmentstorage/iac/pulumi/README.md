# AzureContainerAppEnvironmentStorage - Pulumi Module

Pulumi implementation for the AzureContainerAppEnvironmentStorage
kind.

## Architecture

```
containerapp.EnvironmentStorage (one file-share registration)
```

## Key Design Decisions

- **Exactly one protocol** -- SMB (`account_name` + `access_key`) XOR
  NFS (`nfs_server_url`), branched explicitly in Go; the NFS path
  requires a VNet-injected environment.
- **Access mode maps to ARM wire values** (`READ_ONLY` -> `ReadOnly`,
  `READ_WRITE` -> `ReadWrite`) so a vocabulary drift fails loudly at
  preview.
- **Only the SMB access key updates in place** (rotation); every other
  change recreates the registration and briefly breaks active mounts --
  the module keeps that ForceNew surface honest.
- **No tags** -- ARM does not support tags on
  `managedEnvironments/storages`.

## Provider

Built via the shared `pulumiazureprovider.Get` builder -- static client
secret, keyless web identity, or ambient chain, resolved from the IaC
input. Never construct the provider inline.
