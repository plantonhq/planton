# DigitalOcean Database Kafka Schema
#
# Registers one schema subject in a DigitalOcean managed Kafka cluster's
# schema registry -- the complete
# digitalocean_database_kafka_schema_registry resource surface.
#
# EVERY argument is create-only: the provider has no update path, so any
# change destroys the subject and re-registers it, which DROPS all previously
# registered versions. Treat schema evolution as a deliberate replacement.
#
# The registry stores every JSON-typed schema (avro, json) in canonical form
# -- object keys sorted, no whitespace -- and the provider's Read stores that
# canonical text verbatim, so a manifest written in any other key order or
# with any whitespace would re-plan a REPLACE on every refreshed plan. This
# module renders JSON schemas into the same canonical form before sending
# (jsonencode sorts object keys and emits no whitespace), so the manifest and
# the registry agree on the first plan and every one after. Protobuf schemas
# are text, not JSON, and are sent verbatim (see the GUIDE: the registry
# reformats them too, and nothing this side of the provider can absorb that).

locals {
  # Canonicalize JSON-typed schemas; pass protobuf text through untouched.
  # An avro/json schema that is not valid JSON fails here, at plan time,
  # instead of at the registry.
  schema = var.spec.schema_type == "protobuf" ? var.spec.schema : jsonencode(jsondecode(var.spec.schema))
}

resource "digitalocean_database_kafka_schema_registry" "schema" {
  cluster_id   = var.spec.cluster
  subject_name = var.spec.subject_name
  schema_type  = var.spec.schema_type
  schema       = local.schema
}
