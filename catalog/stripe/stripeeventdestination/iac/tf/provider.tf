terraform {
  required_version = ">= 1.0"

  required_providers {
    # Pinned exactly: the provider is generated from Stripe's OpenAPI spec, is 0.x, and publishes
    # no changelog, so every version move is re-proven live before it lands.
    stripe = {
      source  = "stripe/stripe"
      version = "0.3.0"
    }
  }
}
