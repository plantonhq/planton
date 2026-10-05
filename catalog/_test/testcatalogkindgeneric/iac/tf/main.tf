# The one "resource" this module manages. terraform_data is a Terraform
# builtin: it stores its input in state and triggers replacement when the
# input changes -- a real lifecycle (create, update-in-place vs replace,
# destroy) with no cloud behind it. Every spec field feeds the input so a
# change to ANY field class is visible in the plan diff.
resource "terraform_data" "this" {
  input = {
    metadata = var.metadata
    spec     = var.spec
  }
}
