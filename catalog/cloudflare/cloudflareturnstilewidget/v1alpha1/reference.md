# CloudflareTurnstileWidget

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareTurnstileWidgetSpec provisions a Cloudflare Turnstile widget: a
privacy-preserving CAPTCHA alternative. A widget yields a public site key
(embedded in the page's frontend) and a secret key (used server-side to call
the /siteverify endpoint). The secret is exported as a sensitive output
so a Worker or backend that validates tokens can reference it.

Allowed-value sets (mode, clearance_level, region) are validated with CEL using
the provider's exact strings rather than proto enums: Turnstile's `mode`
includes the hyphenated value "non-interactive", and "managed" is a valid value
for both `mode` and `clearance_level` (proto enum value names are file-scoped
and could not collide), so plain strings keep the spec faithful to the API
without a lossy enum-to-string mapping in the modules.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareTurnstileWidget
metadata:
  name: test-widget
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: login-form
  domains:
    - example.com
  mode: managed
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.domains` | `[]string` | yes |  |  |
| `spec.mode` | `string` | yes |  |  |
| `spec.clearanceLevel` | `string` |  |  |  |
| `spec.botFightMode` | `bool` |  |  |  |
| `spec.ephemeralId` | `bool` |  |  |  |
| `spec.offlabel` | `bool` |  |  |  |
| `spec.region` | `string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns this widget.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.name

`string` · required

A human-readable widget name (not unique). Set something meaningful so the
widget is easy to identify and locate where it is used.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.domains

`[]string` · required

The domains the widget may be served on (e.g. "example.com"). At least one
is required. Use "localhost" for local development.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.mode

`string` · required

Widget mode: "non-interactive", "invisible", or "managed" (Cloudflare
decides the challenge — the recommended default).

- rule: mode must be one of non-interactive, invisible, managed
- rule: {"required":true}

### spec.clearanceLevel

`string`

Clearance level granted when the widget is embedded on a Cloudflare-proxied
site: "no_clearance", "jschallenge", "managed", or "interactive". Leave
empty to use the provider's behavior.

- rule: clearance_level must be empty or one of no_clearance, jschallenge, managed, interactive

### spec.botFightMode

`bool`

Issue computationally-expensive challenges to malicious bots (Enterprise
only).

### spec.ephemeralId

`bool`

Return the Ephemeral ID in /siteverify responses (Enterprise only).

### spec.offlabel

`bool`

Hide all Cloudflare branding on the widget (Enterprise only).

### spec.region

`string`

Region the widget can be used in: "world" (default) or "china". Immutable —
cannot be changed after creation. Leave empty to use the provider default
("world").

- rule: region must be empty or one of world, china

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareTurnstileWidget, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.sitekey` | `string` | The public site key. Embed this in the page's frontend Turnstile widget. |
| `status.outputs.secret` | `string` (sensitive) | The secret key used server-side to validate tokens via /siteverify. Sensitive — exported as a secret so a Worker or backend can reference it. |
| `status.outputs.created_on` | `string` | RFC3339 timestamp of when the widget was created. |
| `status.outputs.modified_on` | `string` | RFC3339 timestamp of when the widget was last modified. |

## See Also

- [Overview](../README.md)
