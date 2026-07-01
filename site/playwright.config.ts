import { defineConfig, devices } from "@playwright/test";

/**
 * Smoke-test config for planton.dev. Reuses a running `yarn dev` if present,
 * otherwise starts one. Kept intentionally small — these are fast guardrails
 * (render, no console errors, working links), not exhaustive E2E.
 */
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  reporter: "list",
  use: {
    baseURL: "http://localhost:3000",
    trace: "on-first-retry",
  },
  projects: [
    { name: "chromium", use: { ...devices["Desktop Chrome"] } },
    { name: "mobile", use: { ...devices["Pixel 7"] } },
  ],
  webServer: {
    command: "yarn dev",
    url: "http://localhost:3000",
    reuseExistingServer: true,
    timeout: 120_000,
  },
});
