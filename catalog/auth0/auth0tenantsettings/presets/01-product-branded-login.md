# Product-Branded Login

## Pattern

The tenant introduces itself as your product: its name, its logo, and where to get help.

## What It Does

- Sets the tenant's friendly name, so Universal Login reads "Log in to Example to continue to *application*".
- Shows your logo on the login and consent pages instead of Auth0's.
- Offers your support address and page to people who can't sign in.

## When to Use

- Any tenant real people sign in through: a product's production tenant, or a pre-production tenant prospects or customers see.

## Customization

- Leave out any field you don't want managed; the tenant keeps its current value for it.
- Name each application for the second half of the sentence with the Auth0 Client kind's `name`.

## Placeholders to Replace

| Placeholder | Description |
|---|---|
| `metadata.org` | Your Planton organization |
| `spec.friendlyName` | Your product's name |
| `spec.pictureUrl` | A public HTTPS URL of your logo (about 150 x 150 pixels) |
| `spec.supportEmail` | Where people ask for help |
| `spec.supportUrl` | Your help page |
