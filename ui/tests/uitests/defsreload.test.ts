import { test, expect } from '../testutils';

const CHANGED_MOCK = JSON.stringify([
  {
    active: true,
    comment: 'changed on disk',
    method: 'GET',
    urlpart: '/api/changed',
    statuscode: 503,
    freezetimems: 0,
    payload: { changed: true },
    payloadFromFile: '',
  },
]);

test.describe('defs reload tests', () => {
  test('should re-render the mock list when the file changed on disk', async ({ mockedPage: page }) => {
    await expect(page.locator('#comment_mock_0')).toHaveValue('test123 mock');

    await page.route('**/moxyadminui/mockdef', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: CHANGED_MOCK });
    });

    await page.evaluate("reloadDefsFromDisk('mockdef.json')");

    await expect(page.locator('#comment_mock_0')).toHaveValue('changed on disk');
    await expect(page.locator('#urlpart_mock_0')).toHaveValue('/api/changed');
  });

  test('should still re-render when a field merely has focus', async ({ mockedPage: page }) => {
    // wait for the initial fetch before swapping the route, otherwise they race
    await expect(page.locator('#comment_mock_0')).toHaveValue('test123 mock');

    await page.route('**/moxyadminui/mockdef', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: CHANGED_MOCK });
    });

    await page.locator('#urlpart_mock_0').focus();

    await page.evaluate("reloadDefsFromDisk('mockdef.json')");

    await expect(page.locator('#urlpart_mock_0')).toHaveValue('/api/changed');
  });

  test('should defer the re-render while a field has unsaved keystrokes', async ({ mockedPage: page }) => {
    await expect(page.locator('#comment_mock_0')).toHaveValue('test123 mock');

    await page.route('**/moxyadminui/mockdef', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: CHANGED_MOCK });
    });

    await page.locator('#urlpart_mock_0').focus();
    await page.keyboard.type('-half-typed');

    const result = await page.evaluate("reloadDefsFromDisk('mockdef.json')");

    expect(String(result)).toContain('will reload when you finish editing');
    await expect(page.locator('#urlpart_mock_0')).toHaveValue('-half-typed/api/sometest');
  });

  test('should reload once the edited field loses focus', async ({ mockedPage: page }) => {
    await expect(page.locator('#comment_mock_0')).toHaveValue('test123 mock');

    await page.route('**/moxyadminui/mockdef', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({ status: 200, body: '' });
        return;
      }
      await route.fulfill({ status: 200, contentType: 'application/json', body: CHANGED_MOCK });
    });

    await page.locator('#urlpart_mock_0').focus();
    await page.keyboard.type('-half-typed');
    await page.evaluate("reloadDefsFromDisk('mockdef.json')");

    // put the field back so blurring does not fire a save that would race the deferred reload
    await page.evaluate("document.getElementById('urlpart_mock_0').value = '/api/sometest'");
    await page.locator('#urlpart_mock_0').blur();

    await expect(page.locator('#urlpart_mock_0')).toHaveValue('/api/changed');
  });

  test('should do nothing when the file content is unchanged', async ({ mockedPage: page }) => {
    await expect(page.locator('#comment_mock_0')).toHaveValue('test123 mock');

    const result = await page.evaluate("reloadDefsFromDisk('mockdef.json')");

    expect(result).toBeNull();
    await expect(page.locator('#comment_mock_0')).toHaveValue('test123 mock');
  });
});
