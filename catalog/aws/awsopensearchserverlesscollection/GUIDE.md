# AwsOpenSearchServerlessCollection — Kind Guide

The authored wisdom layer for this kind: internal conventions, judgment
calls, and operational judgment earned while building it. The reference for
fields is `v1alpha1/reference.md`; this file explains the decisions the
schema alone cannot.

## Design decisions

- **Collection-scoped policies, typed.** The provider models the four
  policy documents as opaque JSON strings; this kind models them as
  typed spec fields (encryption key choice, network posture, data-access
  rules, retention rules) and the modules render the JSON — scoped to
  exactly this collection (`collection/<name>`, `index/<name>/<pattern>`).
  The account-wide pattern-matching form those policies also support is a
  different tool (one policy governing many collections) and deliberately
  outside this kind's contract; it belongs to future standalone policy
  kinds if demand appears.
- **The encryption policy is ALWAYS rendered** — AWS rejects
  CreateCollection without a matching encryption policy, so an omitted
  `encryption` block means the AWS-owned-key document, never no policy. The
  collection depends on it explicitly (create after, destroy before).
- **The collection's inline `encryption_config` argument is deliberately
  not sent.** It is the same key choice's collection-group-era twin at the
  CreateCollection API; the security-policy path is the universally
  supported arm at the pin. Recorded as a one-arg-two-homes exclusion in
  the parity manifest. Live probe (2026-08-13): CreateCollection WITH
  inline `encryptionConfig` (AWS-owned key) and NO encryption policy in
  the account was accepted and reached ACTIVE, and AWS materialized no
  policy behind it — the inline argument genuinely stands alone.
  Simplifying the module to the inline arm is a recorded follow-up, kept
  deliberate rather than automatic: the policy path remains the one that
  also covers pre-collection-group accounts and keeps the key choice
  visible/auditable as a policy object, and switching arms is a breaking
  re-create for existing collections.
- **An omitted `network` block renders the PUBLIC posture** (the AWS
  console's easy-create default). This is reachability only — every
  request must be SigV4-signed and authorized by a data-access rule, so
  the default is usable-by-default without being open-by-default.
- **No data-access rules means a data-proof collection** — deliberately
  legal (matching AWS, where policies always trail the collection) and
  loudly documented on the spec field: IAM identity permissions alone
  grant NOTHING in OpenSearch Serverless.

## Operational judgment

- **Policy names share the collection's name across all four objects** —
  types are separate namespaces at AWS, so `encryption`/`network` policies
  and the `data`/`retention` policies can all carry it. Everything the
  module owns is discoverable by one name.
- **The OCU floor bills from collection ACTIVE to delete** — 0.5+0.5 OCU
  with standby DISABLED (dev), 2+2 with ENABLED (production default).
  Standby replicas are ForceNew; a dev collection cannot be upgraded in
  place to the HA posture.
- **Collection-group membership requires matching standby settings** — AWS
  rejects a create where the collection's standby_replicas differs from
  the group's.
- **VPC endpoints in `network.vpcEndpointIds` are OpenSearch Serverless's
  OWN endpoint objects** (`aws_opensearchserverless_vpc_endpoint`), not
  ordinary Interface Endpoints — a vpce- id from `aws_vpc_endpoint` will
  not authorize.

## Coverage decisions

- `security_config` (SAML / IAM Identity Center / IAM federation for
  Dashboards sign-in) is account-level identity-provider surface shared by
  many collections — re-judged to a future
  AwsOpenSearchServerlessSecurityConfig kind (recorded in the dispositions
  ledger this session).
- `vpc_endpoint` is the consumer-side network object referenced by many
  collections' network policies — re-judged to a future
  AwsOpenSearchServerlessVpcEndpoint kind (the managed-OpenSearch
  vpc_endpoint precedent).
- `collection_group` is a shared capacity container for many collections —
  re-judged to a future AwsOpenSearchServerlessCollectionGroup kind; this
  spec models the membership side (`collectionGroupName`).
- Live E2E defers: customer-managed KMS (offline-proven KmsARN arm),
  VPC-endpoint network access and collection groups (need the future
  kinds as fixtures), vector acceleration (GPU capacity cost).
