# reCAPTCHA Enterprise Attested

A site that proves a real browser on your domain produced each request
before the project's backends answer it, using Google Cloud's reCAPTCHA
Enterprise, with a referrer-restricted key and a localhost debug token.

## What it configures

- The registration with `apiKeyId` referencing a referrer-restricted
  `GcpApiKey`.
- App Check with reCAPTCHA Enterprise: the PUBLIC site key (the same value
  the page embeds; Enterprise has no secret) at Google's default one-hour
  token lifetime.
- One debug token for local development held as a managed secret
  reference.
- `deletionPolicy: PREVENT`.

## Adjust before deploying

- **`siteKey`** -- a score-based reCAPTCHA Enterprise key created in the
  Google Cloud console for the site's domains.
- **The debug token** -- point the secret reference at a UUID4 you
  generated; the page sets it as `self.FIREBASE_APPCHECK_DEBUG_TOKEN`.
- **Cost** -- Enterprise assessments are free to 10,000 a month, then
  billed on the reCAPTCHA service; `recaptchaV3` (a site SECRET instead of
  a site key) attests for free if the site does not need Enterprise.
- Enforcement is the project's decision: list the backends on the
  `GcpFirebaseProject`'s `appCheck.serviceConfigs`, `UNENFORCED` first.

## When to choose something else

A notifications-only site with no backend to protect takes the **Push Web
Client** preset.
