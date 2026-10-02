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
  description = "GcpFirebaseAppleApp specification"
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

    # The app's bundle identifier, as registered with Apple (e.g.
    # com.example.app). This is the app's IDENTITY in Firebase and in
    # GoogleService-Info.plist: it is immutable (a change replaces the
    # registration -- a new app id, a new plist), and a project accepts each
    # bundle id exactly once. Apple allows letters, digits, hyphens, and
    # periods; the reverse-DNS shape is Apple's convention, not a rule.
    bundle_id = string

    # The app's Apple ID in App Store Connect -- the numeric identifier Apple
    # assigns when the app record is created (the number in the App Store
    # URL). Used by Dynamic Links and App Store redirects; not needed for
    # push. Updatable; leave empty until the app has an App Store record.
    app_store_id = optional(string, "")

    # The Apple Developer Team ID that signs the app: ten uppercase letters
    # and digits, from the Membership page of the Apple Developer account.
    # REQUIRED when app_check configures App Attest or DeviceCheck (Apple's
    # attestations validate against the signing team); otherwise optional.
    # Updatable.
    team_id = optional(string, "")

    # The API key the app presents to Firebase, by the key's Google-assigned
    # UID: a literal, or a reference to a GcpApiKey resource (its uid
    # output). The key is a CLIENT IDENTIFIER that ships inside the app
    # bundle in GoogleService-Info.plist -- not a secret -- and it must be
    # valid for this app: unrestricted, or restricted to this bundle id with
    # API restrictions that include the Firebase APIs the app uses (Firebase
    # Installations and FCM Registration for push). Google's recommended
    # practice is one restricted key per app. When omitted, Firebase
    # associates an existing valid key or provisions a new, unrestricted
    # one; the api_key_id output reports which. Updatable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    api_key_id = optional(string, "")

    # Firebase App Check for this app: the attestation providers that prove
    # requests come from the genuine app on a genuine device (App Attest,
    # with DeviceCheck as the fallback for devices that cannot use it), and
    # debug tokens that let development builds and CI pass App Check without
    # a real device. Whether the project's backends ENFORCE App Check is
    # configured on the GcpFirebaseProject.
    app_check = optional(object({
      # Attest with App Attest -- Apple's attestation for iOS 14+ that proves
      # the app is genuine and unmodified. Google's recommended primary
      # provider for Apple apps; configure DeviceCheck beside it as the
      # fallback for devices that cannot use App Attest. Omit the block to
      # leave the provider unconfigured.
      app_attest = optional(object({
        # Whether App Attest attestation is configured for the app. Defaults to
        # true when the block is present, so `appAttest: {}` means "configured,
        # with Google's default token lifetime"; set false to declare the block
        # and leave the provider unconfigured. The wire has no switch of its own
        # -- the configuration's existence is the switch -- so this field is how
        # the manifest states the intent explicitly.
        enabled = optional(bool)

        # How long an App Check token exchanged from an App Attest artifact
        # stays valid, as a duration in seconds with an "s" suffix (e.g. "3600s",
        # "1800s", "604800s"). Google accepts 30 minutes to 7 days inclusive and
        # assumes 1 hour when unset.
        token_ttl = optional(string, "")
      }))

      # Attest with DeviceCheck -- Apple's attestation available on every iOS
      # 11+ device, verified server-side with a DeviceCheck private key you
      # generate in the Apple Developer account. Omit the block to leave the
      # provider unconfigured.
      device_check = optional(object({
        # The Key ID of the DeviceCheck private key, as shown in the Apple
        # Developer account (Certificates, Identifiers & Profiles > Keys): ten
        # uppercase letters and digits. This identifies the key; it is not
        # secret. Updatable (rotate by generating a new key and setting both
        # fields).
        key_id = string

        # The contents of the DeviceCheck private key file (.p8) Apple issued for
        # key_id -- the PEM text, including its BEGIN and END lines. This is a
        # SECRET Google never returns after it is set (the module reports only
        # whether it is set), so the value must be a managed secret reference on
        # the platform. NOTE: this is the DeviceCheck key, NOT the APNs
        # authentication key -- Apple issues them separately, and the APNs key is
        # uploaded in the Firebase console, not declared here.
        private_key = string

        # How long an App Check token exchanged from a DeviceCheck token stays
        # valid, as a duration in seconds with an "s" suffix (e.g. "3600s").
        # Google accepts 30 minutes to 7 days inclusive and assumes 1 hour when
        # unset.
        token_ttl = optional(string, "")
      }))

      # Debug tokens: pre-shared secrets a development build, a simulator, or a
      # CI lane presents in place of a device attestation. Each token is a
      # SECRET -- anyone holding it passes App Check as this app -- so it is a
      # managed secret reference on the platform, never plaintext. Revoke by
      # removing the entry.
      debug_tokens = optional(list(object({
        # A name for the token as the Firebase console shows it -- the
        # developer's machine, the simulator, or the CI lane it is for. Unique
        # within the app; updatable.
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
    #                this path skips that window. The bundle id becomes
    #                registrable again at once.
    #   "PREVENT" -- destroy FAILS; use for a shipped app whose users hold
    #                its GoogleService-Info.plist
    #   "ABANDON" -- the app is left registered and removed from management
    # The App Attest and DeviceCheck configurations are per-app singletons
    # Google never deletes: they are removed from management on destroy and
    # disappear only with the app itself.
    deletion_policy = optional(string, "")
  })
}
