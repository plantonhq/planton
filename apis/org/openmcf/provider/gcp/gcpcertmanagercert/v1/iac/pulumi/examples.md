# Pulumi Examples - GCP Cert Manager Cert

This document provides examples of using the Pulumi implementation directly (without OpenMCF CLI).

## Prerequisites

- Pulumi CLI installed
- Go 1.21+
- GCP credentials configured
- Cloud DNS zone created

## Basic Pulumi Usage

### 1. Initialize Stack

```bash
cd apis/project/planton/provider/gcp/gcpcertmanagercert/v1/iac/pulumi
pulumi stack init my-org/my-project/dev
```

### 2. Create Stack Input File

Create `stack-input.json`:

```json
{
  "target": {
    "apiVersion": "gcp.openmcf.org/v1",
    "kind": "GcpCertManagerCert",
    "metadata": {
      "name": "my-cert",
      "id": "cert-001",
      "org": "my-org",
      "env": {
        "id": "production"
      }
    },
    "spec": {
      "gcpProjectId": "my-gcp-project",
      "primaryDomainName": "example.com",
      "alternateDomainNames": ["www.example.com"],
      "cloudDnsZoneId": {
        "value": "example-com-zone"
      },
      "certificateType": 0
    }
  },
  "providerConfig": {
    "serviceAccountKeyBase64": "base64-encoded-key"
  }
}
```

### 3. Preview Changes

```bash
STACK_INPUT_FILE=stack-input.json pulumi preview
```

### 4. Deploy

```bash
STACK_INPUT_FILE=stack-input.json pulumi up
```

### 5. View Outputs

```bash
pulumi stack output certificate-id
pulumi stack output certificate-name
```

### 6. Destroy

```bash
pulumi destroy
```

## Example: Certificate Manager Certificate

```json
{
  "target": {
    "apiVersion": "gcp.openmcf.org/v1",
    "kind": "GcpCertManagerCert",
    "metadata": {
      "name": "cert-manager-example",
      "id": "cm-cert-001",
      "org": "acme-corp",
      "env": {
        "id": "production"
      }
    },
    "spec": {
      "gcpProjectId": "acme-production",
      "primaryDomainName": "api.acme.com",
      "alternateDomainNames": [
        "www.api.acme.com",
        "*.services.acme.com"
      ],
      "cloudDnsZoneId": {
        "value": "acme-com-zone"
      },
      "certificateType": 0,
      "validationMethod": "DNS"
    }
  },
  "providerConfig": {
    "serviceAccountKeyBase64": "YOUR_BASE64_KEY_HERE"
  }
}
```

## Example: Load Balancer Certificate

```json
{
  "target": {
    "apiVersion": "gcp.openmcf.org/v1",
    "kind": "GcpCertManagerCert",
    "metadata": {
      "name": "lb-cert-example",
      "id": "lb-cert-001",
      "org": "acme-corp",
      "env": {
        "id": "production"
      }
    },
    "spec": {
      "gcpProjectId": "acme-production",
      "primaryDomainName": "lb.acme.com",
      "alternateDomainNames": ["www.lb.acme.com"],
      "cloudDnsZoneId": {
        "value": "acme-com-zone"
      },
      "certificateType": 1
    }
  },
  "providerConfig": {
    "serviceAccountKeyBase64": "YOUR_BASE64_KEY_HERE"
  }
}
```

## Example: External DNS (No Cloud DNS Zone)

When your DNS zone is hosted outside GCP (e.g. AWS Route 53), omit `cloudDnsZoneId`. The module creates DNS authorizations but skips record insertion. Validation records are exported as the `dns-validation-records` stack output.

```json
{
  "target": {
    "apiVersion": "gcp.openmcf.org/v1",
    "kind": "GcpCertManagerCert",
    "metadata": {
      "name": "external-dns-example",
      "id": "ext-cert-001",
      "org": "acme-corp",
      "env": {
        "id": "production"
      }
    },
    "spec": {
      "gcpProjectId": "acme-production",
      "primaryDomainName": "api.acme.com",
      "alternateDomainNames": [
        "www.api.acme.com"
      ],
      "certificateType": 0,
      "validationMethod": "DNS"
    }
  },
  "providerConfig": {
    "serviceAccountKeyBase64": "YOUR_BASE64_KEY_HERE"
  }
}
```

After deploying, read the validation records and create them in your DNS provider:

```bash
pulumi stack output dns-validation-records
```

## Using Makefile

### Preview

```bash
make preview
```

### Deploy

```bash
make up
```

### Destroy

```bash
make destroy
```

## Debugging

Use the debug script for local development:

```bash
./debug.sh
```

This will:
- Build the Pulumi program
- Run a preview
- Show what changes would be made

## Environment Variables

Set these before running Pulumi:

```bash
export PULUMI_CONFIG_PASSPHRASE="your-passphrase"
export GOOGLE_CREDENTIALS="path/to/service-account-key.json"
export STACK_INPUT_FILE="stack-input.json"
```

## Certificate Types

### Type 0: MANAGED (Certificate Manager)

- Uses Google Certificate Manager
- Creates DNS authorizations
- Adds validation records to Cloud DNS
- Recommended for most use cases

### Type 1: LOAD_BALANCER (SSL Certificate)

- Uses Google-managed SSL certificates
- Optimized for load balancers
- Simpler validation process
- Best for LB-specific scenarios

## Checking Results

After deployment, verify in GCP Console:

1. **Certificate Manager**: Go to Certificate Manager section
2. **DNS Records**: Check Cloud DNS for validation records
3. **Status**: Verify certificate status is ACTIVE

Or use gcloud CLI:

```bash
# List certificates
gcloud certificate-manager certificates list --project=my-project

# Describe specific certificate
gcloud certificate-manager certificates describe my-cert --project=my-project
```

## Troubleshooting

### DNS Validation Issues

If validation is pending:
- Check DNS records are created correctly
- Verify domain ownership
- Wait for DNS propagation (can take up to 10 minutes)

### Permission Errors

Ensure service account has:
- `roles/certificatemanager.editor`
- `roles/dns.admin`
- `roles/compute.loadBalancerAdmin` (for LB certs)

### Build Errors

```bash
go mod download
go mod tidy
```

## Next Steps

- See main [README.md](README.md) for architecture details
- Check [overview.md](../overview.md) for IaC comparison
- Use OpenMCF CLI for production deployments

