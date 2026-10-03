# GCP AlloyDB Instance - Pulumi Module

Pulumi Go module for `GcpAlloydbInstance`. Enables `alloydb.googleapis.com`, then creates `alloydb.Instance` with the released-provider surface: machine config, read pools, query insights, client connection config, connection pooling, public IP arms, and PSC.

## Outputs

| Output | Description |
|--------|-------------|
| `instance_name` | Full instance resource path |
| `ip_address` | Private IP |
| `state` | Instance state |
