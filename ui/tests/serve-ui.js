#!/usr/bin/env node
/**
 * Minimal static file server for the moxy Admin UI.
 * Serves the moxy project root so that paths like /ui/static/main.js
 * (as referenced in index.html) resolve correctly.
 *
 * Usage: node serve-ui.js [port]
 * Default port: 9099 (avoids conflict with moxy's default 9097)
 */
const http = require('http');
const fs = require('fs');
const path = require('path');

const PORT = parseInt(process.argv[2] || '9099', 10);
// Serve from the project root (two levels above ui/tests/)
const ROOT_DIR = path.resolve(__dirname, '../..');

const MIME_TYPES = {
  '.html': 'text/html',
  '.js':   'application/javascript',
  '.css':  'text/css',
  '.json': 'application/json',
  '.png':  'image/png',
  '.ico':  'image/x-icon',
};

const server = http.createServer((req, res) => {
  // Strip query string for file lookup
  const urlPath = req.url.split('?')[0];
  const filePath = path.join(ROOT_DIR, urlPath === '/' ? 'ui/index.html' : urlPath);

  // Security: prevent directory traversal outside ROOT_DIR
  if (!filePath.startsWith(ROOT_DIR)) {
    res.writeHead(403);
    res.end('Forbidden');
    return;
  }

  fs.readFile(filePath, (err, data) => {
    if (err) {
      // Don't log 404s for moxyadminui paths — those are intercepted by Playwright routes
      if (!urlPath.includes('moxyadminui')) {
        console.error('404:', urlPath);
      }
      res.writeHead(404);
      res.end('Not found: ' + urlPath);
      return;
    }

    const ext = path.extname(filePath);
    const contentType = MIME_TYPES[ext] || 'application/octet-stream';
    res.writeHead(200, { 'Content-Type': contentType });
    res.end(data);
  });
});

server.listen(PORT, '127.0.0.1', () => {
  // Print to stdout so Playwright's webServer knows the server is ready
  console.log(`UI server listening on http://127.0.0.1:${PORT}`);
});
