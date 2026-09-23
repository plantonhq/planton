# Website Advanced Site Search

## Use Case

Search over your own public website with advanced site search: the domain is verified in Search Console, Google reads the sitemap, the archive is excluded from the crawl, and the index quota is the larger one. The store enrolls in search and chat, so a search engine with generated answers or a chat engine can sit on top of it.

## When to Use

- A documentation, support, or marketing site you own
- When you need extractive answers, summaries, or a chat experience over the site
- When basic site search's index quota is not enough

## What This Creates

- A `PUBLIC_WEBSITE` data store in the `global` location with `createAdvancedSiteSearch`
- Two target sites: `docs.example.com/*` included, `docs.example.com/archive/*` excluded
- One sitemap at `https://docs.example.com/sitemap.xml`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `targetSites[].providedUriPattern` | `docs.example.com/*` | Your site; add INCLUDE patterns per section and EXCLUDE patterns for what should never surface. |
| `sitemapUris` | one sitemap | Every sitemap Google should read. |
| `advancedSiteSearchConfig.disableInitialIndex` | `false` | `true` to create the store without crawling until you trigger it. |
| `createAdvancedSiteSearch` | `true` | `false` for basic site search over public pages with no domain verification (and no sitemaps). |

Advanced site search needs the domain verified in Search Console by the project's owner before Google indexes it; basic site search does not. Target sites and sitemaps are immutable: a change replaces that entry.
