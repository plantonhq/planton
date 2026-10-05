# AzureSearchService Guide

Judgment and internal conventions for the AI Search kind -- what
the schema alone cannot carry.

## Parity accounting

Modeled from `azurerm_search_service` plus its composed child
`azurerm_search_shared_private_link_service` at the pinned azurerm
v5.0.0, full surface, ZERO parity exceptions (the classic Pulumi
SDK's `search.Service` / `search.SharedPrivateLinkService` carry
every v5 argument -- verified field-by-field, module compile is the
standing verifier). Family renames (`region` -> `location`,
`resource_group` -> `resource_group_name`) plus the child's specRoot
are recorded in `iac/provider-parity.yaml`.

## The SKU update contract is update-time, not manifest-time

The provider's CustomizeDiff allows in-place SKU changes ONLY along
basic -> standard -> standard2 -> standard3 (strictly upward); every
other change -- downgrades, free, storage-optimized, hops -- forces
replacement. A manifest-time rule cannot see the PREVIOUS value, so
this is deliberately NOT CEL; it is recorded on the `sku` field and
here. Composers should treat SKU as semi-immutable: pick the family
right, grow within it.

## Per-SKU caps ARE manifest-time

Every cap the provider enforces in its create/update code is spec
CEL: free <= 1 partition/replica, basic <= 3/3, high-density <= 3
partitions, high-density only on standard3, semantic not on free,
failure mode only with local auth. These fire in seconds instead of
at deploy.

## The name has no format rule -- deliberately

The provider's schema and docs carry no name validation; ARM owns
it. The spec mirrors that honestly (no invented regex). The name IS
globally unique (it forms the endpoint DNS name) -- a name-taken
error at deploy is a genuine global collision, not a ghost (search
services have no soft-delete class).

## The default query key is a single output, not a map

The service creates exactly ONE query key at provisioning, and its
name is EMPTY. A name-keyed map (the ER-port authorization_keys
precedent) would carry the empty string as its only key -- hostile to
the dot-flattening output path -- so the spec exports
`default_query_key` as a single sensitive output instead. Additional
query keys are data-plane objects created per application.

## Credentials in outputs

The admin keys and query key are service-minted with no vault
indirection -- the narrow outputs exception applies (exported,
marked sensitive in both engines, empty in the RBAC-only posture).

## Quota realities

`standard2`, `standard3`, and both storage-optimized SKUs require a
Microsoft quota increase before ARM accepts them (the provider docs'
own note). The canonical manifest uses standard3 for the
high-density seam OFFLINE only; the live scenario stays on
`standard`.

## E2E shape

The smoke scenario is a standard-SKU service with one shared private
link to the fixture storage account (blob) -- the provider's own
SPL-with-standard acceptance pairing. The link sits "Pending"
(nothing approves it in the lane); the verifier asserts ARM state,
never connection health. RBAC-only, high-density, and CMK-enforcement
arms are offline-proven only.
