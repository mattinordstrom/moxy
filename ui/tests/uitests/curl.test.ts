import { test, expect } from '../testutils';
import { execFileSync } from 'node:child_process';
import { existsSync, mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

test.describe('copy as curl tests', () => {
  test('should keep hostile headers and body literal when the copied curl is run', async ({ mockedPage: page }) => {
    const dir = mkdtempSync(join(tmpdir(), 'moxy-curl-'));
    const marker = join(dir, 'pwned');
    const evil = `'; touch ${marker}; echo '$(touch ${marker})\`touch ${marker}\``;
    const event = {
      type: 'proxy',
      message: '/api/x → http://localhost:8080/api/x',
      timestamp: 'now',
      extras: { httpMethod: 'POST', headers: `X-Evil: ${evil} | Content-Type: application/json`, body: `{"a":"${evil}"}` },
    };

    try {
      await page.evaluate((data) => showCurl(null, encodeURIComponent(data)), JSON.stringify(event));
      const copied = await page.evaluate(() => curlCopyText());

      const args = execFileSync('bash', ['-c', 'curl() { printf "%s\\n" "$@"; }\n' + copied]).toString().split('\n');

      expect(existsSync(marker)).toBe(false);
      expect(args).toContain(`X-Evil: ${evil}`);
      expect(args).toContain(`{"a":"${evil}"}`);
      expect(args).toContain('http://localhost:8080/api/x');
    } finally {
      rmSync(dir, { recursive: true, force: true });
    }
  });
});
