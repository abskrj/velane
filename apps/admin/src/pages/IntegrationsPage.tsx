import { useState, useEffect, useMemo } from 'react'
import { Plug, Search, CheckCircle2, Loader2, X, Pencil, Plus } from 'lucide-react'
import Nango from '@nangohq/frontend'
import { api } from '../lib/api'
import type { Connection, NangoProvider, IntegrationConfig } from '../types'

// Auth modes that require operator OAuth app setup before users can connect.
const OAUTH_MODES = new Set(['OAUTH2', 'OAUTH2_CC', 'OAUTH1', 'APP', 'MCP_OAUTH2', 'MCP_OAUTH2_GENERIC'])

function isOAuthMode(authMode: string) {
  return OAUTH_MODES.has(authMode)
}

type ManualField = {
  key: string
  field: {
    title: string
    description?: string
    example?: string
    optional?: boolean
    prefix?: string
    secret?: boolean
  }
}

const FALLBACK_FIELDS_BY_MODE: Record<string, ManualField[]> = {
  OAUTH2: [
    {
      key: 'client_id',
      field: {
        title: 'Client ID',
        description: 'OAuth app client ID',
      },
    },
    {
      key: 'client_secret',
      field: {
        title: 'Client Secret',
        description: 'OAuth app client secret',
        secret: true,
      },
    },
    {
      key: 'scopes',
      field: {
        title: 'Scopes',
        optional: true,
      },
    },
  ],
  OAUTH2_CC: [
    {
      key: 'client_id',
      field: {
        title: 'Client ID',
        description: 'Client credentials app client ID',
      },
    },
    {
      key: 'client_secret',
      field: {
        title: 'Client Secret',
        description: 'Client credentials app client secret',
        secret: true,
      },
    },
    {
      key: 'scopes',
      field: {
        title: 'Scopes',
        optional: true,
      },
    },
  ],
  BASIC: [
    {
      key: 'username',
      field: {
        title: 'Username',
      },
    },
    {
      key: 'password',
      field: {
        title: 'Password',
        secret: true,
      },
    },
  ],
  APP: [
    {
      key: 'app_id',
      field: {
        title: 'App ID',
        description: 'Provider application ID',
      },
    },
    {
      key: 'app_link',
      field: {
        title: 'App Link',
        description: 'Public URL of your application in the provider dashboard',
      },
    },
    {
      key: 'private_key',
      field: {
        title: 'Private Key (RSA PEM)',
        description: 'Paste the full RSA private key including BEGIN/END lines',
        secret: true,
      },
    },
  ],
  CUSTOM: [
    {
      key: 'client_id',
      field: {
        title: 'Client ID',
      },
    },
    {
      key: 'client_secret',
      field: {
        title: 'Client Secret',
        secret: true,
      },
    },
    {
      key: 'app_id',
      field: {
        title: 'App ID',
      },
    },
    {
      key: 'app_link',
      field: {
        title: 'App Link',
      },
    },
    {
      key: 'private_key',
      field: {
        title: 'Private Key (RSA PEM)',
        secret: true,
      },
    },
  ],
}

// ---- Configure Modal -------------------------------------------------------

interface ConfigureModalProps {
  provider: NangoProvider
  existing?: IntegrationConfig
  callbackUrl: string
  onClose: () => void
  onSaved: () => void
}

function ConfigureModal({ provider, existing, callbackUrl, onClose, onSaved }: ConfigureModalProps) {
  const [alias, setAlias] = useState(existing?.alias ?? 'default')
  const [profileName, setProfileName] = useState(existing?.name ?? '')
  const [isEditingName, setIsEditingName] = useState(false)
  const [extraFields, setExtraFields] = useState<Record<string, string>>({})
  const [saving, setSaving]         = useState(false)
  const [error, setError]           = useState<string | null>(null)

  // Non-automated credentials fields the operator can provide for this profile.
  const mode = provider.auth_mode.toUpperCase().trim()
  const providerCredentialFields = provider.credentials ? Object.entries(provider.credentials) : []
  const providerConnectionConfigFields = provider.connection_config ? Object.entries(provider.connection_config) : []
  const mergedSchemaFields = useMemo(() => {
    const seen = new Set<string>()
    const merged: ManualField[] = []
    const includeProviderCredentials = true
    if (includeProviderCredentials) {
      for (const [key, field] of providerCredentialFields) {
        if ((field as any).automated) continue
        if (seen.has(key)) continue
        merged.push({ key, field })
        seen.add(key)
      }
    }
    for (const [key, field] of providerConnectionConfigFields) {
      if ((field as any).automated) continue
      if (seen.has(key)) continue
      merged.push({ key, field })
      seen.add(key)
    }
    if (includeProviderCredentials && (mode === 'OAUTH2' || mode === 'OAUTH2_CC')) {
      const hasClientID = seen.has('client_id') || seen.has('oauth_client_id')
      const hasClientSecret = seen.has('client_secret') || seen.has('oauth_client_secret')
      if (!hasClientID || !hasClientSecret) {
        for (const entry of FALLBACK_FIELDS_BY_MODE[mode] ?? []) {
          if (seen.has(entry.key)) continue
          merged.push(entry)
          seen.add(entry.key)
        }
      }
    }
    return merged
  }, [provider.auth_mode, providerCredentialFields, providerConnectionConfigFields, mode])
  const effectiveCredentialFields = mergedSchemaFields

  const configFields = useMemo(() => {
    return effectiveCredentialFields
      .filter((entry) => !entry.field.optional)
      .map((entry) => ({ key: entry.key, field: entry.field }))
  }, [effectiveCredentialFields])

  const optionalConfigFields = useMemo(() => {
    return effectiveCredentialFields
      .filter((entry) => Boolean(entry.field.optional))
      .map((entry) => ({ key: entry.key, field: entry.field }))
  }, [effectiveCredentialFields])

  const credentialsType = useMemo(() => {
    const mode = provider.auth_mode.toUpperCase().trim()
    if (!mode) return 'OAUTH2'
    return mode
  }, [provider.auth_mode])

  const defaultScopeValue = useMemo(() => {
    if (existing?.oauth_scopes?.trim()) return existing.oauth_scopes.trim()
    const scopes = provider.default_scopes ?? []
    if (scopes.length === 0) return ''
    return scopes.join(' ')
  }, [existing?.oauth_scopes, provider.default_scopes])

  useEffect(() => {
    if (!defaultScopeValue) return
    const hasScopeField = effectiveCredentialFields.some((entry) => entry.key === 'scopes' || entry.key === 'oauth_scopes')
    if (!hasScopeField) return
    setExtraFields((prev) => {
      if ((prev.scopes ?? '').trim() || (prev.oauth_scopes ?? '').trim()) {
        return prev
      }
      if (prev.scopes !== undefined) {
        return { ...prev, scopes: defaultScopeValue }
      }
      if (prev.oauth_scopes !== undefined) {
        return { ...prev, oauth_scopes: defaultScopeValue }
      }
      return { ...prev, scopes: defaultScopeValue }
    })
  }, [defaultScopeValue, effectiveCredentialFields])

  async function handleSave() {
    setSaving(true)
    setError(null)
    try {
      await api.configureIntegration({
        provider: provider.unique_key,
        alias: alias || 'default',
        name: profileName || alias || 'default',
        credentials_type: credentialsType,
        credentials: Object.fromEntries(
          Object.entries(extraFields).filter(([, value]) => value.trim() !== ''),
        ),
        is_default: existing?.is_default ?? false,
      })
      onSaved()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      <div className="fixed inset-0 z-40 bg-black/30" onClick={onClose} />
      <div className="fixed inset-0 z-50 flex items-start justify-center p-4 sm:items-center sm:p-6">
        <div className="flex w-full max-w-lg max-h-[88vh] flex-col overflow-hidden rounded-xl bg-white shadow-xl">
          {/* Header */}
          <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
            <div>
              <p className="text-sm font-semibold text-gray-900">
                Configure {provider.name}
              </p>
              <p className="text-xs text-gray-500 mt-0.5">
                {isOAuthMode(provider.auth_mode)
                  ? 'Enter your OAuth app credentials'
                  : 'Register this integration so users can connect'}
              </p>
            </div>
            <div className="flex items-center gap-2">
              {isEditingName ? (
                <input
                  type="text"
                  value={profileName}
                  onChange={e => setProfileName(e.target.value)}
                  onBlur={() => setIsEditingName(false)}
                  onKeyDown={e => {
                    if (e.key === 'Enter' || e.key === 'Escape') {
                      setIsEditingName(false)
                    }
                  }}
                  placeholder={alias || 'default'}
                  autoFocus
                  className="w-44 rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 focus:border-gray-400 focus:outline-none"
                />
              ) : (
                <button
                  type="button"
                  onClick={() => setIsEditingName(true)}
                  className="flex items-center gap-1 rounded-md px-2 py-1 text-xs text-gray-600 hover:bg-gray-100"
                  title="Edit profile name"
                >
                  <span className="max-w-40 truncate">{profileName || alias || 'default'}</span>
                  <Pencil size={12} />
                </button>
              )}
              <button onClick={onClose} className="rounded-md p-1 text-gray-400 hover:bg-gray-100">
                <X size={16} />
              </button>
            </div>
          </div>

          {/* Form */}
          <div className="flex-1 space-y-4 overflow-y-auto overflow-x-hidden px-6 py-5">
            {error && (
              <p className="rounded-lg bg-red-50 px-3 py-2 text-xs text-red-600">{error}</p>
            )}

            <Field
              label="Credential alias"
              required
              value={alias}
              onChange={setAlias}
              placeholder="default / sandbox / prod"
              description="Alias used when connecting and in integration('provider', { alias })"
            />

            {isOAuthMode(provider.auth_mode) && callbackUrl && (
              <div className="rounded-lg border border-blue-100 bg-blue-50 px-4 py-3 space-y-1">
                <p className="text-xs font-medium text-blue-800">
                  OAuth redirect URL
                </p>
                <p className="text-xs text-blue-700">
                  Register this as the callback / redirect URI in your {provider.name} OAuth app settings before pasting credentials below.
                </p>
                <div className="flex items-center gap-2 mt-1">
                  <code className="flex-1 rounded bg-white border border-blue-200 px-2 py-1 text-xs font-mono text-blue-900 select-all break-all">
                    {callbackUrl}
                  </code>
                  <button
                    type="button"
                    onClick={() => navigator.clipboard.writeText(callbackUrl)}
                    className="shrink-0 rounded px-2 py-1 text-xs text-blue-700 hover:bg-blue-100"
                  >
                    Copy
                  </button>
                </div>
              </div>
            )}

            {configFields.map(({ key, field }) => (
              <Field
                key={key}
                label={field.title || key}
                required={!field.optional}
                value={extraFields[key] ?? ''}
                onChange={v => setExtraFields(prev => ({ ...prev, [key]: v }))}
                placeholder={field.example ? `e.g. ${field.example}` : undefined}
                description={field.description}
                prefix={field.prefix}
                secret={Boolean((field as any).secret)}
              />
            ))}

            {optionalConfigFields.map(({ key, field }) => (
              <Field
                key={key}
                label={field.title || key}
                value={extraFields[key] ?? ''}
                onChange={v => setExtraFields(prev => ({ ...prev, [key]: v }))}
                placeholder={field.example ? `e.g. ${field.example}` : undefined}
                description={field.description}
                prefix={field.prefix}
                secret={Boolean((field as any).secret)}
              />
            ))}

            {configFields.length === 0 && optionalConfigFields.length === 0 && (
              <p className="text-xs text-gray-500">
                {isOAuthMode(provider.auth_mode)
                  ? 'No manual setup fields were provided for this provider. Save this profile, then click Connect.'
                  : 'No configuration needed. Users will enter their credentials when they connect.'}
              </p>
            )}

            {provider.docs && (
              <a
                href={provider.docs}
                target="_blank"
                rel="noopener noreferrer"
                className="block text-xs text-indigo-600 hover:underline"
              >
                View {provider.name} developer docs →
              </a>
            )}
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-2 border-t border-gray-200 px-6 py-4">
            <button
              onClick={onClose}
              className="rounded-lg px-4 py-2 text-sm text-gray-600 hover:bg-gray-100"
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              disabled={saving}
              className="flex items-center gap-1.5 rounded-lg bg-gray-900 px-4 py-2 text-sm text-white hover:bg-gray-800 disabled:opacity-50"
            >
              {saving && <Loader2 size={13} className="animate-spin" />}
              Save
            </button>
          </div>
        </div>
      </div>
    </>
  )
}

interface FieldProps {
  label: string
  value: string
  onChange: (v: string) => void
  required?: boolean
  secret?: boolean
  placeholder?: string
  description?: string
  prefix?: string
}

function Field({ label, value, onChange, required, secret, placeholder, description, prefix }: FieldProps) {
  return (
    <div>
      <label className="mb-1 block text-xs font-medium text-gray-700">
        {label} {required && <span className="text-red-500">*</span>}
      </label>
      <div className="flex items-center rounded-lg border border-gray-200 focus-within:border-gray-400">
        {prefix && (
          <span className="select-none border-r border-gray-200 bg-gray-50 px-3 py-2 text-xs text-gray-500">
            {prefix}
          </span>
        )}
        <input
          type={secret ? 'password' : 'text'}
          value={value}
          onChange={e => onChange(e.target.value)}
          placeholder={placeholder}
          className="w-full rounded-lg bg-transparent px-3 py-2 text-sm text-gray-900 placeholder-gray-400 focus:outline-none"
        />
      </div>
      {description && <p className="mt-1 text-xs text-gray-400">{description}</p>}
    </div>
  )
}

type IntegrationProfileRow = {
  id: string
  providerKey: string
  providerName: string
  providerAuthMode: string
  providerLogoURL?: string
  alias: string
  profileName: string
  isDefault: boolean
  config: IntegrationConfig
  connection?: Connection
  status: 'connected' | 'configured'
}

interface AddIntegrationModalProps {
  providers: NangoProvider[]
  onSelectProvider: (provider: NangoProvider) => void
  onClose: () => void
}

function AddIntegrationModal({ providers, onSelectProvider, onClose }: AddIntegrationModalProps) {
  const [search, setSearch] = useState('')
  const [results, setResults] = useState<NangoProvider[]>(providers.slice(0, 6))
  const [loadingResults, setLoadingResults] = useState(false)
  const [resultsError, setResultsError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    const timer = setTimeout(async () => {
      setLoadingResults(true)
      try {
        const list = await api.listProviders(search, 6)
        if (cancelled) return
        setResults(list)
        setResultsError(null)
      } catch (e: any) {
        if (cancelled) return
        setResultsError(e?.message ?? 'Failed to load integrations')
      } finally {
        if (!cancelled) setLoadingResults(false)
      }
    }, 200)

    return () => {
      cancelled = true
      clearTimeout(timer)
    }
  }, [search])

  return (
    <>
      <div className="fixed inset-0 z-40 bg-black/30" onClick={onClose} />
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div className="w-full max-w-2xl rounded-xl bg-white shadow-xl">
          <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
            <h2 className="text-base font-semibold text-gray-900">Add Integration</h2>
            <button onClick={onClose} className="rounded-md p-1 text-gray-400 hover:bg-gray-100">
              <X size={16} />
            </button>
          </div>

          <div className="p-6">
            <div className="relative">
              <Search size={16} className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400" />
              <input
                autoFocus
                type="text"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search integrations..."
                className="w-full rounded-xl border border-gray-300 py-4 pl-12 pr-4 text-base text-gray-900 placeholder-gray-400 focus:border-gray-400 focus:outline-none"
              />
            </div>

            <div className="mt-4 max-h-[55vh] overflow-y-auto rounded-lg border border-gray-200">
              {resultsError && (
                <p className="px-4 py-3 text-sm text-red-600">{resultsError}</p>
              )}
              {loadingResults && (
                <p className="px-4 py-3 text-sm text-gray-500">Searching...</p>
              )}
              {!loadingResults && results.length === 0 ? (
                <p className="px-4 py-6 text-sm text-gray-500">No matching integrations.</p>
              ) : (
                <div className="divide-y divide-gray-100">
                  {results.map((provider) => (
                    <button
                      key={provider.unique_key}
                      type="button"
                      onClick={() => onSelectProvider(provider)}
                      className="flex w-full items-center justify-between px-4 py-3 text-left hover:bg-gray-50"
                    >
                      <div className="flex min-w-0 items-center gap-2.5">
                        {provider.logo_url ? (
                          <img
                            src={provider.logo_url}
                            alt={provider.name}
                            className="h-6 w-6 shrink-0 object-contain"
                          />
                        ) : (
                          <Plug size={14} className="shrink-0 text-gray-400" />
                        )}
                        <div className="min-w-0">
                          <p className="truncate text-sm font-medium text-gray-900">{provider.name}</p>
                          <p className="truncate text-xs text-gray-500">{provider.unique_key} · {provider.auth_mode}</p>
                        </div>
                      </div>
                      <Plus size={14} className="shrink-0 text-gray-400" />
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

type ProfilesTab = 'connected' | 'ready' | 'all'

// ---- Main Page -------------------------------------------------------------

export default function IntegrationsPage() {
  const PAGE_SIZE = 10
  const [providers,    setProviders]    = useState<NangoProvider[]>([])
  const [configs,      setConfigs]      = useState<IntegrationConfig[]>([])
  const [loading,      setLoading]      = useState(true)
  const [tableLoading, setTableLoading] = useState(false)
  const [busy,         setBusy]         = useState<string | null>(null)
  const [error,        setError]        = useState<string | null>(null)
  const [showAddModal, setShowAddModal] = useState(false)
  const [configModal,  setConfigModal]  = useState<{ provider: NangoProvider; existing?: IntegrationConfig } | null>(null)
  const [callbackUrl,  setCallbackUrl]  = useState('')
  const [hasAnyProfiles, setHasAnyProfiles] = useState(false)
  const [activeTab, setActiveTab] = useState<ProfilesTab>('connected')
  const [connectedPage, setConnectedPage] = useState(1)
  const [readyPage, setReadyPage] = useState(1)
  const [allPage, setAllPage] = useState(1)
  const [hasNextByTab, setHasNextByTab] = useState<Record<ProfilesTab, boolean>>({
    connected: false,
    ready: false,
    all: false,
  })

  const currentPageByTab: Record<ProfilesTab, number> = {
    connected: connectedPage,
    ready: readyPage,
    all: allPage,
  }
  const setCurrentPageByTab: Record<ProfilesTab, (page: number) => void> = {
    connected: setConnectedPage,
    ready: setReadyPage,
    all: setAllPage,
  }

  async function reload(tab = activeTab, page = currentPageByTab[tab]) {
    const status = tab === 'ready' ? 'configured' : tab
    const offset = (Math.max(1, page) - 1) * PAGE_SIZE
    const [p, cfg, anyConfigured] = await Promise.all([
      providers.length > 0 ? Promise.resolve(providers) : api.listProviders(),
      api.listConfigured(undefined, PAGE_SIZE + 1, offset, status),
      api.listConfigured(undefined, 1, 0, 'all'),
    ])
    setProviders(p)
    setConfigs(cfg.slice(0, PAGE_SIZE))
    setHasAnyProfiles(anyConfigured.length > 0)
    setHasNextByTab((prev) => ({
      ...prev,
      [tab]: cfg.length > PAGE_SIZE,
    }))
  }

  useEffect(() => {
    Promise.all([reload('connected', 1), api.getConnectInfo().then(info => setCallbackUrl(info.oauth_callback_url))])
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    if (loading) return
    setTableLoading(true)
    reload(activeTab, currentPageByTab[activeTab])
      .catch((e: any) => setError(e?.message ?? 'Something went wrong'))
      .finally(() => setTableLoading(false))
  }, [activeTab, connectedPage, readyPage, allPage])

  const providerByKey = useMemo(() => new Map(providers.map((p) => [p.unique_key, p])), [providers])
  const profileRows = useMemo<IntegrationProfileRow[]>(() => {
    const rows = configs.map((config) => {
      const provider = providerByKey.get(config.provider)
      const status: IntegrationProfileRow['status'] = config.connected ? 'connected' : 'configured'
      return {
        id: config.id,
        providerKey: config.provider,
        providerName: provider?.name ?? config.provider,
        providerAuthMode: provider?.auth_mode ?? config.credentials_type,
        providerLogoURL: provider?.logo_url,
        alias: config.alias,
        profileName: config.name?.trim() || config.alias,
        isDefault: config.is_default,
        config,
        connection: undefined,
        status,
      }
    })
    return rows.sort((a, b) => {
      if (a.profileName !== b.profileName) return a.profileName.localeCompare(b.profileName)
      return a.alias.localeCompare(b.alias)
    })
  }, [configs, providerByKey])

  const activePage = currentPageByTab[activeTab]
  const hasNextPage = hasNextByTab[activeTab]

  async function handleConnect(row: IntegrationProfileRow) {
    setBusy(`${row.providerKey}::${row.alias}::connect`)
    setError(null)
    try {
      const { session_token, api_url, credential_profile_id, alias } =
        await api.createConnectionSession(row.providerKey, row.alias, row.config.id)
      const nango = new Nango({
        connectSessionToken: session_token,
        host: api_url,
      })
      try {
        await nango.auth(row.config.nango_provider_config_key, {
          detectClosedAuthWindow: true,
        })
      } catch (err: any) {
        if (err?.type === 'window_closed' || err?.message?.toLowerCase().includes('window')) {
          throw new Error('cancelled')
        }
        throw err
      }
      await api.recordConnection(row.providerKey, '', alias, credential_profile_id)
      await reload()
    } catch (e: any) {
      if (e?.message !== 'cancelled') setError(e?.message ?? 'Something went wrong')
    } finally {
      setBusy(null)
    }
  }

  async function handleDisconnect(row: IntegrationProfileRow) {
    setBusy(`${row.providerKey}::${row.alias}::disconnect`)
    setError(null)
    try {
      await api.disconnectProvider(row.providerKey)
      await reload()
    } catch (e: any) {
      setError(e?.message ?? 'Something went wrong')
    } finally {
      setBusy(null)
    }
  }

  async function handleRemove(row: IntegrationProfileRow) {
    setBusy(`${row.providerKey}::${row.alias}::remove`)
    setError(null)
    try {
      await api.deleteIntegrationConfig(row.config.id)
      await reload()
    } catch (e: any) {
      setError(e?.message ?? 'Something went wrong')
    } finally {
      setBusy(null)
    }
  }

  function openAddModal() {
    setShowAddModal(true)
  }

  function openConfigureProvider(provider: NangoProvider, existing?: IntegrationConfig) {
    setConfigModal({ provider, existing })
  }

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 size={20} className="animate-spin text-gray-400" />
      </div>
    )
  }

  return (
    <div>
      <div className="mb-6 flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold text-gray-900">Integrations</h1>
          <p className="mt-1 text-sm text-gray-500">
            Manage saved integration profiles by name and alias, then connect each profile separately.
          </p>
        </div>
        <button
          type="button"
          onClick={openAddModal}
          className="inline-flex items-center gap-1.5 rounded-lg bg-gray-900 px-3 py-2 text-sm text-white hover:bg-gray-800"
        >
          <Plus size={14} />
          Add Integration
        </button>
      </div>

      <p className="mb-6 text-sm text-gray-500">
        Use <code className="rounded bg-gray-100 px-1 py-0.5 text-xs">{`integration('provider', { alias: "..." })`}</code> in snippets.
      </p>

      {error && (
        <div className="mb-4 flex items-center gap-2 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          <span className="flex-1">{error}</span>
          <button onClick={() => setError(null)}><X size={14} /></button>
        </div>
      )}

      {!hasAnyProfiles && (
        <div className="rounded-xl border border-dashed border-gray-300 bg-gray-50 px-6 py-8 text-center">
          <h3 className="text-sm font-semibold text-gray-900">No integration profiles yet</h3>
          <p className="mt-1 text-sm text-gray-500">Create your first profile to configure credentials and connect.</p>
          <button
            type="button"
            onClick={openAddModal}
            className="mt-4 inline-flex items-center gap-1.5 rounded-lg bg-gray-900 px-3 py-2 text-sm text-white hover:bg-gray-800"
          >
            <Plus size={14} />
            Add Integration
          </button>
        </div>
      )}

      {hasAnyProfiles && (
        <div className="rounded-lg border border-gray-200 bg-white shadow-sm">
          <div className="flex items-center gap-2 border-b border-gray-200 px-4 py-3">
            <button
              type="button"
              onClick={() => setActiveTab('connected')}
              className={`rounded-md px-3 py-1.5 text-sm font-medium ${
                activeTab === 'connected' ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              Connected
            </button>
            <button
              type="button"
              onClick={() => setActiveTab('ready')}
              className={`rounded-md px-3 py-1.5 text-sm font-medium ${
                activeTab === 'ready' ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              Ready to Connect
            </button>
            <button
              type="button"
              onClick={() => setActiveTab('all')}
              className={`rounded-md px-3 py-1.5 text-sm font-medium ${
                activeTab === 'all' ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              All Profiles
            </button>
          </div>

          <div className="min-h-[calc(100vh-340px)]">
            <div className="h-full overflow-auto">
              <table className="w-full text-sm">
                <thead className="bg-gray-50 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  <tr>
                    <th className="px-4 py-3">Profile</th>
                    <th className="px-4 py-3">Provider</th>
                    <th className="px-4 py-3">Alias</th>
                    <th className="px-4 py-3">Auth</th>
                    <th className="px-4 py-3">Status</th>
                    <th className="px-4 py-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {tableLoading && (
                    <tr>
                      <td colSpan={6} className="px-4 py-8 text-center text-sm text-gray-500">
                        <span className="inline-flex items-center gap-2">
                          <Loader2 size={14} className="animate-spin text-gray-400" />
                          Loading profiles...
                        </span>
                      </td>
                    </tr>
                  )}
                  {!tableLoading && profileRows.length === 0 && (
                    <tr>
                      <td colSpan={6} className="px-4 py-8 text-center text-sm text-gray-500">
                        No profiles in this tab.
                      </td>
                    </tr>
                  )}
                  {!tableLoading && profileRows.map((row) => (
                    <tr key={row.id} className="hover:bg-gray-50">
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-2">
                          {row.providerLogoURL ? (
                            <img src={row.providerLogoURL} alt={row.providerName} className="h-6 w-6 shrink-0 object-contain" />
                          ) : (
                            <Plug size={14} className="shrink-0 text-gray-400" />
                          )}
                          <div className="min-w-0">
                            <p className="truncate font-medium text-gray-900">{row.profileName}</p>
                            {row.isDefault && <p className="text-xs text-gray-500">Default</p>}
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-gray-700">{row.providerName}</td>
                      <td className="px-4 py-3 font-mono text-xs text-gray-600">{row.alias}</td>
                      <td className="px-4 py-3 text-gray-600">{row.providerAuthMode}</td>
                      <td className="px-4 py-3">
                        {row.status === 'connected' ? (
                          <span className="inline-flex items-center gap-1 text-green-600">
                            <CheckCircle2 size={13} />
                            Connected
                          </span>
                        ) : (
                          <span className="text-blue-600">Ready</span>
                        )}
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex justify-end gap-1.5">
                          {row.status === 'connected' ? (
                            <button
                              onClick={() => handleDisconnect(row)}
                              disabled={busy?.startsWith(`${row.providerKey}::${row.alias}`) ?? false}
                              className="rounded-md border border-gray-200 px-2.5 py-1 text-xs text-gray-700 hover:border-red-200 hover:bg-red-50 hover:text-red-600 disabled:opacity-50"
                            >
                              Disconnect
                            </button>
                          ) : (
                            <button
                              onClick={() => handleConnect(row)}
                              disabled={busy?.startsWith(`${row.providerKey}::${row.alias}`) ?? false}
                              className="rounded-md bg-gray-900 px-2.5 py-1 text-xs text-white hover:bg-gray-800 disabled:opacity-50"
                            >
                              Connect
                            </button>
                          )}
                          <button
                            onClick={() => {
                              const provider = providerByKey.get(row.providerKey)
                              if (provider) openConfigureProvider(provider, row.config)
                            }}
                            disabled={busy?.startsWith(`${row.providerKey}::${row.alias}`) ?? false}
                            className="rounded-md border border-gray-200 px-2.5 py-1 text-xs text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => handleRemove(row)}
                            disabled={busy?.startsWith(`${row.providerKey}::${row.alias}`) ?? false}
                            className="rounded-md border border-gray-200 px-2.5 py-1 text-xs text-gray-700 hover:border-red-200 hover:bg-red-50 hover:text-red-600 disabled:opacity-50"
                          >
                            Remove
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          <div className="flex items-center justify-between border-t border-gray-200 px-4 py-3 text-sm">
            <p className="text-gray-500">
              Page {activePage}
            </p>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setCurrentPageByTab[activeTab](Math.max(1, activePage - 1))}
                disabled={tableLoading || activePage <= 1}
                className="rounded-md border border-gray-200 px-3 py-1 text-gray-700 hover:bg-gray-50 disabled:opacity-50"
              >
                Previous
              </button>
              <button
                type="button"
                onClick={() => setCurrentPageByTab[activeTab](activePage + 1)}
                disabled={tableLoading || !hasNextPage}
                className="rounded-md border border-gray-200 px-3 py-1 text-gray-700 hover:bg-gray-50 disabled:opacity-50"
              >
                Next
              </button>
            </div>
          </div>
        </div>
      )}

      {showAddModal && (
        <AddIntegrationModal
          providers={providers}
          onSelectProvider={(provider) => {
            setShowAddModal(false)
            openConfigureProvider(provider)
          }}
          onClose={() => setShowAddModal(false)}
        />
      )}

      {configModal && (
        <ConfigureModal
          provider={configModal.provider}
          existing={configModal.existing}
          callbackUrl={callbackUrl}
          onClose={() => setConfigModal(null)}
          onSaved={async () => {
            setConfigModal(null)
            await reload()
          }}
        />
      )}
    </div>
  )
}
