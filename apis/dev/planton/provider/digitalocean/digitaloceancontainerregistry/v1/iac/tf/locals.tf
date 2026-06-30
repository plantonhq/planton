locals {
  # Normalize subscription tier to DigitalOcean API format
  # protobuf enum: "STARTER" → "starter"
  # protobuf enum: "BASIC" → "basic"
  # protobuf enum: "PROFESSIONAL" → "professional"
  subscription_tier_slug = lower(var.spec.subscription_tier)
  
  # Normalize region to DigitalOcean API format (lowercase)
  region_slug = lower(var.spec.region)
}

