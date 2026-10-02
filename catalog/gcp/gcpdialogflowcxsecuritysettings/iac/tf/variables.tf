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
  description = "GcpDialogflowCxSecuritySettings specification"
  type = object({
    # The GCP project the settings live in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Dialogflow CX location, e.g. "global" or "us-central1". The
    # settings apply only to agents in this same location, and the inspect
    # and de-identify templates must live in this region too. Immutable.
    location = string

    # Human-readable name, unique within the location. Defaults to
    # metadata.name. Mutable in place.
    display_name = optional(string, "")

    # How Dialogflow redacts sensitive data before it persists anything:
    #   REDACT_WITH_SERVICE -- call Sensitive Data Protection to scrub the
    #                          data (the inspect and de-identify templates
    #                          below shape what is found and how it is
    #                          replaced)
    # Empty means no redaction.
    redaction_strategy = optional(string, "")

    # Which data redaction applies to:
    #   REDACT_DISK_STORAGE -- everything written to disk or other durable
    #                          storage, temporary files included
    # Empty means nothing is redacted, even with a redaction strategy set.
    redaction_scope = optional(string, "")

    # A Sensitive Data Protection inspect template deciding WHAT counts as
    # sensitive (info types, likelihood, custom detectors):
    # projects/{project}/locations/{location}/inspectTemplates/{template} or
    # organizations/{org}/locations/{location}/inspectTemplates/{template},
    # in the same region as these settings. Empty uses Google's default
    # inspect configuration.
    inspect_template = optional(string, "")

    # A Sensitive Data Protection de-identify template deciding HOW found
    # data is replaced (masking, tokenization, bucketing):
    # projects/{project}/locations/{location}/deidentifyTemplates/{template}
    # or the organizations/ form, in the same region. Empty replaces
    # sensitive values with "[redacted]".
    deidentify_template = optional(string, "")

    # What the retention rule purges when it fires. DIALOGFLOW_HISTORY
    # (conversation history) is the only data type Google defines.
    purge_data_types = optional(list(string), [])

    # Retain sensitive conversation data only while the conversation lasts:
    #   REMOVE_AFTER_CONVERSATION -- removed when the conversation ends (a
    #                                session without an explicit
    #                                conversation ends with the session)
    # This also turns off audio export and Insights export. Set this or
    # retention_window_days, never both.
    retention_strategy = optional(string, "")

    # Keep sensitive conversation data for this many days. Only values below
    # Dialogflow's default TTL (365 days; 30 for Agent Assist traffic) take
    # effect -- a larger value is ignored and the default applies, as does 0
    # or unset. Set this or retention_strategy, never both.
    retention_window_days = optional(number, 0)

    # Record telephony audio into a Cloud Storage bucket.
    audio_export_settings = optional(object({
      # The bucket the audio lands in: a GcpGcsBucket reference or a literal
      # bucket name. Setting it makes Google grant the Dialogflow service agent
      # roles/storage.objectCreator on the bucket, so whoever applies this
      # needs storage.buckets.setIamPolicy there. Objects then follow the
      # bucket's own retention and lifecycle rules.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      gcs_bucket = optional(string, "")

      # The object-name pattern for exported audio files. Empty uses Google's
      # default naming.
      audio_export_pattern = optional(string, "")

      # The file format of exported audio -- telephony recordings only today:
      #   MULAW -- G.711 mu-law PCM at 8 kHz (the telephone-native format)
      #   MP3   -- MP3
      #   OGG   -- OGG Vorbis
      # Empty leaves Google's default.
      audio_format = optional(string, "")

      # Redact sensitive spoken content in the exported audio as well as in the
      # transcripts.
      enable_audio_redaction = optional(bool, false)
    }))

    # Send each finished conversation to Conversational Insights and run its
    # analyzers. Ignored under REMOVE_AFTER_CONVERSATION.
    enable_insights_export = optional(bool, false)

    # What happens to the settings when this resource is destroyed:
    #   "" / "DELETE" -- deleted (an agent that references the settings is
    #                    destroyed first, because it depends on them)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the settings leave management and stay in GCP
    deletion_policy = optional(string, "")
  })
}
