# Public Static Website Bucket

This preset creates a bucket that serves a static website directly over S3's website endpoint. Public access is granted deliberately and precisely: the policy grants read on objects, and only the two public-access guards that would block that policy are relaxed.

## When to Use

- Internal tools, previews, and throwaway sites where CloudFront would be overkill
- Redirect buckets and simple HTTP-only content
- Learning/demo environments

For production websites, prefer a **private** bucket behind CloudFront with Origin Access Control — that adds TLS, caching, and keeps the bucket unexposed. See the `AwsCloudFront` kind.

## Key Configuration Choices

- **Deliberate public access** — `publicAccessBlock` keeps the ACL guards ON (`blockPublicAcls`, `ignorePublicAcls`) and relaxes only the policy guards (`blockPublicPolicy`, `restrictPublicBuckets`), so the ONLY public surface is the explicit policy statement
- **Scoped read-only policy** (`policy`) — grants anonymous `s3:GetObject` on objects only; the bucket itself (listing, configuration) stays private
- **Website hosting** (`website`) — `index.html` for directory requests, `error.html` for 4XX responses; the website endpoint appears in `status.outputs.website_endpoint`
- **CORS for browsers** (`corsRules`) — allows cross-origin GET/HEAD so pages on other origins can load assets from this bucket

## Placeholders to Replace

- `<aws-region>` — the region for the bucket
- `my-public-website-bucket` — rename to your site's bucket name, and update the `Resource` ARN inside the policy to match (`arn:aws:s3:::<your-name>/*`)

## Common Additions

- `website.routingRules` — conditional redirects (e.g., send `docs/` to another host)
- A second redirect bucket (`website.redirectAllRequestsTo`) to point the apex domain at `www`
- An `AwsRoute53Zone` alias record targeting `status.outputs.website_domain`

## Related Presets

- **01-private-encrypted** — the default posture for every non-website bucket
- **03-log-archive-lifecycle** — a log destination with aggressive storage tiering and expiration
