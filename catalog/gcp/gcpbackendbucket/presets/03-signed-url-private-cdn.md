# Private Content via CDN Signed URLs

A backend bucket serving private media through Cloud CDN, gated by signed URLs: expiring, tamper-proof links minted with the named signing key. The application signs each URL; the edge verifies the signature and serves from cache.

## When to Use

- Paid or per-user media (video, downloads, exports) that should ride the CDN but never be publicly listable
- Any content where "who can fetch this" is decided per-request by the application, not by bucket ACLs

## Remix Notes

- **The key value is a secret**: anyone holding it can mint valid URLs. Supply it as a managed-secret reference, never plaintext in the manifest; it never appears in outputs.
- **Rotation**: add a second key (`signedUrlKeys` holds up to 3), switch the application to sign with it, then remove the old one. Keys are immutable in GCP — this add/re-sign/remove dance is the designed rotation path.
- The application must sign URLs with the SAME key name and value — see GCP's signed-URL signing documentation for the token format.
- `signedUrlCacheMaxAgeSec` trades origin load against revalidation freshness for signed responses; the signature expiry still gates access.
