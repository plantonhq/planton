import coreWebVitals from "eslint-config-next/core-web-vitals";
import typescript from "eslint-config-next/typescript";

const eslintConfig = [
  { ignores: ["packages/*/dist/**"] },
  ...coreWebVitals,
  ...typescript,
  {
    rules: {
      // Every image says what it shows, and an image the build knows goes
      // through the framework's image element so its size is declared. An
      // image the build cannot know (a document's own picture, a record's
      // avatar) is a plain <img> with a one-line exception at the site
      // saying why. These rules hold everywhere; there is no allowlist.
      "@next/next/no-img-element": "error",
      "jsx-a11y/alt-text": "error",
      "@typescript-eslint/no-explicit-any": "error",
      "@next/next/no-page-custom-font": "off",
      "@next/next/google-font-display": "error",
      "@typescript-eslint/no-unused-vars": ["warn", { argsIgnorePattern: "^_", varsIgnorePattern: "^_", caughtErrorsIgnorePattern: "^_" }],
      "react-hooks/exhaustive-deps": "warn",
    },
  },
];

export default eslintConfig;
