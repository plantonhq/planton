# Folder Dedicated Key Project

## Use Case

Give every project in a folder customer-managed keys that all live in one key project only the key administrators control -- the separation-of-duties model.

## When to Use

- A landing-zone folder of production workloads under a central security team
- Compliance regimes that require keys to sit apart from the data they protect

## What This Creates

- The Cloud KMS API on the key project
- The folder's Autokey configuration pointing at the key project, with destroy blocked

Before the first key handle, create the key project's Cloud KMS service agent and grant it `roles/cloudkms.admin` on the key project (see the GUIDE).

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `scope.folderId` | `123456789012` | Reference a `GcpFolder`. |
| `keyProject` | `security-keys` | Reference the key project's `GcpProject`. |
| `deletionPolicy` | `PREVENT` | `DELETE` only when retiring Autokey for the folder on purpose. |
