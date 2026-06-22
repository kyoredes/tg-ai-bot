import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';
import { ServicesStatusPanel } from '../components/ServicesStatusPanel';

export function DashboardPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['stats'],
    queryFn: async () => (await api.getStats()).stats,
  });

  return (
    <div>
      <h1 className="page-title">Dashboard</h1>

      <ServicesStatusPanel />

      {isLoading && <div className="loading">Loading stats...</div>}
      {error && <div className="error">Failed to load stats</div>}

      {!isLoading && !error && (
        <div className="card-grid">
          <div className="card">
            <div className="stat-value">{data?.users.total ?? 0}</div>
            <div className="stat-label">Total Users</div>
          </div>
          <div className="card">
            <div className="stat-value">{data?.users.new7d ?? 0}</div>
            <div className="stat-label">New Users (7d)</div>
          </div>
          <div className="card">
            <div className="stat-value">{data?.subscriptions.active ?? 0}</div>
            <div className="stat-label">Active Subscriptions</div>
          </div>
          <div className="card">
            <div className="stat-value">{data?.subscriptions.expired ?? 0}</div>
            <div className="stat-label">Expired Subscriptions</div>
          </div>
          <div className="card">
            <div className="stat-value">{data?.chat.sessions ?? 0}</div>
            <div className="stat-label">Chat Sessions</div>
          </div>
          <div className="card">
            <div className="stat-value">{data?.profileRoasts.sessions ?? 0}</div>
            <div className="stat-label">Profile Roast Sessions</div>
          </div>
        </div>
      )}
    </div>
  );
}
