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
  description = "GcpModelArmorFloorSetting specification"
  type = object({
    # Whose floor this is. Omit for the provider's default project: an empty
    # scope and a project_id naming that same project are the same floor.
    # A deploy whose provider has no default project is refused by both
    # engines before anything is sent. Changing the scope replaces the
    # floor (see above).
    scope = optional(object({
      # Project floor: a literal project ID or a GcpProject reference.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      project_id = optional(string, "")

      # Folder floor: the folder's numeric ID, without the folders/ prefix --
      # a literal or a GcpFolder reference. Every project and folder beneath
      # it inherits the floor.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      folder_id = optional(string, "")

      # Organization floor: the numeric organization ID, without the
      # organizations/ prefix. The floor for the whole estate.
      organization_id = optional(string, "")
    }))

    # The location of the floor setting. Google manages floor settings at
    # "global", the default; set a region only if Google documents a
    # regional floor for your case. Changing it replaces the floor (see
    # above).
    location = optional(string, "")

    # Turn the floor on. While false the floor is recorded but nothing is
    # checked or screened -- the state to apply before walking away from a
    # floor, since destroy leaves the last applied floor in force.
    enable_floor_setting_enforcement = optional(bool, false)

    # The Google services whose traffic the floor screens directly:
    #   "AI_PLATFORM"       -- Vertex AI model calls (configure the behavior
    #                          in ai_platform_floor_setting, required with it)
    #   "GOOGLE_MCP_SERVER" -- Google-hosted MCP servers (configure it in
    #                          google_mcp_server_floor_setting, required with
    #                          it)
    # Each service and its setting come together: a service without its
    # setting, or a setting without its service, is refused. Empty: the
    # floor only governs templates.
    integrated_services = optional(list(string), [])

    # The minimum filters. Required, but an empty block is accepted and sets
    # no minimum: a filter constrains templates and traffic only once its
    # own block is set.
    filter_config = object({
      # Malicious URL detection.
      malicious_uri_filter_settings = optional(object({
        # "ENABLED" or "DISABLED". Unset is Google's default, which is disabled.
        filter_enforcement = optional(string, "")
      }))

      # Prompt injection and jailbreak detection.
      pi_and_jailbreak_filter_settings = optional(object({
        # "ENABLED" or "DISABLED". Unset is Google's default, which is disabled.
        filter_enforcement = optional(string, "")

        # The weakest threshold a template may use, and the threshold integrated
        # services' traffic is screened at: "LOW_AND_ABOVE" (strictest),
        # "MEDIUM_AND_ABOVE", or "HIGH". Unset is Google's default, which is
        # LOW_AND_ABOVE.
        confidence_level = optional(string, "")
      }))

      # Responsible AI content filters.
      rai_settings = optional(object({
        # One entry per required content category.
        rai_filters = list(object({
          # "SEXUALLY_EXPLICIT", "HATE_SPEECH", "HARASSMENT", or "DANGEROUS".
          filter_type = string

          # The weakest threshold a template may use for this category, and the
          # threshold integrated services' traffic is screened at:
          # "LOW_AND_ABOVE" (strictest), "MEDIUM_AND_ABOVE", or "HIGH". Unset is
          # Google's default, which is LOW_AND_ABOVE.
          confidence_level = optional(string, "")
        }))
      }))

      # Sensitive Data Protection.
      sdp_settings = optional(object({
        # Google's predefined sensitive-data detectors.
        basic_config = optional(object({
          # "ENABLED" or "DISABLED". Unset is Google's default, which is disabled.
          filter_enforcement = optional(string, "")
        }))

        # Your own Sensitive Data Protection templates.
        advanced_config = optional(object({
          # Inspect template:
          # projects/{project}/locations/{location}/inspectTemplates/{template}.
          inspect_template = optional(string, "")

          # De-identify template:
          # projects/{project}/locations/{location}/deidentifyTemplates/{template}.
          # Every info type it names must also be in the inspect template.
          deidentify_template = optional(string, "")
        }))
      }))
    })

    # How the floor treats Vertex AI traffic. Set exactly when AI_PLATFORM
    # is in integrated_services.
    ai_platform_floor_setting = optional(object({
      # What happens when the floor's filters trip on the service's traffic:
      #   "INSPECT_ONLY"      -- the verdict is recorded and the call proceeds
      #                          (the safe way to roll a floor out)
      #   "INSPECT_AND_BLOCK" -- the call is blocked
      # Required: Google needs exactly one of the two.
      enforcement_type = string

      # Write the floor's verdicts for this service to Cloud Logging.
      enable_cloud_logging = optional(bool, false)
    }))

    # How the floor treats Google MCP server traffic. Set exactly when
    # GOOGLE_MCP_SERVER is in integrated_services.
    google_mcp_server_floor_setting = optional(object({
      # What happens when the floor's filters trip on the service's traffic:
      #   "INSPECT_ONLY"      -- the verdict is recorded and the call proceeds
      #                          (the safe way to roll a floor out)
      #   "INSPECT_AND_BLOCK" -- the call is blocked
      # Required: Google needs exactly one of the two.
      enforcement_type = string

      # Write the floor's verdicts for this service to Cloud Logging.
      enable_cloud_logging = optional(bool, false)
    }))

    # Screen prompts in languages other than English under the floor. Sent
    # only when true.
    enable_multi_language_detection = optional(bool, false)
  })
}
