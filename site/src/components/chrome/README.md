# `components/chrome`

Shared site chrome — the header and footer used by both the landing page and the
docs, plus the CTAs they contain. Defining these once keeps brand and navigation
consistent across every route.

- `SiteHeader` = `HeaderBrand` + `HeaderNav` + `HeaderActions`. Accepts optional
  `leading` (e.g. docs mobile menu) and `slot` (e.g. docs search) so docs can
  reuse it without duplicating the brand.
- `SiteFooter` = `FooterBrand` + `FooterLinks`.
- `DownloadButton` — the primary CTA; targets `DOWNLOAD_HREF` from `@/site`.
- `GitHubStars` — quiet star pill with a live, best-effort count.

All links/targets come from `@/site`; nothing here hardcodes a URL.
