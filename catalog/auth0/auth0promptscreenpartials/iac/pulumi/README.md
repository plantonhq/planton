# Auth0PromptScreenPartials — Pulumi Module

Pulumi Go module that manages the HTML fragments one Universal Login prompt inserts on its screens.

## What It Creates

- `auth0.PromptScreenPartials` — every partial of `promptType`: one `screenPartials` entry per screen, in screen-name order (the order the provider reads them back in), each with its `insertionPoints`. An empty insertion point is nil and never sent. Auth0 replaces the prompt's whole set on every write, so a screen or insertion point the spec leaves out renders nothing. The resource's delete writes an empty set, removing every partial of the prompt.

The single-screen `auth0.PromptScreenPartial` is never declared: this module owns the prompt's whole set, and Auth0 warns against mixing the two on one prompt.

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/install/)
- [Go 1.21+](https://golang.org/dl/)
- Auth0 credentials (domain, client_id, client_secret) for an application holding `read:prompts` and `update:prompts`.

## Environment Variables

When `provider_config` is not set in the IaC input, the module falls back to environment variables:

| Variable | Description |
|---|---|
| `AUTH0_DOMAIN` | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | M2M application client ID |
| `AUTH0_CLIENT_SECRET` | M2M application client secret |

## Structure

- `module/locals.go` orders the screens and maps each empty insertion point to nil, the twin of `iac/tf/locals.tf`; `module/locals_test.go` pins both rules.
- `module/prompt_screen_partials.go` sets the partials and `module/outputs.go` exports the prompt.
