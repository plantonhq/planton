locals {
  tags = {
    "planton.org/resource"      = "true"
    "planton.org/organization"  = var.metadata.org
    "planton.org/environment"   = var.metadata.env
    "planton.org/resource-kind" = "AwsBatchComputeEnvironment"
    "planton.org/resource-id"   = var.metadata.id
  }

  is_ec2  = contains(["EC2", "SPOT"], var.spec.compute_resources.type)
  is_spot = var.spec.compute_resources.type == "SPOT"
}
