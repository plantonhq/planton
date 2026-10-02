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
  description = "GcpFirebaseWebApp specification"
  type = object({
    # The GCP project the app is registered in -- the Firebase-enabled
    # project. Its VALUE is the GCP project id (the same string every GCP
    # kind's project_id carries), so a literal works; the default REFERENCE
    # is a GcpFirebaseProject resource, because an app registration exists
    # only inside a project's Firebase enablement: referencing the enablement
    # orders this app after it in a chart and places it inside the Firebase
    # project on diagrams. If omitted, the provider's default project is
    # used. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The user-assigned display name of the app, as the Firebase console
    # shows it. The only identity a web app has; updatable.
    display_name = string

    # The API key the app presents to Firebase, by the key's Google-assigned
    # UID: a literal, or a reference to a GcpApiKey resource (its uid
    # output). The key is a CLIENT IDENTIFIER that ships in the page as
    # firebaseConfig.apiKey -- not a secret -- and it must be valid for this
    # app: unrestricted, or restricted to the site's HTTP referrers with API
    # restrictions that include the Firebase APIs the app uses (Firebase
    # Installations and FCM Registration for push). Google's recommended
    # practice is one referrer-restricted key per web app. When omitted,
    # Firebase associates an existing valid key or provisions a new,
    # unrestricted one; the api_key_id and api_key outputs report which.
    # Updatable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    api_key_id = optional(string, "")

    # Firebase App Check for this app: the attestation provider that proves
    # requests come from your genuine site (reCAPTCHA v3 or reCAPTCHA
    # Enterprise), and debug tokens that let local development pass App
    # Check. Whether the project's backends ENFORCE App Check is configured
    # on the GcpFirebaseProject.
    app_check = optional(object({
      # Attest with reCAPTCHA v3 -- the score-based, no-interaction reCAPTCHA
      # registered at google.com/recaptcha (site type "reCAPTCHA v3"). Google
      # verifies the client's reCAPTCHA token with the site SECRET configured
      # here. Omit the block to leave the provider unconfigured.
      recaptcha_v3 = optional(object({
        # The reCAPTCHA v3 SITE SECRET from the reCAPTCHA admin console -- the
        # server-side half of the key pair (the site key goes in the page). A
        # SECRET Google never returns after it is set (the module reports only
        # whether it is set), so the value must be a managed secret reference on
        # the platform. Updatable (rotate by setting a new secret).
        site_secret = string

        # How long an App Check token exchanged from a reCAPTCHA v3 token stays
        # valid, as a duration in seconds with an "s" suffix (e.g. "3600s",
        # "1800s", "604800s"). Google accepts 30 minutes to 7 days inclusive and
        # assumes 1 hour when unset.
        token_ttl = optional(string, "")
      }))

      # Attest with reCAPTCHA Enterprise -- the Google Cloud reCAPTCHA product
      # with its own assessments, billing, and console. Google verifies the
      # client's token against the site KEY (a public key) configured here.
      # Omit the block to leave the provider unconfigured.
      recaptcha_enterprise = optional(object({
        # The reCAPTCHA Enterprise SITE KEY, created in the Google Cloud console
        # for the site's domains (a score-based key). This is the PUBLIC half of
        # the key -- the same value the page embeds -- so it is an identifier,
        # not a secret; reCAPTCHA Enterprise has no site secret. Updatable.
        site_key = string

        # How long an App Check token exchanged from a reCAPTCHA Enterprise
        # assessment stays valid, as a duration in seconds with an "s" suffix
        # (e.g. "3600s"). Google accepts 30 minutes to 7 days inclusive and
        # assumes 1 hour when unset.
        token_ttl = optional(string, "")
      }))

      # Debug tokens: pre-shared secrets a local development build presents in
      # place of a reCAPTCHA attestation. Each token is a SECRET -- anyone
      # holding it passes App Check as this app -- so it is a managed secret
      # reference on the platform, never plaintext. Revoke by removing the
      # entry.
      debug_tokens = optional(list(object({
        # A name for the token as the Firebase console shows it -- the
        # developer's machine or the CI lane it is for. Unique within the app;
        # updatable.
        display_name = string

        # The token value itself: a UUID (version 4), case-insensitive, that the
        # development build registers with the App Check debug provider
        # (self.FIREBASE_APPCHECK_DEBUG_TOKEN). Google never returns it after
        # creation, and it cannot be changed -- a new value is a new token. It
        # is a SECRET: the value must be a managed secret reference on the
        # platform.
        token = string
      })), [])
    }))

    # What destroying this resource does to the app registration (and to
    # the composed debug tokens):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the app is removed IMMEDIATELY and PERMANENTLY: Firebase
    #                normally keeps a removed app recoverable for 30 days;
    #                this path skips that window.
    #   "PREVENT" -- destroy FAILS; use for a shipped site whose pages carry
    #                this app's firebaseConfig
    #   "ABANDON" -- the app is left registered and removed from management
    # The reCAPTCHA configurations are per-app singletons Google never
    # deletes: they are removed from management on destroy and disappear
    # only with the app itself.
    deletion_policy = optional(string, "")
  })
}
