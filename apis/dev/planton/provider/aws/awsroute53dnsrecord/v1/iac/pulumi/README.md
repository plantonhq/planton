# AWS Route53 DNS Record Pulumi Module

This Pulumi module creates DNS records in AWS Route53 hosted zones.

## Usage

### As a Go Module

```go
package main

import (
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    awsroute53dnsrecordv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsroute53dnsrecord/v1"
    "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsroute53dnsrecord/v1/iac/pulumi/module"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &awsroute53dnsrecordv1.AwsRoute53DnsRecordStackInput{
            // Note: zone_id and alias_target fields use StringValueOrRef
            // The CLI resolves value_from references before passing to Pulumi
        }
        return module.Resources(ctx, stackInput)
    })
}
```

### Via Planton CLI

```bash
planton pulumi up --manifest dns-record.yaml
```

## Inputs

The module accepts `AwsRoute53DnsRecordStackInput` which contains:

- `target`: The `AwsRoute53DnsRecord` resource definition
  - `spec.zone_id`: Route53 zone ID (StringValueOrRef - resolved before module)
  - `spec.name`: DNS record name (FQDN or subdomain)
  - `spec.type`: Record type (A, AAAA, CNAME, MX, TXT, etc.)
  - `spec.ttl`: Time to live in seconds (ignored for alias records)
  - `spec.values`: Record values (for standard records)
  - `spec.alias_target`: Alias target configuration (StringValueOrRef fields)
  - `spec.routing_policy`: Advanced routing configuration
- `provider_config`: AWS credentials configuration

## Outputs

| Output | Description |
|--------|-------------|
| `fqdn` | Fully qualified domain name of the record |
| `record_type` | DNS record type (A, AAAA, CNAME, etc.) |
| `zone_id` | Route53 hosted zone ID |
| `is_alias` | Whether this is an alias record |
| `set_identifier` | Routing policy set identifier |

## StringValueOrRef Fields

The following fields support both literal values and resource references:

- `spec.zone_id`: Default kind is `AwsRoute53Zone`, field path is `status.outputs.zone_id`
- `alias_target.dns_name`: Default kind is `AwsAlb`, field path is `status.outputs.load_balancer_dns_name`
- `alias_target.zone_id`: Default kind is `AwsAlb`, field path is `status.outputs.load_balancer_hosted_zone_id`

The CLI resolves `value_from` references before invoking the Pulumi module.

## Development

```bash
# Build
make build

# Test locally
pulumi preview --stack dev
```
