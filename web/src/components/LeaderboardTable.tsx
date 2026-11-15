import React from 'react';
import { ModelRating } from '../api/client';
import { Trophy, CheckCircle2, TrendingUp, DollarSign, Clock, ShieldCheck } from 'lucide-react';

interface Props {
  ratings: ModelRating[];
  loading: boolean;
}

export const LeaderboardTable: React.FC<Props> = ({ ratings, loading }) => {
  if (loading && ratings.length === 0) {
    return (
      <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-muted)' }}>
        <div className="animate-pulse-subtle">Computing Bradley-Terry ELO ratings & bootstrap intervals...</div>
      </div>
    );
  }

  if (ratings.length === 0) {
    return (
      <div style={{ padding: '3rem', textAlign: 'center', background: 'var(--bg-card)', borderRadius: '12px', border: '1px solid var(--border-subtle)' }}>
        <Trophy style={{ width: 40, height: 40, margin: '0 auto 1rem', color: 'var(--text-dim)' }} />
        <h3 style={{ fontSize: '1.1rem', fontWeight: 600, marginBottom: '0.5rem' }}>No Tournament Matches Yet</h3>
        <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', maxWidth: 440, margin: '0 auto' }}>
          Execute a tournament round via the CLI or click "Run Tournament" above to generate pairwise completions and ELO ratings.
        </p>
      </div>
    );
  }

  // Calculate ELO min and max for CI visual bar scaling
  const minELO = Math.min(...ratings.map(r => r.ci_lower || r.elo)) - 30;
  const maxELO = Math.max(...ratings.map(r => r.ci_upper || r.elo)) + 30;
  const eloRange = Math.max(1, maxELO - minELO);

  return (
    <div style={{ background: 'var(--bg-card)', borderRadius: '12px', border: '1px solid var(--border-subtle)', overflow: 'hidden' }}>
      <div style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--border-subtle)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h2 style={{ fontSize: '1.15rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Trophy style={{ width: 20, height: 20, color: 'var(--warning)' }} />
            Bradley-Terry Model Leaderboard
          </h2>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '0.2rem' }}>
            Latent skill ratings solved via Maximum Likelihood Estimation with 95% bootstrap confidence intervals.
          </p>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '0.8rem', color: 'var(--text-dim)' }}>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem' }}>
            <ShieldCheck style={{ width: 14, height: 14, color: 'var(--success)' }} /> Position-Debiased
          </span>
        </div>
      </div>

      <div style={{ overflowX: 'auto' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.9rem' }}>
          <thead>
            <tr style={{ background: 'rgba(255,255,255,0.02)', borderBottom: '1px solid var(--border-subtle)', color: 'var(--text-dim)', fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              <th style={{ padding: '0.85rem 1rem 0.85rem 1.5rem', width: 60 }}>Rank</th>
              <th style={{ padding: '0.85rem 1rem' }}>Model Candidate</th>
              <th style={{ padding: '0.85rem 1rem', textAlign: 'right' }}>ELO Rating</th>
              <th style={{ padding: '0.85rem 1.5rem', width: 240 }}>95% Confidence Interval</th>
              <th style={{ padding: '0.85rem 1rem', textAlign: 'center' }}>W / L / T</th>
              <th style={{ padding: '0.85rem 1rem', textAlign: 'right' }}>Win Rate</th>
              <th style={{ padding: '0.85rem 1rem', textAlign: 'right' }}>$/1M Tokens</th>
              <th style={{ padding: '0.85rem 1rem', textAlign: 'right' }}>Latency</th>
              <th style={{ padding: '0.85rem 1.5rem 0.85rem 1rem', textAlign: 'center' }}>Pareto</th>
            </tr>
          </thead>
          <tbody>
            {ratings.map((r, idx) => {
              const ciLeftPct = Math.max(0, ((r.ci_lower - minELO) / eloRange) * 100);
              const ciWidthPct = Math.max(4, (((r.ci_upper - r.ci_lower) / eloRange) * 100));
              const eloPointPct = Math.max(0, Math.min(100, ((r.elo - minELO) / eloRange) * 100));

              return (
                <tr 
                  key={r.model_id}
                  style={{
                    borderBottom: '1px solid var(--border-subtle)',
                    transition: 'background 0.15s ease',
                  }}
                  onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--bg-card-hover)')}
                  onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
                >
                  <td style={{ padding: '1rem 1rem 1rem 1.5rem', fontWeight: 700, color: idx === 0 ? 'var(--warning)' : 'var(--text-dim)' }}>
                    {idx === 0 ? '★ 1' : idx + 1}
                  </td>
                  <td style={{ padding: '1rem' }}>
                    <div style={{ fontWeight: 600, color: 'var(--text-main)' }}>{r.model_id}</div>
                    <div style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }}>{r.total_battles} matches recorded</div>
                  </td>
                  <td style={{ padding: '1rem', textAlign: 'right', fontWeight: 700, color: idx === 0 ? 'var(--warning)' : 'var(--text-main)', fontSize: '1rem' }} className="font-mono">
                    {Math.round(r.elo)}
                  </td>
                  <td style={{ padding: '1rem 1.5rem' }}>
                    <div style={{ position: 'relative', height: 20, display: 'flex', alignItems: 'center' }}>
                      {/* Range track */}
                      <div style={{ width: '100%', height: 4, background: 'var(--bg-subtle)', borderRadius: 2, position: 'relative' }}>
                        {/* 95% CI Bar */}
                        <div style={{
                          position: 'absolute',
                          left: `${ciLeftPct}%`,
                          width: `${ciWidthPct}%`,
                          height: '100%',
                          background: 'rgba(99, 102, 241, 0.4)',
                          borderRadius: 2,
                        }} />
                        {/* Point estimate */}
                        <div style={{
                          position: 'absolute',
                          left: `${eloPointPct}%`,
                          top: -4,
                          width: 12,
                          height: 12,
                          background: idx === 0 ? 'var(--warning)' : 'var(--accent)',
                          borderRadius: '50%',
                          transform: 'translateX(-50%)',
                          boxShadow: '0 0 8px rgba(99,102,241,0.5)',
                        }} />
                      </div>
                    </div>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.7rem', color: 'var(--text-dim)', marginTop: 2 }} className="font-mono">
                      <span>{Math.round(r.ci_lower)}</span>
                      <span>{Math.round(r.ci_upper)}</span>
                    </div>
                  </td>
                  <td style={{ padding: '1rem', textAlign: 'center', color: 'var(--text-muted)' }} className="font-mono">
                    {r.wins} / {r.losses} / {r.ties}
                  </td>
                  <td style={{ padding: '1rem', textAlign: 'right', fontWeight: 600, color: r.win_rate >= 0.5 ? 'var(--success)' : 'var(--text-muted)' }}>
                    {(r.win_rate * 100).toFixed(1)}%
                  </td>
                  <td style={{ padding: '1rem', textAlign: 'right', color: 'var(--text-main)' }} className="font-mono">
                    ${r.avg_cost_per_m.toFixed(2)}
                  </td>
                  <td style={{ padding: '1rem', textAlign: 'right', color: 'var(--text-muted)' }} className="font-mono">
                    {Math.round(r.avg_latency_ms)}ms
                  </td>
                  <td style={{ padding: '1rem 1.5rem 1rem 1rem', textAlign: 'center' }}>
                    {r.is_pareto_optimal ? (
                      <span style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: '0.3rem',
                        padding: '0.25rem 0.6rem',
                        borderRadius: '9999px',
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        background: 'rgba(16, 185, 129, 0.12)',
                        color: 'var(--success)',
                        border: '1px solid rgba(16, 185, 129, 0.25)',
                      }}>
                        <CheckCircle2 style={{ width: 12, height: 12 }} /> Optimal
                      </span>
                    ) : (
                      <span style={{
                        padding: '0.25rem 0.6rem',
                        borderRadius: '9999px',
                        fontSize: '0.75rem',
                        color: 'var(--text-dim)',
                        background: 'rgba(255, 255, 255, 0.03)',
                      }}>
                        Dominated
                      </span>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
};
