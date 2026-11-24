import { test, expect } from '../testutils';

test.describe('proxyentity tests', () => {
  test('should show and toggle preserveHost on a proxy', async ({ mockedPage: page }) => {
    const postedBodies: string[] = [];
    await page.route('**/moxyadminui/proxydef', async (route) => {
      if (route.request().method() === 'POST') {
        postedBodies.push(route.request().postData() ?? '');
        await route.fulfill({ status: 200, body: '' });
        return;
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          { active: true, preserveHost: true, target: 'http://route-utv.fnox.se', urlpart: '/api/test', verbose: false },
        ]),
      });
    });
    await page.reload();

    const checkbox = page.locator('#preserveHost_0');
    await expect(checkbox).toBeChecked();

    await checkbox.click();

    await expect.poll(() => postedBodies.length).toBeGreaterThan(0);
    expect(JSON.parse(postedBodies[postedBodies.length - 1])[0].preserveHost).toBe(false);
  });

  test('should show preserve host for the default route in the header', async ({ mockedPage: page }) => {
    await page.route('**/moxyadminui/settings', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          port: 9097,
          defaultRoute: 'http://route-utv.fnox.se',
          defaultPreserveHost: true,
          payloadFiles: [],
          payloadPath: '/tmp/',
          maxLogEntries: null,
        }),
      });
    });
    await page.reload();

    await expect(page.locator('#header-route')).toHaveText('http://route-utv.fnox.se (preserve host)');
  });
});
