# CloudflareDnsRecord Pulumi Module

This Pulumi module provisions a Cloudflare DNS record.

## Prerequisites

- Go 1.21+
- Pulumi CLI
- Cloudflare API token with DNS:Edit permissions

## Usage

### As Part of Planton

This module is typically invoked through the Planton CLI:

```bash
planton apply -f manifest.yaml
```

### Standalone Usage

Run this module directly with the planton CLI's Pulumi commands. From this directory (it holds `Pulumi.yaml`, so the CLI runs this module):

```bash
planton pulumi init --manifest manifest.yaml --stack <org>/<project>/<stack>
planton pulumi preview --manifest manifest.yaml --stack <org>/<project>/<stack> -p cloudflare-provider-config.yaml
planton pulumi update --manifest manifest.yaml --stack <org>/<project>/<stack> -p cloudflare-provider-config.yaml
```

The CLI builds a `CloudflareDnsRecordIacInput` (the manifest as `target`, the provider config file as `provider_config`) and hands it to the module through `IAC_INPUT_YAML_FILE`.

## Environment Variables

The module reads its `CloudflareDnsRecordIacInput` from the Pulumi config key `planton:iac-input`, or else from one of these:

| Variable | Description | Required |
|----------|-------------|----------|
| `IAC_INPUT_YAML` | `CloudflareDnsRecordIacInput` as YAML content (`target` plus optional `provider_config`) | One of these two, when `planton:iac-input` is not set |
| `IAC_INPUT_YAML_FILE` | Path to a YAML file holding the `CloudflareDnsRecordIacInput` | One of these two, when `planton:iac-input` is not set |
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token (used when the IaC input carries no `provider_config`) | No |

A bare manifest is not an IaC input: wrap it under `target`.

## Module Structure

```
.
├── main.go           # Pulumi entry point
├── Pulumi.yaml       # Pulumi project configuration
├── BUILD.bazel       # Bazel build target
├── README.md         # This file
└── module/
    ├── main.go       # Resource orchestration
    ├── locals.go     # Data transformations
    ├── outputs.go    # Output constants
    └── dns_record.go # DNS record creation logic
```

## Outputs

| Output | Description |
|--------|-------------|
| `record_id` | Cloudflare DNS record ID |
| `record_name` | The record name as declared (relative to the zone; Cloudflare answers reads with the full FQDN) |
| `record_type` | DNS record type |
| `proxied` | Whether the record is proxied |
| `zone_id` | The zone the record lives in (a record's API identity is zone_id + record_id) |

## Debugging

Preview against the test manifest from this directory:

```bash
planton pulumi preview --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> -p cloudflare-provider-config.yaml
```

## Troubleshooting

### "iac-input not found"

The module found no input. Run it through `planton pulumi`, or set `IAC_INPUT_YAML_FILE` (or `IAC_INPUT_YAML`) to a `CloudflareDnsRecordIacInput` with the manifest under `target`.

### "authentication failed"

Verify your Cloudflare API token has the required permissions:
- Zone:DNS:Edit

### "zone not found"

Verify the `zone_id` in your manifest matches an existing Cloudflare zone.
