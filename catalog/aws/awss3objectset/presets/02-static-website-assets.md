# Static Website Assets

This preset uploads a small static-site asset set — an HTML entry page, a fingerprint-friendly stylesheet, and a redirect marker for a moved page — with the cache posture each asset class deserves. Public access comes from the bucket's policy and website-hosting configuration on the referenced `AwsS3Bucket`, not from per-object ACLs (modern bucket ownership disables ACLs).

## When to Use

- Seeding or updating a static website whose bucket is managed as an `AwsS3Bucket` resource
- Serving assets through CloudFront where per-asset Cache-Control drives CDN behavior
- Keeping redirect stubs for moved pages alongside the live content

## Key Configuration Choices

- **Bucket by reference** (`valueFrom`) -- The set follows the bucket resource; no hardcoded names to drift
- **`no-cache` on HTML, immutable year-long cache on assets** -- The standard split: pages revalidate, fingerprinted assets never do
- **`websiteRedirect` marker object** -- A GET of the old key redirects when the bucket has static website hosting enabled

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region where the bucket lives (e.g., `us-west-2`) | Must match the bucket's region |
| `<s3-bucket-resource-name>` | Name of the `AwsS3Bucket` resource providing the bucket | Your infra stack's resource list |
| `<project-name>` | Project tag applied to every object in the set | Your tagging convention |
| `<site-title>` | The HTML title of the entry page | Your site content |
