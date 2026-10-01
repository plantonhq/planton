# GcpBackendBucket Guide

The judgment this guide protects: the backend bucket is deliberately a
separate node from the GCS bucket it serves — the indirection is what
makes static releases boring. Keep it thin, keep the origin swappable.

## The origin swap is the release mechanism

`bucketName` is the one REQUIRED field and it is deliberately MUTABLE:
pointing at a different bucket is an in-place update. That makes
blue/green static releases a one-field edit — upload the new build to a
fresh bucket, swap the origin, roll back by swapping again. Never deploy
by overwriting objects in the live bucket; the cache makes that path
unpredictable.

## CDN caching is a contract with your headers

`CACHE_ALL_STATIC` trusts content types; `USE_ORIGIN_HEADERS` trusts you;
`FORCE_CACHE_ALL` trusts nothing and caches everything — never point it
at a bucket with private or per-user content. When responses look stale
after a release, check `signedUrlCacheMaxAgeSec` and the negative-caching
codes before blaming the swap: cached 404s from the deploy window are the
usual culprit.

## Signed-URL key rotation is add-then-remove

GCP caps keys at 3 per bucket precisely so one can rotate while another
stays live: add the new key, re-sign URLs with it, then remove the old.
Each key is immutable — changing a `keyValue` replaces that key resource,
which is the rotation semantics signed URLs need. The key material is
secret in Pulumi state and sensitive in OpenTofu's (whose state file is
encrypted whole when given a key); it never appears in outputs. Remember the
CDN also needs `roles/storage.objectViewer` on the origin bucket for
signed serving of private objects.

## Cloud CDN and INTERNAL_MANAGED do not mix

`loadBalancingScheme: INTERNAL_MANAGED` serves cross-region internal
ALBs and is incompatible with `enableCdn` — the spec rejects the
combination pre-deploy. The scheme is immutable; choosing it later is a
recreate, though with a stable URL map reference that is a quick swap.

## Teardown discipline

One `deletionPolicy` governs the backend bucket AND every signed-URL key
— they have no independent life. GCP refuses to delete a backend bucket
a URL map still references, so `DELETE` fails safely mid-chain; `PREVENT`
also covers the window after the URL map is gone. `ABANDON` leaves the
CDN origin serving unmanaged. The GCS bucket behind it is never touched —
it belongs to its own kind.
