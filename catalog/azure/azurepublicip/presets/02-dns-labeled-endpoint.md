# DNS-Labeled Endpoint

This preset creates a zone-redundant Standard public IP with an Azure-managed DNS name: `{label}.{region}.cloudapp.azure.com`. The scope-hashed label (`domainNameLabelScope`) lets the same label recur safely across tenants -- Azure hashes it into the FQDN, which defends against dangling-DNS subdomain takeover when the address is later deleted.

## When to Use

- Endpoints that need a stable DNS name without managing your own DNS zone
- Dev/test environments where teams reuse the same label conventions
- Any address you would otherwise CNAME to from your own domain

## Key Configuration Choices

- **`domainNameLabel`** -- 3-63 characters: lowercase letters, digits, and hyphens; starts with a letter, ends with a letter or digit
- **`domainNameLabelScope: TENANT_REUSE`** -- the hashed label is reusable across tenants; use `NO_REUSE` for the strictest takeover defense, or omit the scope entirely for the classic region-unique label (the resulting FQDN is then exactly `{label}.{region}.cloudapp.azure.com` with no hash)
- **The `fqdn` output** carries the final Azure-assigned name -- reference it instead of reconstructing the FQDN

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match the resource this IP attaches to) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `myapp-prod` (the `domainNameLabel` value) | The DNS label to publish | Your naming convention |
