# Enables the project as a Shared VPC host.
#
# The provider resource is a flag on the project: its only argument is the
# project, which is immutable (a different project is a different host).
# Destroying it disables the host role, which Google refuses while any
# service project is still attached -- the registry places every
# GcpSharedVpcServiceProject downstream of the host it references, so a
# chart's teardown detaches first. deletion_policy is sent only when set so
# the provider default (DELETE) stays the provider's.
resource "google_compute_shared_vpc_host_project" "this" {
  project = local.project_id

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
