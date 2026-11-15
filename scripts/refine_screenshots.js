const puppeteer = require('puppeteer');
const path = require('path');
const fs = require('fs');

async function main() {
  const imagesDir = path.resolve(__dirname, '../docs/images');
  if (!fs.existsSync(imagesDir)) fs.mkdirSync(imagesDir, { recursive: true });

  console.log('Launching browser for screenshot refinement...');
  const browser = await puppeteer.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--window-size=1440,1100'],
  });

  const page = await browser.newPage();
  await page.setViewport({ width: 1440, height: 1100, deviceScaleFactor: 2 });

  await page.goto('http://127.0.0.1:7070', { waitUntil: 'domcontentloaded' });
  await new Promise(r => setTimeout(r, 2000));

  // 1. Leaderboard Full View (scroll down to show Leaderboard + Calibration)
  console.log('Capturing Leaderboard Table & Calibration Scorecard...');
  const lbTable = await page.$('table');
  if (lbTable) {
    await page.evaluate(() => {
      window.scrollTo(0, 480);
    });
    await new Promise(r => setTimeout(r, 600));
    await page.screenshot({ path: path.join(imagesDir, 'evalarena_leaderboard.png') });
  }

  // 2. Match Inspector with expanded row and diff
  console.log('Capturing Match Inspector with expanded row & dual-pass reasoning...');
  const tabs = await page.$$('.header-nav button');
  if (tabs.length >= 5) {
    await tabs[4].click(); // Inspector tab
    await new Promise(r => setTimeout(r, 1000));

    // Find first match row and click it
    const rowHeader = await page.$('.responsive-card > div > div > div');
    if (rowHeader) {
      await rowHeader.click();
      console.log('Clicked row to expand details');
      await new Promise(r => setTimeout(r, 600));
    }

    // Toggle Diff Mode
    const diffBtn = await page.$('button');
    const buttons = await page.$$('button');
    for (const b of buttons) {
      const txt = await page.evaluate(el => el.textContent, b);
      if (txt && (txt.includes('Toggle Diff') || txt.includes('Diff Active'))) {
        await b.click();
        console.log('Toggled diff mode');
        await new Promise(r => setTimeout(r, 500));
        break;
      }
    }

    await page.screenshot({ path: path.join(imagesDir, 'evalarena_inspector.png') });
  }

  await browser.close();
  console.log('✓ Successfully refined screenshots!');
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
