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
  description = "GcpVertexAiFeatureOnlineStore specification"
  type = object({
    # The GCP project the store lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Vertex AI location (region) the store lives in, e.g.
    # "us-central1". Feature groups its views serve must be in the same
    # location. Immutable.
    location = string

    # The store's id -- the last segment of its resource name. Up to 60
    # characters of lowercase letters, digits, and underscores; the first
    # character cannot be a digit (hyphens are not allowed, so
    # metadata.name cannot stand in). Unique within the project and
    # location. Immutable.
    feature_online_store_id = string

    # Labels on the store. The platform attribution labels are merged in and
    # win on key conflicts.
    labels = optional(map(string), {})

    # Serve from a managed Bigtable instance. Exactly one of bigtable or
    # optimized.
    bigtable = optional(object({
      # Node bounds for the managed Bigtable instance.
      auto_scaling = object({
        # Fewest nodes kept running -- at least 1. Each node bills around the
        # clock.
        min_node_count = optional(number, 0)

        # Most nodes Google may scale to -- at least min_node_count and at most
        # ten times it.
        max_node_count = optional(number, 0)

        # The CPU utilization (10-80 percent) Bigtable scales to hold: above it
        # nodes are added, well below it nodes are removed. Google defaults to
        # 50. Sent only when set.
        cpu_utilization_target = optional(number)
      })

      # True lets clients read the managed Bigtable instance directly, beside
      # the Feature Store serving API.
      enable_direct_bigtable_access = optional(bool, false)

      # The zone the Bigtable instance is created in, e.g. "us-central1-a".
      # Google picks one in the store's region when empty. Sent only when set.
      zone = optional(string, "")
    }))

    # True serves from Google's Optimized online serving (a dedicated
    # endpoint, the lowest latency). Google's optimized block carries no
    # settings, so the choice is a flag. Exactly one of bigtable or
    # optimized.
    optimized = optional(bool, false)

    # The store's dedicated serving endpoint. Set it to serve over Private
    # Service Connect; omit to keep Google's default (a public dedicated
    # endpoint on Optimized stores). Sent only when set.
    dedicated_serving_endpoint = optional(object({
      # Private Service Connect for the dedicated endpoint.
      private_service_connect_config = optional(object({
        # True serves the dedicated endpoint only through a PSC service
        # attachment (the `service_attachment` output) that consumers target
        # with a forwarding rule; false keeps the public endpoint.
        enable_private_service_connect = optional(bool, false)

        # Projects (IDs or numbers) allowed to create forwarding rules that
        # target the service attachment.
        project_allowlist = optional(list(string), [])
      }))
    }))

    # Customer-managed encryption key protecting both the online and the
    # offline data: a GcpKmsKey reference or a literal
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
    # in the store's region. Omit to use Google-managed encryption.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # True lets a destroy delete the store even when it still holds feature
    # views and features the block does not manage (views created outside
    # this block, for example). Without it Google refuses to delete a
    # non-empty store. The views declared below are deleted first either
    # way.
    force_destroy = optional(bool, false)

    # The feature views the store serves, each keyed by feature_view_id.
    feature_views = optional(list(object({
      # The view's id -- what serving clients name. Up to 60 characters of
      # lowercase letters, digits, and underscores; the first character cannot
      # be a digit. Unique within the store. Immutable.
      feature_view_id = string

      # Labels on the view. The platform attribution labels are merged in and
      # win on key conflicts.
      labels = optional(map(string), {})

      # Materialize a BigQuery table or view directly.
      big_query_source = optional(object({
        # The BigQuery table or view materialized on each sync: a
        # GcpBigQueryTable reference (its {project}.{dataset}.{table} name), a
        # literal "project.dataset.table", or the bq:// form Google stores. The
        # modules add the bq:// prefix when it is missing.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        uri = string

        # The columns that form the entity ID the view serves by. Google
        # supports exactly one today.
        entity_id_columns = list(string)
      }))

      # Serve features registered in feature groups.
      feature_registry_source = optional(object({
        # The feature groups and features the view serves.
        feature_groups = list(object({
          # The feature group: a GcpVertexAiFeatureGroup reference or its literal
          # id. The group must be in the store's location.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          feature_group_id = string

          # The features to serve, by feature_id within the group.
          feature_ids = list(string)
        }))

        # The number of the project that owns the feature groups, when it is not
        # the store's project: a GcpProject reference (its project number) or a
        # literal number.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        project_number = optional(string, "")
      }))

      # When the view copies fresh values in. Omit for Google's default.
      sync_config = optional(object({
        # A cron schedule for batch syncs, e.g. "0 */6 * * *"; prefix
        # "CRON_TZ=America/New_York " (or "TZ=") to pin a time zone. Sent only
        # when set.
        cron = optional(string, "")

        # True syncs continuously as the source changes instead of on a
        # schedule.
        continuous = optional(bool, false)
      }))
    })), [])

    # What happens to the store and its feature views when this resource is
    # destroyed:
    #   "" / "DELETE" -- the views and the store are deleted (the sources are
    #                    never touched)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- everything leaves management and keeps serving (and
    #                    billing)
    deletion_policy = optional(string, "")
  })
}
