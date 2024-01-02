const path = require('path');
const { setupPreConditions, assertCondition } = require('./testutils');

(async () => {
  await (async () => {
    const scenario = "Remove mock entity";
    const {page, browser} = await setupPreConditions();
    //////////////////////////////////////////////////////////////////////////////////

    await page.waitForSelector('#comment_mock_0', { visible: true });

    // Click the remove btn
    const buttonSelector = '#x_btn_0';
    await page.waitForSelector(buttonSelector);
    await page.click(buttonSelector);

    await page.waitForSelector('#comment_mock_0', { hidden: true });
    const element = await page.$('#comment_mock_0');

    //////////////////////////////////////////////////////////////////////////////////
    assertCondition(!element, path.basename(__filename), scenario, browser);
  })();

})();