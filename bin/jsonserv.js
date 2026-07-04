#!/usr/bin/env node
const { spawn } = require('child_process');
const path = require('path');
const os = require('os');
const fs = require('fs');

const platform = os.platform();
const arch = os.arch();

const binaryMap = {
  'linux-x64':   'jsonserv-linux-amd64',
  'linux-arm64': 'jsonserv-linux-arm64',
  'darwin-x64':  'jsonserv-darwin-amd64',
  'darwin-arm64':'jsonserv-darwin-arm64',
  'win32-x64':   'jsonserv-windows-amd64.exe',
  'win32-arm64': 'jsonserv-windows-arm64.exe',
};

const key = `${platform}-${arch}`;
const binaryName = binaryMap[key];

if (!binaryName) {
  console.error(`[jsonserv] unsupported platform: ${platform} ${arch}`);
  console.error(`[jsonserv] supported: linux x64/arm64, darwin x64/arm64, win32 x64/arm64`);
  process.exit(1);
}

const binaryPath = path.join(__dirname, '..', 'binary', binaryName);

if (!fs.existsSync(binaryPath)) {
  console.error(`[jsonserv] binary not found: ${binaryPath}`);
  console.error(`[jsonserv] run "npm run build" to compile binaries, or install the package with prebuilt binaries.`);
  process.exit(1);
}

const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  windowsHide: true,
});

child.on('exit', (code) => {
  process.exit(code ?? 0);
});

child.on('error', (err) => {
  console.error(`[jsonserv] failed to start: ${err.message}`);
  process.exit(1);
});
