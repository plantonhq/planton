---
title: Branding — the Installation's Own Logo in Place of Planton's Mark
description: How a self-hosted Planton shows the company's own logo everywhere its people look (the console's corner and tab, its loading, setup, invitation and device sign-in screens, a desktop window connected to it, and the sign-in page), who may change it, the version for the dark theme, and the `planton branding` verbs. Read when a person asks to put their logo on Planton, why the sign-in page still shows Planton's mark, why the logo is hard to see on the dark theme, or how an organization on planton.ai sets its own.
---

# Branding

Read this when someone says "put our logo on Planton", "the sign-in page still shows the P", or "our logo disappears in dark mode". The people-facing steps are the console's Settings → Branding page; this file is the agent's version, with the reasons.

## The doctrine in one sentence

A self-hosted installation carries one logo, held by the product (not by the PlantonPlatform declaration), uploaded from the console or the CLI, and every surface of that installation draws it in place of Planton's mark -- the product's name stays Planton.

## Where it shows

The console's top-left corner and browser tab; the loading, setup, provisioning and device sign-in screens; invitations; a desktop window connected to the installation, its sign-in screens included; and the identity server's sign-in page. Organizations on the installation share the one logo. With no logo set, every surface shows Planton's mark exactly as before.

Planton's mark stays where it means "Planton did this" (platform-performed actions, Planton-managed connections, the runners mark, the diagram export watermark), and emails stay text-only by design.

## Who changes it, and how

Changing it needs **installation administration** access (Platform Owner). Reading it needs nothing: the logo is display data every screen draws before anyone signs in, so `show` by someone without that access prints the logo without who set it.

```
planton branding show                                  # what is shown, the served addresses, who set it and when
planton branding set-logo logo.png                     # the logo
planton branding set-logo logo.png --dark-theme light.png   # the logo and its version for the dark theme, in one change
planton branding set-logo --dark-theme light.png       # only the version for the dark theme
planton branding clear-logo                            # back to Planton's mark (removes both versions)
planton branding clear-logo --dark-theme               # only the dark-theme version; the logo shows on both themes
```

The CLI acts on the installation it is connected to (`planton instance show` reads `deployment_kind: self_hosted`) and says so in every answer; `--org` is refused there, because organizations share the installation's logo. Setting both versions in one command is one change: one activity record, never a half-applied logo. In the console, the same lives at **Settings → Branding**, shown to administrators; everyone else sees who manages it.

`set-logo` and `clear-logo` are mutations: confirm with the person first, as with any change.

## The picture

PNG, JPEG or WebP at any size; the CLI converts it to a PNG at most 512 pixels on its longest side (and at least 64) before uploading. The console also takes SVG and converts it in the browser; the CLI refuses SVG and asks for a PNG export. The server accepts only PNG it can check before decoding, and keeps its own re-encoded copy -- never the uploaded bytes, never an outside address. A square mark reads best in the corner.

## The dark theme

The console defaults to light, and anyone may switch to dark, where a dark logo vanishes. When the logo would be hard to see on either theme, `set-logo` and `show` say so and print the exact `--dark-theme` command; the console's editor goes further and offers a fix it can make (a recolored version of a one-color mark, a removed background). The version for the dark theme is drawn in the logo's place whenever the theme is dark.

## The sign-in page

The identity server's sign-in page is light by design (the console's default mode), so it draws the **logo itself, else Planton's mark** -- never the version for the dark theme, which would vanish there. It reads them from the installation itself (the control plane serves `/storage/v1/branding/sign-in.png` on the same front door), so a logo change reaches it within a minute -- no operator change, no identity-server restart, nothing fetched from a third party. A sign-in page still showing Planton's mark after a change means either no logo is set (`planton branding show`), or the operator predates the release that serves it (`references/self-hosted.upgrading.md`).

## On planton.ai

Each organization carries its own logo inside its workspace, set by its Owner or Infrastructure Admin from **Settings → General → Logo**, or `planton branding set-logo logo.png` with the organization as the current context (or `--org <slug>`). The hosted sign-in page stays Planton's: no organization is known before sign-in.

## What not to do

- Never put a logo into the PlantonPlatform declaration or the theme; there is no field for it, by design.
- Never offer an image address in place of an upload; the product holds the bytes so the logo works air-gapped and never tells a third party who opened the console.
