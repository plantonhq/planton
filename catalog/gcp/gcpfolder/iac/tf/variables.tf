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
  description = "GcpFolder specification"
  type = object({
    # Where the folder sits in the hierarchy: directly under the
    # organization or inside another folder. Exactly one arm. Changing the
    # parent MOVES the folder in place (Google's folders.move) -- projects and
    # sub-folders travel with it and nothing is recreated -- but every IAM
    # grant and organization policy the folder inherited from its old parent
    # stops applying and the new parent's start applying at once.
    parent = object({
      # A top-level folder: the numeric organization ID (from `gcloud
      # organizations list`), without the `organizations/` prefix.
      organization_id = optional(string, "")

      # A nested folder: the parent folder's numeric ID -- a literal, or a
      # reference to another GcpFolder resource (its folder_id output). The
      # reference is how a chart builds a hierarchy: the child waits for the
      # parent to exist and nests inside it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      folder_id = optional(string, "")
    })

    # The name shown in the console and in `gcloud resource-manager folders
    # list`. Defaults to metadata.name when empty. Google's rules: 3-30
    # characters; letters, digits, spaces, hyphens, and underscores; must
    # start and end with a letter or digit; and UNIQUE among the parent's
    # direct children (two sibling folders cannot share a display name).
    # Mutable: a rename is an in-place update.
    #
    # A deleted folder keeps its display name reserved for the 30-day
    # soft-delete window (folders are recoverable via undelete in that
    # window), so a fresh folder with the same name under the same parent
    # fails until the window closes or the old one is purged.
    display_name = optional(string, "")

    # Client-side destroy guard. While true (Google's provider default), any
    # destroy -- including the platform's own teardown flows -- FAILS until
    # this field is set to false AND applied first; only then does a second
    # destroy delete the folder. Both engines always send the value
    # explicitly, so the spec is the single source of truth. The guard exists
    # because deleting a folder is a hierarchy-wide act: Google refuses to
    # delete a folder that still holds projects or folders, but a folder that
    # was just emptied is one accidental destroy away from a 30-day recovery
    # exercise.
    #
    # Ordering quirk in the provider: deletion_policy ABANDON is evaluated
    # BEFORE this guard, so abandoning a protected folder still works;
    # PREVENT is evaluated before both.
    deletion_protection = optional(bool)

    # Resource Manager tags bound to the folder at CREATE TIME only, as
    # `tagKeys/{numeric_id}` -> `tagValues/{numeric_id}` (the `name` outputs
    # of GcpTagKey and GcpTagValue). Tags are what organization policies
    # (`resource.matchTag`), IAM conditions, and firewall policies key on.
    #
    # Changing this map after creation RECREATES the folder, which for a
    # folder holding projects is not a move but a failure (Google refuses to
    # delete a non-empty folder). Use it only when the tag must exist at
    # create time -- typically so an organization policy conditioned on the
    # tag governs the folder from its first second. For every other case,
    # bind tags after creation with GcpTagBinding, which attaches and
    # detaches without touching the folder.
    tags = optional(map(string), {})

    # What destroying this resource does to the folder in GCP:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the folder is deleted (Google soft-deletes it for 30
    #                days, during which it can be undeleted and its display
    #                name stays reserved); fails while deletion_protection is
    #                true or the folder still holds projects or folders
    #   "PREVENT" -- destroy FAILS; the guard for the folders a landing zone
    #                is built on
    #   "ABANDON" -- the folder is removed from management but keeps existing
    #                in GCP with everything inside it; bypasses
    #                deletion_protection
    deletion_policy = optional(string, "")
  })
}
