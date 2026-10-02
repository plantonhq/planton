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
  description = "AwsMemorydbUser specification"
  type = object({
    # The AWS region where the user is created. MemoryDB users are regional
    # resources — an ACL can only contain users from its own region.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Redis ACL access string scoping what this user may do — the same syntax
    # as the Redis ACL SETUSER rule list. Composed of switches ("on"/"off"
    # enables or disables the user), key patterns ("~app1:*" grants access to
    # matching keys, "~*" to all keys), and command categories ("+@read"
    # grants the read category, "+@all" everything, "-@dangerous" subtracts
    # the dangerous category).
    #
    # Common shapes:
    #   "on ~* +@all"                — full access (an admin user)
    #   "on ~app1:* +@read +@write"  — read/write scoped to one key prefix
    #   "on ~* +@read -@dangerous"   — read-only, no admin commands
    #
    # Updates apply in place — a tightened access string takes effect on new
    # connections without recreating the user.
    #
    # AWS stores the string NORMALIZED (live-verified 2026-08-13): the API
    # echoes it with Redis ACL defaults made explicit — e.g.
    # "on ~orders:* +@read" comes back as
    # "on ~orders:* resetchannels -@all +@read". The provider reconciles the
    # normalized form, so manifests never see drift; just don't expect a
    # byte-identical echo when comparing DescribeUsers output to this field.
    access_string = string

    # How clients prove they are this user. Exactly one authentication type,
    # with passwords carried inline only for the "password" type.
    authentication_mode = object({
      # Authentication mechanism. Values:
      #
      # - "password": the client presents one of the passwords below in the AUTH
      #   command. Two passwords may be set simultaneously so credentials rotate
      #   with zero downtime (add the new password, roll clients, remove the old).
      #
      # - "iam": the client signs a short-lived token with its AWS IAM identity —
      #   no long-lived secret anywhere. AWS requires TLS on the cluster for
      #   IAM-authenticated connections, and the IAM principal needs
      #   `memorydb:Connect` on both the user ARN and the cluster ARN.
      type = string

      # Passwords for the "password" authentication type. One or two entries,
      # each 16–128 characters; two entries enable zero-downtime rotation.
      # Must be empty for the "iam" type. Write-only at AWS: the API never
      # returns passwords, so drift changed outside this manifest is
      # undetectable, and a user adopted by import carries no password state —
      # re-assert the passwords here to manage them.
      passwords = optional(list(string), [])
    })
  })
}
