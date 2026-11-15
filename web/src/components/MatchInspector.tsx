import React, { useState } from 'react';
import { MatchRecord } from '../api/client';
import { ListFilter, ChevronDown, ChevronRight, ShieldAlert, CheckCircle2, Clock, DollarSign } from 'lucide-react';

interface Props {
  matches: MatchRecord[];
}

export const MatchInspector: React.FC<Props> = ({ matches }) => {
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [filterBiasOnly, setFilterBiasOnly] = useState<boolean>(false);

  const displayedMatches = filterBiasOnly
    ? matches.filter(m => m.position_bias_detected)
    : matches;

  if (matches.length === 0) {
    return (
      <div style={{ padding: '3rem', textAlign: 'center', background: 'var(--bg-card)', borderRadius: '12px', border: '1px solid var(--border-subtle)' }}>
        <ListFilter style={{ width: 36, height: 36, margin: '0 auto 1rem', color: 'var(--text-dim)' }} />
        <h3 style={{ fontSize: '1.1rem', fontWeight: 600, marginBottom: '0.5rem' }}>No Matches Recorded</h3>
        <p style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>
          Matches will appear here as the tournament engine runs pairwise battles.
        </p>
      </div>
    );
  }

  return (
    <div style={{ background: 'var(--bg-card)', borderRadius: '12px', border: '1px solid var(--border-subtle)', padding: '1.5rem' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h2 style={{ fontSize: '1.15rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <ListFilter style={{ width: 20, height: 20, color: 'var(--accent)' }} />
            Match Inspector & Position Bias Auditing
          </h2>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '0.2rem' }}>
            Inspect prompt-by-prompt completions, dual-pass judge reasoning, and detected position biases.
          </p>
        </div>

        <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.85rem', cursor: 'pointer', color: 'var(--text-muted)' }}>
          <input
            type="checkbox"
            checked={filterBiasOnly}
            onChange={(e) => setFilterBiasOnly(e.target.checked)}
            style={{ accentColor: 'var(--warning)', cursor: 'pointer' }}
          />
          Show only Position-Biased Inconsistencies ({matches.filter(m => m.position_bias_detected).length})
        </label>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
        {displayedMatches.map((m) => {
          const isExpanded = expandedId === m.id;

          return (
            <div
              key={m.id}
              style={{
                background: 'var(--bg-subtle)',
                borderRadius: '8px',
                border: m.position_bias_detected ? '1px solid rgba(245, 158, 11, 0.3)' : '1px solid var(--border-subtle)',
                overflow: 'hidden',
              }}
            >
              {/* Row summary */}
              <div
                onClick={() => setExpandedId(isExpanded ? null : m.id)}
                style={{
                  padding: '1rem 1.25rem',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  cursor: 'pointer',
                  userSelect: 'none',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                  {isExpanded ? <ChevronDown style={{ width: 16, height: 16, color: 'var(--text-dim)' }} /> : <ChevronRight style={{ width: 16, height: 16, color: 'var(--text-dim)' }} />}
                  <span style={{ fontWeight: 600, color: 'var(--text-main)', fontSize: '0.9rem' }}>
                    {m.model_a_id} vs {m.model_b_id}
                  </span>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }} className="font-mono">
                    [{m.prompt_id}]
                  </span>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', fontSize: '0.8rem' }}>
                  {m.position_bias_detected ? (
                    <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.3rem', color: 'var(--warning)', background: 'rgba(245, 158, 11, 0.1)', padding: '0.2rem 0.5rem', borderRadius: '4px' }}>
                      <ShieldAlert style={{ width: 13, height: 13 }} /> Bias Mitigated (Tie)
                    </span>
                  ) : (
                    <span style={{ color: 'var(--success)', fontWeight: 600 }}>
                      Winner: {m.final_winner === 'tie' ? 'Tie' : m.final_winner}
                    </span>
                  )}

                  {m.human_vote && (
                    <span style={{ color: 'var(--accent-light)', background: 'rgba(99, 102, 241, 0.15)', padding: '0.2rem 0.5rem', borderRadius: '4px' }}>
                      Human: {m.human_vote}
                    </span>
                  )}
                </div>
              </div>

              {/* Expanded details */}
              {isExpanded && (
                <div style={{ padding: '0 1.25rem 1.25rem', borderTop: '1px solid var(--border-subtle)', paddingTop: '1rem' }}>
                  {/* Prompt */}
                  <div style={{ marginBottom: '1rem' }}>
                    <div style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-dim)', textTransform: 'uppercase', marginBottom: '0.3rem' }}>
                      Prompt
                    </div>
                    <div style={{ fontSize: '0.85rem', color: 'var(--text-main)', background: 'rgba(0,0,0,0.2)', padding: '0.75rem', borderRadius: '6px' }}>
                      {m.prompt_text}
                    </div>
                  </div>

                  {/* Side-by-side responses */}
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
                    <div style={{ background: 'rgba(0,0,0,0.15)', padding: '0.85rem', borderRadius: '6px' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.4rem', fontSize: '0.8rem', fontWeight: 600, color: 'var(--accent-light)' }}>
                        <span>{m.model_a_id}</span>
                        <span className="font-mono">{m.latency_a_ms}ms · {m.tokens_a} tok</span>
                      </div>
                      <div style={{ fontSize: '0.82rem', color: 'var(--text-muted)', lineHeight: 1.5, whiteSpace: 'pre-wrap' }}>
                        {m.response_a}
                      </div>
                    </div>

                    <div style={{ background: 'rgba(0,0,0,0.15)', padding: '0.85rem', borderRadius: '6px' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.4rem', fontSize: '0.8rem', fontWeight: 600, color: 'var(--accent-light)' }}>
                        <span>{m.model_b_id}</span>
                        <span className="font-mono">{m.latency_b_ms}ms · {m.tokens_b} tok</span>
                      </div>
                      <div style={{ fontSize: '0.82rem', color: 'var(--text-muted)', lineHeight: 1.5, whiteSpace: 'pre-wrap' }}>
                        {m.response_b}
                      </div>
                    </div>
                  </div>

                  {/* Dual Pass Reasoning */}
                  <div style={{ background: 'rgba(0,0,0,0.25)', padding: '0.85rem', borderRadius: '6px', fontSize: '0.8rem' }}>
                    <div style={{ fontWeight: 700, color: 'var(--text-dim)', textTransform: 'uppercase', fontSize: '0.7rem', marginBottom: '0.5rem' }}>
                      LLM Judge Dual-Pass Audit
                    </div>
                    <div style={{ marginBottom: '0.4rem' }}>
                      <strong style={{ color: 'var(--text-muted)' }}>Pass 1 (Order A/B): </strong>
                      <span style={{ color: 'var(--warning)' }}>Winner = {m.pass1_winner}</span> — {m.pass1_reasoning}
                    </div>
                    <div>
                      <strong style={{ color: 'var(--text-muted)' }}>Pass 2 (Order B/A swapped): </strong>
                      <span style={{ color: 'var(--warning)' }}>Winner = {m.pass2_winner}</span> — {m.pass2_reasoning}
                    </div>
                  </div>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};
