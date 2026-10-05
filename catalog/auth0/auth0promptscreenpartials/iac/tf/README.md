# Auth0PromptScreenPartials — Terraform Module

Terraform/OpenTofu module that manages the HTML fragments one Universal Login prompt inserts on its screens.

## What It Creates

- `auth0_prompt_screen_partials` — every partial of `prompt_type`: one `screen_partials` block per screen, declared in screen-name order (the order the provider reads them back in), each with its `insertion_points`. An empty insertion point is null and never sent. Auth0 replaces the prompt's whole set on every write, so a screen or insertion point the spec leaves out renders nothing. Destroy writes an empty set, removing every partial of the prompt.

The single-screen `auth0_prompt_screen_partial` is never declared: this module owns the prompt's whole set, and Auth0 warns against mixing the two on one prompt.

## Prerequisites

- [Terraform](https://www.terraform.io/downloads) >= 1.0 or [OpenTofu](https://opentofu.org/)
- Auth0 credentials, supplied to the provider via the `AUTH0_DOMAIN`, `AUTH0_CLIENT_ID`, and `AUTH0_CLIENT_SECRET` environment variables. The application needs `read:prompts` and `update:prompts`.

## Inputs

| Name | Description |
|---|---|
| `metadata` | Catalog object metadata (`name`, `org`, `env`, ...) |
| `spec` | `prompt_type`, and `screen_partials` (one entry per screen: `screen_name` and its `insertion_points`) |

## Outputs

| Name | Description |
|---|---|
| `prompt_type` | The prompt whose screens the partials extend |
