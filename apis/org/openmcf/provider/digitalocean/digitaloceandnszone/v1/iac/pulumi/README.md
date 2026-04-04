# DigitalOcean DNS Zone - Pulumi Implementation

## Overview

This directory contains the Pulumi (Go) implementation for provisioning DigitalOcean DNS Zones through the OpenMCF API. The implementation reads a `DigitalOceanDnsZone` protobuf manifest and deploys the corresponding DNS zone and records using the DigitalOcean Pulumi provider.

## Architecture

### Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi entrypoint - loads manifest and invokes module
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment automation
├── debug.sh          # Helper script for local debugging
├── README.md         # This file
└── module/
    ├── main.go       # Module initialization and orchestration
    ├── locals.go     # Shared state and helper functions
    ├── outputs.go    # Stack output definitions
    └── dns_zone.go   # Core DNS zone and record provisioning logic
```

### Component Flow

1. **main.go**: Entry point that:
   - Loads the `DigitalOceanDnsZone` manifest from YAML/JSON
   - Initializes the Pulumi stack
   - Calls the module package to provision resources

2. **module/main.go**: Orchestrates the deployment:
   - Creates the DigitalOcean provider with API credentials
   - Initializes locals (shared state)
   - Calls `dnsZone()` to create the domain and records

3. **module/dns_zone.go**: Core provisioning logic:
   - Creates the DigitalOcean domain (zone)
   - Iterates over spec.records to create DNS records
   - Handles record-type-specific fields (priority, weight, port, flags, tag)
   - Exports stack outputs

4. **module/outputs.go**: Defines stack outputs:
   - Zone name
   - Zone ID
   - DigitalOcean nameservers

## Prerequisites

### Required Tools

- **Go 1.21+**: For building and running Pulumi programs
- **Pulumi CLI**: Install from https://www.pulumi.com/docs/get-started/install/
- **DigitalOcean API Token**: Generate from https://cloud.digitalocean.com/account/api/tokens

### Authentication

Set the DigitalOcean API token as an environment variable:

```bash
export DIGITALOCEAN_TOKEN="dop_v1_xxxxxxxxxxxxxxxxxxxx"
```

Alternatively, configure it as a Pulumi secret:

```bash
pulumi config set digitalocean:token dop_v1_xxxxxxxxxxxxxxxxxxxx --secret
```

## Usage

### Local Development

#### 1. Prepare Manifest

Create a manifest file (e.g., `manifest.yaml`) with your DNS zone specification:

```yaml
apiVersion: code2cloud.planton.ai/v1
kind: DigitalOceanDnsZone
metadata:
  name: my-website-dns
spec:
  digitalOceanCredentialId: do-prod-cred
  domainName: example.com
  records:
    - name: "@"
      type: dns_record_type_a
      values:
        - value: "192.0.2.1"
      ttlSeconds: 3600
    - name: "www"
      type: dns_record_type_cname
      values:
        - value: "example.com."
      ttlSeconds: 3600
```

#### 2. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

#### 3. Configure Credentials

```bash
# Option 1: Environment variable (recommended for local dev)
export DIGITALOCEAN_TOKEN="dop_v1_xxxxxxxxxxxxxxxxxxxx"

# Option 2: Pulumi config secret
pulumi config set digitalocean:token dop_v1_xxxxxxxxxxxxxxxxxxxx --secret
```

#### 4. Preview Changes

```bash
pulumi preview --stack dev
```

#### 5. Deploy

```bash
pulumi up --stack dev
```

#### 6. Verify Outputs

```bash
pulumi stack output zone_name
pulumi stack output zone_id
pulumi stack output name_servers
```

### Using the Debug Script

The `debug.sh` script simplifies local testing:

```bash
# Set your DigitalOcean token
export DIGITALOCEAN_TOKEN="dop_v1_xxxxxxxxxxxxxxxxxxxx"

# Run the debug script with your manifest
./debug.sh /path/to/manifest.yaml
```

The debug script automatically:
- Initializes a temporary Pulumi stack
- Deploys the resources
- Shows stack outputs
- Cleans up the stack after testing (optional)

### Production Deployment

For production, use the OpenMCF CLI instead of running Pulumi directly:

```bash
# Apply a DNS zone manifest
openmcf apply -f manifest.yaml

# Destroy a DNS zone
openmcf destroy -f manifest.yaml
```

## Implementation Details

### Record Type Mapping

The Pulumi implementation maps protobuf enum values to DigitalOcean record types:

| Protobuf Enum | DigitalOcean Type |
|---------------|-------------------|
| `dns_record_type_a` | `A` |
| `dns_record_type_aaaa` | `AAAA` |
| `dns_record_type_cname` | `CNAME` |
| `dns_record_type_mx` | `MX` |
| `dns_record_type_txt` | `TXT` |
| `dns_record_type_srv` | `SRV` |
| `dns_record_type_caa` | `CAA` |
| `dns_record_type_ns` | `NS` |

### Record-Specific Fields

The implementation conditionally sets fields based on record type:

**MX Records:**
```go
if rec.Type.String() == "MX" {
    recordArgs.Priority = pulumi.Int(int(rec.Priority))
}
```

**SRV Records:**
```go
if rec.Type.String() == "SRV" {
    recordArgs.Priority = pulumi.Int(int(rec.Priority))
    recordArgs.Weight = pulumi.Int(int(rec.Weight))
    recordArgs.Port = pulumi.Int(int(rec.Port))
}
```

**CAA Records:**
```go
if rec.Type.String() == "CAA" {
    recordArgs.Flags = pulumi.Int(int(rec.Flags))
    recordArgs.Tag = pulumi.String(rec.Tag)
}
```

### Multi-Value Records

When a record has multiple values (e.g., multiple MX servers), the implementation creates one `digitalocean.DnsRecord` resource per value:

```go
for valIdx, val := range rec.Values {
    resourceName := fmt.Sprintf("%s-%d-%d", rec.Name, recIdx, valIdx)
    // Create separate DnsRecord for each value
}
```

This allows proper priority assignment for MX records and avoids DigitalOcean API limitations.

### Default Values

The implementation applies defaults from the spec:

- **TTL**: Defaults to 3600 seconds (1 hour) if not specified
- **Priority**: Defaults to 0 for MX/SRV if not specified
- **Weight**: Defaults to 0 for SRV if not specified
- **Port**: Defaults to 0 for SRV if not specified
- **Flags**: Defaults to 0 for CAA if not specified

## Stack Outputs

After successful deployment, the following outputs are available:

| Output | Description | Example |
|--------|-------------|---------|
| `zone_name` | The domain name | `"example.com"` |
| `zone_id` | DigitalOcean zone ID | `"example.com"` |
| `name_servers` | DigitalOcean nameservers | `["ns1.digitalocean.com", "ns2.digitalocean.com", "ns3.digitalocean.com"]` |

Access outputs programmatically:

```bash
pulumi stack output zone_name -s dev
```

Or in Go:

```go
ctx.Export("zone_name", pulumi.String(locals.DigitalOceanDnsZone.Spec.DomainName))
```

## Troubleshooting

### Common Errors

#### Error: "domain already exists"

**Cause**: The domain is already registered in your DigitalOcean account.

**Solution**: Either:
- Delete the existing domain from DigitalOcean control panel
- Import the existing domain into Pulumi state:
  ```bash
  pulumi import digitalocean:index/domain:Domain dns_zone example.com
  ```

#### Error: "Invalid authentication token"

**Cause**: The `DIGITALOCEAN_TOKEN` environment variable is not set or is invalid.

**Solution**:
```bash
# Verify token is set
echo $DIGITALOCEAN_TOKEN

# If empty, set it
export DIGITALOCEAN_TOKEN="dop_v1_xxxxxxxxxxxxxxxxxxxx"
```

#### Error: "record already exists"

**Cause**: Attempting to create a duplicate DNS record.

**Solution**: Check for duplicate records in your manifest. Each unique combination of (name, type, value) should appear only once.

#### Error: "priority is required for MX records"

**Cause**: MX record missing priority field in manifest.

**Solution**: Add priority to your MX record:
```yaml
- name: "@"
  type: dns_record_type_mx
  priority: 10  # Add this
  values:
    - value: "mail.example.com."
```

### DNS Propagation Issues

After deployment, DNS changes may take time to propagate:

**Immediate verification** (query DigitalOcean nameservers directly):
```bash
dig @ns1.digitalocean.com example.com
```

**Global propagation check**:
```bash
# Check from multiple locations
dig example.com @8.8.8.8
dig example.com @1.1.1.1
```

**Wait times**:
- DigitalOcean API: Immediate (seconds)
- Local ISP: TTL value (default 3600s = 1 hour)
- Global: 24-48 hours for full propagation

### Debugging Pulumi State

View current stack state:
```bash
pulumi stack --show-ids
```

View detailed resource information:
```bash
pulumi stack export > state.json
cat state.json | jq '.deployment.resources'
```

Refresh state from DigitalOcean:
```bash
pulumi refresh
```

## Advanced Usage

### Importing Existing DNS Zones

If you have existing DigitalOcean DNS zones, import them into Pulumi:

```bash
# Import the domain
pulumi import digitalocean:index/domain:Domain dns_zone example.com

# Import DNS records (repeat for each record)
pulumi import digitalocean:index/dnsRecord:DnsRecord www-0-0 example.com,123456789
```

Where `123456789` is the record ID from DigitalOcean API.

### Cross-Resource References

Reference outputs from other Pulumi stacks:

```yaml
- name: "@"
  type: dns_record_type_a
  values:
    - valueFromResourceOutput:
        resourceIdRef:
          name: my-load-balancer
        outputKey: ip_address
```

The Pulumi module resolves `StringValueOrRef` automatically.

### Stack Management

Create multiple stacks for different environments:

```bash
# Development stack
pulumi stack init dev
pulumi config set digitalocean:token $DO_DEV_TOKEN --secret
pulumi up

# Production stack
pulumi stack init prod
pulumi config set digitalocean:token $DO_PROD_TOKEN --secret
pulumi up
```

## Performance Considerations

### API Rate Limits

DigitalOcean API has a rate limit of **250 requests per minute**.

For large DNS zones (100+ records), the deployment may take several minutes. The Pulumi provider automatically handles rate limiting with retries.

### Parallelism

Pulumi creates resources in parallel by default. To reduce API load:

```bash
pulumi up --parallel 10  # Limit to 10 concurrent operations
```

## References

- **Pulumi DigitalOcean Provider**: https://www.pulumi.com/registry/packages/digitalocean/
- **DigitalOcean DNS API**: https://docs.digitalocean.com/reference/api/api-reference/#tag/Domains
- **OpenMCF Docs**: ../../docs/README.md
- **Examples**: ../../examples.md

## Contributing

When modifying the Pulumi implementation:

1. Update the module code in `module/`
2. Run `go mod tidy` to update dependencies
3. Test locally with `debug.sh`
4. Update this README if behavior changes
5. Update `overview.md` for architectural changes

## License

This implementation is part of the OpenMCF monorepo and follows the same license.

