const puppeteer = require('puppeteer');
const path = require('path');
const fs = require('fs');

async function main() {
  const outputPath = path.resolve(__dirname, '../docs/images/evalarena_cli.png');
  const browser = await puppeteer.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--window-size=1200,820'],
  });

  const page = await browser.newPage();
  await page.setViewport({ width: 1200, height: 820, deviceScaleFactor: 2 });

  const terminalHtml = `<!DOCTYPE html>
<html>
<head>
  <style>
    body {
      margin: 0;
      padding: 32px;
      background: #0a0e17;
      font-family: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 100vh;
      box-sizing: border-box;
    }
    .terminal {
      width: 100%;
      max-width: 1100px;
      background: #0d121f;
      border-radius: 12px;
      border: 1px solid rgba(255, 255, 255, 0.1);
      box-shadow: 0 20px 50px rgba(0, 0, 0, 0.6);
      overflow: hidden;
    }
    .header {
      background: #151c2e;
      padding: 12px 18px;
      display: flex;
      align-items: center;
      gap: 8px;
      border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    }
    .dot {
      width: 12px;
      height: 12px;
      border-radius: 50%;
    }
    .dot-red { background: #ef4444; }
    .dot-yellow { background: #f59e0b; }
    .dot-green { background: #10b981; }
    .title {
      margin-left: 12px;
      color: #94a3b8;
      font-size: 13px;
      font-weight: 600;
    }
    .content {
      padding: 24px 28px;
      color: #f1f5f9;
      font-size: 13.5px;
      line-height: 1.6;
      white-space: pre;
    }
    .prompt-line { color: #6366f1; font-weight: 700; }
    .cmd { color: #ffffff; }
    .gold { color: #fbbf24; font-weight: 700; }
    .cyan { color: #38bdf8; }
    .green { color: #34d399; font-weight: 600; }
    .muted { color: #64748b; }
    .box { color: #94a3b8; }
  </style>
</head>
<body>
  <div class="terminal">
    <div class="header">
      <div class="dot dot-red"></div>
      <div class="dot dot-yellow"></div>
      <div class="dot dot-green"></div>
      <div class="title">evalarena CLI — Production CI/CD Gating & Autonomous Tournament</div>
    </div>
    <div class="content"><span class="prompt-line">$</span> <span class="cmd">evalarena run --config examples/evalarena.yaml --rounds 2</span>
<span class="cyan">Starting tournament "Enterprise LLM Pairwise Benchmark" with 4 models on 5 prompts (2 rounds)...</span>
<span class="green">✓ Completed 60 pairwise battles in 467ms across universal provider pool.</span>

<span class="gold">=== evalarena Leaderboard: Enterprise LLM Pairwise Benchmark ===</span>
<span class="box">┌──────┬───────────────────────┬──────┬───────────────┬───────────┬───────┬────────────┬─────────┬───────────────┐
│ RANK │         MODEL         │ ELO  │   95 %  CI    │ W / L / T │ WIN % │ $/ 1 M TOK │ LATENCY │    PARETO     │
├──────┼───────────────────────┼──────┼───────────────┼───────────┼───────┼────────────┼─────────┼───────────────┤</span>
│ <span class="gold">★ 1</span>  │ <span class="cmd">gpt-4o</span>                │ <span class="gold">1092</span> │ [1051 - 1138] │ 10/0/20   │ 66.7% │ $1.42      │ 30ms    │ <span class="green">Yes (Optimal)</span> │
│ 2    │ claude-3-5-sonnet     │ 1000 │ [976 - 1023]  │ 0/0/30    │ 50.0% │ $2.12      │ 35ms    │ No            │
│ 3    │ <span class="cmd">gemini-2.0-flash</span>      │ 1000 │ [974 - 1025]  │ 0/0/30    │ 50.0% │ $0.06      │ 15ms    │ <span class="green">Yes (Optimal)</span> │
│ 4    │ <span class="cmd">llama-3.1-8b-instruct</span> │ 908  │ [861 - 947]   │ 0/10/20   │ 33.3% │ $0.01      │ 10ms    │ <span class="green">Yes (Optimal)</span> │
<span class="box">└──────┴───────────────────────┴──────┴───────────────┴───────────┴───────┴────────────┴─────────┴───────────────┘</span>

<span class="cyan">--- Pareto Frontier Analysis & Recommendations ---</span>
• <span class="cmd">gpt-4o</span>: Top Quality Baseline (ELO 1092, $1.42/1M tokens, 30ms latency).
• <span class="green">gemini-2.0-flash</span>: High-Efficiency Alternative: Delivers 91.6% of peak ELO while reducing cost by 96.1% (15ms).
• <span class="green">llama-3.1-8b-instruct</span>: High-Efficiency Alternative: Delivers 83.2% of peak ELO while reducing cost by 99.1% (10ms).

<span class="prompt-line">$</span> <span class="cmd">evalarena check --baseline gpt-4o --candidate claude-3-5-sonnet --max-elo-drop 100</span>
<span class="cyan">evalarena CI/CD Quality Gate Check</span>
• Baseline (gpt-4o):              ELO 1091.7
• Candidate (claude-3-5-sonnet): ELO 1000.0
• ELO Difference:                -91.7 (Allowed max drop: -100.0)

<span class="green">[GATE PASSED] Candidate model meets production quality threshold.</span></div>
  </div>
</body>
</html>`;

  await page.setContent(terminalHtml);
  await page.screenshot({ path: outputPath });
  await browser.close();
  console.log('✓ Successfully rendered docs/images/evalarena_cli.png');
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
