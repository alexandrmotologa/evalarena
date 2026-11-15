import React, { useState, useEffect } from 'react';
import { Header } from './components/Header';
import { LeaderboardTable } from './components/LeaderboardTable';
import { ParetoFrontier } from './components/ParetoFrontier';
import { BlindArenaVote } from './components/BlindArenaVote';
import { WinRateHeatmap } from './components/WinRateHeatmap';
import { MatchInspector } from './components/MatchInspector';
import {
  ModelRating,
  WinRateMatrix,
  ParetoPoint,
  Recommendation,
  MatchRecord,
  fetchLeaderboard,
  fetchMatrix,
  fetchPareto,
  fetchMatches,
  startTournament,
  subscribeToEvents,
} from './api/client';

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'leaderboard' | 'arena' | 'matrix' | 'matches'>('leaderboard');
  const [ratings, setRatings] = useState<ModelRating[]>([]);
  const [matrix, setMatrix] = useState<WinRateMatrix | null>(null);
  const [paretoPoints, setParetoPoints] = useState<ParetoPoint[]>([]);
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [matches, setMatches] = useState<MatchRecord[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [isRunning, setIsRunning] = useState<boolean>(false);
  const [notification, setNotification] = useState<string | null>(null);

  const loadAllData = async () => {
    try {
      const [lb, mtx, par, mtchs] = await Promise.all([
        fetchLeaderboard().catch(() => []),
        fetchMatrix().catch(() => null),
        fetchPareto().catch(() => ({ points: [], recommendations: [] })),
        fetchMatches().catch(() => []),
      ]);

      setRatings(lb);
      setMatrix(mtx);
      setParetoPoints(par.points || []);
      setRecommendations(par.recommendations || []);
      setMatches(mtchs);
    } catch (err) {
      console.error('Error fetching tournament data:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAllData();

    // Subscribe to real-time Server-Sent Events
    const unsubscribe = subscribeToEvents((event) => {
      if (event.type === 'match_start') {
        setNotification(`Round in progress: ${event.payload.model_a} vs ${event.payload.model_b}`);
      } else if (event.type === 'match_finish') {
        loadAllData();
      } else if (event.type === 'tournament_complete') {
        setIsRunning(false);
        setNotification('Tournament round completed successfully!');
        loadAllData();
        setTimeout(() => setNotification(null), 5000);
      }
    });

    return () => {
      unsubscribe();
    };
  }, []);

  const handleRunTournament = async () => {
    if (isRunning) return;
    setIsRunning(true);
    setNotification('Dispatching tournament round across model providers...');
    try {
      await startTournament();
    } catch (err: any) {
      setNotification(`Failed to start tournament: ${err.message}`);
      setIsRunning(false);
    }
  };

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <Header
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        onRunTournament={handleRunTournament}
        isRunning={isRunning}
        totalMatches={matches.length}
      />

      {/* Live notification pill */}
      {notification && (
        <div style={{
          background: 'var(--bg-subtle)',
          borderBottom: '1px solid var(--border-focus)',
          padding: '0.5rem 1rem',
          textAlign: 'center',
          fontSize: '0.82rem',
          color: 'var(--accent-light)',
          fontWeight: 600,
        }}>
          {notification}
        </div>
      )}

      {/* Main Container */}
      <main style={{ maxWidth: 1380, width: '100%', margin: '0 auto', padding: '2rem 1.5rem', flex: 1 }}>
        {activeTab === 'leaderboard' && (
          <div>
            <ParetoFrontier points={paretoPoints} recommendations={recommendations} />
            <LeaderboardTable ratings={ratings} loading={loading} />
          </div>
        )}

        {activeTab === 'arena' && (
          <div>
            <BlindArenaVote />
          </div>
        )}

        {activeTab === 'matrix' && (
          <div>
            <WinRateHeatmap matrix={matrix} />
          </div>
        )}

        {activeTab === 'matches' && (
          <div>
            <MatchInspector matches={matches} />
          </div>
        )}
      </main>

      {/* Footer */}
      <footer style={{
        borderTop: '1px solid var(--border-subtle)',
        padding: '1.25rem 2rem',
        textAlign: 'center',
        fontSize: '0.8rem',
        color: 'var(--text-dim)',
      }}>
        evalarena — Self-Hosted Pairwise Blind Model Tournament & Bradley-Terry ELO Benchmarking Platform
      </footer>
    </div>
  );
};
