# HTTP Source

flagd polling a flag definition document over HTTPS, with the Authorization header held as a managed secret. Use it when flags are published by another system -- a flag management service, a build pipeline, a static site -- rather than declared in Planton.

## When to Use

- Flags are published as a JSON or YAML document at a URL
- The document is protected by a bearer token or basic authentication
- A 30-second propagation delay is acceptable

## Key Configuration Choices

- `sources[0].http.authHeader` is the complete header value (for example `Bearer <token>`); it is sensitive and renders only into the module-owned `<name>-sources` Secret
- `sources[0].http.intervalSeconds: 30` sets how often the URL is polled (flagd's default is 5)
- `sources[0].http.intervalSeed` offsets the poll schedule so this flagd does not poll in lockstep with other flagd deployments reading the same document (every replica of one deployment shares the seed)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<namespace>` | Namespace for flagd | Your cluster's namespace plan |
| `https://flags.example.com/flags.json` | URL of the flag definition document | The system that publishes your flags |
| `<flag-document-auth-header>` | Managed secret holding the full Authorization header value | Your organization's secret store |

## Related Presets

- **Flag File Source** -- flags declared in Planton as a KubernetesFlagdFlagFile
