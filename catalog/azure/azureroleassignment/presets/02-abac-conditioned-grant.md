# ABAC-Conditioned Data Grant

This preset grants a data-plane role narrowed by an Azure attribute-based
access control (ABAC) condition -- the role's permissions apply only when the
condition evaluates true. The template shows the canonical pattern: blob read
access limited to blobs carrying a specific tag, so one storage account can
serve multiple teams with tag-level isolation instead of per-team accounts.

## When to Use

- Restricting data access within a shared resource (per-project blobs in one
  storage account)
- Meeting least-privilege requirements that resource-level scoping cannot
  express
- Gradually tightening broad grants without re-architecting storage layout

## Key Configuration Choices

- **Condition syntax** -- conditions are supported on storage data-plane roles
  and a growing set of others; an unsupported role + condition pair fails at
  deploy with a descriptive ARM error
- **Condition version** -- `"2.0"` is the generally available syntax; Azure
  applies it by default when the version is omitted
- **Principal as literal** -- this preset passes the object ID directly (users
  and groups are managed in Azure AD, not deployed by charts); switch to a
  `valueFrom` reference when the grantee is a deployed managed identity

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-storage-account>` | Metadata name of the AzureStorageAccount being granted on | Your Infra Chart / resource list |
| `<principal-object-id>` | Azure AD OBJECT ID of the grantee (not the client ID) | Entra ID portal → the user/group/identity's Object ID |
| `<tag-name>` / `<tag-value>` | Blob index tag the condition matches | Your data-classification convention |
| `<why-this-grant-exists>` | Audit note shown in the portal's IAM blade | Your runbook / change ticket |
