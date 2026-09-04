const sharp = require('sharp');
const path = require('path');
const outDir = path.join(__dirname);

async function createGradient(filename, color1, color2, w = 1440, h = 810) {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${w}" height="${h}">
    <defs>
      <linearGradient id="g" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" style="stop-color:${color1}"/>
        <stop offset="100%" style="stop-color:${color2}"/>
      </linearGradient>
    </defs>
    <rect width="100%" height="100%" fill="url(#g)"/>
  </svg>`;
  await sharp(Buffer.from(svg)).png().toFile(path.join(outDir, filename));
  console.log('Created', filename);
}

async function createIconCircle(filename, color, size = 256) {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}">
    <circle cx="${size/2}" cy="${size/2}" r="${size/2 - 4}" fill="${color}"/>
  </svg>`;
  await sharp(Buffer.from(svg)).png().toFile(path.join(outDir, filename));
  console.log('Created', filename);
}

async function createDollarIcon(filename, size = 256) {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24">
    <circle cx="12" cy="12" r="12" fill="#F59E0B"/>
    <text x="12" y="17" text-anchor="middle" font-family="Arial" font-size="16" font-weight="bold" fill="white">$</text>
  </svg>`;
  await sharp(Buffer.from(svg)).png().toFile(path.join(outDir, filename));
  console.log('Created', filename);
}

async function createQuizIcon(filename, size = 256) {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24">
    <circle cx="12" cy="12" r="12" fill="#6366F1"/>
    <text x="12" y="17" text-anchor="middle" font-family="Arial" font-size="16" font-weight="bold" fill="white">?</text>
  </svg>`;
  await sharp(Buffer.from(svg)).png().toFile(path.join(outDir, filename));
  console.log('Created', filename);
}

async function createLightningIcon(filename, size = 256) {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24">
    <circle cx="12" cy="12" r="12" fill="#10B981"/>
    <text x="12" y="17" text-anchor="middle" font-family="Arial" font-size="16" font-weight="bold" fill="white">!</text>
  </svg>`;
  await sharp(Buffer.from(svg)).png().toFile(path.join(outDir, filename));
  console.log('Created', filename);
}

async function createNigeriaIcon(filename, size = 256) {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24">
    <circle cx="12" cy="12" r="12" fill="#008751"/>
    <text x="12" y="17" text-anchor="middle" font-family="Arial" font-size="12" font-weight="bold" fill="white">NG</text>
  </svg>`;
  await sharp(Buffer.from(svg)).png().toFile(path.join(outDir, filename));
  console.log('Created', filename);
}

async function createCheckIcon(filename, size = 256) {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24">
    <circle cx="12" cy="12" r="12" fill="#10B981"/>
    <text x="12" y="17" text-anchor="middle" font-family="Arial" font-size="18" font-weight="bold" fill="white">V</text>
  </svg>`;
  await sharp(Buffer.from(svg)).png().toFile(path.join(outDir, filename));
  console.log('Created', filename);
}

async function createKweeksLogo(filename, size = 256) {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24">
    <circle cx="12" cy="12" r="12" fill="#F59E0B"/>
    <text x="12" y="16" text-anchor="middle" font-family="Arial" font-size="11" font-weight="bold" fill="#1F2937">K</text>
  </svg>`;
  await sharp(Buffer.from(svg)).png().toFile(path.join(outDir, filename));
  console.log('Created', filename);
}

(async () => {
  // Title slide bg - dark gold to dark
  await createGradient('bg-title.png', '#1F2937', '#111827');
  // Problem slide bg
  await createGradient('bg-problem.png', '#991B1B', '#7F1D1D');
  // Solution slide bg
  await createGradient('bg-solution.png', '#1E3A5F', '#0F172A');
  // Market slide bg
  await createGradient('bg-market.png', '#1E293B', '#0F172A');
  // Competitor slide bg
  await createGradient('bg-competitor.png', '#312E81', '#1E1B4B');
  // Advantage slide bg
  await createGradient('bg-advantage.png', '#064E3B', '#022C22');
  // Demo slide bg
  await createGradient('bg-demo.png', '#1E3A5F', '#0F172A');
  // Tech slide bg
  await createGradient('bg-tech.png', '#1F2937', '#111827');
  // Vision slide bg
  await createGradient('bg-vision.png', '#4C1D95', '#2E1065');
  // Thank you slide bg
  await createGradient('bg-thankyou.png', '#F59E0B', '#D97706');
  // Stats bar bg
  await createGradient('bg-statsbar.png', '#F59E0B', '#D97706');
  // Dark card bg
  await sharp(Buffer.from(`<svg xmlns="http://www.w3.org/2000/svg" width="600" height="300">
    <rect width="600" height="300" rx="12" fill="#1F2937"/>
  </svg>`)).png().toFile(path.join(outDir, 'card-dark.png'));
  console.log('Created card-dark.png');
  
  // Light card bg
  await sharp(Buffer.from(`<svg xmlns="http://www.w3.org/2000/svg" width="600" height="300">
    <rect width="600" height="300" rx="12" fill="#F9FAFB"/>
  </svg>`)).png().toFile(path.join(outDir, 'card-light.png'));
  console.log('Created card-light.png');

  // Yellow accent bar
  await sharp(Buffer.from(`<svg xmlns="http://www.w3.org/2000/svg" width="8" height="120">
    <rect width="8" height="120" rx="4" fill="#F59E0B"/>
  </svg>`)).png().toFile(path.join(outDir, 'accent-bar.png'));
  console.log('Created accent-bar.png');

  await createKweeksLogo('kweeks-logo.png', 256);
  await createDollarIcon('icon-dollar.png', 256);
  await createQuizIcon('icon-quiz.png', 256);
  await createLightningIcon('icon-lightning.png', 256);
  await createNigeriaIcon('icon-nigeria.png', 256);
  await createCheckIcon('icon-check.png', 256);
  await createIconCircle('circle-amber.png', '#F59E0B', 256);
  await createIconCircle('circle-indigo.png', '#6366F1', 256);
  await createIconCircle('circle-emerald.png', '#10B981', 256);
  await createIconCircle('circle-red.png', '#EF4444', 256);
  console.log('All assets created!');
})();
