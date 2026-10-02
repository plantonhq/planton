variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "GcpBillingBudget specification"
  type = object({
    # The Cloud Billing account the budget belongs to, as its ID
    # (`012345-6789AB-CDEF01`) or `billingAccounts/{id}`. Required; the
    # module normalizes to the resource name. Immutable.
    billing_account = string

    # Name shown in the Cloud Billing console, up to 60 characters. Defaults
    # to metadata.name.
    display_name = optional(string, "")

    # The budgeted amount: a fixed amount or last period's spend. Required.
    amount = object({
      # A fixed amount in the billing account's currency.
      specified_amount = optional(object({
        # The 3-letter ISO 4217 currency code. Must match the billing account's
        # currency; unset takes the account's currency.
        currency_code = optional(string, "")

        # The whole units of the amount in currency_code, e.g. 1000 for a budget
        # of one thousand US dollars when the currency is USD. 0 with a non-zero
        # nanos is a sub-unit budget.
        units = optional(number, 0)

        # Fractional part of the amount in nano units (10^-9), 0 to 999,999,999.
        # 750,000,000 with units 1 is 1.75 in the currency.
        nanos = optional(number, 0)
      }))

      # Budget the previous calendar period's spend: the budget for this period
      # is what was spent last period. Only with a calendar_period filter,
      # never with a custom_period.
      last_period_amount = optional(bool, false)
    })

    # Which spend the budget counts. Omit to count everything the billing
    # account pays for, reset monthly.
    budget_filter = optional(object({
      # Only usage from these projects counts: GcpProject references (resolved
      # to project numbers) or `projects/{number}` literals. Omitted means every
      # project the billing account pays for.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      projects = optional(list(string), [])

      # Only usage under these folders or organizations counts: GcpFolder
      # references (resolved to `folders/{id}`) or `organizations/{id}`
      # literals.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      resource_ancestors = optional(list(string), [])

      # Only usage of these services counts, as `services/{service_id}` (the
      # id from the Cloud Billing catalog, e.g. services/6F81-5844-456A for
      # Compute Engine). Unset sends nothing.
      services = optional(list(string), [])

      # Only usage from these subaccounts counts, as
      # `billingAccounts/{account_id}`; naming the parent account includes
      # usage from the parent and every subaccount.
      subaccounts = optional(list(string), [])

      # Only usage carrying this label counts. One key with one value (the API
      # accepts a single pair). Unset sends nothing.
      labels = optional(map(string), {})

      # How credits count against the budget: INCLUDE_ALL_CREDITS (the
      # default -- spend net of every credit), EXCLUDE_ALL_CREDITS (gross
      # spend), INCLUDE_SPECIFIED_CREDITS (net of the credit_types listed).
      credit_types_treatment = optional(string)

      # The credit types subtracted under INCLUDE_SPECIFIED_CREDITS, e.g.
      # COMMITTED_USAGE_DISCOUNT, SUSTAINED_USAGE_DISCOUNT, PROMOTION,
      # FREE_TIER.
      credit_types = optional(list(string), [])

      # The recurring period the budget resets on: MONTH (the default when
      # neither period is set), QUARTER, or YEAR. Alternative to custom_period.
      calendar_period = optional(string, "")

      # A fixed date range instead of a recurring period. Alternative to
      # calendar_period; incompatible with last_period_amount.
      custom_period = optional(object({
        start_date = object({
          year  = number
          month = number
          day   = number
        })
        end_date = optional(object({
          year  = number
          month = number
          day   = number
        }))
      }))
    }))

    # The thresholds that alert, each a share of the budget against current
    # or forecasted spend. Omit for a budget that only reports.
    threshold_rules = optional(list(object({
      # The share of the budget that triggers the alert, as a 1.0-based
      # fraction: 0.5 is 50%, 1.0 is 100%, 1.2 is 120% (over budget). Must be
      # greater than 0.
      threshold_percent = number

      # What the threshold is compared against: CURRENT_SPEND (actual spend so
      # far, the default) or FORECASTED_SPEND (Google's projection of the
      # period's total, which alerts before the money is spent).
      spend_basis = optional(string)
    })), [])

    # Where alerts go beyond the default administrator emails.
    notifications = optional(object({
      # The Pub/Sub topic budget notifications are published to (the full
      # budget state on every update, as JSON): a GcpPubSubTopic reference or
      # `projects/{project}/topics/{topic}`. The Billing service agent must be
      # a publisher on it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      pubsub_topic = optional(string, "")

      # Cloud Monitoring notification channels (email, SMS, Slack, PagerDuty)
      # that receive threshold alerts: GcpMonitoringNotificationChannel
      # references or `projects/{project}/notificationChannels/{id}`. Up to 5.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      monitoring_notification_channels = optional(list(string), [])

      # Stop the default emails to the billing account's administrators and
      # users; only the channels above are notified.
      disable_default_iam_recipients = optional(bool, false)

      # Also email the owners of the projects the budget filters on (only
      # with a single-project filter).
      enable_project_level_recipients = optional(bool, false)

      # The JSON schema version of the Pub/Sub notification. Only "1.0"
      # exists; sent as such.
      schema_version = optional(string)
    }))

    # Who may read the budget's data: ALL_USERS (anyone with billing.budgets
    # permissions on the account) or BILLING_ACCOUNT (only billing account
    # administrators). Unset lets Google apply its default.
    ownership_scope = optional(string, "")

    # What destroy does to the budget:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the budget and its alerts are deleted
    #   "PREVENT" -- destroy FAILS; keeps a guardrail production spend
    #                depends on
    #   "ABANDON" -- the budget leaves management but keeps alerting
    deletion_policy = optional(string, "")
  })
}
