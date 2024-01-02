const path = require('path');
const { setupPreConditions, assertCondition } = require('./testutils');

(async () => {
  await (async () => {
    const scenario = "Viewing default file path in modal";
    const {page, browser} = await setupPreConditions();
    //////////////////////////////////////////////////////////////////////////////////

    // Click the payload files button
    const buttonSelector = '.payload-files-btn';
    await page.waitForSelector(buttonSelector);
    await page.click(buttonSelector);

    // Assert dialog is visible and that it contains the path string
    const elementSelector = '#payloadFiles';
    await page.waitForSelector(elementSelector, { visible: true });
    const textContent = await page.$eval(elementSelector, el => el.textContent);

    //////////////////////////////////////////////////////////////////////////////////
    assertCondition(textContent.includes('/home/testuser/projects/moxyarchive/payloads/'), 
      path.basename(__filename), 
      scenario,
      browser);
  })();

  await (async () => {
    const scenario = "Viewing default file path in modal when clicking pen icon";
    const {page, browser} = await setupPreConditions();
    //////////////////////////////////////////////////////////////////////////////////

    // Click the payload files button (pen icon on first mock)
    const buttonSelector = '#mock_file_edit_0';
    await page.waitForSelector(buttonSelector);
    await page.click(buttonSelector);

    // Assert dialog is visible and that it contains the path string
    const elementSelector = '#payloadFiles';
    await page.waitForSelector(elementSelector, { visible: true });
    const textContent = await page.$eval(elementSelector, el => el.textContent);

    //////////////////////////////////////////////////////////////////////////////////
    assertCondition(textContent.includes('/home/testuser/projects/moxyarchive/payloads/'), 
      path.basename(__filename), 
      scenario,
      browser);
  })();

})();