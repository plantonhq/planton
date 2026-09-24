# DigitalOcean CDN -- Operational Guide

What experience with this component teaches that the field reference cannot.

## Reference certificates by name -- the UUID is a moving target

Every Let's Encrypt renewal mints a NEW certificate UUID; only the name is stable. This spec's certificate reference carries the name (the same value DigitalOceanCertificate exports as `certificate_id`), which is also all the provider's current schema wants. The provider's deprecated numeric `certificate_id` argument hides a real footgun -- a configuration using only it silently DETACHES the certificate whenever the custom domain changes -- and is deliberately unrepresentable here.

## The origin is create-only; the endpoint hostname is an API

Changing `origin` replaces the CDN endpoint, which changes the `endpoint` hostname -- and every CNAME and hard-coded URL pointing at the old one goes stale. Treat the endpoint hostname like an API surface: front it with your own custom domain early if consumers will hard-code URLs, so a future origin swap only re-points your CNAME.

## The edge serves what the bucket allows

The CDN does not bypass bucket permissions: private objects return errors at the edge exactly as they do at the origin. Public-read objects (or a bucket policy granting read) are what make CDN delivery work. If content 403s at the edge, the fix is on the bucket, not here.

## Cache TTL is a blast-radius decision

The TTL applies endpoint-wide -- there are no per-path rules on DigitalOcean's CDN. A long TTL (the 3600 default and up) is right for fingerprinted assets; content updated in place needs a short TTL and tolerance for staleness windows. An explicit zero is unrepresentable (the provider drops zeros on the wire), which is DigitalOcean's way of saying this is a cache, not a pass-through.

## Deleting the endpoint is a traffic event

Destroy tears down edge delivery immediately: the endpoint hostname stops serving and custom-domain CNAMEs dangle. The origin bucket and its content are untouched -- consumers can fall back to the bucket's own domain if they know it. Re-point DNS before destroying when a custom domain fronts real traffic. The API forgets the endpoint fast -- it answers 404 within two seconds of the delete (measured) -- so a refresh right after a destroy reads a clean absence.

## What is deliberately NOT here

The provider's deprecated `certificate_id` argument (the detach footgun above -- name-based reference only); per-path cache rules, WAF, and rate limiting (no provider surface -- DigitalOcean's CDN is a cache tier, not an edge security layer); and a `created_at` output (a timestamp with no operational consumer).
