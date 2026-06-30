locals {
  # Normalize instance size slug (replace underscores with hyphens)
  instance_size_slug = replace(var.spec.instance_size_slug, "_", "-")
  
  # Determine if using git or image source
  is_git_source   = var.spec.git_source != null
  is_image_source = var.spec.image_source != null
  
  # Build environment variables array
  env_vars = [
    for key, value in var.spec.env : {
      key   = key
      value = value
      scope = "RUN_AND_BUILD_TIME"
      type  = "GENERAL"
    }
  ]
  
  # Determine if autoscaling should be configured
  enable_autoscaling = var.spec.enable_autoscale
  
  # Resource labels
  resource_labels = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      "managed-by" = "planton"
      "component"  = "digitalocean-app-platform-service"
    }
  )
}

