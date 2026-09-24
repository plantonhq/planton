# An App Platform image from DOCR declares its registry as a relationship

## What changed

- **`DigitalOceanAppImageSource.registry` and `registry_type` teach how the dependency on a Planton-managed registry is declared.** With `registry_type: docr` the image is pulled from the account's own DigitalOcean Container Registry; App Platform resolves it itself, the API requires `registry` to be empty, and so no spec field names the registry. A manifest that wants the registry to deploy first and the dependency to be drawn declares it under `metadata.relationships` with type `uses` -- the mechanism the catalog has for exactly the case where one resource consumes another that no field carries. The App guide gains the same paragraph.
- Nothing about the field's type or its rules changed.

## Why

The gap this closes was real -- no line reached a registry from the app that pulls from it -- and the obvious fix was wrong. Typing `registry` as a reference would have shipped a field whose value the module has to discard for DOCR, because DigitalOcean never receives a registry name for its own registry; and the same field would have carried a hostname for Docker Hub and GHCR and a registry name for DOCR. A value the provider never receives is not configuration. A dependency is what it is, and the catalog already has the word for it.

## How to check

```bash
grep -n 'metadata.relationships' catalog/digitalocean/app_spec.proto catalog/digitalocean/digitaloceanapp/GUIDE.md
go test ./catalog/digitalocean/digitaloceanapp/v1alpha1/    # the docr_registry_empty rule stands
```
