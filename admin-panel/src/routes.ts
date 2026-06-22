export const routes = {
  login: '/admin/login',
  dashboard: '/admin',
  users: '/admin/users',
  user: (id: string) => `/admin/users/${id}`,
  subscriptions: '/admin/subscriptions',
  chat: '/admin/chat',
  profileRoasts: '/admin/profile-roasts',
  llm: '/admin/llm',
} as const;
