# SES account settings for one region -- a settings singleton: AWS
# keeps exactly one SES account object per account+region, and this
# module manages its suppression-list and VDM attributes.
# metadata.name never reaches AWS.
#
# Lifecycle facts the render below depends on:
#   - each arm renders ONLY when its spec message is present (an
#     omitted arm leaves the account's current setting untouched --
#     that omission is meaningful and deliberate);
#   - suppression switched off (enabled: false) is a real posture: it
#     writes the EMPTY reason list, which turns account-level
#     auto-suppression OFF (the required upstream set
#     argument accepts []);
#   - destroy semantics DIFFER per arm: suppression PERSISTS after
#     destroy (the provider's delete is a no-op; the last-applied
#     reasons stay), while the VDM resource's delete resets VDM to
#     DISABLED;
#   - VDM's dashboard/guardian sub-toggles are presence-typed: unset
#     sends nothing (AWS keeps its default), set maps to the
#     ENABLED/DISABLED FeatureStatus strings.

# The account id feeds the account_id output regardless of which arms
# render (both upstream resources are account-scoped).
data "aws_caller_identity" "this" {}

resource "aws_sesv2_account_suppression_attributes" "this" {
  count = var.spec.suppression != null ? 1 : 0

  # Suppression switched off is written as the empty reason list (SES's
  # spelling of "no auto-suppression"); the API already refuses a
  # switched-on arm with no events.
  suppressed_reasons = coalesce(var.spec.suppression.enabled, true) ? var.spec.suppression.reasons : []
}

resource "aws_sesv2_account_vdm_attributes" "this" {
  count = var.spec.vdm != null ? 1 : 0

  vdm_enabled = var.spec.vdm.enabled ? "ENABLED" : "DISABLED"

  dynamic "dashboard_attributes" {
    for_each = var.spec.vdm.engagement_metrics != null ? [1] : []
    content {
      engagement_metrics = var.spec.vdm.engagement_metrics ? "ENABLED" : "DISABLED"
    }
  }

  dynamic "guardian_attributes" {
    for_each = var.spec.vdm.optimized_shared_delivery != null ? [1] : []
    content {
      optimized_shared_delivery = var.spec.vdm.optimized_shared_delivery ? "ENABLED" : "DISABLED"
    }
  }
}
