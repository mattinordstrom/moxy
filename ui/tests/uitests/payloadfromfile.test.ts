import { test, expect } from '../testutils';

test.describe('payloadfromfile tests', () => {
  test('Viewing default file path in modal', async ({ mockedPage: page }) => {
    // Two elements share the .payload-files-btn class; target the specific button by its accessible name.
    await page.getByRole('button', { name: 'Payload files' }).click();
    await expect(page.locator('#payloadFiles')).toContainText(
      '/home/testuser/projects/moxyarchive/payloads/'
    );
  });

  test('Viewing default file path in modal when clicking pen icon', async ({ mockedPage: page }) => {
    await page.locator('#mock_file_edit_0').click();
    await expect(page.locator('#payloadFiles')).toContainText(
      '/home/testuser/projects/moxyarchive/payloads/'
    );
  });
});
