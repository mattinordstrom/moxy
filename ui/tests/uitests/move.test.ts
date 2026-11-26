import { test, expect } from '../testutils';

const mock = (comment: string) => ({
  active: true,
  comment,
  method: 'GET',
  urlpart: '/api/' + comment,
  statuscode: 200,
  freezetimems: 0,
  payload: { name: comment },
  payloadFromFile: '',
});

test.describe('move tests', () => {
  test('should not undo a move when the button is clicked twice quickly', async ({ mockedPage: page }) => {
    let disk = JSON.stringify(['a', 'b', 'c', 'd'].map(mock));
    const saved: string[][] = [];

    await page.route('**/moxyadminui/mockdef', async (route) => {
      if (route.request().method() === 'POST') {
        disk = route.request().postData() ?? '';
        saved.push(JSON.parse(disk).map((m: { comment: string }) => m.comment));
        await route.fulfill({ status: 200, body: '' });
        return;
      }
      await route.fulfill({ status: 200, contentType: 'application/json', body: disk });
    });
    await page.reload();
    await expect(page.locator('#comment_mock_3')).toHaveValue('d');

    // Both clicks hit the same button before the first move is saved, like a real double click
    await page.locator('#movemock_up_btn_3').evaluate((button: HTMLElement) => {
      button.click();
      button.click();
    });

    await expect.poll(() => saved.length).toBe(1);
    expect(saved[0]).toEqual(['a', 'b', 'd', 'c']);
    await expect(page.locator('#comment_mock_2')).toHaveValue('d');
  });
});
