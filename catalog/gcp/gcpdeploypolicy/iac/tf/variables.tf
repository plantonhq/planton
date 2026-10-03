variable "metadata" {
  description = "Catalog object metadata"
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
  description = "GcpDeployPolicy specification"
  type = object({
    # The project the policy lives in -- the project of the pipelines and
    # targets it governs: a literal project ID or a GcpProject reference.
    # Empty means the provider's default project. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The region the policy lives in, e.g. "us-central1" -- the region of
    # the pipelines and targets it governs. Required. Immutable.
    location = string

    # The policy's ID, unique in the project and location: 1-63 lowercase
    # letters, digits, and hyphens, starting with a letter and not ending
    # with a hyphen. Defaults to metadata.name. Immutable.
    deploy_policy_id = optional(string, "")

    # A description of the policy, up to 255 characters -- say what the
    # freeze is for and who to ask for an override.
    description = optional(string, "")

    # Labels on the policy resource. The platform attribution labels are
    # added on top and win on key conflicts. These label the policy itself;
    # to select pipelines or targets by label, use selectors[].
    labels = optional(map(string), {})

    # Annotations on the policy (user metadata Cloud Deploy never reads).
    # Only the keys declared here are managed.
    annotations = optional(map(string), {})

    # True keeps the policy but stops enforcing it: actions it would block
    # go through. Use it to lift a freeze early without deleting it.
    suspended = optional(bool, false)

    # What the policy restricts. At least one rule; a blocked action is one
    # that any rule matches.
    rules = list(object({
      # A rollout restriction: which actions are blocked, for whom, and when.
      # It is the only kind of rule Cloud Deploy defines, so set it on every
      # rule -- a rule without one restricts nothing.
      rollout_restriction = optional(object({
        # The restriction's ID, unique in the policy and shown in the violation
        # message: 1-63 lowercase letters, digits, and hyphens, starting with a
        # letter and not ending with a hyphen. Required.
        id = string

        # The rollout actions blocked. Empty blocks every action:
        #   "ADVANCE"          -- advancing a rollout to its next phase
        #   "APPROVE"          -- approving a rollout
        #   "CANCEL"           -- cancelling a rollout
        #   "CREATE"           -- creating a rollout (a deploy or a promotion)
        #   "IGNORE_JOB"       -- ignoring a failed job
        #   "RETRY_JOB"        -- retrying a failed job
        #   "ROLLBACK"         -- rolling a target back
        #   "TERMINATE_JOBRUN" -- terminating a running job
        actions = optional(list(string), [])

        # Who is blocked. Empty blocks both:
        #   "USER"              -- a person or a script calling Cloud Deploy
        #   "DEPLOY_AUTOMATION" -- the pipeline's own automations (promotions,
        #                          advances, repairs)
        invokers = optional(list(string), [])

        # When the actions are blocked. Required.
        time_windows = object({
          # The IANA time zone every window is read in, e.g. "America/New_York"
          # or "Europe/Berlin". Required.
          time_zone = string

          # Dated windows that happen once, e.g. a year-end freeze from December
          # 20 at 17:00 to January 3 at 09:00.
          one_time_windows = optional(list(object({
            # The first day of the window. Required.
            start_date = object({
              # The year, 1-9999.
              year = optional(number, 0)

              # The month, 1-12.
              month = optional(number, 0)

              # The day of the month, 1-31 and valid for the month.
              day = optional(number, 0)
            })

            # The time on start_date the window opens (inclusive); 00:00 is the
            # beginning of the day. Required.
            start_time = object({
              # Hours, 0-23; 24 is accepted for the end of the day.
              hours = optional(number, 0)

              # Minutes, 0-59.
              minutes = optional(number, 0)

              # Seconds, 0-59 (60 where a leap second is allowed).
              seconds = optional(number, 0)

              # Fractions of a second in nanoseconds, 0-999999999.
              nanos = optional(number, 0)
            })

            # The last day of the window. Required.
            end_date = object({
              # The year, 1-9999.
              year = optional(number, 0)

              # The month, 1-12.
              month = optional(number, 0)

              # The day of the month, 1-31 and valid for the month.
              day = optional(number, 0)
            })

            # The time on end_date the window closes (exclusive); 24:00 is the end
            # of the day. Required.
            end_time = object({
              # Hours, 0-23; 24 is accepted for the end of the day.
              hours = optional(number, 0)

              # Minutes, 0-59.
              minutes = optional(number, 0)

              # Seconds, 0-59 (60 where a leap second is allowed).
              seconds = optional(number, 0)

              # Fractions of a second in nanoseconds, 0-999999999.
              nanos = optional(number, 0)
            })
          })), [])

          # Windows that recur every week, e.g. every Friday from 15:00 to 24:00,
          # or all of Saturday and Sunday.
          weekly_windows = optional(list(object({
            # The days the window recurs on: MONDAY, TUESDAY, WEDNESDAY, THURSDAY,
            # FRIDAY, SATURDAY, SUNDAY. Empty means every day.
            days_of_week = optional(list(string), [])

            # The time each day the window opens (inclusive); 00:00 is the
            # beginning of the day. Set together with end_time; leave both unset to
            # block the whole of each day.
            start_time = optional(object({
              # Hours, 0-23; 24 is accepted for the end of the day.
              hours = optional(number, 0)

              # Minutes, 0-59.
              minutes = optional(number, 0)

              # Seconds, 0-59 (60 where a leap second is allowed).
              seconds = optional(number, 0)

              # Fractions of a second in nanoseconds, 0-999999999.
              nanos = optional(number, 0)
            }))

            # The time each day the window closes (exclusive); 24:00 is midnight at
            # the end of the day. Set together with start_time.
            end_time = optional(object({
              # Hours, 0-23; 24 is accepted for the end of the day.
              hours = optional(number, 0)

              # Minutes, 0-59.
              minutes = optional(number, 0)

              # Seconds, 0-59 (60 where a leap second is allowed).
              seconds = optional(number, 0)

              # Fractions of a second in nanoseconds, 0-999999999.
              nanos = optional(number, 0)
            }))
          })), [])
        })
      }))
    }))

    # Which pipelines and targets the policy applies to. At least one
    # selector; the policy applies when ANY selector matches, and within a
    # selector EVERY attribute given must match (its delivery_pipeline and
    # its target, by ID and by labels).
    selectors = list(object({
      # The delivery pipelines this selector matches.
      delivery_pipeline = optional(object({
        # The pipeline's ID (the last segment of its name, in the policy's
        # project and location): a GcpDeliveryPipeline reference, a literal ID,
        # or "*" for every pipeline in the location. Empty matches by labels
        # alone.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        id = optional(string, "")

        # Labels a pipeline must carry, all of them, to match.
        labels = optional(map(string), {})
      }))

      # The targets this selector matches.
      target = optional(object({
        # The target's ID (the last segment of its name, in the policy's
        # project and location): a GcpDeployTarget reference, a literal ID, or
        # "*" for every target in the location. Empty matches by labels alone.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        id = optional(string, "")

        # Labels a target must carry, all of them, to match.
        labels = optional(map(string), {})
      }))
    }))

    # What destroy does:
    #   "" / "DELETE" -- the policy is deleted
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the policy leaves management and stays in Google
    deletion_policy = optional(string, "")
  })
}
