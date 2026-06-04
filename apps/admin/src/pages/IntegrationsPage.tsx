import { useState, useEffect, useMemo } from 'react'
import { Plug, Search, CheckCircle2, Loader2, X, Settings, ChevronDown, ChevronUp, Pencil } from 'lucide-react'
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
        description: 'Optional — leave blank to use provider defaults',
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
        description: 'Optional — leave blank to use provider defaults',
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
  const providerCredentialFields = provider.credentials
    ? Object.entries(provider.credentials)
    : []
  const providerConnectionConfigFields = provider.connection_config
    ? Object.entries(provider.connection_config)
    : []
  const shouldUseFallbackFields =
    providerCredentialFields.length === 0 &&
    providerConnectionConfigFields.length === 0 &&
    Boolean(FALLBACK_FIELDS_BY_MODE[mode])
  const effectiveCredentialFields = shouldUseFallbackFields
    ? FALLBACK_FIELDS_BY_MODE[mode]
    : providerCredentialFields.map(([key, field]) => ({ key, field }))

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

            {!isOAuthMode(provider.auth_mode) && configFields.length === 0 && (
              <p className="text-xs text-gray-500">
                No configuration needed. Users will enter their credentials when they connect.
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

// ---- Provider Card ---------------------------------------------------------

type CardAction = 'connect' | 'setup' | 'disconnect' | 'remove-config'

interface ProviderCardProps {
  provider: NangoProvider
  status: 'connected' | 'configured' | 'available'
  connectingOrBusy: boolean
  onAction: (action: CardAction, provider: NangoProvider) => void
}

function ProviderCard({ provider, status, connectingOrBusy, onAction }: ProviderCardProps) {
  const canConnect = status === 'configured' || !isOAuthMode(provider.auth_mode)
  const [imgError, setImgError] = useState(false)

  return (
    <div className="flex flex-col gap-3 rounded-lg border border-gray-200 bg-white p-4">
      {/* Header */}
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-center gap-2.5 min-w-0">
          {/* Logo with fallback */}
          {provider.logo_url && !imgError ? (
            <img
              src={provider.logo_url}
              alt={provider.name}
              className="h-8 w-8 shrink-0 object-contain"
              onError={() => setImgError(true)}
            />
          ) : (
            <Plug size={18} className="shrink-0 text-gray-400" />
          )}
          <div className="min-w-0 flex-1">
            <p className="text-sm font-medium text-gray-900 leading-tight" style={{
              display: '-webkit-box',
              WebkitLineClamp: 2,
              WebkitBoxOrient: 'vertical',
              overflow: 'hidden',
            }}>{provider.name}</p>
            <p className="text-xs text-gray-400 mt-0.5">{provider.auth_mode}</p>
          </div>
        </div>
        {status === 'connected' && (
          <CheckCircle2 size={14} className="mt-0.5 shrink-0 text-green-500" />
        )}
        {status === 'configured' && (
          <span className="shrink-0 rounded-full bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-600">
            Setup
          </span>
        )}
      </div>

      {/* Category */}
      {(provider.categories?.length ?? 0) > 0 && (
        <p className="text-xs capitalize text-gray-400">
          {provider.categories!.slice(0, 2).join(' · ').replace(/-/g, ' ')}
        </p>
      )}

      {/* Actions */}
      <div className="mt-auto flex gap-1.5">
        {status === 'connected' ? (
          <button
            onClick={() => onAction('disconnect', provider)}
            disabled={connectingOrBusy}
            className="flex-1 rounded-md border border-gray-200 px-3 py-1.5 text-xs font-medium text-gray-600 hover:border-red-200 hover:bg-red-50 hover:text-red-600 disabled:opacity-50"
          >
            {connectingOrBusy ? <Loader2 size={11} className="mx-auto animate-spin" /> : 'Disconnect'}
          </button>
        ) : canConnect ? (
          <>
            <button
              onClick={() => onAction('connect', provider)}
              disabled={connectingOrBusy}
              className="flex-1 rounded-md bg-gray-900 px-3 py-1.5 text-xs font-medium text-white hover:bg-gray-800 disabled:opacity-50"
            >
              {connectingOrBusy ? <Loader2 size={11} className="mx-auto animate-spin" /> : 'Connect'}
            </button>
            {status === 'configured' && (
              <button
                onClick={() => onAction('remove-config', provider)}
                className="rounded-md border border-gray-200 p-1.5 text-gray-400 hover:border-gray-300 hover:text-gray-600"
                title="Remove OAuth app config"
              >
                <X size={12} />
              </button>
            )}
          </>
        ) : (
          <button
            onClick={() => onAction('setup', provider)}
            disabled={connectingOrBusy}
            className="flex-1 flex items-center justify-center gap-1 rounded-md border border-gray-200 px-3 py-1.5 text-xs font-medium text-gray-600 hover:bg-gray-50 disabled:opacity-50"
          >
            <Settings size={11} />
            Setup
          </button>
        )}
      </div>
    </div>
  )
}

// ---- Main Page -------------------------------------------------------------

const ALL_CATEGORIES = [
  { key: 'all', label: 'All' },
  { key: 'crm', label: 'CRM' },
  { key: 'developer-tools', label: 'Dev Tools' },
  { key: 'communication', label: 'Communication' },
  { key: 'email', label: 'Email' },
  { key: 'calendar', label: 'Calendar' },
  { key: 'storage', label: 'Storage' },
  { key: 'payments', label: 'Payments' },
  { key: 'marketing', label: 'Marketing' },
  { key: 'hr', label: 'HR' },
  { key: 'analytics', label: 'Analytics' },
  { key: 'productivity', label: 'Productivity' },
]

export default function IntegrationsPage() {
  const [providers,    setProviders]    = useState<NangoProvider[]>([])
  const [configs,      setConfigs]      = useState<IntegrationConfig[]>([])
  const [connections,  setConnections]  = useState<Connection[]>([])
  const [loading,      setLoading]      = useState(true)
  const [search,       setSearch]       = useState('')
  const [category,     setCategory]     = useState('all')
  const [showAllAvail, setShowAllAvail] = useState(false)
  const [busy,         setBusy]         = useState<string | null>(null)
  const [error,        setError]        = useState<string | null>(null)
  const [configModal,  setConfigModal]  = useState<NangoProvider | null>(null)
  const [callbackUrl,  setCallbackUrl]  = useState('')

  const configuredSet = useMemo(() => new Set(configs.map(c => c.provider)), [configs])
  const connectedSet  = useMemo(() => new Set(connections.map(c => c.provider)), [connections])

  async function reload() {
    const [p, cfg, conn] = await Promise.all([
      api.listProviders(),
      api.listConfigured(),
      api.listConnections(),
    ])
    setProviders(p)
    setConfigs(cfg)
    setConnections(conn)
  }

  useEffect(() => {
    Promise.all([reload(), api.getConnectInfo().then(info => setCallbackUrl(info.oauth_callback_url))])
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  function getStatus(providerKey: string): 'connected' | 'configured' | 'available' {
    if (connectedSet.has(providerKey))  return 'connected'
    if (configuredSet.has(providerKey)) return 'configured'
    return 'available'
  }

  const filtered = useMemo(() => providers.filter(p => {
    const matchSearch = !search ||
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.unique_key.toLowerCase().includes(search.toLowerCase())
    const matchCat = category === 'all' || (p.categories ?? []).includes(category)
    return matchSearch && matchCat
  }), [providers, search, category])

  const connected  = filtered.filter(p => getStatus(p.unique_key) === 'connected')
  const configured = filtered.filter(p => getStatus(p.unique_key) === 'configured')
  const available  = filtered.filter(p => getStatus(p.unique_key) === 'available')
  const visibleAvailable = showAllAvail ? available : available.slice(0, 24)

  async function handleAction(action: CardAction, provider: NangoProvider) {
    setBusy(provider.unique_key)
    setError(null)
    try {
      if (action === 'setup') {
        setConfigModal(provider)
        return
      }

      if (action === 'connect') {
        const providerConfigs = configs
          .filter(c => c.provider === provider.unique_key)
          .sort((a, b) => Number(b.is_default) - Number(a.is_default))
        if (providerConfigs.length === 0) {
          setConfigModal(provider)
          return
        }
        const connectedAliases = new Set(
          connections.filter(c => c.provider === provider.unique_key).map(c => c.alias),
        )
        const selectedConfig =
          providerConfigs.find(c => !connectedAliases.has(c.alias)) ??
          providerConfigs.find(c => c.is_default) ??
          providerConfigs[0]

        const { session_token, api_url, credential_profile_id, alias } =
          await api.createConnectionSession(provider.unique_key, selectedConfig.alias, selectedConfig.id)
        const nango = new Nango({
          connectSessionToken: session_token,
          host: api_url,
        })
        try {
          await nango.auth(selectedConfig.nango_provider_config_key, {
            detectClosedAuthWindow: true,
          })
        } catch (err: any) {
          if (err?.type === 'window_closed' || err?.message?.toLowerCase().includes('window')) {
            throw new Error('cancelled')
          }
          throw err
        }
        await api.recordConnection(provider.unique_key, '', alias, credential_profile_id)
        await reload()
        return
      }

      if (action === 'disconnect') {
        await api.disconnectProvider(provider.unique_key)
        await reload()
        return
      }

      if (action === 'remove-config') {
        const defaultProfile = configs.find(
          c => c.provider === provider.unique_key && c.is_default,
        ) ?? configs.find(c => c.provider === provider.unique_key)
        if (!defaultProfile) {
          return
        }
        await api.deleteIntegrationConfig(defaultProfile.id)
        await reload()
      }
    } catch (e: any) {
      if (e?.message !== 'cancelled') setError(e?.message ?? 'Something went wrong')
    } finally {
      setBusy(null)
    }
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
      <div className="mb-6">
        <h1 className="text-xl font-semibold text-gray-900">Integrations</h1>
        <p className="mt-1 text-sm text-gray-500">
          Setup OAuth apps, then connect. Use{' '}
          <code className="rounded bg-gray-100 px-1 py-0.5 text-xs">integration('provider')</code>
          {' '}in snippets to call connected APIs.
        </p>
      </div>

      {error && (
        <div className="mb-4 flex items-center gap-2 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          <span className="flex-1">{error}</span>
          <button onClick={() => setError(null)}><X size={14} /></button>
        </div>
      )}

      {/* Search + filters */}
      <div className="mb-5 flex flex-col gap-3 sm:flex-row">
        <div className="relative flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            placeholder="Search providers..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="w-full rounded-lg border border-gray-200 py-2 pl-9 pr-3 text-sm placeholder-gray-400 focus:border-gray-400 focus:outline-none"
          />
        </div>
        <div className="flex flex-wrap gap-1.5">
          {ALL_CATEGORIES.map(cat => (
            <button
              key={cat.key}
              onClick={() => setCategory(cat.key)}
              className={`rounded-md px-3 py-1.5 text-xs font-medium transition-colors ${
                category === cat.key
                  ? 'bg-gray-900 text-white'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              {cat.label}
            </button>
          ))}
        </div>
      </div>

      {/* Connected */}
      {connected.length > 0 && (
        <Section title={`Connected (${connected.length})`}>
          {connected.map(p => (
            <ProviderCard key={p.unique_key} provider={p} status="connected"
              connectingOrBusy={busy === p.unique_key} onAction={handleAction} />
          ))}
        </Section>
      )}

      {/* Configured (ready to connect, not yet connected) */}
      {configured.length > 0 && (
        <Section title={`Ready to connect (${configured.length})`}>
          {configured.map(p => (
            <ProviderCard key={p.unique_key} provider={p} status="configured"
              connectingOrBusy={busy === p.unique_key} onAction={handleAction} />
          ))}
        </Section>
      )}

      {/* Available */}
      <Section title={`Available (${available.length})`}>
        {available.length === 0 ? (
          <p className="col-span-full text-sm text-gray-400">No providers match your search.</p>
        ) : (
          <>
            {visibleAvailable.map(p => (
              <ProviderCard key={p.unique_key} provider={p} status="available"
                connectingOrBusy={busy === p.unique_key} onAction={handleAction} />
            ))}
            {available.length > 24 && (
              <div className="col-span-full flex justify-center">
                <button
                  onClick={() => setShowAllAvail(v => !v)}
                  className="flex items-center gap-1 text-sm text-gray-500 hover:text-gray-800"
                >
                  {showAllAvail
                    ? <><ChevronUp size={14} /> Show less</>
                    : <><ChevronDown size={14} /> Show all {available.length} providers</>
                  }
                </button>
              </div>
            )}
          </>
        )}
      </Section>

      {/* Configure modal */}
      {configModal && (
        <ConfigureModal
          provider={configModal}
          existing={configs.find(c => c.provider === configModal.unique_key && c.is_default) ?? configs.find(c => c.provider === configModal.unique_key)}
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

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="mb-8">
      <h2 className="mb-3 text-xs font-semibold uppercase tracking-wide text-gray-500">{title}</h2>
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
        {children}
      </div>
    </section>
  )
}
