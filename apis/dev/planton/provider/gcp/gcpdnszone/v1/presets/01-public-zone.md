# Public DNS Zone

This preset creates a Cloud DNS managed zone with IAM permissions granted to service accounts for cert-manager and external-dns. DNS records are managed separately via `GcpDnsRecord` resources, keeping the zone definition clean and composable.

## When to Use

- Hosting a public DNS zone for your domain on Google Cloud DNS
- Environments where cert-manager needs DNS01 challenge access for TLS certificates
- Environments where external-dns automatically manages DNS records from Kubernetes ingresses

## Key Configuration Choices

- **Zone-only** -- no inline DNS records; records are managed via standalone `GcpDnsRecord` resources
- **IAM for automation** (`iamServiceAccounts`) -- grants DNS record management permissions to cert-manager and external-dns service accounts
- **Zone name derived from metadata** -- the `metadata.name` is used as the zone name (must match your domain in kebab-case)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<cert-manager-sa-email>` | Service account email for cert-manager | `GcpServiceAccount` outputs or GKE Workload Identity setup |
| `<external-dns-sa-email>` | Service account email for external-dns | `GcpServiceAccount` outputs or GKE Workload Identity setup |
