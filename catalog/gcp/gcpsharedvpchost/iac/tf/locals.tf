locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Unlike most GCP resources, the Shared VPC host resource
  # REQUIRES an explicit project argument, so the fallback is made concrete by
  # reading the provider's resolved project from google_client_config instead
  # of passing null.
  project_id = (
    var.spec.project_id != ""
    ? var.spec.project_id
    : data.google_client_config.current[0].project
  )
}

# The provider's own resolved configuration -- the source of the default
# project when spec.project_id is omitted. Count-gated on that one case so
# every plan that names its project runs credential-free.
data "google_client_config" "current" {
  count = var.spec.project_id == "" ? 1 : 0
}
