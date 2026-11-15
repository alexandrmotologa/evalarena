import React, { useState } from 'react';
import { ParetoPoint, Recommendation } from '../api/client';
import { Target, Zap, DollarSign, ArrowRight, ShieldAlert } from 'lucide-react';

interface Props {
  points: ParetoPoint[];
  recommendations: Recommendation[];
}

export const ParetoFrontier: React.FC<Props> = ({ points, recommendations }) => {
  const [hoveredPoint, setHoveredPoint] = useState<ParetoPoint | null>(null);

  if (points.length === 0) {
    return null;
  }

  // Chart dimensions
  const width = 640;
  const height = 320;
  const padding = { top: 30, right: 40, bottom: 40, left: 60 };

  const plotW = width - padding.left - padding.right;
  const plotH = height - padding.top - padding.bottom;

  // Determine axes domain
  const minCost = 0;
  const maxCost = Math.max(...points.map(p => p.cost_per_m), 1.0) * 1.15;

  const minELO = Math.min(...points.map(p => p.elo)) - 30;
  const maxELO = Math.max(...points.map(p => p.elo)) + 30;
  const eloRange = Math.max(1, maxELO - minELO);

  const getX = (cost: number) => padding.left + (cost / maxCost) * plotW;
  const getY = (elo: number) => padding.top + plotH - ((elo - minELO) / eloRange) * plotH;

  // Sort optimal points by cost to draw the frontier curve
  const optimalPoints = points
    .filter(p => p.is_optimal)
    .sort((a, b) => a.cost_per_m - b.cost_per_m);

  return (
    <div style={{ background: 'var(--bg-card)', borderRadius: '12px', border: '1px solid var(--border-subtle)', padding: '1.5rem', marginBottom: '1.5rem' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.25rem' }}>
        <div>
          <h2 style={{ fontSize: '1.15rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Target style={{ width: 20, height: 20, color: 'var(--accent)' }} />
            Pareto Efficiency Frontier (Quality vs. Cost)
          </h2>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '0.2rem' }}>
            Identifies non-dominated models that maximize ELO rating while minimizing token expenditure and latency.
          </p>
        </div>
      </div>

      <div style={{ position: 'relative', width: '100%', overflowX: 'auto' }}>
        <svg viewBox={`0 0 ${width} ${height}`} style={{ width: '100%', height: 'auto', maxHeight: 360, overflow: 'visible' }}>
          {/* Grid lines */}
          {[0, 0.25, 0.5, 0.75, 1].map((ratio) => {
            const y = padding.top + plotH * ratio;
            const eloVal = Math.round(maxELO - ratio * eloRange);
            return (
              <g key={`y-grid-${ratio}`}>
                <line x1={padding.left} y1={y} x2={padding.left + plotW} y2={y} stroke="rgba(255,255,255,0.05)" strokeDasharray="4 4" />
                <text x={padding.left - 10} y={y + 4} textAnchor="end" fill="var(--text-dim)" fontSize="10" className="font-mono">
                  {eloVal}
                </text>
              </g>
            );
          })}

          {[0, 0.25, 0.5, 0.75, 1].map((ratio) => {
            const x = padding.left + plotW * ratio;
            const costVal = (ratio * maxCost).toFixed(1);
            return (
              <g key={`x-grid-${ratio}`}>
                <line x1={x} y1={padding.top} x2={x} y2={padding.top + plotH} stroke="rgba(255,255,255,0.05)" strokeDasharray="4 4" />
                <text x={x} y={padding.top + plotH + 20} textAnchor="middle" fill="var(--text-dim)" fontSize="10" className="font-mono">
                  ${costVal}
                </text>
              </g>
            );
          })}

          {/* Axis Labels */}
          <text x={padding.left + plotW / 2} y={height - 5} textAnchor="middle" fill="var(--text-muted)" fontSize="11" fontWeight="600">
            Blended Cost per 1M Tokens (USD) → (Lower is Better)
          </text>
          <text transform={`rotate(-90)`} x={-(padding.top + plotH / 2)} y={15} textAnchor="middle" fill="var(--text-muted)" fontSize="11" fontWeight="600">
            Bradley-Terry ELO Rating → (Higher is Better)
          </text>

          {/* Pareto Frontier Line */}
          {optimalPoints.length > 1 && (
            <polyline
              fill="none"
              stroke="var(--accent)"
              strokeWidth="2.5"
              strokeDasharray="6 3"
              opacity="0.8"
              points={optimalPoints.map(p => `${getX(p.cost_per_m)},${getY(p.elo)}`).join(' ')}
            />
          )}

          {/* Scatter points */}
          {points.map((p) => {
            const cx = getX(p.cost_per_m);
            const cy = getY(p.elo);
            const isHovered = hoveredPoint?.model_id === p.model_id;

            return (
              <g
                key={p.model_id}
                style={{ cursor: 'pointer' }}
                onMouseEnter={() => setHoveredPoint(p)}
                onMouseLeave={() => setHoveredPoint(null)}
              >
                {/* Halo for optimal */}
                {p.is_optimal && (
                  <circle
                    cx={cx}
                    cy={cy}
                    r={isHovered ? 16 : 12}
                    fill="rgba(16, 185, 129, 0.2)"
                    stroke="rgba(16, 185, 129, 0.6)"
                    strokeWidth="1.5"
                  />
                )}
                {/* Core point */}
                <circle
                  cx={cx}
                  cy={cy}
                  r={isHovered ? 8 : 6}
                  fill={p.is_optimal ? 'var(--success)' : 'var(--text-dim)'}
                  stroke="#ffffff"
                  strokeWidth={isHovered ? 2 : 1}
                />
                {/* Model label tag */}
                <text
                  x={cx + 10}
                  y={cy + 4}
                  fill={p.is_optimal ? 'var(--text-main)' : 'var(--text-dim)'}
                  fontSize="11"
                  fontWeight={p.is_optimal ? '600' : '400'}
                >
                  {p.model_id}
                </text>
              </g>
            );
          })}
        </svg>

        {/* Floating Tooltip */}
        {hoveredPoint && (
          <div style={{
            position: 'absolute',
            top: 10,
            right: 10,
            background: 'var(--bg-subtle)',
            border: '1px solid var(--border-focus)',
            borderRadius: '8px',
            padding: '0.75rem 1rem',
            fontSize: '0.8rem',
            boxShadow: '0 8px 24px rgba(0,0,0,0.5)',
            zIndex: 10,
            minWidth: 200,
          }}>
            <div style={{ fontWeight: 700, color: 'var(--text-main)', marginBottom: '0.3rem' }}>
              {hoveredPoint.model_id}
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-muted)' }}>
              <span>ELO Rating:</span>
              <strong style={{ color: 'var(--warning)' }}>{Math.round(hoveredPoint.elo)}</strong>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-muted)' }}>
              <span>Inference Cost:</span>
              <strong>${hoveredPoint.cost_per_m.toFixed(2)}/1M</strong>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', color: 'var(--text-muted)' }}>
              <span>Latency:</span>
              <strong>{Math.round(hoveredPoint.latency_ms)}ms</strong>
            </div>
            <div style={{ marginTop: '0.4rem', paddingTop: '0.4rem', borderTop: '1px solid var(--border-subtle)', color: hoveredPoint.is_optimal ? 'var(--success)' : 'var(--text-dim)', fontWeight: 600 }}>
              {hoveredPoint.is_optimal ? '✓ Strictly Pareto Optimal' : '✕ Dominated by alternative'}
            </div>
          </div>
        )}
      </div>

      {/* Recommendations Cards */}
      {recommendations.length > 0 && (
        <div style={{ marginTop: '1.5rem', display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1rem' }}>
          {recommendations.map((rec) => (
            <div
              key={rec.model_id}
              style={{
                background: 'var(--bg-subtle)',
                border: '1px solid rgba(99, 102, 241, 0.2)',
                borderRadius: '8px',
                padding: '1rem',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'space-between',
              }}
            >
              <div>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.4rem' }}>
                  <span style={{ fontWeight: 700, color: 'var(--text-main)', fontSize: '0.95rem' }}>{rec.model_id}</span>
                  <span style={{ fontSize: '0.75rem', padding: '0.2rem 0.5rem', borderRadius: '4px', background: 'rgba(99, 102, 241, 0.15)', color: 'var(--accent-light)', fontWeight: 600 }}>
                    Pareto Choice
                  </span>
                </div>
                <p style={{ fontSize: '0.82rem', color: 'var(--text-muted)', lineHeight: 1.4 }}>
                  {rec.summary}
                </p>
              </div>

              <div style={{ marginTop: '0.8rem', paddingTop: '0.6rem', borderTop: '1px solid var(--border-subtle)', display: 'flex', justifyContent: 'space-between', fontSize: '0.75rem', color: 'var(--text-dim)' }}>
                <span>Quality: <strong>{rec.elo_retention_pct.toFixed(0)}%</strong></span>
                {rec.cost_reduction_pct > 0 && (
                  <span style={{ color: 'var(--success)' }}>Cost Savings: <strong>-{rec.cost_reduction_pct.toFixed(0)}%</strong></span>
                )}
                <span>Latency: <strong>{Math.round(rec.latency_ms)}ms</strong></span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
