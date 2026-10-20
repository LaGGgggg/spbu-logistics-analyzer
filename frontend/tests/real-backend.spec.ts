import { expect, test } from "./fixtures";

// Opt-in test against an isolated Go service configured with backend/tests/testdata.
// RUN_BACKEND_SMOKE=1 BACKEND_URL=http://127.0.0.1:18080 npm run test:e2e
// It must not run against a production extract, because building could download gigabytes.
test("real Go backend: submit, poll and render a persisted graph", async ({
  page,
  request,
}, testInfo) => {
  test.skip(
    process.env.RUN_BACKEND_SMOKE !== "1",
    "Requires an isolated fixture backend",
  );
  const response = await request.get("/api/regions");
  expect(response.ok()).toBeTruthy();
  const regions = (await response.json()) as { id: string }[];
  expect(regions.some((r) => r.id === "ui-fixture")).toBeTruthy();
  await page.goto("/");
  await expect(page.getByText("Backend подключён")).toBeVisible();
  await expect(page.locator(".leaflet-pane canvas")).toBeVisible();
  await page.screenshot({
    path: testInfo.outputPath("desktop.png"),
    fullPage: true,
  });
  await page
    .getByRole("button", { name: "Построить граф", exact: true })
    .click();
  await expect(
    page.getByText("Результат backend", { exact: true }),
  ).toBeVisible();
  await expect(page.locator(".stat-card").first()).toContainText("3");
  await expect(page).toHaveURL(/graph=/);
  await page.getByRole("tab", { name: "Качество данных" }).click();
  await expect(page.getByText("freight.osm", { exact: true })).toBeVisible();
  await page.getByRole("tab", { name: "Карта сети" }).click();
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(
    page.getByRole("button", { name: "Параметры территории" }),
  ).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  ).toBe(true);
  await page.screenshot({
    path: testInfo.outputPath("mobile.png"),
    fullPage: true,
  });
});
