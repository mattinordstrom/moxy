const { setupPreConditions } = require('.././testutils');
const assert = require('assert'); 

describe('mockentity tests', function() {
  this.timeout(5000);
  it('should remove the mock entity', async function() {
    const { page, browser } = await setupPreConditions();

    try {
      await page.waitForSelector('#comment_mock_0', { visible: true });

      const buttonSelector = '#x_btn_0';
      await page.waitForSelector(buttonSelector);
      await page.click(buttonSelector);

      await page.waitForSelector('#comment_mock_0', { hidden: true });
      const element = await page.$('#comment_mock_0');

      assert.strictEqual(element, null, 'Mock entity was not removed');
    } finally {
      await browser.close();
    }
  });
});
