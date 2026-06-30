# Preset: Advanced cache key and Cache Reserve

Take full control of the cache key with the `http_request_cache_settings` phase:
include only the query parameters, cookies, headers, and user attributes that should
vary the cached object, and persist eligible objects to Cache Reserve.

## When to use

- Cache personalized-but-shareable pages (e.g. per-locale, per-currency, per-device)
  without fragmenting the cache on irrelevant query parameters.
- Keep large, rarely-changing objects warm with Cache Reserve.

## Key choices

- `cache_key.custom_key.query_string.include.list`: cache varies only on these query
  params (everything else is ignored for the key).
- `cache_key.cache_by_device_type` / `custom_key.user.geo`: vary by device class or
  visitor country.
- `cache_key.ignore_query_strings_order`: treat `?a=1&b=2` and `?b=2&a=1` as one key.
- `cache_reserve.eligible` + `minimum_file_size`: persist large objects to Cache Reserve.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-zone-id>` | The zone the cache ruleset applies to |
