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
  description = "GcpModelArmorTemplate specification"
  type = object({
    # The GCP project the template lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Where the template lives: a region (e.g. "us-central1") or a
    # multi-region ("us", "eu"). Prompts are screened in this location, so
    # pick the one your application's data-residency rules allow; the
    # application calls Model Armor's endpoint for this location. Immutable.
    location = string

    # The template's id -- the last segment of its resource name, what an
    # application passes to Model Armor. Letters, digits, hyphens, and
    # underscores, up to 63 characters. Defaults to metadata.name.
    # Immutable.
    template_id = optional(string, "")

    # Labels on the template. The platform attribution labels are merged in
    # and win on key conflicts.
    labels = optional(map(string), {})

    # Which filters run and how sensitive each one is. Required: a template
    # with no filter configuration screens nothing.
    filter_config = object({
      # Malicious URL detection: flags links to known phishing and malware
      # sites in prompts and responses.
      malicious_uri_filter_settings = optional(object({
        # "ENABLED" or "DISABLED". Unset leaves Google's default (disabled).
        filter_enforcement = optional(string, "")
      }))

      # Prompt injection and jailbreak detection: flags attempts to override
      # the model's instructions or talk it out of its safety rules.
      pi_and_jailbreak_filter_settings = optional(object({
        # "ENABLED" or "DISABLED". Unset leaves Google's default (disabled).
        filter_enforcement = optional(string, "")

        # How confident Model Armor must be before it flags a prompt:
        #   "LOW_AND_ABOVE"    -- flags the most; the strictest setting, with
        #                         the most false positives
        #   "MEDIUM_AND_ABOVE" -- the balanced choice
        #   "HIGH"             -- flags only clear attempts
        confidence_level = optional(string, "")
      }))

      # Responsible AI content filters: sexually explicit, hate speech,
      # harassment, and dangerous content, each with its own threshold.
      rai_settings = optional(object({
        # One entry per content category to screen for; a category not listed
        # is not screened.
        rai_filters = list(object({
          # The content category: "SEXUALLY_EXPLICIT", "HATE_SPEECH",
          # "HARASSMENT", or "DANGEROUS".
          filter_type = string

          # How confident Model Armor must be before it flags content in this
          # category: "LOW_AND_ABOVE" (strictest), "MEDIUM_AND_ABOVE", or "HIGH".
          confidence_level = optional(string, "")
        }))
      }))

      # Sensitive Data Protection: finds (and optionally redacts) personal and
      # confidential data such as card numbers, credentials, and IDs.
      sdp_settings = optional(object({
        # Google's predefined set of sensitive-data detectors (card numbers,
        # government IDs, credentials, and similar). The quick start.
        basic_config = optional(object({
          # "ENABLED" or "DISABLED".
          filter_enforcement = optional(string, "")
        }))

        # Your own Sensitive Data Protection templates: an inspect template
        # decides what counts as sensitive, and an optional de-identify template
        # decides how findings are redacted in the sanitized text.
        advanced_config = optional(object({
          # Inspect template:
          # projects/{project}/locations/{location}/inspectTemplates/{template}.
          # With only an inspect template, Model Armor reports findings.
          inspect_template = optional(string, "")

          # De-identify template:
          # projects/{project}/locations/{location}/deidentifyTemplates/{template}.
          # Adds redaction to the sanitized result. Every info type the
          # de-identify template names must also be in the inspect template.
          deidentify_template = optional(string, "")
        }))
      }))
    })

    # How the template behaves around the filters: whether it blocks or only
    # reports, what an end user sees when a prompt or response is blocked,
    # logging, and which filter version it runs. Omit for Google's defaults.
    template_metadata = optional(object({
      # Log template create, update, and delete operations to Cloud Logging.
      log_template_operations = optional(bool, false)

      # Log every sanitize call (the prompt or response screened and the
      # verdict) to Cloud Logging. Useful for tuning thresholds; the logs
      # carry the screened text.
      log_sanitize_operations = optional(bool, false)

      # Detect and screen prompts in languages other than English. Sent only
      # when true.
      enable_multi_language_detection = optional(bool, false)

      # When one of several detectors fails, return the other detectors'
      # verdicts instead of failing the whole call.
      ignore_partial_invocation_failures = optional(bool, false)

      # The error code a service extension returns to the end user when a
      # prompt trips a filter (for example 403). Sent only when set.
      custom_prompt_safety_error_code = optional(number, 0)

      # The error message returned to the end user when a prompt trips a
      # filter.
      custom_prompt_safety_error_message = optional(string, "")

      # The error code returned to the end user when a model response trips a
      # filter. Sent only when set.
      custom_llm_response_safety_error_code = optional(number, 0)

      # The error message returned to the end user when a model response
      # trips a filter.
      custom_llm_response_safety_error_message = optional(string, "")

      # What happens when a filter trips:
      #   "INSPECT_ONLY"      -- the verdict is reported and the traffic
      #                          passes (a safe way to tune a new template)
      #   "INSPECT_AND_BLOCK" -- the traffic is blocked
      # Unset keeps Google's default.
      enforcement_type = optional(string, "")

      # Which version of Google's filters the template runs. Omit to follow
      # Google's default.
      filter_version_selector = optional(object({
        # Follow a moving version: "FILTER_VERSION_ALIAS_STABLE" (Google's
        # recommended version) or "FILTER_VERSION_ALIAS_LATEST" (newest filters
        # first, verdicts may shift as Google updates them).
        alias = optional(string, "")

        # Pin an exact, immutable filter version such as "v1" or "v2"
        # (case-sensitive), so verdicts never shift under you.
        version = optional(string, "")
      }))
    }))

    # What happens to the template when this resource is destroyed:
    #   "" / "DELETE" -- the template is deleted; applications and engines
    #                    that still name it start failing their screens
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the template leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
