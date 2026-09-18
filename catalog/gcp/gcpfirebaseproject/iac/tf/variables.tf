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
  description = "GcpFirebaseProject specification"
  type = object({
    # The GCP project to enable Firebase on. Can be a literal project ID or a
    # reference to a GcpProject resource. If omitted, the provider's default
    # project is used. Immutable: Firebase enablement is a property of the
    # project itself.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Create the project's DEFAULT Cloud Storage for Firebase bucket in this
    # location (a Cloud Storage location: a multi-region such as US or EU, a
    # dual-region such as NAM4, or a region such as us-central1). The default
    # bucket is the one the Firebase client SDKs use when no bucket is named
    # and is created at most once per project; leave empty to skip it.
    # Immutable: the bucket's location cannot change after creation.
    #
    # Requires the project to be on Firebase's pay-as-you-go (Blaze) plan --
    # that is, linked to a Cloud Billing account. On a project without
    # billing the create fails.
    default_storage_location = optional(string, "")

    # Firebase App Check: verify that requests to your Firebase backends come
    # from your genuine apps. Per-app attestation providers (Play Integrity,
    # App Attest, DeviceCheck, reCAPTCHA) are configured on each app
    # registration; the PROJECT-level part -- which backend services enforce
    # App Check, and per-resource overrides -- lives here.
    app_check = optional(object({
      # Enforcement per Firebase backend service. A service not listed here is
      # in Google's default OFF state (no enforcement, no metrics). Listing a
      # service with enforcement_mode UNENFORCED starts collecting metrics
      # without rejecting anything -- the recommended first step before
      # switching to ENFORCED.
      service_configs = optional(list(object({
        # The service to configure. Google supports exactly these:
        #   firebasestorage.googleapis.com  -- Cloud Storage for Firebase
        #   firebasedatabase.googleapis.com -- Firebase Realtime Database
        #   firestore.googleapis.com        -- Cloud Firestore
        #   identitytoolkit.googleapis.com  -- Firebase Authentication
        # Immutable: to move enforcement to another service, add a new entry.
        service_id = string

        # How the service treats requests without a valid App Check token:
        #   ""           -- OFF: not enforced, no metrics (Google's default for
        #                   an unlisted service; listing a service with this
        #                   value records the intent explicitly)
        #   "UNENFORCED" -- not enforced, but metrics show what WOULD be
        #                   rejected -- the safe first step
        #   "ENFORCED"   -- requests without a valid token are rejected
        enforcement_mode = optional(string, "")
      })), [])

      # Per-resource overrides of a service's enforcement, for the services
      # that support them (today: individual OAuth clients under Google
      # Identity for iOS).
      resource_policies = optional(list(object({
        # The service the resource belongs to. Google supports exactly one today:
        # oauth2.googleapis.com (Google Identity for iOS). Immutable.
        service_id = string

        # The resource the override applies to, in the service's own format --
        # for an iOS OAuth client:
        # //oauth2.googleapis.com/projects/{project_number}/oauthClients/{client_id}
        target_resource = string

        # The enforcement for this resource, overriding the service's setting:
        # "" (OFF), "UNENFORCED", or "ENFORCED" -- same semantics as the
        # service-level enforcement_mode.
        enforcement_mode = optional(string, "")
      })), [])
    }))

    # What destroying this resource does to the COMPOSED settings (the
    # default storage bucket and the App Check configurations). Firebase
    # enablement itself is not deletable and is always detached:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the default bucket is deleted (and its objects with it)
    #                and App Check enforcement is switched OFF for every
    #                configured service
    #   "PREVENT" -- destroy FAILS; use for a production project whose
    #                default bucket holds user data
    #   "ABANDON" -- the bucket and the App Check settings are left in place
    #                and removed from management
    deletion_policy = optional(string, "")
  })
}
