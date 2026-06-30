locals {
  # Normalize certificate type for DigitalOcean API
  # protobuf enum: "lets_encrypt" → "lets_encrypt" (already correct)
  # protobuf enum: "custom" → "custom" (already correct)
  cert_type = var.spec.type
  
  # Determine which certificate source is being used
  is_lets_encrypt = var.spec.type == "lets_encrypt"
  is_custom       = var.spec.type == "custom"
  
  # Extract Let's Encrypt domains if specified
  le_domains = local.is_lets_encrypt && var.spec.lets_encrypt != null ? var.spec.lets_encrypt.domains : []
  
  # Extract custom certificate materials if specified
  custom_leaf_cert  = local.is_custom && var.spec.custom != null ? var.spec.custom.leaf_certificate : ""
  custom_private_key = local.is_custom && var.spec.custom != null ? var.spec.custom.private_key : ""
  custom_cert_chain = local.is_custom && var.spec.custom != null && var.spec.custom.certificate_chain != null ? var.spec.custom.certificate_chain : ""
}

