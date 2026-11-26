import { test, expect } from '../testutils';
import type { Page } from '@playwright/test';

const mock = (comment: string, payload: unknown) => ({
  active: true,
  comment,
  method: 'GET',
  urlpart: '/api/' + comment,
  statuscode: 200,
  freezetimems: 0,
  payload,
  payloadFromFile: '',
});

type MockDefRoute = { posted: string[]; setDisk: (text: string) => void };

async function routeMockDef(page: Page, diskText: string, postStatus = 200): Promise<MockDefRoute> {
  const state = { disk: diskText, posted: [] as string[] };

  await page.route('**/moxyadminui/mockdef', async (route) => {
    if (route.request().method() === 'POST') {
      state.posted.push(route.request().postData() ?? '');
      if (postStatus === 200) {
        state.disk = route.request().postData() ?? '';
      }
      await route.fulfill({ status: postStatus, body: postStatus === 200 ? '' : 'Forbidden' });
      return;
    }
    await route.fulfill({ status: 200, contentType: 'application/json', body: state.disk });
  });

  return { posted: state.posted, setDisk: (text) => { state.disk = text; } };
}

test.describe('data loss tests', () => {
  test('should load mocks whose payload is true and keep them on add', async ({ mockedPage: page }) => {
    const defs = await routeMockDef(page, JSON.stringify([mock('first', true), mock('second', { a: 1 })]));
    await page.reload();

    await expect(page.locator('#payload_0')).toHaveValue('true');
    await expect(page.locator('#comment_mock_1')).toHaveValue('second');

    await page.evaluate('addMock()');

    await expect.poll(() => defs.posted.length).toBe(1);
    const saved = JSON.parse(defs.posted[0]);
    expect(saved).toHaveLength(3);
    expect(saved[1].payload).toBe(true);
  });

  test('should keep big ints, number formatting and key order when saving', async ({ mockedPage: page }) => {
    const payload = '{"b": 1, "2": 2, "big": 12345678901234567890, "f": 1.10}';
    const text = '[' + JSON.stringify(mock('first', null)).replace('"payload":null', '"payload":' + payload) + ']';
    const defs = await routeMockDef(page, text);
    await page.reload();

    await expect(page.locator('#payload_0')).toHaveValue('{\n  "b": 1,\n  "2": 2,\n  "big": 12345678901234567890,\n  "f": 1.10\n}');

    await page.locator('#comment_mock_0').fill('renamed');
    await page.locator('#comment_mock_0').blur();

    await expect.poll(() => defs.posted.length).toBe(1);
    expect(defs.posted[0]).toContain('"payload":{"b":1,"2":2,"big":12345678901234567890,"f":1.10}');
  });

  test('should report a failed save in the log', async ({ mockedPage: page }) => {
    await routeMockDef(page, JSON.stringify([mock('first', { a: 1 })]), 403);
    await page.reload();

    await page.locator('#active_mock_0').click();

    await expect(page.locator('#footer-log')).toContainText('Saving mockdef.json failed (403): Forbidden');
  });

  test('should not save when mockdef.json could not be loaded', async ({ mockedPage: page }) => {
    const defs = await routeMockDef(page, '[{"broken"');
    await page.reload();

    await expect(page.locator('#footer-log')).toContainText('Could not load mockdef.json');

    await page.evaluate('addMock()');

    await expect(page.locator('#footer-log')).toContainText('Not saved, could not read mockdef.json');
    expect(defs.posted).toHaveLength(0);
  });

  test('should apply an edit on top of a change made on disk while typing', async ({ mockedPage: page }) => {
    const defs = await routeMockDef(page, JSON.stringify([mock('first', { a: 1 }), mock('second', { b: 2 })]));
    await page.reload();
    await expect(page.locator('#comment_mock_1')).toHaveValue('second');

    await page.locator('#urlpart_mock_0').fill('/api/typed');
    defs.setDisk(JSON.stringify([mock('first', { a: 1 }), mock('changed on disk', { b: 2 })]));
    await page.evaluate("reloadDefsFromDisk('mockdef.json')");
    await page.locator('#urlpart_mock_0').blur();

    await expect.poll(() => defs.posted.length).toBe(1);
    const saved = JSON.parse(defs.posted[0]);
    expect(saved[0].urlpart).toBe('/api/typed');
    expect(saved[1].comment).toBe('changed on disk');
  });

  test('should reload instead of removing when the file changed on disk', async ({ mockedPage: page }) => {
    const defs = await routeMockDef(page, JSON.stringify([mock('first', { a: 1 })]));
    await page.reload();
    await expect(page.locator('#comment_mock_0')).toHaveValue('first');

    defs.setDisk(JSON.stringify([mock('added elsewhere', { a: 1 }), mock('first', { a: 1 })]));
    await page.locator('#x_btn_0').first().click();

    await expect(page.locator('#footer-log')).toContainText('mockdef.json changed on disk — reloaded, please try again');
    await expect(page.locator('#comment_mock_0')).toHaveValue('added elsewhere');
    expect(defs.posted).toHaveLength(0);
  });

  test('should finish loading with show only mocks on and a single mock', async ({ mockedPage: page }) => {
    await page.evaluate("localStorage.setItem('moxyShowOnlyMocks', 'true')");
    await page.reload();

    await expect(page.locator('#comment_mock_0')).toHaveValue('test123 mock');
    await expect(page.locator('#cover')).toBeHidden();
    await expect(page.locator('.right')).toBeHidden();
  });

  test('should keep payload file numbers intact when editing a file', async ({ mockedPage: page }) => {
    const posted: string[] = [];
    await page.route('**/moxyadminui/editpayloadfile**', async (route) => {
      if (route.request().method() === 'POST') {
        posted.push(route.request().postData() ?? '');
        await route.fulfill({ status: 200, body: '' });
        return;
      }
      await route.fulfill({ status: 200, contentType: 'application/json', body: '{"id": 12345678901234567890, "f": 1.10}' });
    });

    await page.evaluate("PayloadFromFileModule.setPayloadFiles(['example.json'])");
    await page.evaluate("editFile(null, '/tmp/example.json')");
    await expect(page.locator('#payloadedit')).toHaveValue('{\n  "id": 12345678901234567890,\n  "f": 1.10\n}');

    await page.evaluate("updatePayloadFile(document.getElementById('payloadedit'))");

    await expect.poll(() => posted.length).toBe(1);
    expect(posted[0]).toContain('"payload":{"id":12345678901234567890,"f":1.10}');
  });
});
