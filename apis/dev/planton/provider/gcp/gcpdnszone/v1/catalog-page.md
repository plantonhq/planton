# GCP DNS Zone

Deploys a Google Cloud DNS Managed Zone with optional DNS record creation and IAM bindings for service accounts that need to manage records in the zone. The zone is created as a public zone with the domain name derived from `metadata.name`.

## What Gets Created

When you deploy a GcpDnsZone resource, Planton provisions:

- **Cloud DNS Managed Zone** — a public managed zone in the specified GCP project, with the DNS name set to `metadata.name` (a trailing dot is appended automatically)
- **DNS Record Sets** — one `google_dns_record_set` per entry in `records`, each created as a child of the managed zone
- **IAM Binding** — created only when `iamServiceAccounts` is non-empty, grants the `roles/dns.admin` role on the project to every listed service account so they can create, update, and delete records

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A GCP project** where the managed zone will be created
- **A domain name** you own or control, used as `metadata.name` (e.g., `example.com`)
- **Service account emails** if you need automated tools like cert-manager to manage DNS records

## Quick Start

Create a file `dns-zone.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDnsZone
metadata:
  name: example.com
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpDnsZone.example-com
spec:
  projectId: my-gcp-project-123
```

Deploy:

```shell
planton apply -f dns-zone.yaml
```

This creates a public Cloud DNS managed zone for `example.com.` in the specified GCP project.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `string` or `valueFrom` | The GCP project ID where the managed zone is created. Can be a literal value or a reference to a GcpProject resource. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `iamServiceAccounts` | `string[]` | `[]` | GCP service account emails granted `roles/dns.admin` on the project. Typically used for workload identities such as cert-manager. |
| `records` | `GcpDnsRecord[]` | `[]` | DNS records to create in the managed zone. |
| `records[].recordType` | `DnsRecordType` | — | DNS record type. One of: `A`, `AAAA`, `ALIAS`, `CNAME`, `MX`, `NS`, `PTR`, `SOA`, `SRV`, `TXT`, `CAA`. |
| `records[].name` | `string` | — | Fully qualified domain name for the record (e.g., `dev.example.com.`). Must match a valid DNS name pattern. |
| `records[].values` | `string[]` | — | Record values. For `CNAME` records, each value should end with a dot. Minimum 1 item. |
| `records[].ttlSeconds` | `int32` | `60` | Time to live for the record, in seconds. |

## Examples

### Zone with a Single A Record

A DNS zone with one A record pointing a subdomain to an IP address:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDnsZone
metadata:
  name: example.com
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpDnsZone.example-com
spec:
  projectId: my-gcp-project-123
  records:
    - recordType: A
      name: app.example.com.
      values:
        - 203.0.113.10
      ttlSeconds: 300
```

### Zone with IAM Bindings for cert-manager

Grant a Kubernetes workload identity service account permission to manage DNS records, commonly used for DNS-01 ACME challenges:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDnsZone
metadata:
  name: example.com
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpDnsZone.example-com
spec:
  projectId: my-gcp-project-123
  iamServiceAccounts:
    - cert-manager@my-gcp-project-123.iam.gserviceaccount.com
```

### Full-Featured Zone with Multiple Records

Production zone with multiple record types and service account bindings:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDnsZone
metadata:
  name: example.com
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpDnsZone.example-com
spec:
  projectId: my-gcp-project-123
  iamServiceAccounts:
    - cert-manager@my-gcp-project-123.iam.gserviceaccount.com
    - external-dns@my-gcp-project-123.iam.gserviceaccount.com
  records:
    - recordType: A
      name: example.com.
      values:
        - 203.0.113.10
        - 203.0.113.11
      ttlSeconds: 300
    - recordType: CNAME
      name: www.example.com.
      values:
        - example.com.
      ttlSeconds: 300
    - recordType: MX
      name: example.com.
      values:
        - "10 mail.example.com."
        - "20 mail2.example.com."
      ttlSeconds: 3600
    - recordType: TXT
      name: example.com.
      values:
        - "v=spf1 include:_spf.google.com ~all"
      ttlSeconds: 3600
```

### Using a Foreign Key Reference for Project ID

Reference an Planton-managed GcpProject instead of hardcoding the project ID:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpDnsZone
metadata:
  name: example.com
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpDnsZone.example-com
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
  records:
    - recordType: A
      name: api.example.com.
      values:
        - 203.0.113.50
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `zone_id` | `string` | The ID of the created Cloud DNS managed zone |
| `zone_name` | `string` | The name of the managed zone (dots in the domain are replaced with hyphens) |
| `nameservers` | `string[]` | The list of nameservers assigned to the managed zone. Update your domain registrar's NS records to point to these values. |

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project where the managed zone is created
- [GcpDnsRecord](/docs/catalog/gcp/gcpdnsrecord) — manages individual DNS records as standalone resources
- [GcpCertManagerCert](/docs/catalog/gcp/gcpcertmanagercert) — provisions TLS certificates that may use DNS-01 validation against this zone
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — creates service accounts that can be added to `iamServiceAccounts`
