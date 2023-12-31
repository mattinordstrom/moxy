const path = require('path');
const { setupPreConditions } = require('./testutils');

(async () => {
  await (async () => {
    const scenario = "Remove mock entity";
    console.log("## " + path.basename(__filename) + ": " + scenario + " ##");
    const {page, browser} = await setupPreConditions();

    ////////////////////////////////////
    // TEST START

    const elementSelector = '#mock-content-container';

    await page.waitForSelector('#comment_mock_0', { visible: true });

    // Click the remove btn
    const buttonSelector = '#x_btn_0';
    await page.waitForSelector(buttonSelector);
    await page.click(buttonSelector);

    await page.waitForSelector('#comment_mock_0', { hidden: true });
    const element = await page.$('#comment_mock_0');
    if (!element) {
      console.log('Test passed\n');
    } else {
      console.log('Test failed\n');
      throw new Error('Test failed! ' + path.basename(__filename) + ": " + scenario);
    }

    await browser.close();
  })();

})();