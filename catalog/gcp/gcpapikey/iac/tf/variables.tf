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
  description = "GcpApiKey specification"
  type = object({
    # The GCP project the key belongs to. Can be a literal project ID or a
    # reference to a GcpProject resource. If omitted, the provider's default
    # project is used. Immutable: a key cannot move between projects.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The key's resource id -- the last segment of
    # projects/{project}/locations/global/keys/{key_id}. Unique within the
    # project; lowercase letters, digits, and hyphens, starting with a letter,
    # 1-63 characters (RFC 1034 label). Immutable: changing it destroys and
    # recreates the key, which changes the key string every client holds.
    #
    # A deleted key's id stays reserved for 30 days (Google keeps deleted keys
    # recoverable via undelete); a fresh key cannot reuse the id in that
    # window, so an ephemeral key needs a fresh id per lifetime.
    key_id = string

    # Human-readable name shown in the Cloud console's Credentials page.
    # Freely updatable.
    display_name = optional(string, "")

    # Bind the key to a service account, making it a service-account-bound
    # key: requests carrying it are authenticated AS that service account
    # (for APIs that accept API-key authentication in place of OAuth). Can be
    # a literal email or a reference to a GcpServiceAccount resource.
    # Immutable: binding cannot be added, changed, or removed after creation.
    # Leave empty for an ordinary client-identifying key -- the Firebase case.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_account_email = optional(string, "")

    # What may present the key and what it may call. Omit for an
    # unrestricted key (accepted by Google, discouraged: any caller anywhere
    # can spend your quota). Every arm is freely updatable in place -- the
    # key string does not change when restrictions do.
    restrictions = optional(object({
      # Android apps allowed to present the key, each identified by package
      # name AND the SHA-1 of its signing certificate. The pairing is what
      # makes the restriction real: a package name alone is trivially spoofed.
      # For a Firebase Android app, list the app's package name with every
      # signing certificate it ships under (debug, release, Play App Signing).
      android_key_restrictions = optional(object({
        # At least one application; each pairs a package name with a signing
        # certificate SHA-1 fingerprint.
        allowed_applications = list(object({
          # The app's package name, e.g. ai.planton.mobile.
          package_name = string

          # The SHA-1 fingerprint of the certificate the app is signed with, as 40
          # hex characters with or without colon separators
          # (DA:39:A3:EE:... or DA39A3EE...). Get it with
          # `keytool -list -v -keystore <keystore>` or from the Play Console's App
          # signing page. Google stores and returns lowercase hex without colons;
          # both modules send that form, so declare it in whichever shape you have.
          sha1_fingerprint = string
        }))
      }))

      # iOS apps allowed to present the key, by bundle id (the same bundle id a
      # GcpFirebaseAppleApp registers).
      ios_key_restrictions = optional(object({
        # At least one bundle id, e.g. ai.planton.mobile.
        allowed_bundle_ids = list(string)
      }))

      # Websites allowed to present the key, by HTTP referrer pattern (the key
      # a GcpFirebaseWebApp's firebaseConfig carries).
      browser_key_restrictions = optional(object({
        # At least one referrer pattern. Google's referrer grammar: a URL with
        # optional wildcards, e.g. `https://app.example.com/*`,
        # `*.example.com/*`. Requests with no Referer header (native apps,
        # curl) are rejected by a browser-restricted key.
        allowed_referrers = list(string)
      }))

      # Servers allowed to present the key, by caller IP address or CIDR.
      server_key_restrictions = optional(object({
        # At least one caller address: an IPv4/IPv6 address or CIDR block, e.g.
        # 203.0.113.7 or 2001:db8::/32.
        allowed_ips = list(string)
      }))

      # The Google APIs (and optionally methods) the key may call. Empty means
      # every API enabled on the project -- restrict to the APIs the client
      # actually uses (for Firebase Cloud Messaging on a mobile client:
      # `firebaseinstallations.googleapis.com`, `fcmregistrations.googleapis.com`,
      # plus whatever else the app's SDKs call).
      api_targets = optional(list(object({
        # The service's canonical name, e.g. translate.googleapis.com,
        # fcmregistrations.googleapis.com.
        service = string

        # Methods the key may call on the service. Empty means every method. A
        # trailing wildcard is allowed, e.g. `google.cloud.translate.v2.*` or
        # `TranslateText`.
        methods = optional(list(string), [])
      })), [])
    }))

    # What destroying this resource does to the key in GCP:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the key is deleted; Google keeps it recoverable for 30
    #                days (console/undelete), during which the key string
    #                stops working and the key_id stays reserved
    #   "PREVENT" -- destroy FAILS; use for a key baked into a shipped
    #                mobile binary you cannot rotate on demand
    #   "ABANDON" -- the key is removed from management but stays live in
    #                GCP, still presentable by every client that holds it
    deletion_policy = optional(string, "")
  })
}
