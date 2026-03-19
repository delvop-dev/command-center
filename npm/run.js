#!/usr/bin/env node

const { execFileSync } = require('child_process');
const path = require('path');

const binPath = path.join(__dirname, 'bin', 'delvop');

try {
  execFileSync(binPath, process.argv.slice(2), { stdio: 'inherit' });
} catch (err) {
  if (err.status !== undefined) {
    process.exit(err.status);
  }
  console.error('delvop binary not found. Try reinstalling: npm install -g @delvop/cli');
  process.exit(1);
}
