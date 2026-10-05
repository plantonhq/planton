# CloudflareZeroTrustAccessServiceToken guide

The judgment this guide protects you from: the client secret is shown twice in the token's life -- at create, and at each rotation -- and never again. If you did not capture it, the token still exists and policies still match its ID, but no client can authenticate until you rotate.

## Capture the secret in the same change that creates the token

`client_secret` is a computed, sensitive output. Cloudflare returns it on create and when `client_secret_version` changes; Read always overwrites the API response with the value already in state. Import lands with an empty secret. The operational contract is: write `status.outputs.client_id` and `status.outputs.client_secret` to your secret store before the apply is considered done. Logs must never print the secret.

## Rotation is a pair, not a version bump

`client_secret_version` and `previous_client_secret_expires_at` are both-or-neither. Setting one without the other is rejected here (the provider's AlsoRequires pair). Leaving both unset is the normal non-rotating state -- Cloudflare treats the initial secret as version 1 (the field defaults to 1 when omitted).

To rotate:

1. Increment `client_secret_version` (1 → 2 on the first rotation).
2. Set `previous_client_secret_expires_at` to when the OLD secret should die. Extend it into the future so services can migrate; set it in the past to kill a compromised secret immediately.
3. Apply. Capture the **new** `client_secret` output. The `service_token_id` and `client_id` stay the same, so Access policy `service_token` rules keep matching.

Do not destroy-and-recreate to rotate. That mints a new token ID and every policy that listed the old ID stops matching.

## Duration vs rotation

`duration` is how long the token itself is valid (`8760h` default, or `forever`). Rotation is how you replace the secret inside that lifetime. A `forever` token still needs rotation when the secret leaks.

## API-token auth on the rotation path

The provider's own rotation acceptance tests unset `CLOUDFLARE_API_TOKEN` -- "the Access service does not yet support API tokens" for that path. Measured live 2026-08-26 with an account-owned token (`cfat_`): create, rotation-at-create, destroy, and import all worked with no 403 -- the upstream caveat did not reproduce. If a 403 ever appears on a rotation apply, record the defect rather than switching the harness to a legacy key.

## Adopting an existing token (import)

Import (`accounts/{account_id}/{service_token_id}`) restores the token's identity, name, and duration -- but never `client_secret`: Cloudflare returns it only at create and rotation, so an adopted token's secret is unrecoverable by design. If you need a usable credential after adoption, rotate (increment `client_secret_version` with a `previous_client_secret_expires_at`) and capture the fresh secret from the output. The first post-import apply re-asserts `client_secret_version` (and the expiry, if set) from configuration -- a no-op write, since rotation only triggers when the version increases past the token's real one (measured live 2026-08-26).

## Pairs well with

- [CloudflareZeroTrustAccessPolicy](../cloudflarezerotrustaccesspolicy/README.md) -- a `service_token` include rule that lists this token's ID.
- [CloudflareZeroTrustAccessApplication](../cloudflarezerotrustaccessapplication/README.md) -- the app the machine client calls.
- [CloudflareZeroTrustAccessIdentityProvider](../cloudflarezerotrustaccessidentityprovider/README.md) -- human sign-in, the other door.
