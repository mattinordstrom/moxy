const puppeteer = require('puppeteer');
const fs = require('fs');
const path = require('path');

const handleJsAndCssRequests = (interceptedRequest) => {
    const requestUrl = interceptedRequest.url();
    const filename = requestUrl.split('/').pop();

    if (requestUrl.endsWith('.css')) {
        const cssPath = path.resolve(__dirname, '../static/style/'+filename);
        const cssContent = fs.readFileSync(cssPath, 'utf8');

        interceptedRequest.respond({ contentType: 'text/css', body: cssContent });
    
        return true;
    } else if (requestUrl.endsWith('.js')) {
        const jsPath = path.resolve(__dirname, '../static/'+filename);
        const jsContent = fs.readFileSync(jsPath, 'utf8');

        interceptedRequest.respond({ contentType: 'text/javascript', body: jsContent });
    
        return true;
    }

    return false;
}
  
const handleHttpRequestsMock = (interceptedRequest) => {
    const requestUrl = interceptedRequest.url();

    if (requestUrl.endsWith('moxyadminui/mockdef')) {
        interceptedRequest.respond({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([
              {
                "active": true,
                "comment": "test123 mock",
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

        return true;
      } else if (requestUrl.endsWith('moxyadminui/proxydef')) {
          interceptedRequest.respond({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify([]),
          });

          return true;
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

          return true;
      }

      return false;
}

const setupRequestHandling = async (page) => {
    await page.setRequestInterception(true);

    page.on('request', (interceptedRequest) => {
        if (handleJsAndCssRequests(interceptedRequest)) {
            return;
        }

        if (handleHttpRequestsMock(interceptedRequest)) {
            return;
        }

        interceptedRequest.continue();
    });
}

const setupPreConditions = async () => {
    //const browser = await puppeteer.launch({headless: false, args: ['--no-sandbox', '--disable-setuid-sandbox']});
    const browser = await puppeteer.launch({headless: "new", args: ['--no-sandbox', '--disable-setuid-sandbox']});

    const page = await browser.newPage();
    setupRequestHandling(page);

    const localHtmlPath = path.resolve(__dirname, '../index.html?mode=test');
    
    await page.goto(`file://${localHtmlPath}`);

    //await new Promise(resolve => setTimeout(resolve, 15000));
    return {page, browser};
}


module.exports = { setupPreConditions };
  