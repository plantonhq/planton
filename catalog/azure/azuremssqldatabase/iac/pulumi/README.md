# AzureMssqlDatabase - Pulumi Module

Pulumi implementation for the AzureMssqlDatabase kind.

## Architecture

```
mssql.Database (single resource -- ARM's TDE, retention, and
                threat-detection sub-APIs are argument blocks)
```

## Key Design Decisions

- **Enums map through exhaustive vocabularies** in `locals.go` (create
  mode, secondary type, license, enclave, backup redundancy, import
  credential/auth kinds, Defender state). Unspecified create_mode and
  sku_name are not sent at all, matching the Terraform module's null and
  letting Azure compute its defaults.
- **Presence guards on every optional-with-default field** (collation,
  transparent_data_encryption_enabled, geo_backup_enabled): IaC inputs
  built from a manifest do not materialize proto defaults, so unset
  falls back to the documented default explicitly.
- **`threat_detection_policy.email_account_admins` maps bool → wire
  string** ("Enabled"/"Disabled") -- this resource's contract differs
  from the server-scope policy's bool.
- **The mode↔source matrix lives in spec validation** -- the module
  forwards whichever source is set; a wrong pairing never reaches it.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless (web identity), and ambient
credential chains. Never construct the provider inline.

## Running Locally

```bash
# Build
make build

# Run with Pulumi
make run
```
