# HTTP API Function

This preset creates a public Scaleway Serverless Function using the Node.js 20 runtime. It auto-scales from zero to 20 instances based on incoming HTTP requests. This is the most common serverless function configuration for lightweight APIs and webhooks.

## When to Use

- Lightweight REST or webhook endpoints
- Event-driven APIs that benefit from scale-to-zero (no cost when idle)
- Quick serverless backends without container image management

## Key Configuration Choices

- **Node.js 20 runtime** (`runtime: node20`) -- the most popular serverless runtime; change to `python311`, `go122`, or `rust165` for other languages
- **Public privacy** (`privacy: public`) -- the function endpoint is accessible from the internet; use `private` for internal functions
- **Handler** (`handler: handler.handle`) -- the entry point function in your code; format is `{file}.{export}`
- **256 MB memory** (`memoryLimitMb: 256`) -- suitable for typical API handlers; increase for compute-intensive functions
- **Scale 0-20** (`maxScale: 20`) -- scales to zero when idle; adjust maximum based on expected concurrency
- **5-minute timeout** (`timeoutSeconds: 300`) -- maximum execution time per request

## Placeholders to Replace

No placeholders -- this preset is ready to deploy. Deploy your function code via Scaleway CLI (`scw function deploy`) or by uploading a zip file.

## Related Presets

- **02-scheduled-job** -- Use instead for background tasks triggered on a CRON schedule rather than HTTP requests
