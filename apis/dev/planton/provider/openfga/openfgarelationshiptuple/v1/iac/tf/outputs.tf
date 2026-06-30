# OpenFgaRelationshipTuple Outputs
# This file defines the outputs from the OpenFGA relationship tuple deployment.
# These values are used to populate the stack outputs protobuf message.
#
# Note: Relationship tuples don't have a unique ID in OpenFGA. They are identified
# by the combination of (store_id, user, relation, object).
#
# Reference: https://registry.terraform.io/providers/openfga/openfga/latest/docs/resources/relationship_tuple

output "user" {
  description = "The user of the relationship tuple"
  value       = openfga_relationship_tuple.this.user
}

output "relation" {
  description = "The relation of the relationship tuple"
  value       = openfga_relationship_tuple.this.relation
}

output "object" {
  description = "The object of the relationship tuple"
  value       = openfga_relationship_tuple.this.object
}
