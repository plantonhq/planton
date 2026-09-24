# Verification — 2026-09-23

- Full `make -C site build` passed with the sibling platform checkout supplied: lint, TypeScript, source/route/link/claim guards, static export, generated text, and search index.
- Existing workflow acceptance: 308 checks passed (five-provider rotation, interruption, focus, hover, visibility, reduced motion, icons, layout).
- Homepage and mocked booking acceptance: 68 checks passed. Three mocked webhook requests; no real leads or bookings. Includes navigation/demo/self-service/product-example attribution.
- New experience acceptance: 48 checks passed. Desktop 1366×768, 1440×900, 1920×1080; tablet 768; mobile 320/390. Complete desktop hero above fold, image proportions and full export loading, no horizontal overflow, reduced motion, no JS, pause, replay after completion, and 200% CSS zoom reflow.
- Self-reviewed the hero, product evidence, controls, booking, mobile, and social preview captures. This is not an independent design review.
- Hero MP4 rendered with the shared scene: H.264, 1080×1080, 20 seconds, 0.38 MiB. Export stays outside Git. Existing seven workflow exports remain separate artifacts.
- The complete light-mode product export is preserved unchanged. Desktop/mobile show the original aspect ratio, with a full-size inspection link.
- Performance evidence is local headless Chrome with no network/CPU throttling and external scripts blocked. After: LCP 124–224 ms, CLS 0; sampled interactions 24–32 ms. Initial decoded JS declined from 1,805,441 to 1,709,190 bytes on desktop and 1,055,057 to 1,035,622 on mobile. These are lab comparisons, not field LCP/INP or conversion evidence.
- Browser preview and representative screenshots presented before PR creation/merge. Deployment verification follows the merge.

## Full-screen viewer follow-up

- Full build passed; experience acceptance now passes 80 checks.
- At 320 and 1366px: both image triggers open the dialog without navigation, dialog fills viewport, Close receives focus, Tab wraps inside, Escape restores opener focus, page scrolling is restored, zoom permits scrolling, and reopening resets zoom.
- Native dialog makes the background inert; explicit keyboard wrapping also prevents Tab escaping to browser chrome. Image aspect ratio remains intact at every zoom.
- Founder authorized a direct main follow-up. The image link remains a progressive fallback without JavaScript.
