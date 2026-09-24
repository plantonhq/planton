# The Docs Name the Google Artifact Registry Mirror

**Date**: September 24, 2026
**Type**: Docs
**Components**: Docs (self-hosting), Helm (planton-operator, planton, planton-runner), Skills (planton)

## Summary

Every Planton image and chart release is copied, byte for byte, from ghcr.io to Google Artifact Registry in `asia-south1` at the same path after the host. The self-hosting page, the three chart READMEs, and the upgrading reference now say so, and show the one setting that moves an install: `spec.imageRegistry` on the platform, the operator chart's `image.repository`, and the charts' OCI path. The self-hosting page's install examples declare `v0.0.75`, the oldest release today's operator runs, where they still declared a release it refuses.

## What Changed

- `site/public/docs/self-hosting/index.md`: a "Pull from Google Artifact Registry" section (the two hosts side by side, the two installs from the mirror, the digest comparison, what stays on its upstream); the examples at `v0.0.75`.
- `helm/planton-operator/README.md`, `helm/planton/README.md`, `helm/planton-runner/README.md`: the mirror path for each chart and image.
- `skills/planton/references/self-hosted.upgrading.md`: moving a running install to the mirror, in two declarations.
- `_issues/`: two feature requests for the images the one setting cannot reach yet, Tekton's and the operator's bundled components.

## Verification

- The mirror holds `planton` 0.3.1, `planton-operator` 0.21.0 and 0.22.0, and `planton-runner` 0.7.0; the operator `v0.22.0` and platform `v0.0.75` images carry the same digest on both hosts, read from each registry's `Docker-Content-Digest`.
- The site's internal link gate: every internal link resolves.
