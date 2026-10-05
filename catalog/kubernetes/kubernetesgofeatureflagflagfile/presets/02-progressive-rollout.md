# Progressive Rollout

A flag that ramps itself: staff get it at once, and everyone else moves from 0% to 100% between two dates, with each caller keeping its variation as the share grows. No one edits the flag during the ramp.

## When to Use

- Rolling a change out gradually without a person widening the percentage each day.
- Watching error rates while the audience grows on a schedule.
- Giving an internal audience the feature first.

## Key Configuration Choices

- `targeting[0]` gives staff (matched by email domain) the feature immediately.
- `defaultRule.progressiveRollout` ramps from the `initial` variation (`disabled`) to the `end` variation (`enabled`): before the start date everyone gets `disabled`; between the dates the share on `enabled` grows from the initial `percentage` (0) to the end `percentage` (100); after the end date everyone gets `enabled`.
- Buckets come from the targeting key, so a caller does not flip back and forth as the share grows; set `bucketingKey` to bucket on another attribute, such as a company id.
- To stop the ramp, replace the default rule with a fixed variation.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<namespace>` | Namespace of the rendered ConfigMap | Your relay's namespace |
| `<your-domain>` | Email domain of the internal audience | Your organization |
| `2026-11-01T09:00:00Z` | Ramp start, RFC 3339 | Your release plan |
| `2026-11-15T09:00:00Z` | Ramp end, RFC 3339 | Your release plan |

## Related Presets

- **Release Flags** - on for a named list of organizations.
