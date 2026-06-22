import { routes } from '../routes';

const BASE = import.meta.env.VITE_GATEWAY_URL ?? '/api';

const TOKEN_KEY = 'admin_token';

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(`${BASE}${path}`, { ...options, headers });

  if (response.status === 401) {
    clearToken();
    window.location.href = routes.login;
    throw new Error('Unauthorized');
  }

  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error ?? 'Request failed');
  }
  return data as T;
}

export interface AdminStats {
  users: { total: number; new7d: number };
  subscriptions: { total: number; active: number; expired: number };
  chat: { sessions: number };
  profileRoasts: { sessions: number };
}

export interface ServiceStatus {
  id: string;
  name: string;
  status: 'up' | 'down' | 'degraded';
  latencyMs?: number;
}

export interface ServicesStatusResponse {
  services: ServiceStatus[];
  checkedAt: number;
}

export interface AdminUser {
  userID: string;
  email: string;
  telegramID: string;
  createdAt: number;
}

export interface AdminUserDetail extends AdminUser {
  updatedAt: number;
}

export interface AdminSubscription {
  subscriptionID: string;
  userID: string;
  startsAt: number;
  expiresAt: number;
}

export interface ChatMessage {
  role: string;
  content: string;
}

export interface ChatSession {
  telegramID: string;
  messageCount: number;
}

export interface ProfileRoastItem {
  createdAt: number;
  firstName: string;
  lastName?: string;
  username?: string;
  bio?: string;
  isPremium: boolean;
  languageCode?: string;
  hasPhoto: boolean;
  response: string;
}

export interface ProfileRoastSession {
  telegramID: string;
  roastCount: number;
}

export interface LLMConfig {
  model: string;
  temperature: number;
  maxTokens: number;
  debug: boolean;
  provider: string;
  g4fModels: string[];
  usesLiteLLM: boolean;
}

export interface SystemPrompt {
  prompt: string;
  defaultPrompt: string;
  isCustom: boolean;
}

export const api = {
  login: (username: string, password: string) =>
    request<{ token: string }>('/admin/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),

  getStats: () => request<{ stats: AdminStats }>('/admin/stats'),

  getServicesStatus: () =>
    request<{ servicesStatus: ServicesStatusResponse }>('/admin/services'),

  listUsers: (page = 1, limit = 20, search = '') => {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) });
    if (search) params.set('search', search);
    return request<{ users: AdminUser[]; total: number }>(`/admin/users?${params}`);
  },

  getUser: (id: string) => request<{ user: AdminUserDetail }>(`/admin/users/${id}`),

  updateUser: (id: string, email: string) =>
    request<{ user: AdminUserDetail }>(`/admin/users/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ email }),
    }),

  deleteUser: (id: string) =>
    request<{ status: string }>(`/admin/users/${id}`, { method: 'DELETE' }),

  listSubscriptions: (page = 1, limit = 20, status = '') => {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) });
    if (status) params.set('status', status);
    return request<{ subscriptions: AdminSubscription[]; total: number }>(
      `/admin/subscriptions?${params}`,
    );
  },

  updateSubscription: (id: string, startsAt: number, expiresAt: number) =>
    request<{ subscription: AdminSubscription }>(`/admin/subscriptions/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ startsAt, expiresAt }),
    }),

  deleteSubscription: (id: string) =>
    request<{ status: string }>(`/admin/subscriptions/${id}`, { method: 'DELETE' }),

  listChatSessions: (page = 1, limit = 20) => {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) });
    return request<{ sessions: ChatSession[]; total: number }>(
      `/admin/chat/sessions?${params}`,
    );
  },

  getChatHistory: (telegramId: string) =>
    request<{ history: { telegramID: string; messages: ChatMessage[] } }>(
      `/admin/chat/history/${telegramId}`,
    ),

  clearChatHistory: (telegramId: string) =>
    request<{ status: string }>(`/admin/chat/history/${telegramId}`, { method: 'DELETE' }),

  listProfileRoastSessions: (page = 1, limit = 20) => {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) });
    return request<{ sessions: ProfileRoastSession[]; total: number }>(
      `/admin/profile-roasts/sessions?${params}`,
    );
  },

  getProfileRoastHistory: (telegramId: string) =>
    request<{ history: { telegramID: string; roasts: ProfileRoastItem[] } }>(
      `/admin/profile-roasts/history/${telegramId}`,
    ),

  clearProfileRoastHistory: (telegramId: string) =>
    request<{ status: string }>(`/admin/profile-roasts/history/${telegramId}`, {
      method: 'DELETE',
    }),

  getLLMConfig: () => request<{ config: LLMConfig }>('/admin/llm'),

  getSystemPrompt: () => request<{ systemPrompt: SystemPrompt }>('/admin/llm/prompt'),

  updateSystemPrompt: (prompt: string) =>
    request<{ systemPrompt: SystemPrompt }>('/admin/llm/prompt', {
      method: 'PATCH',
      body: JSON.stringify({ prompt }),
    }),
};
