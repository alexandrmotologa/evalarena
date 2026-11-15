const puppeteer = require('puppeteer');
const path = require('path');
const fs = require('fs');
const { execSync } = require('child_process');

async function main() {
  const imagesDir = path.resolve(__dirname, '../docs/images');
  const framesDir = path.resolve(__dirname, '../docs/images/frames');
  if (!fs.existsSync(imagesDir)) fs.mkdirSync(imagesDir, { recursive: true });
  if (!fs.existsSync(framesDir)) fs.mkdirSync(framesDir, { recursive: true });

  console.log('Launching headless browser...');
  const browser = await puppeteer.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--window-size=1440,960'],
  });

  const page = await browser.newPage();
  await page.setViewport({ width: 1440, height: 960, deviceScaleFactor: 2 });

  console.log('Navigating to http://127.0.0.1:7070...');
  await page.goto('http://127.0.0.1:7070', { waitUntil: 'domcontentloaded', timeout: 15000 });
  await new Promise(r => setTimeout(r, 2000));

  // 1. Capture Dashboard & Pareto Frontier
  console.log('Capturing Dashboard (Leaderboard & Pareto)...');
  await page.screenshot({ path: path.join(imagesDir, 'evalarena_dashboard.png') });

  // 2. Capture Mobile Responsive View (iPhone 14 Pro style: 390x844)
  console.log('Capturing Mobile Responsive View...');
  await page.setViewport({ width: 390, height: 844, deviceScaleFactor: 2 });
  await new Promise(r => setTimeout(r, 800));
  await page.screenshot({ path: path.join(imagesDir, 'evalarena_mobile.png') });

  // Reset viewport to desktop
  await page.setViewport({ width: 1440, height: 960, deviceScaleFactor: 2 });
  await new Promise(r => setTimeout(r, 500));

  // 3. Navigate to Blind Arena Tab
  console.log('Capturing Blind Arena Tab...');
  const tabs = await page.$$('.header-nav button');
  if (tabs.length >= 2) {
    await tabs[1].click(); // Blind Arena tab
    await new Promise(r => setTimeout(r, 1000));
    await page.screenshot({ path: path.join(imagesDir, 'evalarena_blind_arena.png') });
  }

  // 4. Navigate to Quick Battle Tab
  console.log('Capturing Quick Battle Tab...');
  if (tabs.length >= 3) {
    await tabs[2].click(); // Quick Battle
    await new Promise(r => setTimeout(r, 800));
    // Click Run Head-to-Head Battle
    const buttons = await page.$$('button');
    for (const btn of buttons) {
      const text = await page.evaluate(el => el.textContent, btn);
      if (text && text.includes('Run Head-to-Head Battle')) {
        await btn.click();
        console.log('Executing Quick Battle...');
        await new Promise(r => setTimeout(r, 2000));
        break;
      }
    }
    await page.screenshot({ path: path.join(imagesDir, 'evalarena_quick_battle.png') });
  }

  // 5. Navigate to Win Matrix Tab
  console.log('Capturing Win Rate Matrix Tab...');
  if (tabs.length >= 4) {
    await tabs[3].click(); // Matrix
    await new Promise(r => setTimeout(r, 800));
    await page.screenshot({ path: path.join(imagesDir, 'evalarena_win_matrix.png') });
  }

  // 6. Navigate to Match Inspector Tab
  console.log('Capturing Match Inspector Tab...');
  if (tabs.length >= 5) {
    await tabs[4].click(); // Inspector
    await new Promise(r => setTimeout(r, 800));
    // Expand first match
    const rows = await page.$$('.responsive-card > div > div');
    if (rows.length > 0) {
      await rows[0].click();
      await new Promise(r => setTimeout(r, 600));
    }
    await page.screenshot({ path: path.join(imagesDir, 'evalarena_inspector.png') });
  }

  // 7. RECORD DEMO FRAMES FOR ANIMATED GIF
  console.log('Recording animated demo sequence frames...');
  let frameIdx = 0;
  async function captureFrame(durationMs = 200) {
    const framePath = path.join(framesDir, `frame_${String(frameIdx).padStart(4, '0')}.png`);
    await page.screenshot({ path: framePath });
    frameIdx++;
    await new Promise(r => setTimeout(r, durationMs));
  }

  // Go to Leaderboard
  await tabs[0].click();
  await new Promise(r => setTimeout(r, 600));

  // Frame sequence 1: Dashboard overview
  for (let i = 0; i < 5; i++) await captureFrame(150);

  // Switch to Blind Arena
  await tabs[1].click();
  for (let i = 0; i < 4; i++) await captureFrame(150);

  // Toggle diff view in Blind Arena
  const arenaDiffBtn = await page.$('button[style*="SplitSquareVertical"], .responsive-card button');
  if (arenaDiffBtn) {
    await arenaDiffBtn.click();
    for (let i = 0; i < 4; i++) await captureFrame(150);
  }

  // Vote for Option A
  const voteABtn = await page.$('button:has(svg)');
  const allBtns = await page.$$('button');
  for (const b of allBtns) {
    const txt = await page.evaluate(el => el.textContent, b);
    if (txt && txt.includes('Response A is Better')) {
      await b.click();
      console.log('Voted Option A in Blind Arena!');
      break;
    }
  }
  for (let i = 0; i < 8; i++) await captureFrame(150);

  // Switch to Quick Battle
  await tabs[2].click();
  for (let i = 0; i < 4; i++) await captureFrame(150);

  // Click Run Head-to-Head
  for (const b of await page.$$('button')) {
    const txt = await page.evaluate(el => el.textContent, b);
    if (txt && txt.includes('Run Head-to-Head Battle')) {
      await b.click();
      break;
    }
  }
  for (let i = 0; i < 8; i++) await captureFrame(200);

  // Switch to Win Matrix
  await tabs[3].click();
  for (let i = 0; i < 4; i++) await captureFrame(150);

  // Return to Leaderboard
  await tabs[0].click();
  for (let i = 0; i < 6; i++) await captureFrame(150);

  await browser.close();
  console.log(`Captured ${frameIdx} frames. Compiling with ffmpeg into evalarena_demo.gif...`);

  const gifPath = path.join(imagesDir, 'evalarena_demo.gif');
  const palettePath = path.join(framesDir, 'palette.png');

  // Generate optimal color palette
  execSync(`ffmpeg -y -framerate 6 -i "${path.join(framesDir, 'frame_%04d.png')}" -vf "fps=6,scale=1000:-1:flags=lanczos,palettegen=stats_mode=diff" "${palettePath}"`);
  
  // Render GIF using palette
  execSync(`ffmpeg -y -framerate 6 -i "${path.join(framesDir, 'frame_%04d.png')}" -i "${palettePath}" -lavfi "fps=6,scale=1000:-1:flags=lanczos [x]; [x][1:v] paletteuse=dither=bayer:bayer_scale=3" "${gifPath}"`);

  console.log('✓ Successfully generated docs/images/evalarena_demo.gif');

  // Clean up individual frames
  fs.rmSync(framesDir, { recursive: true, force: true });
  console.log('✓ Cleaned up temporary frame sequence');
}

main().catch(err => {
  console.error('Asset capture failed:', err);
  process.exit(1);
});
