#!/usr/bin/env node

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const REPO = 'delvop-dev/delvop';

const PLATFORM_MAP = { darwin: 'darwin', linux: 'linux' };
const ARCH_MAP = { x64: 'amd64', arm64: 'arm64' };

function main() {
  const pkg = require('./package.json');
  const version = pkg.version;
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!platform || !arch) {
    console.error(`Unsupported platform: ${process.platform}-${process.arch}`);
    process.exit(1);
  }

  const binDir = path.join(__dirname, 'bin');
  const binPath = path.join(binDir, 'delvop');

  if (fs.existsSync(binPath)) {
    return;
  }

  fs.mkdirSync(binDir, { recursive: true });

  const tarball = `delvop-${platform}-${arch}.tar.gz`;
  const url = `https://github.com/${REPO}/releases/download/v${version}/${tarball}`;

  console.log(`Downloading delvop v${version} for ${platform}-${arch}...`);

  try {
    execSync(`curl -fsSL "${url}" | tar xz -C "${binDir}"`, { stdio: 'inherit' });
    fs.chmodSync(binPath, 0o755);
    console.log('delvop installed successfully!');
  } catch (err) {
    console.error('Failed to install delvop.');
    console.error('Install manually: https://github.com/delvop-dev/delvop/releases');
    process.exit(1);
  }
}

main();
