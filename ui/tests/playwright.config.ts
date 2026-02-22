import { defineConfig, devices } from '@playwright/test';
import path from 'path';

const UI_SERVER_PORT = 9099;

export default defineConfig({
  testDir: './uitests',
  timeout: 10_000,
  retries: process.env.CI ? 2 : 0,
  fullyParallel: false,
  reporter: [['list'], ['html', { open: 'never' }]],

  // Spin up a lightweight static HTTP server for the Admin UI so that
  // page.route() can intercept all requests (file:// protocol is not
  // interceptable by Playwright's routing layer).
  webServer: {
    command: `node ${path.resolve(__dirname, 'serve-ui.js')} ${UI_SERVER_PORT}`,
    url: `http://127.0.0.1:${UI_SERVER_PORT}`,
    reuseExistingServer: !process.env.CI,
    stdout: 'pipe',
    stderr: 'pipe',
  },

  use: {
    baseURL: `http://127.0.0.1:${UI_SERVER_PORT}`,
    screenshot: 'only-on-failure',
    trace: 'on-first-retry',
    launchOptions: {
      args: ['--no-sandbox', '--disable-setuid-sandbox'],
    },
  },

  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
});
