import React, { useState, useEffect } from 'react';
import { BlindMatch, VoteReveal, fetchNextArenaMatch, submitArenaVote } from '../api/client';
import { EyeOff, ThumbsUp, Scale, RefreshCw, CheckCircle, ShieldAlert, AlertCircle, SplitSquareVertical } from 'lucide-react';
import { computeWordDiff } from '../utils/diff';

export const BlindArenaVote: React.FC = () => {
  const [match, setMatch] = useState<BlindMatch | null>(null);
  const [reveal, setReveal] = useState<VoteReveal | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [submitting, setSubmitting] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [showDiff, setShowDiff] = useState<boolean>(false);

  const loadNext = async () => {
    setLoading(true);
    setReveal(null);
    setError(null);
    setShowDiff(false);
    try {
      const data = await fetchNextArenaMatch();
      setMatch(data);
    } catch (err: any) {
      setError(err.message || 'No matches found in database. Run a tournament first!');
      setMatch(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadNext();
  }, []);

  const handleVote = async (vote: string) => {
    if (!match || submitting) return;
    setSubmitting(true);
    try {
      const rev = await submitArenaVote(match.match_id, vote);
      setReveal(rev);
    } catch (err: any) {
      setError(err.message || 'Failed to submit vote');
    } finally {
      setSubmitting(false);
    }
  };

  const diffResult = match ? computeWordDiff(match.response_a, match.response_b) : null;

  if (loading) {
    return (
      <div style={{ padding: '4rem', textAlign: 'center', color: 'var(--text-muted)' }}>
        <RefreshCw className="animate-spin" style={{ width: 32, height: 32, margin: '0 auto 1rem', color: 'var(--accent)' }} />
        <div>Loading next anonymized pairwise completion...</div>
      </div>
    );
  }

  if (error || !match) {
    return (
      <div style={{ padding: '3rem', textAlign: 'center', background: 'var(--bg-card)', borderRadius: '12px', border: '1px solid var(--border-subtle)' }}>
        <AlertCircle style={{ width: 36, height: 36, margin: '0 auto 1rem', color: 'var(--warning)' }} />
        <h3 style={{ fontSize: '1.1rem', fontWeight: 600, marginBottom: '0.5rem' }}>No Arena Match Available</h3>
        <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', maxWidth: 440, margin: '0 auto 1.5rem' }}>
          {error || 'Run a tournament round first to populate the database with model responses.'}
        </p>
        <button
          onClick={loadNext}
          style={{
            padding: '0.6rem 1.25rem',
            background: 'var(--accent)',
            color: '#fff',
            borderRadius: '6px',
            border: 'none',
            fontWeight: 600,
            cursor: 'pointer',
          }}
        >
          Check Again
        </button>
      </div>
    );
  }

  return (
    <div className="responsive-card" style={{ background: 'var(--bg-card)', borderRadius: '12px', border: '1px solid var(--border-subtle)', padding: '1.75rem' }}>
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h2 style={{ fontSize: '1.2rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <EyeOff style={{ width: 22, height: 22, color: 'var(--accent)' }} />
            Internal Blind Voting Arena
          </h2>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '0.2rem' }}>
            Compare completions side-by-side without knowing model identities. Vote objectively based on prompt adherence.
          </p>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <button
            onClick={() => setShowDiff(!showDiff)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.35rem',
              fontSize: '0.75rem',
              padding: '0.35rem 0.75rem',
              borderRadius: '4px',
              background: showDiff ? 'rgba(99,102,241,0.2)' : 'rgba(255,255,255,0.04)',
              color: showDiff ? 'var(--accent-light)' : 'var(--text-muted)',
              border: showDiff ? '1px solid var(--accent)' : '1px solid var(--border-subtle)',
              cursor: 'pointer',
              fontWeight: 600,
            }}
          >
            <SplitSquareVertical style={{ width: 14, height: 14 }} />
            {showDiff ? 'Diff View Active' : 'Toggle Diff View'}
          </button>

          {match.category && (
            <span style={{ fontSize: '0.75rem', color: 'var(--accent-light)', background: 'rgba(99, 102, 241, 0.15)', padding: '0.3rem 0.6rem', borderRadius: '4px', fontWeight: 600 }}>
              {match.category}
            </span>
          )}

          <div style={{ fontSize: '0.8rem', color: 'var(--text-dim)', background: 'var(--bg-subtle)', padding: '0.3rem 0.75rem', borderRadius: '6px' }}>
            Match: <code className="font-mono">{match.match_id}</code>
          </div>
        </div>
      </div>

      {/* User Prompt Box */}
      <div style={{ background: 'var(--bg-subtle)', borderRadius: '8px', padding: '1rem 1.25rem', marginBottom: '1.5rem', borderLeft: '4px solid var(--accent)' }}>
        <div style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--accent-light)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '0.4rem' }}>
          Enterprise Benchmark Prompt
        </div>
        <div style={{ fontSize: '0.95rem', color: 'var(--text-main)', lineHeight: 1.5 }}>
          {match.prompt}
        </div>
      </div>

      {/* Side-by-side response cards */}
      <div className="arena-grid-2col" style={{ marginBottom: '1.5rem' }}>
        {/* Option A */}
        <div style={{
          background: 'rgba(255, 255, 255, 0.02)',
          border: reveal ? (reveal.vote === 'model_a' ? '2px solid var(--success)' : '1px solid var(--border-subtle)') : '1px solid var(--border-subtle)',
          borderRadius: '8px',
          padding: '1.25rem',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'space-between',
        }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
              <span style={{ fontWeight: 700, fontSize: '0.9rem', color: 'var(--text-main)' }}>
                {reveal ? `Model A: ${reveal.model_a_id}` : 'Response Option A'}
              </span>
              {reveal && (
                <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }} className="font-mono">
                  {reveal.model_a_latency_ms}ms
                </span>
              )}
            </div>
            <div style={{ fontSize: '0.9rem', color: 'var(--text-muted)', lineHeight: 1.6, whiteSpace: 'pre-wrap' }}>
              {showDiff && diffResult ? (
                diffResult.diffA.map((t, idx) => (
                  <span
                    key={idx}
                    style={{
                      background: t.type === 'removed' ? 'rgba(239,68,68,0.2)' : 'transparent',
                      color: t.type === 'removed' ? '#fca5a5' : 'inherit',
                      borderRadius: 2,
                    }}
                  >
                    {t.value}
                  </span>
                ))
              ) : (
                match.response_a
              )}
            </div>
          </div>

          {!reveal && (
            <button
              disabled={submitting}
              onClick={() => handleVote('model_a')}
              style={{
                marginTop: '1.25rem',
                padding: '0.65rem 1rem',
                background: 'rgba(99, 102, 241, 0.15)',
                color: 'var(--accent-light)',
                border: '1px solid rgba(99, 102, 241, 0.3)',
                borderRadius: '6px',
                fontWeight: 600,
                fontSize: '0.85rem',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '0.4rem',
                transition: 'all 0.15s ease',
              }}
            >
              <ThumbsUp style={{ width: 14, height: 14 }} /> Response A is Better
            </button>
          )}
        </div>

        {/* Option B */}
        <div style={{
          background: 'rgba(255, 255, 255, 0.02)',
          border: reveal ? (reveal.vote === 'model_b' ? '2px solid var(--success)' : '1px solid var(--border-subtle)') : '1px solid var(--border-subtle)',
          borderRadius: '8px',
          padding: '1.25rem',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'space-between',
        }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
              <span style={{ fontWeight: 700, fontSize: '0.9rem', color: 'var(--text-main)' }}>
                {reveal ? `Model B: ${reveal.model_b_id}` : 'Response Option B'}
              </span>
              {reveal && (
                <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }} className="font-mono">
                  {reveal.model_b_latency_ms}ms
                </span>
              )}
            </div>
            <div style={{ fontSize: '0.9rem', color: 'var(--text-muted)', lineHeight: 1.6, whiteSpace: 'pre-wrap' }}>
              {showDiff && diffResult ? (
                diffResult.diffB.map((t, idx) => (
                  <span
                    key={idx}
                    style={{
                      background: t.type === 'added' ? 'rgba(16,185,129,0.2)' : 'transparent',
                      color: t.type === 'added' ? '#6ee7b7' : 'inherit',
                      borderRadius: 2,
                    }}
                  >
                    {t.value}
                  </span>
                ))
              ) : (
                match.response_b
              )}
            </div>
          </div>

          {!reveal && (
            <button
              disabled={submitting}
              onClick={() => handleVote('model_b')}
              style={{
                marginTop: '1.25rem',
                padding: '0.65rem 1rem',
                background: 'rgba(99, 102, 241, 0.15)',
                color: 'var(--accent-light)',
                border: '1px solid rgba(99, 102, 241, 0.3)',
                borderRadius: '6px',
                fontWeight: 600,
                fontSize: '0.85rem',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '0.4rem',
                transition: 'all 0.15s ease',
              }}
            >
              <ThumbsUp style={{ width: 14, height: 14 }} /> Response B is Better
            </button>
          )}
        </div>
      </div>

      {/* Auxiliary Voting Controls (Tie / Both Bad) */}
      {!reveal && (
        <div style={{ display: 'flex', justifyContent: 'center', gap: '1rem', marginTop: '1rem' }}>
          <button
            disabled={submitting}
            onClick={() => handleVote('tie')}
            style={{
              padding: '0.5rem 1rem',
              background: 'transparent',
              color: 'var(--text-muted)',
              border: '1px solid var(--border-subtle)',
              borderRadius: '6px',
              fontSize: '0.85rem',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
            }}
          >
            <Scale style={{ width: 14, height: 14 }} /> Both are Equivalent (Tie)
          </button>
          <button
            disabled={submitting}
            onClick={() => handleVote('both_bad')}
            style={{
              padding: '0.5rem 1rem',
              background: 'transparent',
              color: 'var(--danger)',
              border: '1px solid rgba(239, 68, 68, 0.2)',
              borderRadius: '6px',
              fontSize: '0.85rem',
              cursor: 'pointer',
            }}
          >
            Both are Poor
          </button>
        </div>
      )}

      {/* Post-Vote Reveal Card */}
      {reveal && (
        <div style={{
          marginTop: '1.5rem',
          background: 'var(--bg-subtle)',
          borderRadius: '8px',
          padding: '1.25rem',
          border: '1px solid var(--border-focus)',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--success)', fontWeight: 700 }}>
              <CheckCircle style={{ width: 18, height: 18 }} />
              Vote Recorded! Model Identities Revealed.
            </div>
            <button
              onClick={loadNext}
              style={{
                padding: '0.5rem 1.25rem',
                background: 'var(--accent)',
                color: '#fff',
                border: 'none',
                borderRadius: '6px',
                fontWeight: 600,
                fontSize: '0.85rem',
                cursor: 'pointer',
              }}
            >
              Next Pair →
            </button>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '0.75rem', fontSize: '0.85rem' }}>
            <div>
              <span style={{ color: 'var(--text-dim)' }}>Your Blind Vote: </span>
              <strong style={{ color: 'var(--text-main)', textTransform: 'uppercase' }}>{reveal.vote}</strong>
            </div>
            <div>
              <span style={{ color: 'var(--text-dim)' }}>Automated Judge Verdict: </span>
              <strong style={{ color: 'var(--warning)' }}>{reveal.judge_verdict}</strong>
            </div>
            <div>
              <span style={{ color: 'var(--text-dim)' }}>Judge Position Bias: </span>
              <strong style={{ color: reveal.bias_detected ? 'var(--warning)' : 'var(--success)' }}>
                {reveal.bias_detected ? 'Detected & Mitigated' : 'Clean (Consistent)'}
              </strong>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
