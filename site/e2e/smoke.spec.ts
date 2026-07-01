import { test, expect, type Page } from "@playwright/test";

/**
 * Collects console errors and uncaught page errors for a page interaction, so
 * tests can assert a clean console (catches hydration mismatches, bad imports).
 */
function trackErrors(page: Page): string[] {
  const errors: string[] = [];
  page.on("console", (msg) => {
    if (msg.type() === "error") errors.push(msg.text());
  });
  page.on("pageerror", (err) => errors.push(err.message));
  return errors;
}

test("landing renders with no console errors", async ({ page }) => {
  const errors = trackErrors(page);
  await page.goto("/");

  await expect(
    page.getByRole("heading", {
      name: /Deploy real infrastructure to your own cloud/i,
      level: 1,
    }),
  ).toBeVisible();

  // The primary CTA and brand are present.
  await expect(page.getByRole("link", { name: "Planton home" })).toBeVisible();
  await expect(page.getByRole("link", { name: "Download Planton" }).first()).toBeVisible();

  expect(errors, `console errors:\n${errors.join("\n")}`).toEqual([]);
});

test("showcase tabs switch between Desktop and Terminal", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("tab", { name: "Terminal" }).first().click();
  await expect(page.getByText("planton chart install", { exact: false }).first()).toBeVisible();
  await page.getByRole("tab", { name: "Desktop" }).first().click();
  await expect(page.getByText("Planton Desktop").first()).toBeVisible();
});

test("every header and footer link has a real target", async ({ page }) => {
  await page.goto("/");
  const links = page.locator("header a, footer a");
  const count = await links.count();
  expect(count).toBeGreaterThan(0);
  for (let i = 0; i < count; i++) {
    const href = await links.nth(i).getAttribute("href");
    expect(href, "link is missing href").toBeTruthy();
    expect(href, `dead link: ${href}`).not.toBe("#");
  }
});

test("download page renders with no console errors", async ({ page }) => {
  const errors = trackErrors(page);
  await page.goto("/download");
  await expect(page.getByRole("heading", { name: "Download Planton", level: 1 })).toBeVisible();
  await expect(page.getByRole("link", { name: /Download for macOS/i })).toBeVisible();
  expect(errors, `console errors:\n${errors.join("\n")}`).toEqual([]);
});
