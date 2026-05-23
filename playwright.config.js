// @ts-check
const { defineConfig, devices } = require("@playwright/test");

const port = Number(process.env.E2E_PORT || 8090);
const dbPath = process.env.E2E_DB_PATH || `/tmp/bill-of-fare-ui-${port}.db`;
const goCache = process.env.GOCACHE || `/tmp/bill-of-fare-ui-gocache-${port}`;

module.exports = defineConfig({
  testDir: "./tests/ui",
  timeout: 30_000,
  expect: {
    timeout: 5_000,
  },
  fullyParallel: false,
  reporter: process.env.CI ? [["github"], ["html", { open: "never" }]] : "list",
  use: {
    baseURL: `http://127.0.0.1:${port}`,
    trace: "on-first-retry",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  webServer: {
    command: `sh -c 'rm -f "${dbPath}" && GOCACHE="${goCache}" go run ./cmd/seed -db "${dbPath}" -menu seed/menu.yaml && GOCACHE="${goCache}" DB_PATH="${dbPath}" HOST="127.0.0.1" PORT="${port}" go run ./cmd/server'`,
    url: `http://127.0.0.1:${port}`,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
