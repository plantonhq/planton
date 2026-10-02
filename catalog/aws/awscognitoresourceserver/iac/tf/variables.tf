variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsCognitoResourceServer specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The Cognito User Pool this resource server belongs to.
    # Format: "{region}_{poolId}" (e.g., "us-east-1_Ab1Cd2EfG").
    # ForceNew -- a resource server cannot be moved between pools.
    # Accepts a direct pool ID or a reference to an AwsCognitoUserPool resource.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    user_pool_id = string

    # The resource server's unique identifier within the pool -- the prefix of
    # every scope it mints ("{identifier}/{scope_name}") and the value access
    # tokens carry. Conventionally the API's audience URI (e.g.
    # "https://api.example.com"). 1-256 characters. ForceNew.
    identifier = string

    # Human-readable display name shown in the Cognito console. 1-256
    # characters.
    name = string

    # The custom scopes this resource server defines. Each becomes requestable
    # by app clients as "{identifier}/{scope_name}". Maximum 100 scopes.
    scopes = optional(list(object({
      # The scope name (e.g. "read", "orders:write"). 1-256 characters; letters,
      # digits, and most punctuation are allowed, but not spaces, '"', '/', or
      # '\\' (the '/' is reserved as the identifier/scope separator).
      scope_name = string

      # What granting this scope means, shown on consent screens and in the
      # console. 1-256 characters.
      scope_description = string
    })), [])
  })
}
