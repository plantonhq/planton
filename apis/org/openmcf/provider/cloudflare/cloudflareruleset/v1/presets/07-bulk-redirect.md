# Bulk Redirect — Redirect From a List

Apply a large set of URL redirects from a reusable Bulk Redirect list. The list
holds the source → target entries (managed independently as `CloudflareListItem`
resources), and an account-level redirect ruleset consults it on every request.

## When to Use

- Migrating a site and redirecting hundreds or thousands of old URLs to new ones
- Centrally managing redirects that several zones in the account share
- Keeping redirect entries decoupled from the rule that applies them, so the list
  can grow or shrink without touching the ruleset

## How It Composes

This is the bulk-redirect half of a two-resource pattern:

1. A `CloudflareList` of kind `redirect` (the Bulk Redirect list) plus its
   `CloudflareListItem` entries.
2. This ruleset, whose `from_list.name` references that list. Cloudflare resolves
   a Bulk Redirect list **by name**, so the reference points at the list's `name`
   output (`status.outputs.name`) — wire it with `valueFrom` and the platform
   resolves it to the live list name before deployment.

## Key Configuration Choices

- **`ruleset_kind: root` + `account_id`** — Bulk Redirects are account-scoped, so
  the ruleset lives at the account level (not on a single zone).
- **`phase: http_request_redirect`** — the Bulk Redirect phase (distinct from
  `http_request_dynamic_redirect`, which uses inline `from_value` targets).
- **`from_list.name`** — the list to match against. Use a literal list name or a
  `valueFrom` reference to a `CloudflareList`.
- **`from_list.key`** — the expression producing the lookup key per request;
  `http.request.full_uri` matches each request's full URL against the list's
  source URLs.
- **`expression`** — the trigger that decides which requests are evaluated;
  `http.request.full_uri ne ""` evaluates every request, leaving the match
  decision to the list lookup.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-account-id>` | Account ID that owns the ruleset and list | Cloudflare dashboard > account home |
| `my-redirects` | Name of the `CloudflareList` holding the redirect entries | Your list resource |

## Related Presets

- **CloudflareList / CloudflareListItem** — define the redirect list and entries
  this rule consumes.
- **01-origin-rule** — route traffic between origins on a zone.
