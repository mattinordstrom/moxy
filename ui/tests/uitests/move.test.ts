import { test, expect } from '../testutils';
import type { Page } from '@playwright/test';

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

async function routeMocks(page: Page, comments: string[]): Promise<string[][]> {
  let disk = JSON.stringify(comments.map(mock));
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
  await expect(page.locator('#comment_mock_' + (comments.length - 1))).toHaveValue(comments[comments.length - 1]);

  return saved;
}

const shownOrder = (page: Page, count: number) =>
  Promise.all([...Array(count).keys()].map((i) => page.locator('#comment_mock_' + i).inputValue()));

test.describe('move tests', () => {
  test('should move twice when up is clicked twice quickly', async ({ mockedPage: page }) => {
    const saved = await routeMocks(page, ['a', 'b', 'c', 'd']);

    // Both clicks hit the same button before the first move re-renders the list, like a real double click
    await page.locator('#movemock_up_btn_3').evaluate((button: HTMLElement) => {
      button.click();
      button.click();
    });

    await expect.poll(() => saved.at(-1)).toEqual(['a', 'd', 'b', 'c']);
    await expect.poll(() => shownOrder(page, 4)).toEqual(['a', 'd', 'b', 'c']);
    expect(saved).toHaveLength(2);
  });

  test('should apply clicks made before the list re-rendered to the items that were clicked', async ({ mockedPage: page }) => {
    const saved = await routeMocks(page, ['a', 'b', 'c', 'd']);

    // first on c, down on a, last on b, all on the list as it looked before any move finished
    await page.evaluate(() => {
      const buttons = ['movemock_first_btn_2', 'movemock_down_btn_0', 'movemock_last_btn_1'].map((id) => document.getElementById(id));
      buttons.forEach((button) => moveMock(button));
    });

    await expect.poll(() => saved.length).toBe(3);
    expect(saved).toEqual([['c', 'a', 'b', 'd'], ['c', 'b', 'a', 'd'], ['c', 'a', 'd', 'b']]);
    await expect.poll(() => shownOrder(page, 4)).toEqual(['c', 'a', 'd', 'b']);
  });
});
