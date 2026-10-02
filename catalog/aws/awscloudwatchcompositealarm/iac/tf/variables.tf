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
  description = "AwsCloudwatchCompositeAlarm specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Boolean rule expression over the states of other alarms in the same
    # account and region. Alarms are addressed BY NAME inside state functions:
    #
    #   ALARM("my-metric-alarm")                 — true while that alarm is in ALARM
    #   OK("my-metric-alarm")                    — true while it is in OK
    #   INSUFFICIENT_DATA("my-metric-alarm")     — true while it lacks data
    #
    # Functions combine with AND / OR / NOT and parentheses; the constants
    # TRUE and FALSE are also valid (useful when testing a new composite).
    #
    # Example: "ALARM(\"cpu-high\") AND ALARM(\"error-rate-high\")"
    #
    # Compose alarm names from AwsCloudwatchAlarm resources via their exported
    # `alarm_name` stack output. Maximum 10240 characters.
    alarm_rule = string

    # Human-readable description of what this composite alarm represents and
    # what to do when it fires. Maximum 1024 characters.
    alarm_description = optional(string, "")

    # Whether actions execute when the composite alarm changes state. When
    # unset, AWS defaults to true. Unlike metric alarms, this flag is
    # create-time-only on composite alarms (ForceNew): changing it replaces
    # the alarm.
    actions_enabled = optional(bool)

    # Actions to execute when the composite alarm transitions to ALARM state.
    # Composite alarms support SNS topic ARNs and Systems Manager OpsItem
    # actions (not Auto Scaling or EC2 actions — those belong on metric
    # alarms). Maximum 5 actions.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    alarm_actions = optional(list(string), [])

    # Actions to execute when the composite alarm transitions to OK state.
    # Maximum 5 actions.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    ok_actions = optional(list(string), [])

    # Actions to execute when the composite alarm transitions to
    # INSUFFICIENT_DATA state. Maximum 5 actions.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    insufficient_data_actions = optional(list(string), [])

    # Suppresses this composite alarm's actions while a designated suppressor
    # alarm is in ALARM state — the mechanism for maintenance windows and
    # deploy freezes. State transitions still happen and are recorded; only
    # the actions are withheld.
    actions_suppressor = optional(object({
      # The alarm that suppresses actions while it is in ALARM state, addressed
      # by alarm NAME (the CloudWatch API contract). Reference an
      # AwsCloudwatchAlarm's exported `alarm_name` output, or provide a literal
      # alarm name.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      alarm = string

      # Maximum time (in seconds) the composite alarm waits for the suppressor
      # alarm to enter ALARM state after the composite itself transitions,
      # before concluding the suppressor is not firing and executing actions.
      #
      # AWS requires this field whenever a suppressor is configured (the
      # PutCompositeAlarm contract) — both engines always send it with the
      # suppressor. 0 means the composite acts immediately unless the suppressor
      # is already in ALARM.
      wait_period = optional(number, 0)

      # Maximum time (in seconds) actions remain suppressed AFTER the suppressor
      # alarm leaves ALARM state — a grace window that avoids acting on
      # transitions that occur while the suppression is winding down.
      #
      # AWS requires this field whenever a suppressor is configured (the
      # PutCompositeAlarm contract) — both engines always send it with the
      # suppressor. 0 ends suppression the moment the suppressor recovers.
      extension_period = optional(number, 0)
    }))
  })
}
