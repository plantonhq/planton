locals {
  # Resource group name (provided externally)
  resource_group_name = var.spec.resource_group

  # Registry name
  registry_name = var.spec.registry_name

  # SKU mapping from enum to Azure value
  sku_name = var.spec.sku == "BASIC" ? "Basic" : (var.spec.sku == "PREMIUM" ? "Premium" : "Standard")

  # Tags from metadata
  tags = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      Name        = var.metadata.name
      Environment = var.metadata.env != null ? var.metadata.env : "default"
      ManagedBy   = "terraform"
    }
  )
}

