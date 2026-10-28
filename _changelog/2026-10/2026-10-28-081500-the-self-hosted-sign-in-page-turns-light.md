# The Self-Hosted Sign-In Page Turns Light

**Date**: October 28, 2026

## Summary

A self-hosted installation's sign-in page now wears the console's light palette, because light is now the console's default mode: a person who never chose a theme meets one theme from sign-in into the product. The page also draws its logo once. It used to draw it twice, at two sizes, one over the other.

## What Changed

- `operator/internal/keycloaklogintheme/theme/login/resources/css/planton.css`: the `--planton-*` tokens take the light values (each commented with the `light-colors.ts` token it copies). keycloak.v2's own logo slot (`--keycloak-logo-*`, 40px tall, centered) draws the installation's logo. The banner no longer draws a second copy as its background. Secondary buttons (a sign-in provider such as Google) get the border token's hairline instead of PatternFly's primary-ink frame, matching the hosted sign-in page.
- `theme_test.go` pins the new token values. `darkMode=false` stays, now keeping the page light whatever the visitor's OS prefers.
- The public design-system page and the Planton skill's branding reference say light is the default, and that the sign-in page draws the logo itself (never its dark-theme version).

## Impact

The platform's control plane serves the logo this page draws (`/storage/v1/branding/sign-in.png`). In the same release it serves the logo itself and Planton's mark in dark ink. The two must land in the same upgrade: either one alone puts a white mark on white or a dark mark on dark.
