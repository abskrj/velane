import { useState, useEffect, useMemo } from 'react'
import { Plug, Search, CheckCircle2, Loader2, X, Settings, ChevronDown, ChevronUp } from 'lucide-react'
import { api } from '../lib/api'
import type { Connection, NangoProvider, IntegrationConfig } from '../types'

// Auth modes that require operator OAuth app setup before users can connect.
const OAUTH_MODES = new Set(['OAUTH2', 'OAUTH2_CC', 'OAUTH1', 'APP', 'MCP_OAUTH2', 'MCP_OAUTH2_GENERIC'])

function isOAuthMode(authMode: string) {
  return OAUTH_MODES.has(authMode)
}

// ---- Configure Modal -------------------------------------------------------

interface ConfigureModalProps {
  provider: NangoProvider
  existing?: IntegrationConfig
  onClose: () => void
  onSaved: () => void
}

function ConfigureModal({ provider, existing, onClose, onSaved }: ConfigureModalProps) {
  const [clientId, setClientId]     = useState('')
  const [clientSecret, setClientSecret] = useState('')
  const [scopes, setScopes]         = useState(existing?.oauth_scopes ?? '')
  const [extraFields, setExtraFields] = useState<Record<string, string>>({})
  const [saving, setSaving]         = useState(false)
  const [error, setError]           = useState<string | null>(null)

  // Non-automated connection_config fields the operator must fill in.
  const configFields = useMemo(() => {
    if (!provider.connection_config) return []
    return Object.entries(provider.connection_config)
      .filter(([, f]) => !f.automated && !f.optional)
      .map(([key, field]) => ({ key, field }))
  }, [provider])

  const optionalConfigFields = useMemo(() => {
    if (!provider.connection_config) return []
    return Object.entries(provider.connection_config)
      .filter(([, f]) => !f.automated && f.optional)
      .map(([key, field]) => ({ key, field }))
  }, [provider])

  async function handleSave() {
    setSaving(true)
    setError(null)
    try {
      await api.configureIntegration({
        provider: provider.unique_key,
        oauth_client_id: clientId || undefined,
        oauth_client_secret: clientSecret || undefined,
        oauth_scopes: scopes || undefined,
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
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div className="w-full max-w-lg rounded-xl bg-white shadow-xl">
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
            <button onClick={onClose} className="rounded-md p-1 text-gray-400 hover:bg-gray-100">
              <X size={16} />
            </button>
          </div>

          {/* Form */}
          <div className="px-6 py-5 space-y-4">
            {error && (
              <p className="rounded-lg bg-red-50 px-3 py-2 text-xs text-red-600">{error}</p>
            )}

            {isOAuthMode(provider.auth_mode) && (
              <>
                <Field
                  label="Client ID"
                  required
                  value={clientId}
                  onChange={setClientId}
                  placeholder="Paste your OAuth app client ID"
                />
                <Field
                  label="Client Secret"
                  required
                  secret
                  value={clientSecret}
                  onChange={setClientSecret}
                  placeholder="Paste your OAuth app client secret"
                />
                <Field
                  label="Scopes"
                  value={scopes}
                  onChange={setScopes}
                  placeholder="e.g. repo user (space or comma separated)"
                  description="Optional — leave blank to use provider defaults"
                />
              </>
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
          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-gray-50 border border-gray-100 overflow-hidden">
            {provider.logo_url && !imgError ? (
              <img
                src={provider.logo_url}
                alt={provider.name}
                className="h-5 w-5 object-contain"
                onError={() => setImgError(true)}
              />
            ) : (
              <Plug size={13} className="text-gray-400" />
            )}
          </div>
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

  const configuredSet = useMemo(() => new Set(configs.map(c => c.unique_key)), [configs])
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
    reload().catch(e => setError(e.message)).finally(() => setLoading(false))
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
        // For non-OAuth providers, auto-create config entry if missing.
        if (!isOAuthMode(provider.auth_mode) && !configuredSet.has(provider.unique_key)) {
          await api.configureIntegration({ provider: provider.unique_key })
          await reload()
        }
        const { session_token } = await api.createConnectionSession(provider.unique_key)
        const { default: Nango } = await import('@nangohq/frontend')
        const nango = new Nango({ connectSessionToken: session_token })
        await nango.openConnectUI({
          onSuccess: async () => {
            await api.recordConnection(provider.unique_key)
            await reload()
          },
        })
        return
      }

      if (action === 'disconnect') {
        await api.disconnectProvider(provider.unique_key)
        await reload()
        return
      }

      if (action === 'remove-config') {
        await api.deleteIntegrationConfig(provider.unique_key)
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
    <div className="max-w-5xl">
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
          existing={configs.find(c => c.unique_key === configModal.unique_key)}
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
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
        {children}
      </div>
    </section>
  )
}
