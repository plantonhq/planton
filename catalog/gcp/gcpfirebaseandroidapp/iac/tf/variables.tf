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
  description = "GcpFirebaseAndroidApp specification"
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
    # shows it. Updatable.
    display_name = string

    # The canonical package name of the Android app, as it appears in the
    # Play Console (e.g. com.example.app). This is the app's IDENTITY in
    # Firebase and in google-services.json: it is immutable (a change
    # replaces the registration -- a new app id, a new config file), and a
    # project accepts each package name exactly once. Must be a valid Java
    # package name: two or more dot-separated segments, each starting with a
    # letter and made of letters, digits, and underscores.
    package_name = string

    # SHA-1 fingerprints of the app's signing certificates (debug, release,
    # Play App Signing), 40 hex digits, with or without colon separators.
    # Cloud Messaging needs NONE of these; they are required by Google
    # Sign-In, Dynamic Links, and Phone Authentication. Updatable.
    sha1_hashes = optional(list(string), [])

    # SHA-256 fingerprints of the app's signing certificates, 64 hex digits,
    # with or without colon separators. Cloud Messaging needs none; App Check
    # with Play Integrity REQUIRES the fingerprint of the certificate the
    # shipped APK is signed with (Play App Signing's certificate for a Play
    # release) -- without it attestation fails at runtime. Updatable.
    sha256_hashes = optional(list(string), [])

    # The API key the app presents to Firebase, by the key's Google-assigned
    # UID: a literal, or a reference to a GcpApiKey resource (its uid
    # output). The key is a CLIENT IDENTIFIER that ships inside the APK in
    # google-services.json -- not a secret -- and it must be valid for this
    # app: unrestricted, or restricted to this package name and certificate
    # with API restrictions that include the Firebase APIs the app uses
    # (Firebase Installations and FCM Registration for push). Google's
    # recommended practice is one restricted key per app. When omitted,
    # Firebase associates an existing valid key or provisions a new,
    # unrestricted one; the api_key_id output reports which. Updatable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    api_key_id = optional(string, "")

    # Firebase App Check for this app: the attestation provider that proves
    # requests come from the genuine app on a genuine device (Play
    # Integrity), and debug tokens that let development builds and CI pass
    # App Check without a real device. Whether the project's backends
    # ENFORCE App Check is configured on the GcpFirebaseProject.
    app_check = optional(object({
      # Attest with the Play Integrity API -- Google's attestation provider for
      # Android apps distributed through Google Play. Requires the app's
      # SHA-256 signing fingerprint in sha256_hashes. Omit the block to leave
      # the provider unconfigured.
      play_integrity = optional(object({
        # Whether Play Integrity attestation is configured for the app. Defaults
        # to true when the block is present, so `playIntegrity: {}` means
        # "configured, with Google's default token lifetime"; set false to
        # declare the block and leave the provider unconfigured. The wire has no
        # switch of its own -- the configuration's existence is the switch -- so
        # this field is how the manifest states the intent explicitly.
        enabled = optional(bool)

        # How long an App Check token exchanged from a Play Integrity verdict
        # stays valid, as a duration in seconds with an "s" suffix (e.g. "3600s",
        # "1800s", "604800s"). Google accepts 30 minutes to 7 days inclusive and
        # assumes 1 hour when unset. Shorter lifetimes re-attest more often;
        # longer ones spare the device.
        token_ttl = optional(string, "")
      }))

      # Debug tokens: pre-shared secrets a development build or a CI emulator
      # presents in place of a device attestation. Each token is a SECRET --
      # anyone holding it passes App Check as this app -- so it is a managed
      # secret reference on the platform, never plaintext. Revoke by removing
      # the entry.
      debug_tokens = optional(list(object({
        # A name for the token as the Firebase console shows it -- the
        # developer's machine or the CI lane it is for. Unique within the app;
        # updatable.
        display_name = string

        # The token value itself: a UUID (version 4), case-insensitive, that the
        # development build registers with the App Check debug provider. Google
        # never returns it after creation, and it cannot be changed -- a new
        # value is a new token. It is a SECRET: the value must be a managed
        # secret reference on the platform.
        token = string
      })), [])
    }))

    # What destroying this resource does to the app registration (and to
    # the composed debug tokens):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the app is removed IMMEDIATELY and PERMANENTLY: Firebase
    #                normally keeps a removed app recoverable for 30 days;
    #                this path skips that window. The package name becomes
    #                registrable again at once.
    #   "PREVENT" -- destroy FAILS; use for a shipped app whose users hold
    #                its google-services.json
    #   "ABANDON" -- the app is left registered and removed from management
    # The Play Integrity configuration is a per-app singleton Google never
    # deletes: it is removed from management on destroy and disappears only
    # with the app itself.
    deletion_policy = optional(string, "")
  })
}
