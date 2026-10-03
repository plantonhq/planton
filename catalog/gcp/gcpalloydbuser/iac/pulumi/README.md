# GCP AlloyDB User - Pulumi Module

Pulumi Go module for `GcpAlloydbUser`. Enables `alloydb.googleapis.com`, then creates `alloydb.User` with BUILT_IN or IAM authentication and optional database roles.

## Outputs

| Output | Description |
|--------|-------------|
| `name` | Full user resource path |
| `user_id` | User ID |
| `cluster_id` | Cluster resource path |
