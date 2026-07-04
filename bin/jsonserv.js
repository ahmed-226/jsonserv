#!/usr/bin/env node
const { spawn } = require('child_process');
const path = require('path');
const os = require('os');
const fs = require('fs');

const platformMap = {
  darwin: 'darwin',
  linux:  'linux',
  win32:  'windows',
};

const archMap = {
  x64:   'amd64',
  arm64: 'arm64',
};

const platform = platformMap[os.platform()];
const arch = archMap[os.arch()];

if (!platform || !arch) {
  console.error(`[jsonserv] unsupported platform: ${os.platform()} ${os.arch()}`);
  console.error(`[jsonserv] supported: linux x64/arm64, darwin x64/arm64, win32 x64/arm64`);
  process.exit(1);
}

const exe = platform === 'windows' ? '.exe' : '';
const binaryName = `jsonserv-${platform}-${arch}${exe}`;
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

child.on('close', (code) => {
  process.exit(code ?? 0);
});

child.on('error', (err) => {
  console.error(`[jsonserv] failed to start: ${err.message}`);
  process.exit(1);
});
