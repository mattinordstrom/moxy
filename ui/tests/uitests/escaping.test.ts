import { test, expect } from '../testutils';

const XSS = '<img src=x onerror="window.__xss = true">';

test.describe('escaping tests', () => {
  test('should show log messages as text instead of running them as HTML', async ({ mockedPage: page }) => {
    await page.routeWebSocket(/\/moxyws$/, (ws) => {
      const timestamp = '2026-11-26 10:00:00';
      ws.send(JSON.stringify({ type: 'wsmock', message: XSS, timestamp }));
      ws.send(JSON.stringify({ type: 'error', message: XSS, timestamp }));
      ws.send(JSON.stringify({ type: 'mock', message: XSS, timestamp }));
      ws.send(JSON.stringify({
        type: 'proxy',
        message: '/api/x  ➔  http://localhost:8089/api/x',
        timestamp,
        extras: { httpMethod: 'POST', headers: 'X-Test: ' + XSS, body: XSS },
      }));
    });

    // WebSocket routes only apply to pages loaded after they are set up
    await page.reload();
    await page.evaluate('wsSetup()');

    const log = page.locator('#footer-log');
    await expect(log.locator('.log-entry')).toHaveCount(4);
    await expect(log).toContainText(XSS);
    await expect(log.locator('img')).toHaveCount(0);
    expect(await page.evaluate('window.__xss')).toBeUndefined();
  });

  test('should keep quotes and markup in def values intact', async ({ mockedPage: page }) => {
    const comment = '"quoted" </textarea><b>bold</b>';

    await page.route('**/moxyadminui/mockdef', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([{
          active: true, comment, method: 'GET', urlpart: '/api/"x"', statuscode: 200,
          freezetimems: 0, payload: { html: '</textarea><b>x</b>' }, payloadFromFile: '',
        }]),
      });
    });

    await page.evaluate("reloadDefsFromDisk('mockdef.json')");

    await expect(page.locator('#comment_mock_0')).toHaveValue(comment);
    await expect(page.locator('#urlpart_mock_0')).toHaveValue('/api/"x"');
    await expect(page.locator('#payload_0')).toHaveValue(/<\/textarea><b>x<\/b>/);
    await expect(page.locator('#mock-content-container b', { hasText: 'bold' })).toHaveCount(0);
  });
});
