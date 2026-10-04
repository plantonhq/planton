# Response-Content Health Check

This preset creates a web test that not only checks for a 200 but also
asserts the response body contains an expected healthy payload (for example
`"status":"ok"`). A endpoint can return 200 while its body reports a
degraded state; matching on content catches that. Set `passIfTextFound` to
false instead to fail when an error string appears.

## When to Use

- Health/status endpoints that return 200 even when a dependency is
  unhealthy, with the real state in the body
- APIs where "up" means a specific payload, not just a reachable socket

## Key Configuration Choices

- **`content.contentMatch`** -- the exact string that must appear in the
  body for the test to pass
- **`content.passIfTextFound: true`** -- finding the string is success; set
  false to fail when a known error string appears
- **`content.ignoreCase`** -- match case-insensitively when the payload
  casing is not guaranteed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<resource-group-name>` | The resource group to create the test in | The resource group's `status.outputs.resource_group_name` |
| `<application-insights>` | The AzureApplicationInsights component storing results | Your component's Planton resource name |
| `<https://your-api/status>` | The status endpoint to check | Your API's status URL |
