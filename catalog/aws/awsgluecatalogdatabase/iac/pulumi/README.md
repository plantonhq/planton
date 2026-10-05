# Pulumi Module: AWS Glue Catalog Database

Provisions an AWS Glue Data Catalog database using Pulumi (Go).

## Resources Created

- `glue.CatalogDatabase` — The Glue Data Catalog database

## How It Works

The module receives an `AwsGlueCatalogDatabaseIacInput` (the manifest
plus provider credentials), builds the AWS provider through the shared
builder, and renders the database from the spec. Send conditions match
the Terraform module argument-for-argument.

The module carries all three creation shapes — regular, resource link
(`target_database`), and federated (`federated_database`); a database is
exactly one shape and the spec validation enforces the exclusivity. Lake
Formation create-table default permissions render one entry per listed
principal, including the hardening shape (an entry with an empty
permission list) that suppresses the `IAM_ALLOWED_PRINCIPALS` default
grant.

## Outputs

| Name | Description |
|------|-------------|
| `database_name` | Name of the Glue Data Catalog database |
| `database_arn` | ARN of the database |
