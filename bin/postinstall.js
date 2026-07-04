#!/usr/bin/env node
const { chmodSync, readdirSync } = require('fs');
const path = require('path');
const os = require('os');

if (os.platform() === 'win32') return;

const binaryDir = path.join(__dirname, '..', 'binary');

try {
  for (const file of readdirSync(binaryDir)) {
    chmodSync(path.join(binaryDir, file), 0o755);
  }
} catch {
  // skip if we can't chmod (e.g. read-only filesystem)
}
