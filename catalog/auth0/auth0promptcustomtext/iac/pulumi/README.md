# Auth0PromptCustomText — Pulumi Module

Pulumi Go module that manages the words one Universal Login prompt shows in one language.

## What It Creates

- `auth0.PromptCustomText` — the custom text of `prompt` in `language`. The spec's `screens` render into the provider's `body` as `{"<screen>": {"<key>": "<text>"}}`, keys sorted (`json.Marshal`, the same bytes Terraform's `jsonencode` writes). Auth0 replaces the whole custom text on every write, so a screen or key the spec leaves out shows Auth0's default words. The resource's delete writes `{}`, returning the prompt to Auth0's defaults in that language.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret) for an application holding `read:prompts` and `update:prompts`.

## Environment Variables

When `provider_config` is not set in the stack input, the module falls back to environment variables:

| Variable | Description |
|---|---|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | M2M application client ID |
| `AUTH0_CLIENT_SECRET` | M2M application client secret |

## Structure

- `module/locals.go` renders the screens into the body document, the twin of `iac/tf/locals.tf`; `module/locals_test.go` pins the rendering.
- `module/prompt_custom_text.go` sets the custom text and `module/outputs.go` exports its prompt, language and id.
