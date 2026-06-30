# DigitalOcean Certificate - Pulumi Module

This Pulumi module provisions and manages SSL/TLS certificates on DigitalOcean using the `pulumi-digitalocean` provider.

## Overview

The module implements Planton's `DigitalOceanCertificate` protobuf spec, providing a type-safe interface for:

- **Let's Encrypt Certificates**: Free, auto-renewing certificates (requires DigitalOcean DNS)
- **Custom Certificates**: User-provided certificates (BYOC - Bring Your Own Certificate)

## Module Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint (calls module.Resources)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build and deployment helpers
├── debug.sh             # Debug script for local testing
├── module/
│   ├── main.go          # Module entry point (Resources function)
│   ├── locals.go        # Local variables and context initialization
│   ├── certificate.go   # Certificate resource creation logic
│   └── outputs.go       # Output constant definitions
└── README.md            # This file
```

## Prerequisites

- **Pulumi CLI**: Install from [pulumi.com/docs/install](https://www.pulumi.com/docs/install/)
- **Go**: Version 1.21 or later
- **DigitalOcean API Token**: Set via environment variable `DIGITALOCEAN_TOKEN` or Pulumi config

## Quick Start

### 1. Set Up Environment

```bash
# Navigate to the Pulumi module directory
cd apis/dev/planton/provider/digitalocean/digitaloceancertificate/v1/iac/pulumi

# Set your DigitalOcean API token
export DIGITALOCEAN_TOKEN="your-digitalocean-api-token"
```

### 2. Configure Stack Input

Create a `stack-input.yaml` file with your certificate specification:

**Let's Encrypt Example:**
```yaml
provider_config:
  credential_id: "digitalocean-prod-credential"

target:
  spec:
    certificate_name: prod-web-cert
    type: letsEncrypt
    lets_encrypt:
      domains:
        - example.com
        - www.example.com
    description: "Production web certificate"
    tags:
      - env:production
```

**Custom Certificate Example:**
```yaml
provider_config:
  credential_id: "digitalocean-prod-credential"

target:
  spec:
    certificate_name: prod-custom-cert
    type: custom
    custom:
      private_key: |
        -----BEGIN PRIVATE KEY-----
        ...
        -----END PRIVATE KEY-----
      leaf_certificate: |
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
      certificate_chain: |
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
    description: "Custom EV certificate"
```

### 3. Deploy

```bash
# Initialize Pulumi stack
pulumi stack init dev

# Preview changes
pulumi preview --stack-input stack-input.yaml

# Deploy
pulumi up --stack-input stack-input.yaml
```

### 4. Retrieve Outputs

```bash
# Get certificate ID
pulumi stack output certificate_id

# Get expiration timestamp
pulumi stack output expiry_rfc3339
```

## Module API

### Input: DigitalOceanCertificateStackInput

The module expects a `DigitalOceanCertificateStackInput` protobuf message containing:

```go
type DigitalOceanCertificateStackInput struct {
    ProviderConfig *ProviderConfig                     // DigitalOcean credentials
    Target         *DigitalOceanCertificate            // Certificate specification
}

type DigitalOceanCertificate struct {
    Metadata *CloudResourceMetadata                    // Resource metadata
    Spec     *DigitalOceanCertificateSpec              // Certificate spec
}
```

### Output: Stack Outputs

The module exports the following Pulumi outputs:

| Output | Type | Description |
|--------|------|-------------|
| `certificate_id` | string | DigitalOcean certificate UUID |
| `expiry_rfc3339` | string | Certificate expiration timestamp (RFC 3339) |

## Discriminated Union Implementation

The module correctly handles the discriminated union pattern for certificate types:

```go
// In certificate.go
if locals.DigitalOceanCertificate.Spec.Type == digitaloceancertificatev1.DigitalOceanCertificateType_letsEncrypt {
    // Configure Let's Encrypt parameters
    certArgs.Domains = domains
}

if locals.DigitalOceanCertificate.Spec.Type == digitaloceancertificatev1.DigitalOceanCertificateType_custom {
    // Configure custom certificate parameters
    certArgs.LeafCertificate = pulumi.String(locals.DigitalOceanCertificate.Spec.GetCustom().LeafCertificate)
    certArgs.PrivateKey = pulumi.String(locals.DigitalOceanCertificate.Spec.GetCustom().PrivateKey)
    certArgs.CertificateChain = pulumi.StringPtr(locals.DigitalOceanCertificate.Spec.GetCustom().CertificateChain)
}
```

This ensures type safety and prevents invalid configurations at compile-time.

## Secret Management

### Using Pulumi Config Secrets

Pulumi provides built-in secret encryption for sensitive values like private keys:

```bash
# Set secrets (automatically encrypted in state)
pulumi config set --secret privateKey "$(cat privkey.pem)"
pulumi config set --secret leafCert "$(cat cert.pem)"
pulumi config set --secret certChain "$(cat chain.pem)"
```

Then reference them in your Go code:

```go
config := pulumi.Config(ctx)
privateKey := config.RequireSecret("privateKey")
leafCert := config.RequireSecret("leafCert")
certChain := config.RequireSecret("certChain")
```

### Using External Secret Managers

For production, integrate with Vault, AWS Secrets Manager, or K8s External Secrets:

```go
// Example: Reading from Vault
vaultSecret, err := vault.ReadSecret(ctx, "secret/digitalocean/prod-cert")
if err != nil {
    return err
}

certArgs.PrivateKey = pulumi.String(vaultSecret.Data["private_key"])
certArgs.LeafCertificate = pulumi.String(vaultSecret.Data["leaf_certificate"])
```

## Development Workflow

### Local Testing

Use the provided `debug.sh` script for rapid iteration:

```bash
# Edit debug.sh to configure your test inputs
./debug.sh
```

### Building

```bash
# Build the module
make build

# Run tests
make test
```

### Debugging

Enable Pulumi debug logging:

```bash
export PULUMI_DEBUG_COMMANDS=true
export PULUMI_DEBUG_GRPC=debug.log
pulumi up --logtostderr -v=9
```

## Limitations and Notes

### Current Limitations

1. **Tags Not Supported**: The DigitalOcean Pulumi provider does not expose a `tags` field for certificates. The spec's `tags` field is ignored (noted in code comments).

2. **Auto-Renew Not Configurable**: The `disable_auto_renew` field in the spec is not supported by the provider. Let's Encrypt certificates always auto-renew.

### Provider Quirks

- **Let's Encrypt State**: Certificates transition through `pending` → `verified` states. The module exports outputs immediately, but the certificate may not be usable until `state = verified`.

- **Custom Certificate Chain**: The `certificate_chain` field is optional in the DigitalOcean API, but omitting it can cause browser trust issues. **Always provide it**.

## Integration Examples

### With Load Balancers

Reference the certificate in a DigitalOcean Load Balancer:

```go
lb := digitalocean.NewLoadBalancer(ctx, "lb", &digitalocean.LoadBalancerArgs{
    ForwardingRules: digitalocean.LoadBalancerForwardingRuleArray{
        &digitalocean.LoadBalancerForwardingRuleArgs{
            EntryPort:       pulumi.Int(443),
            EntryProtocol:   pulumi.String("https"),
            TargetPort:      pulumi.Int(80),
            TargetProtocol:  pulumi.String("http"),
            CertificateId:   createdCertificate.ID(),
        },
    },
})
```

### Zero-Downtime Rotation

Pulumi's dependency-aware engine automatically performs `create → update → delete` for resource replacements:

```go
// When certificate materials change, Pulumi will:
// 1. Create a new certificate with new UUID
// 2. Update Load Balancer to reference new UUID
// 3. Delete old certificate
// No manual lifecycle configuration needed!
```

## Troubleshooting

### Certificate Stuck in "Pending"

**Symptom**: Let's Encrypt certificate never reaches `verified` state.

**Cause**: DNS is not managed by DigitalOcean.

**Solution**: Verify DNS zone exists in DigitalOcean DNS. If DNS is external, use the Custom certificate workflow instead.

### "Invalid PEM Format" Error

**Symptom**: Custom certificate fails with PEM validation error.

**Cause**: PEM blocks must have correct headers/footers and base64 content.

**Solution**: Validate PEM format:
```bash
openssl x509 -in cert.pem -text -noout  # Validate certificate
openssl rsa -in privkey.pem -check      # Validate private key
```

### Outputs Not Appearing

**Symptom**: `pulumi stack output` returns empty.

**Cause**: Outputs are only exported after successful resource creation.

**Solution**: Check deployment logs for errors:
```bash
pulumi logs --stack dev
```

## Further Reading

- **Component Overview**: See [../../README.md](../../README.md)
- **Comprehensive Guide**: See [../../docs/README.md](../../docs/README.md)
- **Examples**: See [../../examples.md](../../examples.md)
- **Module Architecture**: See [overview.md](./overview.md)

## Support

For module-specific issues:
1. Check Pulumi logs: `pulumi logs --stack <stack-name>`
2. Enable debug logging: `pulumi up --logtostderr -v=9`
3. Refer to [DigitalOcean Pulumi Docs](https://www.pulumi.com/registry/packages/digitalocean/)

For component issues, see the main [README.md](../../README.md).

