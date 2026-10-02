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
  description = "AwsKmsKey specification"
  type = object({
    # The AWS region the key lives in. Ciphertext is regional: data
    # encrypted under this key is decrypted in this region (use
    # multi_region + replica keys for cross-region decryption with the
    # same key material).
    # Example: "us-west-2", "eu-west-1".
    region = string

    # Free-form description shown in the AWS Console. Up to 8192
    # characters.
    description = optional(string, "")

    # The cryptographic configuration, create-time immutable:
    # "SYMMETRIC_DEFAULT" (AES-256-GCM -- the default and what AWS
    # service integrations require), "RSA_2048"/"RSA_3072"/"RSA_4096"
    # (asymmetric encrypt/decrypt or sign/verify),
    # "ECC_NIST_P256"/"ECC_NIST_P384"/"ECC_NIST_P521" (asymmetric
    # sign/verify or ECDH key agreement),
    # "ECC_NIST_EDWARDS25519" (Ed25519 sign/verify only),
    # "ECC_SECG_P256K1" (sign/verify only -- the blockchain curve),
    # "ML_DSA_44"/"ML_DSA_65"/"ML_DSA_87" (post-quantum ML-DSA
    # sign/verify only), "HMAC_224"/"HMAC_256"/"HMAC_384"/"HMAC_512"
    # (MAC generation), or "SM2" (China regions only). Empty keeps the
    # AWS default (SYMMETRIC_DEFAULT).
    key_spec = optional(string, "")

    # What the key is used for, create-time immutable:
    # "ENCRYPT_DECRYPT" (the default -- required for SYMMETRIC_DEFAULT
    # and the only usage AWS service integrations support),
    # "SIGN_VERIFY" (asymmetric RSA/ECC/ML-DSA/SM2 signing keys),
    # "KEY_AGREEMENT" (ECDH shared-secret derivation -- NIST ECC curves
    # and SM2 only), or "GENERATE_VERIFY_MAC" (HMAC keys). Empty keeps
    # the AWS default (ENCRYPT_DECRYPT). The cross-field rules below
    # mirror AWS's own key-spec / key-usage compatibility matrix.
    key_usage = optional(string, "")

    # The key policy as a JSON document -- the resource-based policy
    # that is the root of access control on the key (IAM policies only
    # work when the key policy delegates to them). Empty keeps the AWS
    # default policy, which grants the account's root user full access
    # and enables IAM-policy delegation -- the right choice for most
    # keys. Set a custom policy for cross-account grants or to restrict
    # administration; keep the account root (or your admin role) as an
    # administrator so the key cannot become unmanageable.
    policy = optional(string, "")

    # Skip AWS's check that the key policy leaves the calling principal
    # able to manage the key. Setting this with a policy that locks out
    # every administrator makes the key PERMANENTLY unmanageable (only
    # AWS Support can recover it) -- leave false unless you are
    # deliberately constructing a lockout and understand the blast
    # radius.
    bypass_policy_lockout_safety_check = optional(bool, false)

    # Create the key in (or flip a live key to) the disabled state --
    # every cryptographic operation under it fails until re-enabled. An
    # operational pause switch, gentler than deletion. False (the
    # default) keeps the key enabled.
    disabled = optional(bool, false)

    # Rotate the key material automatically. AWS supports automatic
    # rotation only for SYMMETRIC_DEFAULT keys with KMS-generated
    # material; rotation is transparent to callers (old material is
    # retained to decrypt old ciphertext). Recommended for long-lived
    # encryption keys. False (the AWS default) never rotates.
    enable_key_rotation = optional(bool, false)

    # How often (in days) the material rotates, 90-2560. 0 keeps the
    # AWS default (365 days). Only meaningful with enable_key_rotation.
    rotation_period_in_days = optional(number, 0)

    # Make this a multi-Region PRIMARY key, create-time immutable.
    # Replica keys in other regions then share its key material --
    # ciphertext encrypted in one region decrypts in another. This kind
    # always creates the primary; replicas are created from it and
    # reference its key_arn.
    multi_region = optional(bool, false)

    # Waiting period (days) between scheduling deletion and the key
    # being destroyed, 7-30 -- the recovery window against accidental
    # deletion (ciphertext under a destroyed key is unrecoverable).
    # 0 keeps the AWS default (30 days).
    deletion_window_days = optional(number, 0)

    # Friendly names for the key, each beginning "alias/" (e.g.
    # "alias/orders-db"). Aliases are how humans and SDK callers
    # address the key without its generated ID; many aliases may point
    # at one key, and each materializes as its own alias resource so
    # list edits add/remove in place. The "alias/aws/" prefix is
    # reserved for AWS-managed keys.
    aliases = optional(list(string), [])

    # The custom key store to create the key in: a CloudHSM key store
    # (backed by your CloudHSM cluster) or an external key store (backed
    # by a key manager outside AWS). Supply a literal key-store id
    # (cks-...); custom key stores are account-level infrastructure the
    # catalog does not provision. Create-time immutable. Custom key
    # store keys must be symmetric encryption keys, never rotate
    # automatically, and cannot be multi-Region (AWS's contract,
    # enforced by the rules below). Empty (the default) keeps the key
    # material in standard AWS KMS.
    custom_key_store_id = optional(string, "")

    # The id of an existing key in the external key manager, for keys
    # created in an EXTERNAL key store -- KMS forwards cryptographic
    # operations under this key to that external key. Requires
    # custom_key_store_id (pointing at an external key store).
    # Create-time immutable. Up to 128 characters.
    xks_key_id = optional(string, "")

    # KMS grants on this key: scoped, revocable permissions that let a
    # principal use the key for specific operations without editing the
    # key policy -- the mechanism for wiring "this workload role may
    # encrypt/decrypt under this key" as a first-class dependency, and
    # for cross-account key usage. Each entry materializes as its own
    # grant resource; grants are create-time immutable (any change
    # replaces the grant, which is safe -- grants carry no state).
    grants = optional(list(object({
      # Friendly name for the grant, shown by ListGrants next to the
      # generated grant id. Optional. Up to 256 characters: letters,
      # digits, and _:/- .
      name = optional(string, "")

      # The IAM principal the grant permits to use the key, in ARN form: an
      # IAM role (the mainstream pattern -- reference an AwsIamRole's
      # role_arn output), an IAM user, an account root, or a federated/
      # assumed-role principal. Cross-account delegation works by naming a
      # principal in another account. NOT a service principal: AWS's
      # CreateGrant takes those through a separate GranteeServicePrincipal
      # parameter (paired with a mandatory SourceArn constraint) that the
      # Terraform provider does not expose -- a bare service principal here
      # is rejected live with "InvalidArnException: GranteePrincipal does
      # not refer to a valid principal" (live-verified 2026-08-11).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      grantee_principal = string

      # The KMS operations the grant allows, at least one. The domain is
      # AWS's own GrantOperation set; which entries make sense depends on
      # the key's shape (e.g. Sign/Verify need a signing key, GenerateMac
      # an HMAC key) -- AWS validates that pairing at grant creation.
      operations = list(string)

      # The IAM principal (ARN form) allowed to retire the grant when it is
      # no longer needed, in addition to the key administrators. Reference
      # an AwsIamRole's role_arn output or pass a literal role/user ARN.
      # Service principals are not accepted here for the same reason as
      # grantee_principal (AWS's RetiringServicePrincipal parameter is not
      # in the provider surface).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      retiring_principal = optional(string, "")

      # Constrain the grant to requests whose encryption context EQUALS
      # exactly these key-value pairs. Only valid for operations that take
      # an encryption context (symmetric keys). Mutually exclusive with
      # encryption_context_subset.
      encryption_context_equals = optional(map(string), {})

      # Constrain the grant to requests whose encryption context CONTAINS
      # these key-value pairs (a subset match -- the request may carry
      # more). Mutually exclusive with encryption_context_equals.
      encryption_context_subset = optional(map(string), {})

      # How teardown releases the grant: false (the default) REVOKES it --
      # the hard stop that denies all further use immediately; true RETIRES
      # it -- the graceful path AWS recommends when the grant's work is
      # done. Both remove the grant; they differ in intent and in the API
      # permission they exercise (RevokeGrant vs RetireGrant).
      retire_on_delete = optional(bool, false)
    })), [])
  })
}
