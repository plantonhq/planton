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
  description = "GcpVertexAiSearchDataConnector specification"
  type = object({
    # The GCP project the collection lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Where the collection lives: "global", "us", or "eu". Engines over its
    # stores must match. Immutable.
    location = string

    # The collection's id -- 1-63 characters, RFC 1034 (lowercase letters,
    # digits, hyphens; starts with a letter). Defaults to metadata.name.
    # Immutable.
    collection_id = optional(string, "")

    # The collection's name in the console (up to 1024 characters). Defaults
    # to metadata.name. Immutable.
    collection_display_name = optional(string, "")

    # The source, as Google names it: first-party "bigquery", "gcp_fhir",
    # "google_mail", "google_drive", "google_calendar", "google_chat";
    # third-party "jira", "confluence", "servicenow", "sharepoint",
    # "onedrive", "outlook", "salesforce", "slack", "notion", "github",
    # "gitlab", "zendesk", "box", "dropbox", "workday", and more (Google's
    # connector documentation is the full list). Immutable.
    data_source = string

    # The source's API version where the connector supports several, e.g.
    # 3 for Jira v3. Sent only when set.
    data_source_version = optional(number)

    # Source connection parameters as string pairs -- instance URI, auth
    # type, client id, and the Secret Manager resource names of the secrets
    # (client_secret, password, refresh_token). Exactly one of params or
    # json_params.
    params = optional(map(string), {})

    # The same parameters as one compact JSON string, for sources whose
    # parameters nest. Exactly one of params or json_params.
    json_params = optional(string, "")

    # How often a full sync runs, as a duration string ("86400s" for daily);
    # 30 minutes to 7 days. Equal to incremental_refresh_interval disables
    # incremental sync.
    refresh_interval = string

    # How often an incremental sync runs (third-party sources), as a
    # duration string; 30 minutes to 7 days. Empty lets Google default
    # (3 hours).
    incremental_refresh_interval = optional(string, "")

    # Pause incremental syncs.
    incremental_sync_disabled = optional(bool, false)

    # Pause full syncs.
    auto_run_disabled = optional(bool, false)

    # PERIODIC (Google's default) or STREAMING. Sent only when set.
    sync_mode = optional(string, "")

    # What the connector does: DATA_INGESTION (index the source), FEDERATED
    # (search the source live without indexing), ACTIONS (act on the source),
    # EUA (end-user authentication), FEDERATED_AND_EUA. Empty lets Google
    # default.
    connector_modes = optional(list(string), [])

    # Give the connector static egress IP addresses (for sources behind an
    # allowlist); the addresses appear in the outputs. Immutable.
    static_ip_enabled = optional(bool, false)

    # Customer-managed encryption key protecting every data store the
    # connector creates: a GcpKmsKey reference or a literal key path. Omit
    # for Google-managed encryption. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # The source entities to ingest; Google creates one data store per
    # entity. Add or remove an entity by editing the list; an entity's name
    # is immutable.
    entities = optional(list(object({
      # The entity's name as the source names it. Supported values depend on
      # the source: Jira -- project, issue, attachment, comment, worklog;
      # Confluence -- Content, Space; Salesforce -- Lead, Opportunity, Contact,
      # Account, Case, Contract, Campaign; ServiceNow -- catalog, incident,
      # knowledge_base. Immutable.
      entity_name = string

      # Entity-specific ingestion parameters as a compact JSON string (Google
      # normalizes it), e.g. {"inclusion_filters":{"knowledgeBaseSysId":["123"]}}.
      params = optional(string, "")

      # Source field -> key property (title, description, ...) so results
      # render a title and description from the right fields.
      key_property_mappings = optional(map(string), {})
    })), [])

    # Where the connector reaches or serves data, keyed by destination.
    destination_configs = optional(list(object({
      # The configuration's key, e.g. "url".
      key = optional(string, "")

      # The destinations for this key.
      destinations = optional(list(object({
        # The destination host, e.g. "https://example.atlassian.net".
        host = optional(string, "")

        # The port the destination accepts.
        port = optional(number)
      })), [])

      # Destination parameters as a compact JSON string, e.g.
      # {"destination_type":"private"}.
      params = optional(string, "")
    })), [])

    # The action side of an ACTIONS-mode connector.
    action_config = optional(object({
      # Connection parameters for the action side, as string pairs. Credentials
      # are named as Secret Manager secret resource names, never pasted.
      action_params = optional(map(string), {})

      # Create a Business Application Platform connection for the actions.
      create_bap_connection = optional(bool, false)
    }))

    # Which actions an ACTIONS-mode connector exposes.
    bap_config = optional(object({
      # Connector modes the BAP connection supports; Google offers ACTIONS.
      supported_connector_modes = optional(list(string), [])

      # Actions enabled on the source, e.g. create_issue, update_issue,
      # change_issue_status, create_comment, update_comment, upload_attachment.
      enabled_actions = optional(list(string), [])
    }))

    # What happens to the connector, its collection, and the data stores it
    # created when this resource is destroyed:
    #   "" / "DELETE" -- deleted, indexed data included
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- everything leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
