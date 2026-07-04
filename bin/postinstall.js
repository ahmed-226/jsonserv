#!/usr/bin/env node
const { chmodSync, existsSync } = require('fs');
const path = require('path');
const os = require('os');

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

if (!binaryName) return;

const binaryPath = path.join(__dirname, '..', 'binary', binaryName);

if (existsSync(binaryPath) && platform !== 'win32') {
  try {
    chmodSync(binaryPath, 0o755);
  } catch {
    // skip if we can't chmod (e.g. read-only filesystem)
  }
}
