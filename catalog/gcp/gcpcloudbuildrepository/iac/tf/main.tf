# The repository link. Every field is immutable; destroy removes the link
# from Cloud Build, never the repository on the code host. No API
# enablement here: the parent connection's block enabled Cloud Build in
# this project.
resource "google_cloudbuildv2_repository" "this" {
  project           = local.project_id
  location          = local.location
  parent_connection = var.spec.parent_connection
  name              = local.repository_id
  remote_uri        = var.spec.remote_uri
  annotations       = length(var.spec.annotations) > 0 ? var.spec.annotations : null
  deletion_policy   = local.deletion_policy
}
