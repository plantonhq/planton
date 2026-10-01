# App Service Origin

This preset creates an origin pointing at an Azure App Service -- the
most common Front Door backend. The defaults are exactly right for it:
the Host header falls back to the origin hostname (which App Service
routes by), certificate-name checking stays on, and ports stay 80/443.

## When to Use

- App Service, Function App, or Container Apps backends (all
  multi-tenant services that route by Host header)
- Any HTTPS backend whose certificate matches its hostname

## Key Configuration Choices

- **No `originHostHeader`** -- deliberately unset: Azure then sends the
  origin's own `hostName`, which is what multi-tenant backends need to
  route the request; set it only when the backend expects the
  client-facing domain instead (and can serve it)
- **`certificateNameCheckEnabled` stays true** (the default) --
  disabling it accepts ANY certificate from the origin, a
  man-in-the-middle door
- **Priority/weight defaults** (1/500) -- right for a single origin;
  they only matter once siblings join the group

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<origin-group-resource-name>` | The AzureFrontDoorOriginGroup's Planton resource name | Your Front Door composition |
| `originName` (example value) | 2-90 chars -- rename to your convention | Your naming convention |
| `<web-app-resource-name>` | The AzureLinuxWebApp whose default hostname the origin targets | Your App Service resource |
