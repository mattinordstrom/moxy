import { test, expect } from '../testutils';

test.describe('header tests', () => {
  test('should show the version from settings', async ({ mockedPage: page }) => {
    await expect(page.locator('#header-version')).toHaveText('v2.0.0');
  });
});
