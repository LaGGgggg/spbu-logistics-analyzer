import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { exampleGraph } from "../src/data/example";

const region = {
  id: "spb-leningrad",
  name: "Санкт-Петербург и Ленинградская область",
  bounds: { west: 29, south: 59, east: 31, north: 61 },
};
const graph = {
  ...exampleGraph,
  id: "real-graph-1",
  region_id: region.id,
  coverage: {
    ...exampleGraph.coverage!,
    warnings: ["Проверка отчёта backend"],
    sources: ["freight.osm"],
  },
};
async function base(page: Page) {
  await page.route("https://tile.openstreetmap.org/**", (route) =>
    route.abort(),
  );
  await page.route("https://fonts.googleapis.com/**", (route) => route.abort());
  await page.route("**/health", (route) =>
    route.fulfill({
      json: { status: "healthy", service: "Logistics Network API" },
    }),
  );
  await page.route("**/api/regions", (route) =>
    route.fulfill({ json: [region] }),
  );
}

test("example is clearly labelled, supports node inspection and never starts a build", async ({
  page,
}) => {
  await base(page);
  let posts = 0;
  await page.route("**/api/graphs", (route) => {
    posts++;
    return route.fulfill({ json: {} });
  });
  await page.goto("/");
  await expect(page.getByText("Backend подключён")).toBeVisible();
  await expect(page.getByText("Пример · локальные данные")).toBeVisible();
  await page.getByRole("tab", { name: "Узлы" }).click();
  await page.getByRole("textbox", { name: "Поиск узла" }).fill("Автово");
  await expect(
    page.getByRole("row").filter({ hasText: "Грузовой терминал Автово" }),
  ).toHaveCount(1);
  await page
    .getByRole("button", { name: "Показать Грузовой терминал Автово на карте" })
    .click();
  await expect(
    page.getByRole("complementary", { name: "Карточка узла" }),
  ).toContainText("example:avtovo");
  expect(posts).toBe(0);
});

test("build, long poll, filter via API, read coverage and export", async ({
  page,
}) => {
  await base(page);
  let polls = 0;
  await page.route("**/api/graphs", async (route) => {
    expect(route.request().postDataJSON()).toEqual({ region_id: region.id });
    await route.fulfill({ status: 202, json: { job_id: "job-1" } });
  });
  await page.route("**/api/jobs/job-1*", async (route) => {
    if (polls > 0) {
      const url = new URL(route.request().url());
      expect(url.searchParams.get("wait")).toBe("25");
      expect(url.searchParams.get("after")).toBe("2026-10-21T00:00:00Z");
      await new Promise((resolve) => setTimeout(resolve, 200));
    }
    return route.fulfill({
      json: {
        id: "job-1",
        region_id: region.id,
        status: ++polls === 1 ? "processing" : "completed",
        graph_id: polls > 1 ? graph.id : undefined,
        updated_at: "2026-10-21T00:00:00Z",
      },
    });
  });
  await page.route(`**/api/graphs/${graph.id}`, (route) =>
    route.fulfill({ json: graph }),
  );
  await page.route(`**/api/graphs/${graph.id}?mode=rail`, (route) =>
    route.fulfill({
      json: {
        ...graph,
        nodes: graph.nodes.filter((n) => n.modes.includes("rail")),
        edges: graph.edges.filter((e) => e.mode === "rail"),
      },
    }),
  );
  await page.route(`**/api/graphs/${graph.id}/coverage`, (route) =>
    route.fulfill({ json: graph.coverage }),
  );
  await page.route(`**/api/graphs/${graph.id}/export/geojson`, (route) =>
    route.fulfill({
      contentType: "application/geo+json",
      body: '{"type":"FeatureCollection","features":[]}',
    }),
  );
  await page.goto("/");
  await page
    .getByRole("button", { name: "Построить граф", exact: true })
    .click();
  await expect(page.getByText("Строится на backend")).toBeVisible();
  await expect(
    page.getByText("Результат backend", { exact: true }),
  ).toBeVisible();
  await expect(page).toHaveURL(new RegExp(`graph=${graph.id}`));
  await page
    .getByRole("button", { name: "Железная дорога", exact: true })
    .click();
  await expect(
    page.locator(".request-path").filter({ hasText: "?mode=rail" }),
  ).toBeVisible();
  await page.getByRole("tab", { name: "Качество данных" }).click();
  await page.getByRole("button", { name: "Запросить coverage" }).click();
  await expect(
    page.getByText("Проверка отчёта backend", { exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Экспорт", exact: true }).click();
  const download = page.waitForEvent("download");
  await page.getByRole("button", { name: "geojson", exact: true }).click();
  expect((await download).suggestedFilename()).toContain(graph.id);
  await expect(
    page.locator(".request-path").filter({ hasText: "/export/geojson" }),
  ).toBeVisible();
});

test("cached result skips polling and survives a page reload", async ({
  page,
}) => {
  await base(page);
  let polls = 0;
  await page.route("**/api/jobs/**", (route) => {
    polls++;
    return route.fulfill({ status: 404, json: { error: "not found" } });
  });
  await page.route("**/api/graphs", (route) =>
    route.fulfill({ json: { graph_id: graph.id } }),
  );
  await page.route(`**/api/graphs/${graph.id}`, (route) =>
    route.fulfill({ json: graph }),
  );
  await page.goto("/");
  await page
    .getByRole("button", { name: "Построить граф", exact: true })
    .click();
  await expect(
    page.getByText("Результат backend", { exact: true }),
  ).toBeVisible();
  await page.reload();
  await expect(
    page.getByText("Результат backend", { exact: true }),
  ).toBeVisible();
  expect(polls).toBe(0);
});

test("validates bounds and surfaces backend errors without replacing the example", async ({
  page,
}) => {
  await base(page);
  await page.route("**/api/graphs", async (route) => {
    expect(route.request().postDataJSON()).toEqual({
      bounds: { west: 30.16, south: 59.85, east: 30.34, north: 59.97 },
    });
    await route.fulfill({
      status: 422,
      json: { error: "Для этой области нет источника данных." },
    });
  });
  await page.goto("/");
  await page.getByRole("button", { name: "Своя область", exact: true }).click();
  await page.getByRole("spinbutton", { name: "Запад" }).fill("100");
  await expect(
    page.getByRole("button", { name: "Построить граф", exact: true }),
  ).toBeDisabled();
  await page.getByRole("spinbutton", { name: "Запад" }).fill("30.16");
  await page
    .getByRole("button", { name: "Построить граф", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText(
    "Для этой области нет источника данных.",
  );
  await expect(page.getByText("Пример · локальные данные")).toBeVisible();
  await expect(
    page.locator(".http-status.failure").filter({ hasText: "422" }),
  ).toBeVisible();
});

test("stopping long polling leaves the server job resumable", async ({ page }) => {
  await base(page);
  let completed = false;
  await page.route("**/api/graphs", (route) =>
    route.fulfill({ status: 202, json: { job_id: "long-job" } }),
  );
  await page.route("**/api/jobs/long-job*", async (route) => {
    if (!completed && new URL(route.request().url()).searchParams.has("wait"))
      await new Promise((resolve) => setTimeout(resolve, 100));
    return route.fulfill({
      json: {
        id: "long-job",
        region_id: region.id,
        status: completed ? "completed" : "processing",
        graph_id: completed ? graph.id : undefined,
        updated_at: "2026-10-21T00:00:00Z",
      },
    });
  });
  await page.route(`**/api/graphs/${graph.id}`, (route) =>
    route.fulfill({ json: graph }),
  );
  await page.goto("/");
  await page
    .getByRole("button", { name: "Построить граф", exact: true })
    .click();
  await expect(page.getByText("Строится на backend")).toBeVisible();
  await page.getByRole("button", { name: "Прекратить ожидание" }).click();
  await expect(page.getByText("Ожидание приостановлено")).toBeVisible();
  completed = true;
  await page.getByRole("button", { name: "Продолжить ожидание" }).click();
  await expect(
    page.getByText("Результат backend", { exact: true }),
  ).toBeVisible();
});

test("offline state is honest and API reference stays usable on mobile", async ({
  page,
}) => {
  await base(page);
  await page.route("**/health", (route) =>
    route.fulfill({ status: 502, json: { error: "offline" } }),
  );
  await page.route("**/api/regions", (route) =>
    route.fulfill({ status: 502, json: { error: "offline" } }),
  );
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto("/");
  await expect(page.getByText("Backend недоступен")).toBeAttached();
  await page.getByRole("tab", { name: "API", exact: true }).click();
  await expect(page.getByText("API, которое стоит за картой")).toBeVisible();
  const download = page.waitForEvent("download");
  await page.getByRole("button", { name: "OpenAPI 3.1" }).click();
  expect((await download).suggestedFilename()).toBe("openapi.yaml");
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  ).toBe(true);
});

test("a failed job displays its error and keeps the previous graph", async ({
  page,
}) => {
  await base(page);
  await page.route("**/api/graphs", (route) =>
    route.fulfill({ status: 202, json: { job_id: "failed-job" } }),
  );
  await page.route("**/api/jobs/failed-job", (route) =>
    route.fulfill({
      json: {
        id: "failed-job",
        region_id: region.id,
        status: "failed",
        error: "Не удалось построить граф из исходных данных.",
      },
    }),
  );
  await page.goto("/");
  await page
    .getByRole("button", { name: "Построить граф", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText(
    "Не удалось построить граф из исходных данных.",
  );
  await expect(
    page.getByText("Ошибка построения", { exact: true }),
  ).toBeVisible();
  await expect(page.getByText("Пример · локальные данные")).toBeVisible();
});

test("map selection fills validated bounds without starting a server job", async ({
  page,
}) => {
  await base(page);
  let submitted = false;
  await page.route("**/api/graphs", (route) => {
    submitted = true;
    return route.fulfill({ json: {} });
  });
  await page.goto("/");
  await page
    .getByRole("button", { name: "Выбрать область", exact: true })
    .click();
  await expect(
    page.getByText("Укажите два противоположных угла области. Esc — отмена."),
  ).toBeVisible();
  await page.locator(".graph-map").click({ position: { x: 160, y: 140 } });
  await page.locator(".graph-map").click({ position: { x: 300, y: 260 } });
  const west = Number(
    await page.getByRole("spinbutton", { name: "Запад" }).inputValue(),
  );
  const east = Number(
    await page.getByRole("spinbutton", { name: "Восток" }).inputValue(),
  );
  const south = Number(
    await page.getByRole("spinbutton", { name: "Юг" }).inputValue(),
  );
  const north = Number(
    await page.getByRole("spinbutton", { name: "Север" }).inputValue(),
  );
  expect(west).toBeLessThan(east);
  expect(south).toBeLessThan(north);
  await expect(
    page.getByRole("button", { name: "Построить граф", exact: true }),
  ).toBeEnabled();
  expect(submitted).toBe(false);
});

test("opening the example does not cancel an in-flight connection check", async ({
  page,
}) => {
  await base(page);
  await page.route("**/health", async (route) => {
    await new Promise((resolve) => setTimeout(resolve, 500));
    await route.fulfill({
      json: { status: "healthy", service: "Logistics Network API" },
    });
  });
  await page.goto("/");
  await page.getByRole("button", { name: /Посмотреть пример/ }).click();
  await expect(page.getByText("Backend подключён")).toBeVisible();
  await expect(page.getByText("Пример · локальные данные")).toBeVisible();
});
