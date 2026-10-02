locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The spec's tier enum is the honest form of Google's three exclusive
  # empty blocks; exactly one is emitted below -- identical to the Pulumi
  # module's switch.
  tier_basic         = var.spec.tier == "BASIC" ? [true] : []
  tier_scaled        = var.spec.tier == "SCALED" ? [true] : []
  tier_unprovisioned = var.spec.tier == "UNPROVISIONED" ? [true] : []
}
