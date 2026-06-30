# OpenFgaRelationshipTuple Local Values
# This file computes local values from the input variables.

locals {
  # Store ID extracted from the StringValueOrRef object
  store_id = var.spec.store_id.value

  # Authorization model ID extracted from the optional StringValueOrRef object
  # Returns null if not specified (uses latest model in the store)
  authorization_model_id = var.spec.authorization_model_id != null ? var.spec.authorization_model_id.value : null

  # Construct the user string from structured input:
  # - Without relation: "type:id" (e.g., "user:anne")
  # - With relation: "type:id#relation" (e.g., "group:engineering#member")
  # The id is a StringValueOrRef object, so we extract the .value field.
  user = (
    var.spec.user.relation != null && var.spec.user.relation != ""
    ? "${var.spec.user.type}:${var.spec.user.id.value}#${var.spec.user.relation}"
    : "${var.spec.user.type}:${var.spec.user.id.value}"
  )

  # The relation field (direct pass-through)
  relation = var.spec.relation

  # Construct the object string from structured input: "type:id"
  # The id is a StringValueOrRef object, so we extract the .value field.
  object = "${var.spec.object.type}:${var.spec.object.id.value}"

  # Condition (optional, pass-through)
  condition = var.spec.condition
}
