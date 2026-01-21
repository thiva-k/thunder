/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const platform = process.platform;
const arch = process.arch;

// Node.js SEA sentinel fuse - standard value that must remain consistent
const SEA_SENTINEL_FUSE = "NODE_SEA_FUSE_fce680ab2cc467b6e072b8b5df1996b2";

const outputDir = path.join(__dirname, "..", "executables");
const distDir = path.join(__dirname, "..", "dist");
const blobPath = path.join(distDir, "sea-prep.blob");

// Verify the SEA blob exists before proceeding
if (!fs.existsSync(blobPath)) {
  console.error("SEA blob not found at:", blobPath);
  console.error('Run "npm run build:blob" first to generate the blob file.');
  process.exit(1);
}

// Check for codesign on macOS (required for SEA binaries to run)
if (platform === "darwin") {
  try {
    execSync("which codesign", { stdio: "ignore" });
  } catch {
    console.error(
      "codesign not found. Please install Xcode Command Line Tools:"
    );
    console.error("  xcode-select --install");
    process.exit(1);
  }
}

// Ensure output directory exists
if (!fs.existsSync(outputDir)) {
  fs.mkdirSync(outputDir, { recursive: true });
}

// Determine executable name based on platform (matching pkg naming convention)
let execName;
let osName;
let archName = arch === "x64" ? "x64" : arch; // Keep arm64 as-is

if (platform === "win32") {
  osName = "win";
  execName = "server-" + osName + "-" + archName + ".exe";
} else if (platform === "darwin") {
  osName = "macos";
  execName = "server-" + osName + "-" + archName;
} else {
  osName = "linux";
  execName = "server-" + osName + "-" + archName;
}

const outputPath = path.join(outputDir, execName);

// Copy the Node.js binary
const nodePath = process.execPath;
console.log("Copying Node.js binary from:", nodePath);
fs.copyFileSync(nodePath, outputPath);

// Make it executable on Unix systems
if (platform !== "win32") {
  fs.chmodSync(outputPath, 0o755);
}

// Inject the SEA blob into the binary
console.log("Injecting SEA blob into executable...");

try {
  if (platform === "darwin") {
    // macOS: Use postject with codesign removal
    execSync(
      `npx postject "${outputPath}" NODE_SEA_BLOB "${blobPath}" --sentinel-fuse ${SEA_SENTINEL_FUSE} --macho-segment-name NODE_SEA`,
      { stdio: "inherit" }
    );
    // Re-sign the binary (ad-hoc signing)
    execSync(`codesign --sign - "${outputPath}"`, { stdio: "inherit" });
  } else if (platform === "win32") {
    // Windows: Use postject with PE resource
    execSync(
      `npx postject "${outputPath}" NODE_SEA_BLOB "${blobPath}" --sentinel-fuse ${SEA_SENTINEL_FUSE}`,
      { stdio: "inherit" }
    );
  } else {
    // Linux: Use postject
    execSync(
      `npx postject "${outputPath}" NODE_SEA_BLOB "${blobPath}" --sentinel-fuse ${SEA_SENTINEL_FUSE}`,
      { stdio: "inherit" }
    );
  }
} catch (error) {
  console.error("Failed to build SEA executable:", error.message);
  // Clean up partial build
  if (fs.existsSync(outputPath)) {
    fs.unlinkSync(outputPath);
  }
  process.exit(1);
}

console.log("Successfully built executable:", outputPath);
