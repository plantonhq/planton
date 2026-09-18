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
  description = "GcpCertificateMap specification"
  type = object({
    # The GCP project to create the map in. Can be a literal project ID or
    # a reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The map name in GCP. Defaults to metadata.name when left empty.
    # Immutable: changing it replaces the map (and every entry — detach the
    # map from proxies first).
    map_name = optional(string, "")

    # What this map routes — which domains, which environments. Shown in
    # the console.
    description = optional(string, "")

    # User labels attached to the map (merged with the platform's standard
    # labels by the module).
    labels = optional(map(string), {})

    # The routing entries. A map with no entries is legal (attach entries
    # later), but a proxy consulting it will fail every handshake until a
    # matching entry (or a PRIMARY fallback) exists.
    entries = optional(list(object({
      # The entry name in GCP (unique within the map). Immutable: changing it
      # replaces the entry.
      entry_name = string

      # The hostname this entry serves: a FQDN (example.com) or a wildcard
      # expression (*.example.com) matched against the client's SNI. Exactly
      # one of hostname or matcher. Immutable. The API validates COVERAGE at
      # entry-create time (live-verified 400: certificate "..." does not
      # cover map entry hostname "...") — every attached certificate's
      # domain set must cover this hostname, regardless of the
      # certificate's provisioning state.
      hostname = optional(string, "")

      # A predefined matcher instead of a hostname. The API's documented
      # value is "PRIMARY" — the fallback entry used when no hostname entry
      # matches the SNI (the value list is API-side and not walled here).
      # Exactly one of hostname or matcher. Immutable.
      matcher = optional(string, "")

      # The certificates presented when this entry matches — 1 to 15 (the
      # API's per-entry cap). Each is a full certificate resource name
      # (projects/{project}/locations/{location}/certificates/{name}) — a
      # literal or a reference to a GcpCertManagerCert resource (its
      # certificate_id output). Mutable: rotate certificates by editing this
      # list in place.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      certificates = list(string)

      # What this entry serves. Shown in the console.
      description = optional(string, "")

      # User labels attached to the entry.
      labels = optional(map(string), {})
    })), [])

    # Deletion policy — what happens when this resource is destroyed
    # (applied to the map and every entry):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- entries and map are deleted; a proxy still referencing
    #                the map fails TLS handshakes (detach first)
    #   "PREVENT" -- destroy FAILS; protects live TLS routing
    #   "ABANDON" -- resources are removed from management but keep serving
    #                in GCP
    deletion_policy = optional(string, "")
  })
}
