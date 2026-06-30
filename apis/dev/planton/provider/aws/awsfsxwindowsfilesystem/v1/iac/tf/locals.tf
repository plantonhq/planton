locals {
  tags = {
    Resource     = "true"
    Organization = var.metadata.org
    Environment  = var.metadata.env
    ResourceKind = "AwsFsxWindowsFileSystem"
    ResourceId   = var.metadata.id
  }
}
