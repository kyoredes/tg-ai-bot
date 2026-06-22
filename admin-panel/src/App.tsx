import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { Layout } from './components/Layout';
import { ProtectedRoute } from './components/ProtectedRoute';
import { AuthProvider } from './hooks/useAuth';
import { ChatHistoryPage } from './pages/ChatHistory';
import { DashboardPage } from './pages/Dashboard';
import { LLMConfigPage } from './pages/LLMConfig';
import { LoginPage } from './pages/Login';
import { ProfileRoastsPage } from './pages/ProfileRoasts';
import { SubscriptionsPage } from './pages/Subscriptions';
import { UserDetailPage } from './pages/UserDetail';
import { UsersPage } from './pages/Users';
import { routes } from './routes';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30_000 },
  },
});

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<Navigate to={routes.dashboard} replace />} />
            <Route path="/login" element={<Navigate to={routes.login} replace />} />
            <Route path={routes.login} element={<LoginPage />} />
            <Route path="/admin" element={<ProtectedRoute />}>
              <Route element={<Layout />}>
                <Route index element={<DashboardPage />} />
                <Route path="users" element={<UsersPage />} />
                <Route path="users/:id" element={<UserDetailPage />} />
                <Route path="subscriptions" element={<SubscriptionsPage />} />
                <Route path="chat" element={<ChatHistoryPage />} />
                <Route path="profile-roasts" element={<ProfileRoastsPage />} />
                <Route path="llm" element={<LLMConfigPage />} />
              </Route>
            </Route>
            <Route path="*" element={<Navigate to={routes.dashboard} replace />} />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </QueryClientProvider>
  );
}
