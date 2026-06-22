import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { api } from '../api/client';
import { formatDate } from '../utils/format';

export function ProfileRoastsPage() {
  const [page, setPage] = useState(1);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const { data: sessions, isLoading } = useQuery({
    queryKey: ['profile-roast-sessions', page],
    queryFn: () => api.listProfileRoastSessions(page, 20),
  });

  const { data: history, isLoading: historyLoading } = useQuery({
    queryKey: ['profile-roast-history', selectedId],
    queryFn: async () => (await api.getProfileRoastHistory(selectedId!)).history,
    enabled: !!selectedId,
  });

  const handleClear = async (telegramId: string) => {
    if (!confirm('Clear profile roast history for this user?')) return;
    await api.clearProfileRoastHistory(telegramId);
    queryClient.invalidateQueries({ queryKey: ['profile-roast-sessions'] });
    queryClient.invalidateQueries({ queryKey: ['profile-roast-history', telegramId] });
    if (selectedId === telegramId) setSelectedId(null);
  };

  return (
    <div>
      <h1 className="page-title">Profile Roasts</h1>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
        <div className="card table-wrap">
          <h3>Sessions</h3>
          {isLoading ? (
            <div className="loading">Loading...</div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>Telegram ID</th>
                  <th>Roasts</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {sessions?.sessions.map((s) => (
                  <tr key={s.telegramID}>
                    <td>{s.telegramID}</td>
                    <td>{s.roastCount}</td>
                    <td className="actions">
                      <button
                        type="button"
                        className="btn btn-secondary btn-sm"
                        onClick={() => setSelectedId(s.telegramID)}
                      >
                        View
                      </button>
                      <button
                        type="button"
                        className="btn btn-danger btn-sm"
                        onClick={() => handleClear(s.telegramID)}
                      >
                        Clear
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
          <div className="pagination">
            <button
              type="button"
              className="btn btn-secondary btn-sm"
              disabled={page <= 1}
              onClick={() => setPage((p) => p - 1)}
            >
              Prev
            </button>
            <span>Page {page}</span>
            <button
              type="button"
              className="btn btn-secondary btn-sm"
              disabled={!sessions || page * 20 >= sessions.total}
              onClick={() => setPage((p) => p + 1)}
            >
              Next
            </button>
          </div>
        </div>

        <div className="card">
          <h3>Roasts {selectedId && `— ${selectedId}`}</h3>
          {!selectedId && <div className="loading">Select a session</div>}
          {selectedId && historyLoading && <div className="loading">Loading...</div>}
          {history?.roasts.map((roast, i) => (
            <div key={i} className="message message-assistant">
              <div className="message-role">
                {formatDate(roast.createdAt)}
                {roast.username ? ` · @${roast.username}` : ''}
                {roast.hasPhoto ? ' · photo' : ''}
              </div>
              {(roast.firstName || roast.bio) && (
                <div style={{ fontSize: '0.85rem', opacity: 0.8, marginBottom: '0.5rem' }}>
                  {[roast.firstName, roast.lastName].filter(Boolean).join(' ')}
                  {roast.bio ? ` — ${roast.bio}` : ''}
                </div>
              )}
              <div style={{ whiteSpace: 'pre-wrap' }}>{roast.response}</div>
            </div>
          ))}
          {selectedId && history && history.roasts.length === 0 && (
            <div className="loading">No roasts</div>
          )}
        </div>
      </div>
    </div>
  );
}
