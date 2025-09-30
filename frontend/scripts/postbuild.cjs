const fs = require('fs');
const path = require('path');

const distDir = path.join(__dirname, '..', 'dist');
const keepFile = path.join(distDir, '.gitkeep');

try {
  fs.mkdirSync(distDir, { recursive: true });
  if (!fs.existsSync(keepFile)) {
    fs.writeFileSync(keepFile, '');
  }
} catch (error) {
  console.error('Failed to ensure dist/.gitkeep:', error);
  process.exitCode = 1;
}
