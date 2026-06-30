# Auth0Action Variables
# Maps to the Auth0ActionSpec protobuf message

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
  description = "Auth0Action specification"
  type = object({
    # supported_trigger defines the single trigger this action targets.
    # Required. Contains id (trigger type) and version (trigger API version).
    supported_trigger = object({
      id      = string
      version = string
    })

    # code is the Node.js source code of the action.
    # Required.
    code = string

    # runtime is the Node.js runtime version (node18 or node22).
    # Optional — Auth0 assigns a default based on the trigger version.
    runtime = optional(string)

    # deploy controls whether the action is automatically deployed.
    # Default: true
    deploy = optional(bool, true)

    # dependencies is a list of npm packages this action depends on.
    # Each entry requires name and version.
    dependencies = optional(list(object({
      name    = string
      version = string
    })), [])

    # secrets is a list of key-value secrets available to the action at runtime.
    # Each entry requires name and value.
    secrets = optional(list(object({
      name  = string
      value = string
    })), [])

    # trigger_binding optionally binds this action to its supported trigger.
    # When set, the action is deployed and attached to the trigger flow.
    # When null, the action is created/deployed but not bound.
    trigger_binding = optional(object({
      display_name = optional(string)
    }))
  })
}
