# Password-Authenticated Drop Box

This preset creates a write-only "drop box" identity: a sender uploads
files over SFTP with a generated password but can never read, list, or
delete anything -- the receive-only posture for inbound feeds from
less-technical counterparties.

## When to Use

- Inbound data feeds from counterparties who can only do
  username+password SFTP
- Receive-only intake where the sender must not see other senders'
  files (or its own past uploads)

## Key Configuration Choices

- **Azure generates the password** -- it lands in the `password`
  output EXACTLY ONCE; there is no way to choose or retrieve it later
  (regenerate by flipping `sshPasswordEnabled` off and on)
- **write + create WITHOUT read/list/delete** is the drop-box grant --
  uploads succeed, everything else is denied
- Prefer the key-auth preset whenever the counterparty can manage a key
  pair

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<storage-account-resource-name>` | The SFTP AzureStorageAccount's Planton resource name | Your exchange composition |
| `vendordropbox` | 3-64 lowercase letters and digits | Your counterparty naming convention |
| `<containerName>` / `<container-resource-name>` | The intake container | Your exchange composition |
