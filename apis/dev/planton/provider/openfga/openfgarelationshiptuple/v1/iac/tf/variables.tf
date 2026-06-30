# OpenFgaRelationshipTuple Variables
# This file defines all input variables for the OpenFgaRelationshipTuple Terraform module.
# These variables map to the OpenFgaRelationshipTupleSpec protobuf message.

variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "OpenFgaRelationshipTuple specification"
  type = object({
    # store_id is the unique identifier of the OpenFGA store this tuple belongs to.
    # This is an object with a 'value' field containing the store ID.
    # The value is resolved by Planton runtime if using foreign key references.
    # Note: Changing the store_id requires replacing the tuple.
    store_id = object({
      value = string
    })

    # authorization_model_id is the optional ID of the authorization model this tuple
    # is associated with. If not specified, uses the latest model.
    # This is an object with a 'value' field containing the model ID.
    # The value is resolved by Planton runtime if using foreign key references.
    # Note: Changing this requires replacing the tuple.
    authorization_model_id = optional(object({
      value = string
    }))

    # user is the subject of the relationship tuple - who is being granted access.
    # Structured as type + id + optional relation.
    # The id is a StringValueOrRef object with a 'value' field containing the user ID.
    # The value is resolved by Planton runtime if using foreign key references.
    # Examples:
    #   - {type: "user", id: {value: "anne"}} -> "user:anne"
    #   - {type: "group", id: {value: "engineering"}, relation: "member"} -> "group:engineering#member"
    #   - {type: "user", id: {value: "*"}} -> "user:*"
    # Note: Changing the user requires replacing the tuple.
    user = object({
      type = string
      id = object({
        value = string
      })
      relation = optional(string)
    })

    # relation is the relationship type between the user and object.
    # Examples: "viewer", "editor", "owner", "member", "admin"
    # Note: Changing the relation requires replacing the tuple.
    relation = string

    # object is the resource the user is being granted access to.
    # Structured as type + id.
    # The id is a StringValueOrRef object with a 'value' field containing the object ID.
    # The value is resolved by Planton runtime if using foreign key references.
    # Examples:
    #   - {type: "document", id: {value: "budget-2024"}} -> "document:budget-2024"
    #   - {type: "folder", id: {value: "reports"}} -> "folder:reports"
    # Note: Changing the object requires replacing the tuple.
    object = object({
      type = string
      id = object({
        value = string
      })
    })

    # condition is an optional condition that must be satisfied for this tuple.
    # Includes name (condition name from model) and context_json (partial context).
    condition = optional(object({
      name         = string
      context_json = optional(string)
    }))
  })
}
