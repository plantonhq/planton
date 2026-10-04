# Privacy Policy and Terms of Service Pages

**Date**: March 25, 2026
**Type**: Feature
**Components**: Legal Pages, Footer, Content Management

## Summary

Added production-ready Privacy Policy and Terms of Service pages at `/privacy` and `/terms`, built with a markdown-content + server-component architecture that keeps legal text maintainable separately from presentation code. Content is tailored to Planton's specific data practices, credential handling, and Google OAuth consent screen requirements. Footer links updated from dead placeholders to the new pages.

## Problem Statement / Motivation

Planton.ai had no privacy policy or terms of service pages. The footer had dead links (`/`) for Privacy, Legal, and Terms & Conditions. This was a blocker for:

- Google OAuth consent screen approval (requires a publicly accessible privacy policy on the same domain that describes Google user data handling)
- General legal compliance for a SaaS platform processing user accounts, infrastructure credentials, and payment information
- Professional credibility with enterprise prospects evaluating the platform

### Pain Points

- Footer links for Privacy, Legal, and Terms all pointed to `/` (dead links)
- No public documentation of data handling practices, especially the reference-based secret management architecture
- Google OAuth consent review requires explicit mention of Google user data usage and compliance with the Google API Services User Data Policy

## Solution / What's New

### Architecture

Markdown files in `content/legal/` are read at build time by server components and rendered through a lightweight `LegalContent` client component. This keeps legal text editable without touching React code while maintaining consistent site styling.

### New Files

| File | Purpose |
|------|---------|
| `content/legal/privacy.md` | Privacy Policy content (Planton-specific, Google-compliant) |
| `content/legal/terms.md` | Terms of Service content (covers InfraHub, ServiceHub, Runner, Agent Fleet, OpenMCF) |
| `src/components/legal/LegalContent.tsx` | Lightweight ReactMarkdown renderer with prose-invert dark theme styling |
| `src/app/(root)/privacy/page.tsx` | Server component for `/privacy` route with SEO metadata |
| `src/app/(root)/terms/page.tsx` | Server component for `/terms` route with SEO metadata |

### Modified Files

| File | Change |
|------|--------|
| `src/components/layout/footer.tsx` | Updated `termsLinks` to point to `/privacy` and `/terms`, removed dead "Legal" entry |

### Privacy Policy Highlights

- **Section 1.C (Third-Party Sign-In Providers)**: Explicitly describes data received from Google and GitHub sign-in, which is what Google's OAuth reviewers look for
- **Section 2 (Infrastructure Credentials)**: Documents the reference-based secret management architecture -- Planton never stores plaintext credentials
- **Section 5 (Google User Data)**: Dedicated section declaring compliance with Google API Services User Data Policy and Limited Use requirements
- **Section 9 (Cookies)**: Cookie categories table (Essential, Functional, Analytics)

### Terms of Service Highlights

- **Section 1.1**: Full Service description covering InfraHub, ServiceHub, Runner, Agent Fleet (Beta), and OpenMCF
- **Section 1.3**: Explicit model training opt-in -- Planton will not use Content to train AI models without consent
- **Section 4.3**: OpenMCF carved out under Apache License 2.0, independent of the proprietary platform terms
- **Section 5.2**: Three security models documented (standard credentials, cross-account roles, customer-hosted runner)
- **Section 12**: Binding arbitration with 30-day opt-out window

## Implementation Details

- Pages live under the `(root)` route group, inheriting `MainLayout` (Header + Footer) automatically
- Content is read via `fs.readFileSync` at build time, compatible with Next.js `output: 'export'` static generation
- `LegalContent.tsx` uses `ReactMarkdown` with `remark-gfm` and `rehype-raw` (same plugins as the existing `MDXRenderer`) but stripped of docs-specific features (sidebar, next-article, page actions)
- Heading anchors use the existing `HeadingWithAnchor` and `generateHeadingId` components from `@/components/docs`
- Company details sourced from incorporation records: Planton Cloud, Inc., Delaware C-Corp, principal address in Fresno, CA

## Benefits

- Unblocks Google OAuth consent screen approval
- Provides legally required privacy and terms documentation for a SaaS platform
- Markdown-based content is easy for legal counsel to review and edit
- Footer links are now functional instead of dead
- Clean separation of legal content from presentation code

## Impact

- **Users**: Can now review Planton's data practices and service terms before signing up
- **Google OAuth**: Privacy policy meets Google's requirements for consent screen verification
- **Footer**: Three dead links replaced with two functional ones (`/privacy`, `/terms`)
- **Build**: Two new static pages added to the export, verified with `yarn build`

## Related Work

- ACME demo site already had placeholder privacy/terms pages at `/acme/privacy` and `/acme/terms` (hardcoded JSX) -- the new pages follow a different, more maintainable pattern
- Existing docs pipeline (`DocsLayout` + `MDXRenderer`) was intentionally not reused to keep legal pages lightweight

---

**Status**: Live
**Timeline**: Single session
