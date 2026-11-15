import React, { useState } from 'react';
import { QuickBattleRequest, QuickBattleResponse, quickBattle } from '../api/client';
import { Swords, Play, RefreshCw, ShieldAlert, CheckCircle2, SplitSquareVertical, Clock, DollarSign, Sparkles } from 'lucide-react';
import { computeWordDiff } from '../utils/diff';

interface Props {
  models: string[];
}

export const QuickBattle: React.FC<Props> = ({ models }) => {
  const [modelA, setModelA] = useState<string>(models[0] || 'gpt-4o');
  const [modelB, setModelB] = useState<string>(models[1] || 'gemini-2.0-flash');
  const [prompt, setPrompt] = useState<string>(
    'A customer wants to return an item purchased 16 days ago (14-day policy limit). Draft a polite refusal explaining the policy and offer a 15% discount on their next order.'
  );
  const [systemPrompt, setSystemPrompt] = useState<string>('');
  const [rubric, setRubric] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);
  const [result, setResult] = useState<QuickBattleResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [showDiff, setShowDiff] = useState<boolean>(false);

  const availableModels = models.length > 0 ? models : ['gpt-4o', 'claude-3-5-sonnet', 'gemini-2.0-flash', 'llama-3.1-8b-instruct'];

  const handleExecute = async () => {
    if (!prompt.trim() || !modelA || !modelB) return;
    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const data = await quickBattle({
        prompt,
        system_prompt: systemPrompt,
        model_a_id: modelA,
        model_b_id: modelB,
        rubric,
      });
      setResult(data);
    } catch (err: any) {
      setError(err.message || 'Battle execution failed');
    } finally {
      setLoading(false);
    }
  };

  const diffResult = result ? computeWordDiff(result.response_a, result.response_b) : null;

  return (
    <div className="responsive-card" style={{ background: 'var(--bg-card)', borderRadius: '12px', border: '1px solid var(--border-subtle)', padding: '1.75rem' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h2 style={{ fontSize: '1.2rem', fontWeight: 700, display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Swords style={{ width: 22, height: 22, color: 'var(--accent)' }} />
            Ad-hoc Quick Battle Playground
          </h2>
          <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginTop: '0.2rem' }}>
            Test prompts interactively across any two foundation models with live double-blind LLM judge debiasing.
          </p>
        </div>
      </div>

      {/* Model Selection and Setup Grid */}
      <div className="arena-grid-2col" style={{ marginBottom: '1rem' }}>
        <div>
          <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-dim)', textTransform: 'uppercase', marginBottom: '0.3rem' }}>
            Contender Model A
          </label>
          <select
            value={modelA}
            onChange={(e) => setModelA(e.target.value)}
            style={{
              width: '100%',
              padding: '0.6rem 0.75rem',
              background: 'var(--bg-subtle)',
              border: '1px solid var(--border-subtle)',
              borderRadius: '6px',
              color: 'var(--text-main)',
              fontSize: '0.9rem',
              fontWeight: 600,
            }}
          >
            {availableModels.map(m => (
              <option key={`a-${m}`} value={m}>{m}</option>
            ))}
          </select>
        </div>

        <div>
          <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-dim)', textTransform: 'uppercase', marginBottom: '0.3rem' }}>
            Contender Model B
          </label>
          <select
            value={modelB}
            onChange={(e) => setModelB(e.target.value)}
            style={{
              width: '100%',
              padding: '0.6rem 0.75rem',
              background: 'var(--bg-subtle)',
              border: '1px solid var(--border-subtle)',
              borderRadius: '6px',
              color: 'var(--text-main)',
              fontSize: '0.9rem',
              fontWeight: 600,
            }}
          >
            {availableModels.map(m => (
              <option key={`b-${m}`} value={m}>{m}</option>
            ))}
          </select>
        </div>
      </div>

      {/* Prompt Textarea */}
      <div style={{ marginBottom: '1.25rem' }}>
        <label style={{ display: 'block', fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-dim)', textTransform: 'uppercase', marginBottom: '0.3rem' }}>
          Evaluation Prompt
        </label>
        <textarea
          rows={3}
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          placeholder="Enter prompt to evaluate..."
          style={{
            width: '100%',
            padding: '0.75rem',
            background: 'var(--bg-subtle)',
            border: '1px solid var(--border-subtle)',
            borderRadius: '6px',
            color: 'var(--text-main)',
            fontSize: '0.9rem',
            lineHeight: 1.5,
            resize: 'vertical',
          }}
        />
      </div>

      {/* Submit Button */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '1.5rem' }}>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button
            type="button"
            onClick={() => setPrompt('Explain how to prevent Server-Side Request Forgery (SSRF) in a Go webhook dispatcher service.')}
            style={{
              fontSize: '0.75rem',
              padding: '0.3rem 0.6rem',
              background: 'rgba(255,255,255,0.04)',
              color: 'var(--text-muted)',
              border: '1px solid var(--border-subtle)',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Preset: SSRF Security
          </button>
          <button
            type="button"
            onClick={() => setPrompt('Write a PostgreSQL query for 90-day churn cohort analysis with window functions.')}
            style={{
              fontSize: '0.75rem',
              padding: '0.3rem 0.6rem',
              background: 'rgba(255,255,255,0.04)',
              color: 'var(--text-muted)',
              border: '1px solid var(--border-subtle)',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
          >
            Preset: SQL Churn
          </button>
        </div>

        <button
          disabled={loading || !prompt.trim()}
          onClick={handleExecute}
          style={{
            padding: '0.65rem 1.4rem',
            background: loading ? 'var(--bg-subtle)' : 'var(--accent)',
            color: '#fff',
            border: 'none',
            borderRadius: '6px',
            fontWeight: 700,
            fontSize: '0.9rem',
            cursor: loading ? 'not-allowed' : 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
            boxShadow: '0 4px 14px rgba(99,102,241,0.3)',
          }}
        >
          {loading ? (
            <>
              <RefreshCw className="animate-spin" style={{ width: 16, height: 16 }} />
              Executing Battle...
            </>
          ) : (
            <>
              <Play style={{ width: 16, height: 16 }} />
              Run Head-to-Head Battle
            </>
          )}
        </button>
      </div>

      {error && (
        <div style={{ padding: '0.75rem 1rem', background: 'rgba(239,68,68,0.1)', border: '1px solid rgba(239,68,68,0.3)', borderRadius: '6px', color: 'var(--danger)', fontSize: '0.85rem', marginBottom: '1.5rem' }}>
          {error}
        </div>
      )}

      {/* Battle Results */}
      {result && (
        <div>
          {/* Controls header */}
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
            <span style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-muted)' }}>
              Side-by-Side Completions
            </span>
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
          </div>

          <div className="arena-grid-2col" style={{ marginBottom: '1.25rem' }}>
            {/* Contender A output */}
            <div style={{
              background: 'rgba(255,255,255,0.02)',
              border: result.final_winner === result.model_a_id ? '2px solid var(--success)' : '1px solid var(--border-subtle)',
              borderRadius: '8px',
              padding: '1.25rem',
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
                <span style={{ fontWeight: 700, color: 'var(--text-main)', fontSize: '0.95rem' }}>
                  {result.model_a_id}
                </span>
                <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }} className="font-mono">
                  {result.latency_a_ms}ms · {result.tokens_a} tok · ${result.cost_a.toFixed(5)}
                </span>
              </div>
              <div style={{ fontSize: '0.88rem', color: 'var(--text-muted)', lineHeight: 1.6, whiteSpace: 'pre-wrap' }}>
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
                  result.response_a
                )}
              </div>
            </div>

            {/* Contender B output */}
            <div style={{
              background: 'rgba(255,255,255,0.02)',
              border: result.final_winner === result.model_b_id ? '2px solid var(--success)' : '1px solid var(--border-subtle)',
              borderRadius: '8px',
              padding: '1.25rem',
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
                <span style={{ fontWeight: 700, color: 'var(--text-main)', fontSize: '0.95rem' }}>
                  {result.model_b_id}
                </span>
                <span style={{ fontSize: '0.75rem', color: 'var(--text-dim)' }} className="font-mono">
                  {result.latency_b_ms}ms · {result.tokens_b} tok · ${result.cost_b.toFixed(5)}
                </span>
              </div>
              <div style={{ fontSize: '0.88rem', color: 'var(--text-muted)', lineHeight: 1.6, whiteSpace: 'pre-wrap' }}>
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
                  result.response_b
                )}
              </div>
            </div>
          </div>

          {/* Double-Blind LLM Judge Audit Card */}
          <div style={{
            background: 'var(--bg-subtle)',
            borderRadius: '8px',
            padding: '1.25rem',
            border: '1px solid var(--border-focus)',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.75rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontWeight: 700, fontSize: '0.95rem' }}>
                <CheckCircle2 style={{ width: 18, height: 18, color: 'var(--success)' }} />
                <span>Authoritative Verdict: </span>
                <strong style={{ color: 'var(--warning)', fontSize: '1.05rem' }}>
                  {result.final_winner === 'tie' ? 'Tie / Equal' : result.final_winner}
                </strong>
              </div>

              {result.position_bias_detected ? (
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.3rem', fontSize: '0.75rem', color: 'var(--warning)', background: 'rgba(245,158,11,0.1)', padding: '0.2rem 0.5rem', borderRadius: '4px' }}>
                  <ShieldAlert style={{ width: 13, height: 13 }} /> Position Bias Detected & Mitigated
                </span>
              ) : (
                <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.3rem', fontSize: '0.75rem', color: 'var(--success)' }}>
                  ✓ Consistent Across Order Inversion
                </span>
              )}
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.82rem', color: 'var(--text-muted)' }}>
              <div>
                <strong style={{ color: 'var(--text-main)' }}>Pass 1 (Direct Order A/B): </strong>
                Winner was <strong style={{ color: 'var(--warning)' }}>{result.pass1_winner}</strong> — {result.pass1_reasoning}
              </div>
              <div>
                <strong style={{ color: 'var(--text-main)' }}>Pass 2 (Swapped Order B/A): </strong>
                Winner was <strong style={{ color: 'var(--warning)' }}>{result.pass2_winner}</strong> — {result.pass2_reasoning}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
