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
  description = "AwsCertManagerCert specification"
  type = object({
    # The AWS region the certificate is created in. Regional services
    # (ALB, API Gateway, OpenSearch, ...) require the certificate in
    # their own region. CloudFront is the notable exception: it only
    # accepts certificates from "us-east-1", regardless of where the
    # distribution's origins live.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # The main domain the certificate covers -- an apex ("example.com"),
    # a subdomain ("app.example.com"), or a wildcard ("*.example.com",
    # which covers one label level only). Setting this selects the
    # requested (Amazon-issued) mode, or the private mode when
    # certificate_authority_arn is also set. Leave empty only for
    # imported certificates (ACM derives their domains from the
    # certificate body).
    primary_domain_name = optional(string, "")

    # Additional domains (Subject Alternative Names) the certificate
    # covers, each validated independently. A common pairing is the apex
    # plus its wildcard ("example.com" + "*.example.com") so one
    # certificate serves both the bare domain and every subdomain. Do
    # not repeat primary_domain_name here -- ACM already includes it.
    alternate_domain_names = optional(list(string), [])

    # How ACM verifies domain ownership for requested certificates:
    # "DNS" (recommended -- a CNAME record per domain, kept in place so
    # renewals are fully automatic), "EMAIL" (approval mail to the
    # domain's WHOIS/admin addresses; renewal requires re-approval every
    # time, so prefer DNS), or "HTTP" (serve a validation token from the
    # domain -- for domains whose DNS you cannot touch; renewals repeat
    # the challenge). Empty keeps DNS. Not applicable to imported or
    # private certificates.
    validation_method = optional(string, "")

    # Per-domain overrides for where the validation request is sent.
    # The classic use is EMAIL-validating a subdomain at its parent:
    # validating "app.example.com" by mailing the owners of
    # "example.com". Rarely needed with DNS validation.
    validation_options = optional(list(object({
      # The certificate domain this override applies to (the primary
      # domain or one of the alternate domains).
      domain_name = string

      # The domain the validation request is sent to -- must be the
      # domain itself or one of its ancestors. Example: validate
      # "app.example.com" via "example.com".
      validation_domain = string
    })), [])

    # The key algorithm for the certificate's key pair, create-time
    # immutable: "RSA_2048" (the default -- universally compatible),
    # "RSA_3072", "RSA_4096", "EC_prime256v1" (smaller/faster TLS
    # handshakes; check client compatibility), "EC_secp384r1", or
    # "EC_secp521r1". Empty keeps the ACM default (RSA_2048). Not
    # applicable to imported certificates (the algorithm is baked into
    # the imported key material).
    key_algorithm = optional(string, "")

    # The Route53 public hosted zone where DNS validation records are
    # created automatically. When set (DNS validation only), the module
    # creates the validation CNAMEs in this zone and -- unless
    # wait_for_validation is false -- waits for the certificate to be
    # ISSUED before finishing. When unset, the certificate is created in
    # PENDING_VALIDATION and the required records are exported as the
    # domain_validation_records output for you to create in your
    # external DNS; the deployment does not wait. The zone must be
    # authoritative for every domain on the certificate.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    route53_hosted_zone_id = optional(string, "")

    # Whether the deployment waits for the certificate to reach ISSUED
    # after creating the validation records (DNS validation with
    # route53_hosted_zone_id only -- issuance typically lands within a
    # few minutes). Set false to create the certificate and records
    # without blocking; downstream resources that require an ISSUED
    # certificate (CloudFront, listeners) will then fail until issuance
    # completes, so keep the default unless you are staging DNS ahead of
    # time. Ignored when the module does not manage the validation
    # records.
    wait_for_validation = optional(bool)

    # Certificate options for Amazon-issued (requested and private)
    # certificates.
    options = optional(object({
      # Whether the certificate is recorded in public Certificate
      # Transparency logs: "ENABLED" (the ACM default -- browsers
      # increasingly require CT-logged certificates, so keep it unless
      # you must hide internal hostnames) or "DISABLED". Empty keeps
      # ENABLED.
      certificate_transparency_logging_preference = optional(string, "")

      # Whether the certificate's private key may be exported ("ENABLED"
      # or "DISABLED", the ACM default). Exportable public certificates
      # let you run the same certificate on non-AWS infrastructure, and
      # incur an additional AWS charge per certificate. Empty keeps
      # DISABLED.
      export = optional(string, "")
    }))

    # Bring-your-own certificate material issued by an external CA.
    # Setting this block selects the imported mode: ACM stores and
    # distributes the certificate but never renews it -- re-import new
    # material before expiry (updates re-import in place, keeping the
    # same ARN so consumers are undisturbed).
    imported = optional(object({
      # The PEM-encoded certificate. Public material, not a secret.
      certificate_body = string

      # The PEM-encoded, unencrypted private key matching the certificate.
      private_key = string

      # The PEM-encoded intermediate/root chain, if the issuing CA is not
      # already trusted by AWS. Optional for certificates issued by
      # well-known public CAs. Public material, not a secret.
      certificate_chain = optional(string, "")
    }))

    # The AWS Private Certificate Authority (ACM-PCA) that issues this
    # certificate. Setting this together with primary_domain_name
    # selects the private mode: the CA issues the certificate directly,
    # no public validation happens, and validation_method must stay
    # unset. Private certificates are for internal TLS (service meshes,
    # internal ALBs) where clients trust your private root. Can
    # reference an AwsPrivateCa resource or pass a literal
    # certificate-authority ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    certificate_authority_arn = optional(string, "")

    # How long before expiry ACM starts the managed renewal of this
    # PRIVATE certificate -- either an RFC 3339 duration ("P90D",
    # "P3M") or a Go-style duration ("2160h"). Durations under 60 days
    # have no effect (AWS's floor). Private-CA certificates only:
    # publicly validated certificates renew on ACM's own schedule (kept
    # automatic by leaving the DNS validation records in place), and
    # imported certificates never renew.
    early_renewal_duration = optional(string, "")
  })
}
