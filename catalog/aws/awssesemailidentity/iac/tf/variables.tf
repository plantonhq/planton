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
  description = "AwsSesEmailIdentity specification"
  type = object({
    # The AWS region where the identity is created. SES identities are
    # regional: verify the same domain in each region you send from.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # The identity to verify: a domain ("example.com") for the production
    # shape, or a single email address ("sender@example.com") for
    # mailbox-verified sending. IMMUTABLE: changing it replaces the
    # identity (the new identity re-verifies from scratch).
    email_identity = string

    # The configuration set applied by default to every message sent from
    # this identity -- the delivery/tracking/event-publishing rules defined
    # once and inherited here. Reference an AwsSesConfigurationSet's
    # configuration_set_name output or pass a literal set name. Can be
    # attached, swapped, or removed in place.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    configuration_set = optional(string, "")

    # DKIM signing configuration -- Easy DKIM (AWS-managed keys) or BYODKIM
    # (bring your own key pair). Only DOMAIN identities carry DKIM
    # configuration; leave unset for email-address identities (they inherit
    # the domain's DKIM when the domain is also verified) and to accept
    # Easy DKIM with a 2048-bit key, the AWS default, on domains.
    dkim_signing = optional(object({
      # Easy DKIM: the RSA key length AWS generates for the NEXT signing key
      # -- "RSA_1024_BIT" or "RSA_2048_BIT" (prefer 2048, the AWS default;
      # 1024 exists for DNS providers with 255-character TXT limits).
      # Setting it on a live identity rotates the key. Mutually exclusive
      # with the BYODKIM pair (empty when BYODKIM is used).
      next_signing_key_length = optional(string, "")

      # BYODKIM: the base64-encoded RSA PRIVATE key SES signs with (PKCS #8,
      # headers stripped). A signing secret -- never logged, never exported.
      # Requires domain_signing_selector.
      domain_signing_private_key = optional(string, "")

      # BYODKIM: the DKIM selector under which YOU publish the public key
      # ("<selector>._domainkey.<domain>" TXT record). Requires
      # domain_signing_private_key.
      domain_signing_selector = optional(string, "")
    }))

    # Custom MAIL FROM domain configuration. By default SES uses its own
    # bounce domain (amazonses.com) as the envelope MAIL FROM, which fails
    # strict DMARC alignment on SPF; a custom MAIL FROM subdomain (e.g.
    # "mail.example.com" with its MX + SPF records, composable via
    # AwsRoute53DnsRecord) aligns the envelope with the sending domain.
    mail_from = optional(object({
      # The custom MAIL FROM domain. Must be a subdomain of the identity's
      # domain (e.g. "mail.example.com" for "example.com"), must not be used
      # to receive mail, and needs an MX record pointing at the regional SES
      # feedback endpoint plus an SPF TXT record -- both composable with
      # AwsRoute53DnsRecord.
      mail_from_domain = string

      # What SES does when the MAIL FROM domain's MX record is missing or
      # broken:
      #   USE_DEFAULT_VALUE -- fall back to amazonses.com and keep sending
      #                        (the AWS default; mail flows but loses DMARC
      #                        SPF alignment).
      #   REJECT_MESSAGE    -- fail the send with MailFromDomainNotVerified
      #                        (strict: nothing leaves unaligned).
      behavior_on_mx_failure = optional(string)
    }))

    # Whether bounce and complaint notifications are forwarded by email to
    # the identity's mailbox. Tri-state: leave UNSET to accept AWS's own
    # default (forwarding on) with no managed setting at all; set true or
    # false to pin the position explicitly -- the modules materialize the
    # feedback sub-resource only when a position is taken. Turn it off once
    # event destinations or SNS feedback handle bounces -- forwarding is the
    # fallback channel, not the production one. One retention caveat
    # (live-verified 2026-08-12): SES remembers the LAST-WRITTEN forwarding
    # value per identity name, surviving even DeleteEmailIdentity and
    # re-creation -- so on an identity that was ever explicitly managed,
    # clearing this field stops managing the setting but does NOT restore
    # AWS's default; set true explicitly to turn forwarding back on.
    email_forwarding_enabled = optional(bool)

    # Named authorization policies on this identity -- the cross-account
    # sending grants that let another AWS account or role send mail AS this
    # identity (SendEmail with this identity as the source). Each policy is
    # its own AWS sub-resource keyed by name, materialized per-name by the
    # modules. Names must be unique.
    policies = optional(any, [])
  })
}
