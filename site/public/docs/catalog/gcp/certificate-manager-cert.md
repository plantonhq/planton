---
title: "Certificate Manager Cert"
description: "Certificate Manager Cert deployment documentation"
icon: "package"
order: 100
componentName: "gcpcertmanagercert"
---

# GCP Certificate Manager Cert

Provisions a Google-managed SSL/TLS certificate with automatic DNS validation through Cloud DNS. The component supports two certificate backends: Certificate Manager (newer, with DNS authorization) and classic Google-managed SSL certificates for load balancers.

## What Gets Created

When you deploy a GcpCertManagerCert resource, OpenMCF provisions:

- **Certificate Manager DNS Authorizations** (MANAGED type) — one `google_certificate_manager_dns_authorization` per domain (primary + alternates), each proving domain ownership
- **Cloud DNS Validation Records** (MANAGED type) — one `google_dns_record_set` per domain in the specified Cloud DNS zone, populated automatically from the DNS authorization challenge data
- **Certificate Manager Certificate** (MANAGED type) — a `google_certificate_manager_certificate` with managed DNS authorization references covering all specified domains
- **Google-Managed SSL Certificate** (LOAD_BALANCER type) — a `google_compute_managed_ssl_certificate` with the specified domains, designed for use with GCP load balancers

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** where the certificate and DNS resources will be created
- **A Cloud DNS managed zone** that is authoritative for the domain(s) you are requesting certificates for
- **DNS configured** so the Cloud DNS zone is serving live DNS for the domain (nameservers delegated at the registrar)
- **IAM permissions** to create Certificate Manager resources, Compute SSL certificates, and DNS record sets in the target project

## Quick Start

Create a file `cert.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCertManagerCert
metadata:
  name: my-cert
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpCertManagerCert.my-cert
spec:
  gcpProjectId: my-gcp-project-123
  primaryDomainName: example.com
  cloudDnsZoneId:
    value: example-com
```

Deploy:

```shell
openmcf apply -f cert.yaml
```

This creates a Certificate Manager certificate for `example.com` with automatic DNS validation records in the `example-com` Cloud DNS zone.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `gcpProjectId` | `string` | GCP project ID where the certificate and DNS resources are created. | Required |
| `primaryDomainName` | `string` | Main domain name for the certificate. Supports apex domains (`example.com`) and wildcards (`*.example.com`). | Required, must match domain pattern |
| `cloudDnsZoneId` | `string` or `valueFrom` | Cloud DNS managed zone ID where validation records are created. Can reference a GcpDnsZone resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `alternateDomainNames` | `string[]` | `[]` | Subject Alternative Names (SANs) for the certificate. Each entry follows the same pattern as `primaryDomainName`. Must not contain duplicates or repeat the primary domain. |
| `certificateType` | `enum` | `MANAGED` | Type of certificate to create. `MANAGED` uses Certificate Manager with DNS authorization. `LOAD_BALANCER` uses classic Google-managed SSL certificates for load balancers. |
| `validationMethod` | `string` | `DNS` | Domain ownership validation method. Currently only `DNS` is supported. |

## Examples

### Single-Domain Certificate

A basic certificate for a single apex domain:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCertManagerCert
metadata:
  name: api-cert
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCertManagerCert.api-cert
spec:
  gcpProjectId: my-prod-project
  primaryDomainName: api.example.com
  cloudDnsZoneId:
    value: example-com
```

### Wildcard Certificate with Alternate Domains

A wildcard certificate that also covers the apex domain and a subdomain:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCertManagerCert
metadata:
  name: wildcard-cert
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCertManagerCert.wildcard-cert
spec:
  gcpProjectId: my-prod-project
  primaryDomainName: "*.example.com"
  alternateDomainNames:
    - example.com
    - "*.staging.example.com"
  cloudDnsZoneId:
    value: example-com
  certificateType: MANAGED
```

### Load Balancer SSL Certificate

A classic Google-managed SSL certificate for use with GCP load balancers:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCertManagerCert
metadata:
  name: lb-cert
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCertManagerCert.lb-cert
spec:
  gcpProjectId: my-prod-project
  primaryDomainName: app.example.com
  alternateDomainNames:
    - www.example.com
  cloudDnsZoneId:
    value: example-com
  certificateType: LOAD_BALANCER
```

### Using Foreign Key References

Reference an OpenMCF-managed GcpDnsZone instead of hardcoding the zone ID:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCertManagerCert
metadata:
  name: ref-cert
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpCertManagerCert.ref-cert
spec:
  gcpProjectId: my-prod-project
  primaryDomainName: "*.example.com"
  alternateDomainNames:
    - example.com
  cloudDnsZoneId:
    valueFrom:
      kind: GcpDnsZone
      name: example.com
      fieldPath: status.outputs.zone_name
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `certificate_id` | `string` | The identifier of the created certificate resource. For Certificate Manager: the certificate ID. For Load Balancer: the SSL certificate ID. |
| `certificate_name` | `string` | The full resource name of the created certificate (e.g., `projects/my-project/locations/global/certificates/my-cert`). |
| `certificate_domain_name` | `string` | The primary domain name for which the certificate was issued. |
| `certificate_status` | `string` | The provisioning status of the certificate. Possible values include `ACTIVE`, `PROVISIONING`, `FAILED`. |

## Related Components

- [GcpDnsZone](/docs/catalog/gcp/dns-zone) — provides the Cloud DNS managed zone where validation records are created
- [GcpGkeCluster](/docs/catalog/gcp/gke-cluster) — GKE clusters that may use certificates for ingress TLS termination
- [GcpServiceAccount](/docs/catalog/gcp/service-account) — service accounts that may need permissions to manage certificates
