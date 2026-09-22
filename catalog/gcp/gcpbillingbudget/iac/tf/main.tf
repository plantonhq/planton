# One Cloud Billing budget: an amount for a period, the spend it measures,
# the thresholds that alert, and where the alerts go. Everything but the
# billing account changes in place.
resource "google_billing_budget" "this" {
  billing_account = local.billing_account
  display_name    = local.display_name

  # Who may read the budget's data; unset lets Google apply its default.
  ownership_scope = var.spec.ownership_scope != "" ? var.spec.ownership_scope : null

  # Exactly one arm (spec CEL): a fixed amount, or last period's spend.
  amount {
    dynamic "specified_amount" {
      for_each = var.spec.amount.specified_amount != null ? [var.spec.amount.specified_amount] : []
      content {
        # Optional+Computed: unset takes the billing account's currency.
        currency_code = specified_amount.value.currency_code != "" ? specified_amount.value.currency_code : null
        units         = tostring(specified_amount.value.units)
        nanos         = specified_amount.value.nanos != 0 ? specified_amount.value.nanos : null
      }
    }
    last_period_amount = var.spec.amount.last_period_amount ? true : null
  }

  # Which spend counts. Every leaf is optional; the defaulted enums are
  # sent explicitly so the manifest states them (the loader applies the
  # proto defaults before either engine runs).
  dynamic "budget_filter" {
    for_each = var.spec.budget_filter != null ? [var.spec.budget_filter] : []
    content {
      projects               = length(local.filter_projects) > 0 ? local.filter_projects : null
      resource_ancestors     = length(budget_filter.value.resource_ancestors) > 0 ? budget_filter.value.resource_ancestors : null
      services               = length(budget_filter.value.services) > 0 ? budget_filter.value.services : null
      subaccounts            = length(budget_filter.value.subaccounts) > 0 ? budget_filter.value.subaccounts : null
      labels                 = length(budget_filter.value.labels) > 0 ? budget_filter.value.labels : null
      credit_types_treatment = coalesce(budget_filter.value.credit_types_treatment, "INCLUDE_ALL_CREDITS")
      credit_types           = length(budget_filter.value.credit_types) > 0 ? budget_filter.value.credit_types : null
      calendar_period        = budget_filter.value.calendar_period != "" ? budget_filter.value.calendar_period : null

      dynamic "custom_period" {
        for_each = budget_filter.value.custom_period != null ? [budget_filter.value.custom_period] : []
        content {
          start_date {
            year  = custom_period.value.start_date.year
            month = custom_period.value.start_date.month
            day   = custom_period.value.start_date.day
          }
          dynamic "end_date" {
            for_each = custom_period.value.end_date != null ? [custom_period.value.end_date] : []
            content {
              year  = end_date.value.year
              month = end_date.value.month
              day   = end_date.value.day
            }
          }
        }
      }
    }
  }

  # The thresholds that alert; spend_basis defaults to CURRENT_SPEND and is
  # sent explicitly.
  dynamic "threshold_rules" {
    for_each = var.spec.threshold_rules
    content {
      threshold_percent = threshold_rules.value.threshold_percent
      spend_basis       = coalesce(threshold_rules.value.spend_basis, "CURRENT_SPEND")
    }
  }

  # Where alerts go beyond the default administrator emails (the provider's
  # all_updates_rule block).
  dynamic "all_updates_rule" {
    for_each = var.spec.notifications != null ? [var.spec.notifications] : []
    content {
      pubsub_topic                     = all_updates_rule.value.pubsub_topic != "" ? all_updates_rule.value.pubsub_topic : null
      monitoring_notification_channels = length(all_updates_rule.value.monitoring_notification_channels) > 0 ? all_updates_rule.value.monitoring_notification_channels : null
      disable_default_iam_recipients   = all_updates_rule.value.disable_default_iam_recipients
      enable_project_level_recipients  = all_updates_rule.value.enable_project_level_recipients
      schema_version                   = coalesce(all_updates_rule.value.schema_version, "1.0")
    }
  }

  # What destroy does to the guardrail: DELETE (default), PREVENT (refuse),
  # or ABANDON (drop from state, keep alerting).
  deletion_policy = local.deletion_policy
}
