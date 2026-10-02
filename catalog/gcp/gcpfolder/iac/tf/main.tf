# The Resource Manager folder.
#
# The parent is mutable: changing it MOVES the folder (Google's folders.move)
# with every project and sub-folder inside it. The create-time tags are the
# one immutable input -- the provider recreates the folder when they change,
# which Google refuses for a non-empty folder; the spec steers users to
# GcpTagBinding for everything but create-time tagging.
#
# deletion_protection is always sent (see locals.tf). deletion_policy is sent
# only when set so the provider's default (DELETE) stays the provider's.
resource "google_folder" "this" {
  parent       = local.parent
  display_name = local.display_name

  deletion_protection = local.deletion_protection
  tags                = length(var.spec.tags) > 0 ? var.spec.tags : null
  deletion_policy     = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
