# From CSR

## Use Case

Sign a certificate signing request someone else produced -- a partner's client, a device, an appliance -- without their private key ever leaving them.

## When to Use

- Systems that generate their own key and CSR
- Year-long client or device certificates that must be revocable

## What This Creates

- A one-year certificate from the `internal-servers` pool for the CSR's subject and names, as the pool's policy allows

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `pemCsr` | a placeholder | The requester's CSR, PEM. |
| `lifetime` | `31536000s` (1 year) | Shorter for anything you cannot revoke quickly enough. |
