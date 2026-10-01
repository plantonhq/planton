# Terraform Module: DigitalOcean Database Kafka Schema

Registers one schema subject in a DigitalOcean managed Kafka cluster's schema registry -- the complete `digitalocean_database_kafka_schema_registry` resource surface.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_database_kafka_schema_registry.schema` | The registered subject: name, type, definition |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanDatabaseKafkaSchemaSpec` proto: `cluster` (flattened reference string), `subject_name`, `schema_type`, `schema`. Authentication uses `digitalocean_token` (sensitive).

## Outputs

Exactly the `DigitalOceanDatabaseKafkaSchemaStackOutputs` contract: `cluster_id`, `subject_name`.

## Behavior notes

- ALL arguments are create-only (the resource has no update function): any change is destroy+recreate and DROPS all previously registered versions of the subject.
- Avro and JSON Schema definitions are rendered into the registry's canonical form (`jsonencode(jsondecode(...))`: keys sorted, no whitespace) before sending, because the registry stores that form and the provider's Read compares it verbatim -- without this, a human-ordered schema would re-plan a REPLACE on every refreshed plan. Protobuf text is sent verbatim and the registry reformats it, so a protobuf subject DOES re-plan a replacement on Terraform until the provider compares normalized text (see the GUIDE).
- Import: excluded -- the provider's importer is defective at the pin (see `iac/import-map.yaml`).
