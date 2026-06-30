# Preset: Configuration rules (set_config)

Override zone settings per request with the `http_config_settings` phase. A
`set_config` rule applies settings (SSL mode, security level, Rocket Loader, Polish,
auto-minify, email obfuscation, …) only to requests matching its expression.

## When to use

- Harden a sensitive path (raise `security_level`, force `ssl: strict`).
- Enable performance features (Rocket Loader, Mirage, Polish, auto-minify) on
  marketing/static pages without affecting the API.

## Key choices

- Only set the fields you want to manage — omitted settings are left untouched, so
  each `set_config` rule is additive.
- `ssl`: `off` / `flexible` / `full` / `strict` / `origin_pull`.
- `security_level`: `off` / `essentially_off` / `low` / `medium` / `high` / `under_attack`.
- `polish`: `off` / `lossless` / `lossy` / `webp`.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-zone-id>` | The zone the configuration ruleset applies to |
