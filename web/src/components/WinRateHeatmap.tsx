import React from 'react';
import { WinRateMatrix } from '../api/client';
import { Grid, HelpCircle } from 'lucide-react';

interface Props {
  matrix: WinRateMatrix | null;
}

export const WinRateHeatmap: React.FC<Props> = ({ matrix }) => {
  if (!matrix || !matrix.models || matrix.models.length < 2) {
    return (
      <div style={{ padding: '3rem', textAlign: 'center', background: 'var(--bg-card)', borderRadius: '12px', border: '1px solid var(--border-subtle)' }}>
        <Grid style={{ width: 36, height: 36, margin: '0 auto 1rem', color: 'var(--text-dim)' }} />
        <h3 style={{ fontSize: '1.1rem', fontWeight: 600, marginBottom: '0.5rem' }}>No Head-to-Head Data</h3>
        <p style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>
          At least two competing models with recorded battles are required to construct the pairwise win-rate matrix.
        </p>
      </div>
    );
  }

  const { models, data } = matrix;

  const getCellColor = (winRate: number, total: number) => {
    if (total === 0) return 'rgba(255, 255, 255, 0.02)';
    if (winRate > 0.65) return 'rgba(16, 185, 129, 0.25)';
    if (winRate > 0.50) return 'rgba(16, 185, 129, 0.12)';
    if (winRate === 0.50) return 'rgba(255, 255, 255, 0.05)';
    if (winRate > 0.35) return 'rgba(239, 68, 68, 0.12)';
    return 'rgba(239, 68, 68, 0.25)';
  };

  return (
    <div style={{ background: 'var(--bg-card)', borderRadius: '12px', border: '1px solid var(--border-subtle)', padding: '1.5rem' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.25rem' }}>
        <div>
          <h2 style={{ fontSize: '1.15rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Grid style={{ width: 20, height: 20, color: 'var(--accent)' }} />
            Pairwise Head-to-Head Win-Rate Matrix
          </h2>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '0.2rem' }}>
            Percentage of times Row Model defeated Column Opponent (ties count as 0.5 win).
          </p>
        </div>
      </div>

      <div style={{ overflowX: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'center', fontSize: '0.85rem' }}>
          <thead>
            <tr>
              <th style={{ padding: '0.75rem 1rem', textAlign: 'left', color: 'var(--text-dim)', fontSize: '0.75rem', textTransform: 'uppercase' }}>
                Row vs. Col
              </th>
              {models.map(m => (
                <th key={m} style={{ padding: '0.75rem 0.5rem', color: 'var(--text-muted)', fontWeight: 600, fontSize: '0.8rem' }}>
                  {m}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {models.map(rowModel => (
              <tr key={rowModel} style={{ borderTop: '1px solid var(--border-subtle)' }}>
                <td style={{ padding: '0.75rem 1rem', textAlign: 'left', fontWeight: 600, color: 'var(--text-main)' }}>
                  {rowModel}
                </td>
                {models.map(colModel => {
                  if (rowModel === colModel) {
                    return (
                      <td key={colModel} style={{ padding: '0.75rem 0.5rem', background: 'rgba(255, 255, 255, 0.01)', color: 'var(--text-dim)' }}>
                        —
                      </td>
                    );
                  }

                  const rec = data[rowModel]?.[colModel] || { wins: 0, losses: 0, ties: 0, total: 0, win_rate: 0 };
                  const cellBg = getCellColor(rec.win_rate, rec.total);

                  return (
                    <td
                      key={colModel}
                      style={{
                        padding: '0.75rem 0.5rem',
                        background: cellBg,
                        border: '1px solid rgba(255, 255, 255, 0.04)',
                        transition: 'background 0.15s ease',
                      }}
                      title={`${rowModel} vs ${colModel}: ${rec.wins}W / ${rec.losses}L / ${rec.ties}T`}
                    >
                      <div style={{ fontWeight: 700, color: rec.total > 0 ? 'var(--text-main)' : 'var(--text-dim)' }} className="font-mono">
                        {rec.total > 0 ? `${(rec.win_rate * 100).toFixed(0)}%` : '0%'}
                      </div>
                      <div style={{ fontSize: '0.7rem', color: 'var(--text-dim)' }} className="font-mono">
                        {rec.wins}-{rec.losses}-{rec.ties}
                      </div>
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
