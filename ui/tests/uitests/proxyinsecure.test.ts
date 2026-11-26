import { test, expect } from '../testutils';

const proxy = { active: true, target: 'https://dev.example', urlpart: '/api/dev', verbose: false, insecure: true };

test.describe('insecure proxy tests', () => {
  test('should show and toggle insecure on a proxy', async ({ mockedPage: page }) => {
    const posted: string[] = [];
    await page.route('**/moxyadminui/proxydef', async (route) => {
      if (route.request().method() === 'POST') {
        posted.push(route.request().postData() ?? '');
        await route.fulfill({ status: 200, body: '' });
        return;
      }
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify([proxy]) });
    });
    await page.reload();

    await expect(page.locator('#insecure_0')).toBeChecked();

    await page.locator('#insecure_0').click();

    await expect.poll(() => posted.length).toBe(1);
    expect(JSON.parse(posted[0])[0].insecure).toBe(false);
  });

  test('should mark an insecure default route in the header', async ({ mockedPage: page }) => {
    await page.route('**/moxyadminui/settings', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ port: 9097, defaultRoute: 'https://dev.example', defaultInsecure: true, payloadFiles: [], payloadPath: '' }),
      });
    });
    await page.reload();

    await expect(page.locator('#header-route')).toHaveText('https://dev.example (insecure TLS)');
  });
});
