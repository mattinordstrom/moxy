import { test, expect } from '../testutils';

const mock = (extra: object) => ({
  active: true, freezetimems: 0, method: 'GET', payload: null,
  payloadFromFile: '/tmp/a.json', statuscode: 200, urlpart: '/api/a', ...extra,
});

const serveMocks = async (page, mocks: object[]) => {
  const posted: string[] = [];
  await page.route('**/moxyadminui/mockdef', async (route) => {
    if (route.request().method() === 'POST') {
      posted.push(route.request().postData() ?? '');
      await route.fulfill({ status: 200, body: '' });
      return;
    }
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(mocks) });
  });
  await page.reload();
  return posted;
};

test.describe('payload source tests', () => {
  test('should show and switch which payload a mock uses', async ({ mockedPage: page }) => {
    const posted = await serveMocks(page, [mock({ usePayloadFromFile: false })]);

    await expect(page.locator('#payloadsource_payload_0')).toBeChecked();
    await expect(page.locator('#payloadfromfileobj_0')).toHaveClass(/payload-unused/);

    await page.locator('#payloadsource_file_0').check();

    await expect(page.locator('#payloadobj_0')).toHaveClass(/payload-unused/);
    await expect(page.locator('#payloadfromfileobj_0')).not.toHaveClass(/payload-unused/);
    await expect.poll(() => posted.length).toBe(1);
    const saved = JSON.parse(posted[0])[0];
    expect(saved.usePayloadFromFile).toBe(true);
    expect(saved.payload).toBeNull();
  });

  test('should pin a mock without the toggle to the source it was shown with', async ({ mockedPage: page }) => {
    const posted = await serveMocks(page, [mock({})]);

    await expect(page.locator('#payloadsource_file_0')).toBeChecked();

    await page.locator('#payload_0').fill('{"a":1}');
    await page.locator('#payload_0').blur();

    await expect.poll(() => posted.length).toBe(1);
    expect(JSON.parse(posted[0])[0].usePayloadFromFile).toBe(true);
  });

  test('should default new mocks to payload', async ({ mockedPage: page }) => {
    const posted = await serveMocks(page, []);

    await page.locator('button[onclick="addMock()"]').click();

    await expect.poll(() => posted.length).toBe(1);
    expect(JSON.parse(posted[0])[0].usePayloadFromFile).toBe(false);
  });
});
