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
  description = "AwsElasticacheUser specification"
  type = object({
    # The AWS region where the user is created. ElastiCache users are regional
    # resources — a user group can only contain users from its own region.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Cache engine this user authenticates against. Values: "redis", "valkey".
    # A user group only accepts users whose engine matches its own.
    # Create-time immutable — changing it destroys and recreates the user.
    engine = string

    # The name clients present in the AUTH command (`AUTH <user_name>
    # <password>`). Unlike the user id (metadata.name), user names need not be
    # unique — AWS unions the credentials of same-named users at
    # authentication time. Every user group must include one user whose name
    # is exactly "default": it defines the permissions of clients that connect
    # without authenticating. Create-time immutable.
    user_name = string

    # Redis ACL access string scoping what this user may do — the same syntax
    # as the Redis ACL SETUSER rule list. Composed of switches ("on"/"off"
    # enables or disables the user), key patterns ("~app1:*" grants access to
    # matching keys, "~*" to all keys), and command categories ("+@read"
    # grants the read category, "+@all" everything, "-@dangerous" subtracts
    # the dangerous category).
    #
    # Common shapes:
    #   "on ~* +@all"                      — full access (an admin user)
    #   "on ~app1:* +@read +@write"        — read/write scoped to one key prefix
    #   "on ~* +@read -@dangerous"         — read-only, no admin commands
    #   "off ~* +@all"                     — a disabled "default" user (clients
    #                                        MUST authenticate as someone else)
    #
    # Updates apply in place — tightening an access string takes effect on new
    # connections without recreating the user.
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
      #   no long-lived secret anywhere. AWS requires the user's `user_name` to
      #   equal its user id (metadata.name) for IAM auth, and the attached cache
      #   must have transit (TLS) encryption enabled.
      #
      # - "no-password-required": the user authenticates with no credential at
      #   all. Intended for the mandatory "default" user when it is switched
      #   "off" in its access string, or for migration windows — never for a
      #   production user that is "on".
      type = string

      # Passwords for the "password" authentication type. One or two entries,
      # each 16–128 printable characters; two entries enable zero-downtime
      # rotation. Must be empty for the "iam" and "no-password-required" types.
      passwords = optional(list(string), [])
    })
  })
}
