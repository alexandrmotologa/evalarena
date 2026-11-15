import React from 'react';
import { Play, RefreshCw, Trophy, EyeOff, Grid, ListFilter, Activity } from 'lucide-react';

interface Props {
  activeTab: 'leaderboard' | 'arena' | 'matrix' | 'matches';
  setActiveTab: (tab: 'leaderboard' | 'arena' | 'matrix' | 'matches') => void;
  onRunTournament: () => void;
  isRunning: boolean;
  totalMatches: number;
}

export const Header: React.FC<Props> = ({ activeTab, setActiveTab, onRunTournament, isRunning, totalMatches }) => {
  return (
    <header style={{
      background: 'var(--bg-card)',
      borderBottom: '1px solid var(--border-subtle)',
      padding: '1.25rem 2rem',
      position: 'sticky',
      top: 0,
      zIndex: 50,
      backdropFilter: 'blur(8px)',
    }}>
      <div style={{ maxWidth: 1380, margin: '0 auto', display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '1rem' }}>
        {/* Brand */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <div style={{
            width: 38,
            height: 38,
            borderRadius: '8px',
            background: 'linear-gradient(135deg, #6366f1, #4338ca)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 0 16px rgba(99,102,241,0.4)',
          }}>
            <Trophy style={{ width: 20, height: 20, color: '#ffffff' }} />
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <span style={{ fontSize: '1.2rem', fontWeight: 800, letterSpacing: '-0.02em', color: '#ffffff' }}>
                evalarena
              </span>
              <span style={{
                fontSize: '0.7rem',
                fontWeight: 600,
                padding: '0.15rem 0.45rem',
                borderRadius: '4px',
                background: 'rgba(99, 102, 241, 0.15)',
                color: 'var(--accent-light)',
                border: '1px solid rgba(99, 102, 241, 0.3)',
              }}>
                STUDIO
              </span>
            </div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              Self-Hosted Blind Pairwise Model Preference Tournament
            </div>
          </div>
        </div>

        {/* Tab Navigation */}
        <nav style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', background: 'var(--bg-main)', padding: '0.25rem', borderRadius: '8px', border: '1px solid var(--border-subtle)' }}>
          <button
            onClick={() => setActiveTab('leaderboard')}
            style={{
              padding: '0.45rem 0.85rem',
              borderRadius: '6px',
              border: 'none',
              background: activeTab === 'leaderboard' ? 'var(--bg-subtle)' : 'transparent',
              color: activeTab === 'leaderboard' ? '#ffffff' : 'var(--text-muted)',
              fontWeight: 600,
              fontSize: '0.85rem',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              transition: 'all 0.15s ease',
            }}
          >
            <Trophy style={{ width: 14, height: 14 }} /> Leaderboard & Pareto
          </button>

          <button
            onClick={() => setActiveTab('arena')}
            style={{
              padding: '0.45rem 0.85rem',
              borderRadius: '6px',
              border: 'none',
              background: activeTab === 'arena' ? 'var(--bg-subtle)' : 'transparent',
              color: activeTab === 'arena' ? '#ffffff' : 'var(--text-muted)',
              fontWeight: 600,
              fontSize: '0.85rem',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              transition: 'all 0.15s ease',
            }}
          >
            <EyeOff style={{ width: 14, height: 14 }} /> Blind Arena
          </button>

          <button
            onClick={() => setActiveTab('matrix')}
            style={{
              padding: '0.45rem 0.85rem',
              borderRadius: '6px',
              border: 'none',
              background: activeTab === 'matrix' ? 'var(--bg-subtle)' : 'transparent',
              color: activeTab === 'matrix' ? '#ffffff' : 'var(--text-muted)',
              fontWeight: 600,
              fontSize: '0.85rem',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              transition: 'all 0.15s ease',
            }}
          >
            <Grid style={{ width: 14, height: 14 }} /> Win-Rate Matrix
          </button>

          <button
            onClick={() => setActiveTab('matches')}
            style={{
              padding: '0.45rem 0.85rem',
              borderRadius: '6px',
              border: 'none',
              background: activeTab === 'matches' ? 'var(--bg-subtle)' : 'transparent',
              color: activeTab === 'matches' ? '#ffffff' : 'var(--text-muted)',
              fontWeight: 600,
              fontSize: '0.85rem',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              transition: 'all 0.15s ease',
            }}
          >
            <ListFilter style={{ width: 14, height: 14 }} /> Match Inspector
          </button>
        </nav>

        {/* Action Controls */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontSize: '0.8rem', color: 'var(--text-dim)' }}>
            <Activity style={{ width: 14, height: 14, color: 'var(--success)' }} />
            <span>{totalMatches} battles</span>
          </div>

          <button
            disabled={isRunning}
            onClick={onRunTournament}
            style={{
              padding: '0.55rem 1.1rem',
              background: isRunning ? 'var(--bg-subtle)' : 'var(--accent)',
              color: '#ffffff',
              border: 'none',
              borderRadius: '6px',
              fontWeight: 600,
              fontSize: '0.85rem',
              cursor: isRunning ? 'not-allowed' : 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '0.4rem',
              boxShadow: isRunning ? 'none' : '0 2px 10px rgba(99,102,241,0.3)',
              transition: 'all 0.15s ease',
            }}
          >
            {isRunning ? (
              <>
                <RefreshCw className="animate-spin" style={{ width: 14, height: 14 }} />
                Executing Round...
              </>
            ) : (
              <>
                <Play style={{ width: 14, height: 14 }} />
                Run Tournament Round
              </>
            )}
          </button>
        </div>
      </div>
    </header>
  );
};
