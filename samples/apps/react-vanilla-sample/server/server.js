// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

const express = require("express");
const path = require("path");
const https = require("https");
const fs = require("fs");

const app = express();
const PORT = process.env.PORT || 3000;

// Use actual working directory to access certs
const isRunningAsExecutable = process.pkg !== undefined;
const certDir = isRunningAsExecutable
  ? path.dirname(process.execPath)
  : path.join(process.cwd());

const keyPath = path.join(certDir, "server.key");
const certPath = path.join(certDir, "server.cert");

// Serve static files from the 'dist' directory
const appDir = path.join(certDir, "app");
app.use(express.static(appDir));

// Handle SPA routing
app.get("*", (req, res) => {
  res.sendFile(path.join(appDir, "index.html"));
});

if (fs.existsSync(keyPath) && fs.existsSync(certPath)) {
  const sslOptions = {
    key: fs.readFileSync(keyPath),
    cert: fs.readFileSync(certPath),
  };

  https.createServer(sslOptions, app).listen(PORT, () => {
    console.log(`✅ HTTPS server running at https://localhost:${PORT}`);
    console.log("Press Ctrl+C to stop the server.");
  });
} else {
  app.listen(PORT, () => {
    console.log(
      `⚠️  HTTPS certs missing. Falling back to HTTP at http://localhost:${PORT}`
    );
    console.log("Press Ctrl+C to stop the server.");
  });
}
