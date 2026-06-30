# CivoCertificate Pulumi Module

## Overview

This directory contains the Pulumi implementation for Civo TLS certificates. The module provides validation and structure for certificate specifications but **cannot provision actual certificates** as the Civo Pulumi provider does not currently expose certificate resources.

**Provider Limitation**: As of 2025, neither the Civo Terraform nor Pulumi provider supports certificate resources. This module validates specifications and logs appropriate warnings. See: https://registry.terraform.io/providers/civo/civo/latest/docs

## Module Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint (stack initialization)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── debug.sh             # Local debugging script
├── README.md            # This file
└── module/              # Core module implementation
    ├── main.go          # Module entry point (Resources function)
    ├── locals.go        # Local variables and initialization
    ├── outputs.go       # Output constant definitions
    └── certificate.go   # Certificate validation and logging
```

## Prerequisites

### Required Tools

- **Go**: 1.21 or later
- **Pulumi CLI**: 3.x or later
- **Civo Account**: With API credentials (for future when provider support is added)

### Civo Provider

The module references the [Pulumi Civo Provider](https://www.pulumi.com/registry/packages/civo/), though certificate resources are not yet available.

## How It Works

### 1. Stack Input

The module receives a `CivoCertificateStackInput` containing:
- `CivoCertificate` resource (metadata + spec)
- Civo provider configuration (API token, region)

### 2. Validation and Logging

The module:
1. **Validates** the certificate specification against buf.validate rules
2. **Initializes** locals (metadata, labels, tags)
3. **Logs warnings** about provider limitations
4. **Logs configuration** details (domains, type, auto-renewal settings)
5. **Exports placeholder outputs** for future use

### 3. Stack Outputs

When provider support is added, the module will export:
- `certificate_id`: Unique identifier (UUID)
- `expiry_rfc3339`: Certificate expiration timestamp

Currently, these are exported as empty strings.

## Local Development

### Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/plantonhq/planton.git
   cd planton/apis/dev/planton/provider/civo/civocertificate/v1/iac/pulumi/
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Set Civo credentials** (for future use):
   ```bash
   export CIVO_TOKEN=your-civo-api-token
   ```

### Running Locally

Use the provided `debug.sh` script for local testing:

```bash
./debug.sh
```

This script:
- Loads the example manifest from `../../hack/manifest.yaml`
- Initializes a local Pulumi stack
- Runs `pulumi up` with the test configuration
- Displays validation results and warnings

**Expected Output**: Warnings about provider limitations, but successful validation.

## Module Implementation Details

### `module/main.go`

Entry point for the module. Calls:
1. `initializeLocals()` - Prepares local variables
2. `certificate()` - Validates and logs certificate configuration

### `module/locals.go`

Initializes the `Locals` struct containing:
- `CivoCertificate`: The Protobuf spec
- `Metadata`: Cloud resource metadata (name, labels)
- `Labels`: Combined metadata labels and spec tags

### `module/outputs.go`

Defines output constants:
- `OpCertificateId`: "certificate_id"
- `OpExpiryRfc3339`: "expiry_rfc3339"

### `module/certificate.go`

Validates certificate configuration and logs:
- Warning about provider limitation
- Certificate type (Let's Encrypt or custom)
- Domains list (for Let's Encrypt)
- Auto-renewal status
- Exports placeholder outputs

## Current Behavior

### Let's Encrypt Certificate

```
⚠️  Certificate 'prod-cert' specification is valid but cannot be provisioned.
    The Civo Pulumi/Terraform provider does not currently support certificate resources.
    Certificates must be managed manually via the Civo dashboard or API.

ℹ️  Let's Encrypt certificate requested for domains: [example.com www.example.com] (auto-renew: true)
```

### Custom Certificate

```
⚠️  Certificate 'custom-cert' specification is valid but cannot be provisioned.
    The Civo Pulumi/Terraform provider does not currently support certificate resources.

ℹ️  Custom certificate provided for 'custom-cert'
```

## Testing

### Unit Tests

Run Go unit tests:
```bash
cd ../../  # Back to v1/ directory
go test -v
```

Tests validate:
- Certificate name length (1-64 chars)
- Type enumeration
- Domain patterns (FQDN and wildcard)
- Custom certificate PEM requirements
- Tags uniqueness
- Description length

### Integration Testing

Since provider support is missing, full integration testing is not possible. When support is added:

```bash
export CIVO_TOKEN=your-token
./debug.sh
```

This will provision actual certificates.

## Workarounds for Certificate Management

Until provider support is added, use these alternatives:

### 1. Civo Dashboard

Navigate to: https://dashboard.civo.com/certificates

1. Click "Add Certificate"
2. Choose Let's Encrypt or Custom
3. Fill in the form matching your spec
4. Manually track certificate IDs

### 2. Civo API

Use `curl` or HTTP client to call Civo API:

```bash
# Create Let's Encrypt certificate
curl -X POST https://api.civo.com/v2/certificates \
  -H "Authorization: Bearer $CIVO_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "prod-cert",
    "domains": ["example.com", "www.example.com"]
  }'
```

### 3. Civo CLI

```bash
# Install
brew install civo

# Login
civo apikey add my-key $CIVO_TOKEN

# Create certificate
civo certificate create prod-cert \
  --domains example.com,www.example.com
```

## Tracking Provider Support

**Monitor these sources for certificate resource support**:
- **Terraform Provider**: https://registry.terraform.io/providers/civo/civo/latest/docs
- **Pulumi Provider**: https://www.pulumi.com/registry/packages/civo/
- **Civo GitHub**: https://github.com/civo/terraform-provider-civo

When support is added, this module will be updated to provision actual certificates.

## Future Implementation

When the Civo provider adds certificate support, the `certificate.go` file will be updated to:

```go
// Future implementation (example)
cert, err := civo.NewCertificate(ctx, "cert", &civo.CertificateArgs{
    Name:    pulumi.String(locals.CivoCertificate.Spec.CertificateName),
    Domains: pulumi.ToStringArray(locals.CivoCertificate.Spec.GetLetsEncrypt().Domains),
    // ... other args
}, pulumi.Provider(civoProvider))

// Export real outputs
ctx.Export(OpCertificateId, cert.ID())
ctx.Export(OpExpiryRfc3339, cert.Expiry)
```

## Security Best Practices

1. **Store private keys securely**: Use Pulumi secrets for custom certificates
   ```bash
   pulumi config set --secret custom-cert-key "$(cat private-key.pem)"
   ```

2. **Never commit secrets**: Use `.gitignore` for PEM files

3. **Use separate credentials**: Different Civo API tokens per environment

4. **Rotate regularly**: Even with Let's Encrypt auto-renewal, monitor and validate

## Performance Considerations

- **Validation**: Instant (no network calls)
- **Future provisioning**: ~30-90 seconds for Let's Encrypt (when supported)
- **Custom cert upload**: ~5-10 seconds (when supported)

## Related Documentation

- **API Specification**: [../../README.md](../../README.md)
- **Examples**: [../../examples.md](../../examples.md)
- **Research**: [../../docs/README.md](../../docs/README.md)
- **Pulumi Civo Provider**: [pulumi.com/registry/packages/civo](https://www.pulumi.com/registry/packages/civo/)

## Support

- **Issues**: [GitHub Issues](https://github.com/plantonhq/planton/issues)
- **Discussions**: [GitHub Discussions](https://github.com/plantonhq/planton/discussions)
- **Civo Support**: [support.civo.com](https://support.civo.com)

## Contributing

When contributing to this module:

1. Follow [Planton contribution guidelines](../../../../../../../../CONTRIBUTING.md)
2. Ensure all tests pass: `go test -v`
3. Run linters: `golangci-lint run`
4. Update this README if adding features
5. Add tests for new validation rules

---

**Note**: This module is ready for certificate provisioning once the Civo provider adds support. The structure, validation, and outputs are complete.
