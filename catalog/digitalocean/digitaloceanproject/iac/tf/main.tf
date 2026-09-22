# DigitalOcean Project
#
# Provisions a DigitalOcean project -- the account-level organizational
# container -- modeling the complete digitalocean_project resource surface.
# Membership is carried here on the project itself as resource URNs; the
# provider's standalone partial-ownership membership resource is
# deliberately not used (one project object owns its full membership list,
# which is also how the API reports it back).
#
# Destroying the project never destroys member resources: the provider
# relocates every member to the account's default project and retries the
# delete through the API's 412 responses while the asynchronous moves
# settle.

resource "digitalocean_project" "project" {
  name = var.spec.project_name

  description = var.spec.description != "" ? var.spec.description : null

  # Unset defers to the provider's default purpose ("Web Application").
  # DigitalOcean stores non-standard purposes as "Other: <text>" and strips
  # the prefix on read, so free text round-trips cleanly; values that
  # themselves start with "Other:" are unrepresentable (spec validation).
  purpose = var.spec.purpose != "" ? var.spec.purpose : null

  # Lowercase canonical (spec validation); DigitalOcean accepts it
  # case-insensitively and reports it back capitalized, which the provider
  # diff-suppresses.
  environment = var.spec.environment != "" ? var.spec.environment : null

  is_default = var.spec.is_default

  # Membership is managed only when declared: an empty list stays null so
  # out-of-band assignments (and the resources' own project selections) are
  # left untouched -- the attribute is Optional+Computed upstream, so
  # omitting it adopts whatever the API reports without drift.
  resources = length(var.spec.resources) > 0 ? var.spec.resources : null

  # The relocation of members is asynchronous on DigitalOcean's side and the
  # provider retries the DELETE through "412 cannot delete a project with
  # resources" only until this timeout. Its 3-minute default was exceeded
  # live: a one-member project stayed non-empty for over 180 seconds after
  # the relocation was accepted, the destroy failed, and a second destroy of
  # the by-then-empty project succeeded. Ten minutes covers the measured lag
  # with room; a retry that ends earlier costs nothing. Twin of the Pulumi
  # module's CustomTimeouts. The tight default and the undocumented knob are
  # reported upstream as digitalocean/terraform-provider-digitalocean#1608.
  timeouts {
    delete = "10m"
  }
}
