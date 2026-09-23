# Launch verification

- Full `make build` passes with PLANTON_PLATFORM_DIR set to the source checkout: lint, types, pricing/entitlement agreement, apex routing, static export, internal links, discovery export, and search generation.
- `node scripts/check-homepage.mjs`: 63 checks pass. Widths 320, 390, 768, 1280, 1680; light-only homepage; three attributed CTAs; coding-agent story; text contrast; keyboard layer selection and FAQ; reduced motion; light desktop-menu and mobile-drawer portals; form validation; server/network retry with fields retained; unchanged payload; light prefilled calendar; focus handoff; confirmed-event deduplication; pending/payment-required events excluded; no personal information in analytics; calendar fallback; dark docs/product/pricing preserved. All external requests are mocked, including three lead submissions. No real lead or meeting was created.
- Existing Product and Pricing pages: pixel-identical to pre-change 1280px captures.
- Local screenshot review: desktop and mobile typography, illustrations, menus, booking form, and light social card inspected. Fixed desktop headline wrapping and mobile booking-form order. Hero construction guides removed. Header menu controls are keyboard-operable and the nested main landmark was removed.
- Social card generated from the editable homepage at 1200 × 630. Static source content is available without interaction; illustrations reserve their dimensions.
- Review evidence: site/content/design-review/homepage-v6. Complete captures at /private/tmp/planton-homepage-review during this session; CI uploads fresh captures as homepage-review.
- Scope: homepage, booking journey, opt-in light shell tokens, homepage discovery and metadata, and PR validation. The shared shell package is consumed from the website workspace; it is not published to npm and the product's installed package/theme is unchanged.
- Remaining limitation: no real booking submission was made; conversion uplift requires production funnel data. The rest of the marketing site intentionally remains dark during the transition.

Release procedure: merge the reviewed PR after all checks pass, watch release.site, then verify public HTML, assets, mobile layout and booking surface. Revert the release commit through a PR if a material regression is found. The prior v5 homepage is retained for composition rollback.
