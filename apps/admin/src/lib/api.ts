import type {
  User,
  Branding,
  TenantMember,
  InviteToken,
  UsageSummary,
  APIKey,
  EgressPolicy,
} from '../types'

const BASE = '/api'

function getToken(): string {
  return localStorage.getItem('sessionToken') ?? ''
}

function getSlug(): string {
  return localStorage.getItem('tenantSlug') ?? ''
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  authType: 'session' | 'apikey' | 'none' = 'session',
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  if (authType === 'session') {
    const token = getToken()
    if (token) headers['Authorization'] = `Bearer ${token}`
  } else if (authType === 'apikey') {
    const key = localStorage.getItem('apiKey') ?? ''
    if (key) headers['Authorization'] = `Bearer ${key}`
  }

  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error ?? res.statusText)
  }

  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  // Auth
  async login(email: string, password: string): Promise<{ session_token: string; expires_at: string }> {
    return request('POST', '/v1/admin/auth/login', { email, password }, 'none')
  },

  async register(
    email: string,
    password: string,
    inviteToken?: string,
  ): Promise<{ user: User; session_token: string }> {
    return request('POST', '/v1/admin/auth/register', { email, password, invite_token: inviteToken }, 'none')
  },

  async logout(): Promise<void> {
    return request('POST', '/v1/admin/auth/logout', undefined, 'session')
  },

  async me(): Promise<User> {
    return request('GET', '/v1/admin/auth/me', undefined, 'session')
  },

  // Branding
  async getBranding(): Promise<Branding> {
    return request('GET', `/v1/tenants/${getSlug()}/branding`, undefined, 'apikey')
  },

  async updateBranding(b: Branding): Promise<Branding> {
    return request('PUT', `/v1/tenants/${getSlug()}/branding`, b, 'apikey')
  },

  // Members
  async listMembers(): Promise<TenantMember[]> {
    return request('GET', `/v1/tenants/${getSlug()}/members`, undefined, 'apikey')
  },

  async inviteMember(email: string, role: string): Promise<{ invite_token: string; expires_at: string }> {
    return request('POST', `/v1/tenants/${getSlug()}/members/invite`, { email, role }, 'apikey')
  },

  async removeMember(userID: string): Promise<void> {
    return request('DELETE', `/v1/tenants/${getSlug()}/members/${userID}`, undefined, 'apikey')
  },

  async listInvites(): Promise<InviteToken[]> {
    return request('GET', `/v1/tenants/${getSlug()}/members/invites`, undefined, 'apikey')
  },

  // Usage
  async getUsage(window: string): Promise<UsageSummary> {
    return request('GET', `/v1/tenants/${getSlug()}/usage?window=${window}`, undefined, 'apikey')
  },

  // API Keys
  async listAPIKeys(): Promise<APIKey[]> {
    return request('GET', `/v1/tenants/${getSlug()}/api-keys`, undefined, 'apikey')
  },

  async createAPIKey(name: string, scopes: string[]): Promise<APIKey> {
    return request('POST', `/v1/tenants/${getSlug()}/api-keys`, { name, scopes }, 'apikey')
  },

  async deleteAPIKey(id: string): Promise<void> {
    return request('DELETE', `/v1/tenants/${getSlug()}/api-keys/${id}`, undefined, 'apikey')
  },

  // Egress
  async getEgressPolicy(): Promise<EgressPolicy> {
    return request('GET', `/v1/tenants/${getSlug()}/egress`, undefined, 'apikey')
  },

  async updateEgressPolicy(p: EgressPolicy): Promise<EgressPolicy> {
    return request('PUT', `/v1/tenants/${getSlug()}/egress`, p, 'apikey')
  },
}
