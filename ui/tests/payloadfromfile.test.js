const path = require('path');
const { setupPreConditions } = require('./testutils');

(async () => {
  await (async () => {
    const scenario = "Viewing default file path in modal";
    console.log("## " + path.basename(__filename) + ": " + scenario + " ##");
    const {page, browser} = await setupPreConditions();

    ////////////////////////////////////
    // TEST START

    // Click the payload files button
    const buttonSelector = '.payload-files-btn';
    await page.waitForSelector(buttonSelector);
    await page.click(buttonSelector);

    // Assert dialog is visible and that it contains the path string
    const elementSelector = '#payloadFiles';
    await page.waitForSelector(elementSelector, { visible: true });
    const textContent = await page.$eval(elementSelector, el => el.textContent);

    if (textContent.includes('/home/testuser/projects/moxyarchive/payloads/')) {
      console.log('Test passed\n');
    } else {
      console.log('Test failed\n');
      throw new Error('Test failed! ' + path.basename(__filename) + ": " + scenario);
    }

    await browser.close();
  })();

  await (async () => {
    const scenario = "Viewing default file path in modal when clicking pen icon";
    console.log("## " + path.basename(__filename) + ": " + scenario + " ##");
    const {page, browser} = await setupPreConditions();

    ////////////////////////////////////
    // TEST START

    // Click the payload files button (pen icon on first mock)
    const buttonSelector = '#mock_file_edit_0';
    await page.waitForSelector(buttonSelector);
    await page.click(buttonSelector);

    // Assert dialog is visible and that it contains the path string
    const elementSelector = '#payloadFiles';
    await page.waitForSelector(elementSelector, { visible: true });
    const textContent = await page.$eval(elementSelector, el => el.textContent);

    if (textContent.includes('/home/testuser/projects/moxyarchive/payloads/')) {
      console.log('Test passed\n');
    } else {
      console.log('Test failed\n');
      throw new Error('Test failed! ' + path.basename(__filename) + ": " + scenario);
    }

    await browser.close();
  })();
})();