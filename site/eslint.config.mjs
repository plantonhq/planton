import coreWebVitals from "eslint-config-next/core-web-vitals";
import typescript from "eslint-config-next/typescript";

/**
 * Folders that predate the accessibility and image rules being enforced. Each
 * entry names the work that retires it; when a folder is rebuilt from the story
 * (or deleted), its line goes and the rules apply there too. Nothing new is
 * ever added here: a new page passes the rules or it does not ship.
 */
const LEGACY_ALLOWLIST = [
  "src/components/product/solutions/**", // the Solutions pages rebuilt from personas.ts
  "src/components/product/shared/**",    // the 2025 product kit; dies with the Solutions pages and the invest explainer that still compose it
  "src/components/product/desktop/**",   // the desktop landing's move onto the palette's role classes
  "src/components/pricing/**",        // the pricing page's next content pass
  "src/components/enterprise/**",     // the pricing page's next content pass
  "src/components/landing-page/v1-*/**", // dies with the 2025 hackathon
  "src/components/landing-page/v3-*/**", // rollback folder; deleted when v5 has held for one release
  "src/components/landing-page/v4-*/**", // rollback folder; deleted when v5 has held for one release
  "src/components/hackathon/**",      // the 2025 hackathon's retirement
  "src/components/invest/**",         // invest onto the deck engine
  "src/components/meetings/**",       // the meeting decks' slide kit convergence
  "src/components/demo/**",           // the interactive demo's decision (rebuild on the design system, or retire)
  "src/components/tour/**",           // the interactive demo's decision
  "src/components/blog/**",           // the content pages' pass
  "src/components/docs/**",           // the docs rebuild
  "src/lib/MDXRenderer.tsx",          // the docs rebuild (the markdown renderer)
  "src/lib/mdx.ts",                   // the docs rebuild
  "src/lib/mdx-client.ts",            // the docs rebuild
  "src/components/tutorials/**",      // the content pages' pass
  "src/components/changelog/**",      // the content pages' pass
  "src/components/common/**",         // shrinks as primitives move to components/marketing
  "src/components/book-demo/**",      // the book-demo page's pass
  "src/components/legal/**",          // the legal renderer's pass
  "src/components/branding/**",       // the design-system page's pass
  "src/components/investor-updates/**", // invest onto the deck engine
  "src/app/**",                       // route files; shrinks with each page group rebuilt
];

const eslintConfig = [
  { ignores: ["packages/*/dist/**"] },
  ...coreWebVitals,
  ...typescript,
  {
    rules: {
      // Every image says what it shows, and marketing images go through the
      // framework's image element so sizes are declared. These were switched
      // off for a long time; they are on, with the allowlist above as the debt.
      "@next/next/no-img-element": "error",
      "jsx-a11y/alt-text": "error",
      "@typescript-eslint/no-explicit-any": "error",
      "@next/next/no-page-custom-font": "off",
      "@next/next/google-font-display": "error",
      "@typescript-eslint/no-unused-vars": ["warn", { argsIgnorePattern: "^_", varsIgnorePattern: "^_", caughtErrorsIgnorePattern: "^_" }],
      "react-hooks/exhaustive-deps": "warn",
    },
  },
  {
    files: LEGACY_ALLOWLIST,
    rules: {
      "@next/next/no-img-element": "warn",
      "jsx-a11y/alt-text": "warn",
      "@typescript-eslint/no-explicit-any": "warn",
    },
  },
];

export default eslintConfig;
