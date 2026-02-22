import { test, expect } from '../testutils';

test.describe('mockentity tests', () => {
  test('should remove the mock entity', async ({ mockedPage: page }) => {
    await expect(page.locator('#comment_mock_0')).toBeVisible();
    await page.locator('#x_btn_0').click();
    await expect(page.locator('#comment_mock_0')).toBeHidden();
  });
});
