# Auth0PromptCustomText — Terraform Module

Terraform/OpenTofu module that manages the words one Universal Login prompt shows in one language.

## What It Creates

- `auth0_prompt_custom_text` — the custom text of `prompt` in `language`. The spec's `screens` render into the provider's `body` as `{"<screen>": {"<key>": "<text>"}}`, keys sorted (`jsonencode`). Auth0 replaces the whole custom text on every write, so a screen or key the spec leaves out shows Auth0's default words. Destroy writes `{}`, returning the prompt to Auth0's defaults in that language.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials, supplied to the provider via the `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` environment variables. The application needs `read:prompts` and `update:prompts`.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Catalog object metadata (`name`, `org`, `env`, ...) |
| `spec` | `prompt`, `language`, and `screens` (screen name to `texts`, a map of text key to words) |

## Outputs

| Name | Description |
|---|---|
| `prompt` | The prompt the words belong to |
| `language` | The language the words are shown in |
| `id` | The custom text's identifier, `<prompt>::<language>` (for example `login::en`), also its import id |
