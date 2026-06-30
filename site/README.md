# Planton Website

<!-- Auto-release test: Website change triggers v{semver}-website-{YYYYMMDD}.{N} tag format -->

Next.js App Router project for the planton.dev documentation website.

Key packages:

- Next.js 15 (App Router)
- Tailwind CSS v4
- React 19
- Radix UI (tabs), CVA, tailwind-merge, clsx

Dev commands (Yarn):

```bash
yarn dev
yarn build
yarn start
yarn lint
```

Folder structure:

- `src/app` – Next.js App Router pages and routes (home, robots, sitemap)
- `src/components` – UI primitives and page sections
- `src/lib/utils.ts` – `cn()` helper for classnames
- `public/` – static assets (`icon.png`)

Notes:

- All Base44 SDK references have been removed in this project.
- The UI sections mirror the original Vite site components.

