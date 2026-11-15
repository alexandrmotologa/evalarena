import React from 'react';
import { ModelRating, CalibrationReport } from '../api/client';
import { Trophy, CheckCircle2, TrendingUp, DollarSign, Clock, ShieldCheck, Filter, Award, Scale } from 'lucide-react';

interface Props {
  ratings: ModelRating[];
  loading: boolean;
  categories: string[];
  selectedCategory: string;
  onSelectCategory: (cat: string) => void;
  calibration: CalibrationReport | null;
}

export const LeaderboardTable: React.FC<Props> = ({
  ratings,
  loading,
  categories,
  selectedCategory,
  onSelectCategory,
  calibration,
}) => {
  // Calculate ELO min and max for CI visual bar scaling
  const minELO = ratings.length > 0 ? Math.min(...ratings.map(r => r.ci_lower || r.elo)) - 30 : 900;
  const maxELO = ratings.length > 0 ? Math.max(...ratings.map(r => r.ci_upper || r.elo)) + 30 : 1200;
  const eloRange = Math.max(1, maxELO - minELO);

  return (
    <div>
      {/* Calibration Scorecard (Human vs LLM Judge Alignment) */}
      {calibration && calibration.total_comparisons > 0 && (
        <div style={{
          background: 'var(--bg-card)',
          borderRadius: '12px',
          border: '1px solid var(--border-subtle)',
          padding: '1.25rem 1.5rem',
          marginBottom: '1.5rem',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          flexWrap: 'wrap',
          gap: '1rem',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <div style={{
              width: 38,
              height: 38,
              borderRadius: '8px',
              background: 'rgba(16, 185, 129, 0.15)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}>
              <Scale style={{ width: 20, height: 20, color: 'var(--success)' }} />
            </div>
            <div>
              <div style={{ fontWeight: 700, fontSize: '0.95rem', color: 'var(--text-main)' }}>
                LLM Judge Calibration & Human Consensus
              </div>
              <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                Inter-annotator agreement between team blind votes and automated debiased judgments.
              </div>
            </div>
          </div>

          <div className="calibration-stats-row" style={{ fontSize: '0.85rem' }}>
            <div>
              <span style={{ color: 'var(--text-dim)', display: 'block', fontSize: '0.7rem', textTransform: 'uppercase' }}>Observed Agreement</span>
              <strong style={{ color: 'var(--success)', fontSize: '1.1rem' }} className="font-mono">
                {calibration.observed_agreement_pct.toFixed(1)}%
              </strong>
            </div>

            <div>
              <span style={{ color: 'var(--text-dim)', display: 'block', fontSize: '0.7rem', textTransform: 'uppercase' }}>Cohen's Kappa (κ)</span>
              <strong style={{ color: 'var(--accent-light)', fontSize: '1.1rem' }} className="font-mono">
                {calibration.cohens_kappa.toFixed(2)}
              </strong>
            </div>

            <div>
              <span style={{ color: 'var(--text-dim)', display: 'block', fontSize: '0.7rem', textTransform: 'uppercase' }}>Consensus Level</span>
              <span style={{
                display: 'inline-block',
                padding: '0.2rem 0.5rem',
                borderRadius: '4px',
                background: 'rgba(99, 102, 241, 0.15)',
                color: 'var(--accent-light)',
                fontWeight: 600,
                fontSize: '0.8rem',
              }}>
                {calibration.interpretation}
              </span>
            </div>

            <div>
              <span style={{ color: 'var(--text-dim)', display: 'block', fontSize: '0.7rem', textTransform: 'uppercase' }}>Verified Battles</span>
              <strong style={{ color: 'var(--text-main)', fontSize: '1.1rem' }} className="font-mono">
                {calibration.total_comparisons}
              </strong>
            </div>
          </div>
        </div>
      )}

      {/* Main Table Card */}
      <div style={{ background: 'var(--bg-card)', borderRadius: '12px', border: '1px solid var(--border-subtle)', overflow: 'hidden' }}>
        <div style={{ padding: '1.25rem 1.5rem', borderBottom: '1px solid var(--border-subtle)', display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '1rem' }}>
          <div>
            <h2 style={{ fontSize: '1.15rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Trophy style={{ width: 20, height: 20, color: 'var(--warning)' }} />
              Bradley-Terry Model Leaderboard
            </h2>
            <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '0.2rem' }}>
              Latent skill ratings solved via Maximum Likelihood Estimation with 95% bootstrap confidence intervals.
            </p>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
            {/* Category Slice Filter */}
            {categories.length > 0 && (
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Filter style={{ width: 14, height: 14, color: 'var(--text-dim)' }} />
                <select
                  value={selectedCategory}
                  onChange={(e) => onSelectCategory(e.target.value)}
                  style={{
                    padding: '0.4rem 0.75rem',
                    background: 'var(--bg-subtle)',
                    border: '1px solid var(--border-subtle)',
                    borderRadius: '6px',
                    color: 'var(--text-main)',
                    fontSize: '0.85rem',
                    fontWeight: 600,
                    cursor: 'pointer',
                  }}
                >
                  <option value="">All Benchmark Slices</option>
                  {categories.map((c) => (
                    <option key={c} value={c}>
                      Slice: {c}
                    </option>
                  ))}
                </select>
              </div>
            )}

            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', fontSize: '0.8rem', color: 'var(--text-dim)' }}>
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.25rem' }}>
                <ShieldCheck style={{ width: 14, height: 14, color: 'var(--success)' }} /> Position-Debiased
              </span>
            </div>
          </div>
        </div>

        {loading && ratings.length === 0 ? (
          <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-muted)' }}>
            <div className="animate-pulse-subtle">Computing Bradley-Terry ELO ratings & bootstrap intervals...</div>
          </div>
        ) : ratings.length === 0 ? (
          <div style={{ padding: '3rem', textAlign: 'center' }}>
            <Trophy style={{ width: 40, height: 40, margin: '0 auto 1rem', color: 'var(--text-dim)' }} />
            <h3 style={{ fontSize: '1.1rem', fontWeight: 600, marginBottom: '0.5rem' }}>No Tournament Matches for this Slice</h3>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', maxWidth: 440, margin: '0 auto' }}>
              {selectedCategory ? `No matches recorded under category "${selectedCategory}".` : 'Execute a tournament round to generate rankings.'}
            </p>
          </div>
        ) : (
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
                          <div style={{ width: '100%', height: 4, background: 'var(--bg-subtle)', borderRadius: 2, position: 'relative' }}>
                            <div style={{
                              position: 'absolute',
                              left: `${ciLeftPct}%`,
                              width: `${ciWidthPct}%`,
                              height: '100%',
                              background: 'rgba(99, 102, 241, 0.4)',
                              borderRadius: 2,
                            }} />
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
        )}
      </div>
    </div>
  );
};
