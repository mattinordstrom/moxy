import { test as base, expect, Page } from '@playwright/test';

// The index page is served by the webServer defined in playwright.config.ts.
// The server serves from the project root, so the UI lives at /ui/index.html.
// baseURL is set to http://127.0.0.1:9099, so a relative path is enough.
const INDEX_PATH = '/ui/index.html?mode=test';

async function setupRoutes(page: Page): Promise<void> {
  // Stub the mock definitions endpoint with one test mock entry.
  // The `comment` field is required so the UI renders `#comment_mock_0`.
  await page.route('**/moxyadminui/mockdef', async (route) => {
    if (route.request().method() === 'POST') {
      // Allow the UI's save calls to succeed silently.
      await route.fulfill({ status: 200, body: '' });
      return;
    }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([
        {
          active: true,
          comment: 'test123 mock',
          method: 'GET',
          urlpart: '/api/sometest',
          statuscode: 200,
          freezetimems: 0,
          payload: { test: 123 },
          payloadFromFile: '',
        },
      ]),
    });
  });

  // Stub the proxy definitions endpoint with an empty list.
  await page.route('**/moxyadminui/proxydef', async (route) => {
    if (route.request().method() === 'POST') {
      await route.fulfill({ status: 200, body: '' });
      return;
    }
    await route.fulfill({ status: 200, contentType: 'application/json', body: '[]' });
  });

  // Stub the settings endpoint — intercepted on every call including
  // the re-fetch inside listPayloadFiles() when the modal is opened.
  await page.route('**/moxyadminui/settings', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        port: 9097,
        defaultRoute: 'http://localhost:8089',
        payloadFiles: ['/home/testuser/projects/moxyarchive/payloads/example.json'],
        payloadPath: '/home/testuser/projects/moxyarchive/payloads/',
        maxLogEntries: null,
      }),
    });
  });

  // Block any external CDN requests (e.g. font-awesome) to keep tests offline.
  await page.route(/cdnjs\.cloudflare\.com/, async (route) => {
    await route.fulfill({ status: 200, contentType: 'text/css', body: '' });
  });
}

// Custom fixture that provides a fully mocked and loaded page.
export const test = base.extend<{ mockedPage: Page }>({
  mockedPage: async ({ page }, use) => {
    await setupRoutes(page);
    await page.goto(INDEX_PATH);
    await use(page);
  },
});

export { expect };
