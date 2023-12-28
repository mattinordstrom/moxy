const puppeteer = require('puppeteer');
const fs = require('fs');
const path = require('path');

describe('Payload from file test', () => {
  it('path visible in modal', async () => {
  //const browser = await puppeteer.launch({headless: false});
  const browser = await puppeteer.launch({headless: "new"});
  const page = await browser.newPage();

  // Enable request interception
  await page.setRequestInterception(true);

  // Add event listener to intercept requests
  page.on('request', (interceptedRequest) => {
    const requestUrl = interceptedRequest.url();
    const filename = requestUrl.split('/').pop();

    if (requestUrl.endsWith('.css')) {
        const cssPath = path.resolve(__dirname, '../static/style/'+filename);
        const cssContent = fs.readFileSync(cssPath, 'utf8');

        interceptedRequest.respond({ contentType: 'text/css', body: cssContent });
    } else if (requestUrl.endsWith('.js')) {
        const jsPath = path.resolve(__dirname, '../static/'+filename);
        const jsContent = fs.readFileSync(jsPath, 'utf8');

        interceptedRequest.respond({ contentType: 'text/javascript', body: jsContent });
    } else if (requestUrl.endsWith('moxyadminui/mockdef')) {
      interceptedRequest.respond({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
            {
              "active": true,
              "comment": "test123",
              "freezetimems": 0,
              "method": "GET",
              "payload": {
                "test": 123
              },
              "payloadFromFile": "",
              "statuscode": 200,
              "urlpart": "/api/sometest"
            }
          ]),
      });
    } else if (requestUrl.endsWith('moxyadminui/proxydef')) {
        interceptedRequest.respond({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([]),
        });
    } else if (requestUrl.endsWith('moxyadminui/settings')) {
        interceptedRequest.respond({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            "port": 9097,
            "defaultRoute": "http://localhost:8089",
            "payloadFiles": [],
            "payloadPath": "/home/testuser/projects/moxyarchive/payloads/"
            }),
        });
    } else {
      interceptedRequest.continue();
    }
  });

  const localHtmlPath = path.resolve(__dirname, '../index.html?mode=test');
  await page.goto(`file://${localHtmlPath}`);

  //await new Promise(resolve => setTimeout(resolve, 15000));

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

  if (textContent.includes('/home/testuser/projects/moxxyarchive/payloads/')) {
    console.log('Test passed: Element is visible and contains the correct text');
  } else {
    console.log('Test failed: Element does not contain the correct text');
    throw new Error('Element not found!!!');
  }

  await browser.close();
});
});
