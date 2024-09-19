const { setupPreConditions } = require('.././testutils');
const assert = require('assert'); 

describe('payloadfromfile tests', function() {
  this.timeout(5000);
  it('Viewing default file path in modal', async function() {
    const {page, browser} = await setupPreConditions();
    
    try {
      const buttonSelector = '.payload-files-btn';
      await page.waitForSelector(buttonSelector);
      await page.click(buttonSelector);

      const elementSelector = '#payloadFiles';
      await page.waitForSelector(elementSelector, { visible: true });
      const textContent = await page.$eval(elementSelector, el => el.textContent);

      assert.strictEqual(textContent.includes('/home/testuser/projects/moxyarchive/payloads/'), true, 'file path not visible');
    } finally {
      await browser.close();
    }
  });

 
  it('Viewing default file path in modal when clicking pen icon', async function() {
    const {page, browser} = await setupPreConditions();

    try {
      const buttonSelector = '#mock_file_edit_0';
      await page.waitForSelector(buttonSelector);
      await page.click(buttonSelector);

      const elementSelector = '#payloadFiles';
      await page.waitForSelector(elementSelector, { visible: true });
      const textContent = await page.$eval(elementSelector, el => el.textContent);

      assert.strictEqual(textContent.includes('/home/testuser/projects/moxyarchive/payloads/'), true, 'file path not visible');
    } finally {
      await browser.close();
    }
  });
});