import { test as base, expect } from "@playwright/test";

export const test = base.extend<{ checkBrowserErrors: void }>({
  checkBrowserErrors: [
    async ({ page }, use) => {
      const errors: string[] = [];
      page.on("pageerror", (error) => errors.push(error.message));
      await use();
      expect(
        errors,
        "The page must not produce unhandled JavaScript errors",
      ).toEqual([]);
    },
    { auto: true },
  ],
});
export { expect };
