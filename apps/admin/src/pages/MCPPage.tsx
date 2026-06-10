import { useEffect, useMemo, useState } from 'react'
import { Check, Copy } from 'lucide-react'
import { api } from '../lib/api'
import CopyBox from '../components/CopyBox'

type PlatformID = 'cursor' | 'vscode' | 'claude-code' | 'claude-desktop' | 'codex' | 'gemini'

interface Platform {
  id: PlatformID
  label: string
  configPath: string
  snippetLabel: string
  description: string
}

const PLATFORMS: Platform[] = [
  {
    id: 'cursor',
    label: 'Cursor',
    configPath: '~/.cursor/mcp.json',
    snippetLabel: 'Config JSON',
    description: 'Native remote HTTP MCP configuration for Cursor.',
  },
  {
    id: 'vscode',
    label: 'VS Code',
    configPath: '.vscode/mcp.json',
    snippetLabel: 'Config JSON',
    description: 'Native remote HTTP MCP configuration for VS Code.',
  },
  {
    id: 'claude-code',
    label: 'Claude Code',
    configPath: 'Terminal',
    snippetLabel: 'CLI command',
    description: 'Adds Velane as a remote HTTP MCP server through the Claude Code CLI.',
  },
  {
    id: 'claude-desktop',
    label: 'Claude Desktop',
    configPath: '~/Library/Application Support/Claude/claude_desktop_config.json',
    snippetLabel: 'Config JSON',
    description: 'Uses mcp-remote to bridge Claude Desktop stdio MCP to Velane HTTP MCP.',
  },
  {
    id: 'codex',
    label: 'Codex',
    configPath: '~/.codex/config.toml',
    snippetLabel: 'Config TOML',
    description: 'Uses mcp-remote as a stdio bridge from Codex to Velane HTTP MCP.',
  },
  {
    id: 'gemini',
    label: 'Gemini',
    configPath: '~/.gemini/settings.json',
    snippetLabel: 'Config JSON',
    description: 'Uses mcp-remote as a stdio bridge from Gemini CLI to Velane HTTP MCP.',
  },
]

const API_KEY_PLACEHOLDER = 'vl_YOUR_API_KEY'

function stdioBridgeConfig(mcpURL: string, token: string) {
  return {
    command: 'npx',
    args: ['-y', 'mcp-remote', mcpURL, '--header', 'Authorization:${AUTH_HEADER}'],
    env: {
      AUTH_HEADER: `Bearer ${token}`,
    },
  }
}

function configForPlatform(platform: PlatformID, mcpURL: string, token: string) {
  if (!mcpURL) return ''

  switch (platform) {
    case 'cursor':
      return JSON.stringify(
        {
          mcpServers: {
            velane: {
              name: 'velane',
              type: 'http',
              url: mcpURL,
              headers: {
                Authorization: `Bearer ${token}`,
              },
            },
          },
        },
        null,
        2,
      )
    case 'vscode':
      return JSON.stringify(
        {
          servers: {
            velane: {
              type: 'http',
              url: mcpURL,
              headers: {
                Authorization: `Bearer ${token}`,
              },
            },
          },
        },
        null,
        2,
      )
    case 'claude-code':
      return [
        'claude mcp add velane \\',
        `  --transport http "${mcpURL}" \\`,
        `  --header "Authorization:Bearer ${token}"`,
      ].join('\n')
    case 'claude-desktop':
      return JSON.stringify(
        {
          mcpServers: {
            velane: stdioBridgeConfig(mcpURL, token),
          },
        },
        null,
        2,
      )
    case 'codex':
      return [
        '[mcp_servers.velane]',
        'command = "npx"',
        `args = ["-y", "mcp-remote", "${mcpURL}", "--header", "Authorization:\${AUTH_HEADER}"]`,
        '',
        '[mcp_servers.velane.env]',
        `AUTH_HEADER = "Bearer ${token}"`,
      ].join('\n')
    case 'gemini':
      return JSON.stringify(
        {
          mcpServers: {
            velane: stdioBridgeConfig(mcpURL, token),
          },
        },
        null,
        2,
      )
  }
}

export default function MCPPage() {
  const [activePlatform, setActivePlatform] = useState<PlatformID>('cursor')
  const [mcpURL, setMCPURL] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [creatingKey, setCreatingKey] = useState(false)
  const [createdAPIKey, setCreatedAPIKey] = useState<string | null>(null)
  const [keyError, setKeyError] = useState('')
  const [copied, setCopied] = useState<string | null>(null)

  useEffect(() => {
    async function loadMCPInfo() {
      try {
        const info = await api.getMCPInfo()
        if (!info.mcp_url?.trim()) {
          setError('MCP URL not configured')
          return
        }
        setMCPURL(info.mcp_url)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'MCP URL not configured')
      } finally {
        setLoading(false)
      }
    }
    loadMCPInfo()
  }, [])

  const token = createdAPIKey ?? API_KEY_PLACEHOLDER

  const activePlatformConfig = PLATFORMS.find((p) => p.id === activePlatform) ?? PLATFORMS[0]
  const configSnippet = useMemo(
    () => configForPlatform(activePlatformConfig.id, mcpURL, token),
    [activePlatformConfig.id, mcpURL, token],
  )

  const copyToClipboard = async (value: string, key: string) => {
    if (!value) return
    await navigator.clipboard.writeText(value)
    setCopied(key)
    setTimeout(() => setCopied(null), 2000)
  }

  const handleGenerateKey = async () => {
    setCreatingKey(true)
    setKeyError('')
    try {
      const created = await api.createAPIKey('MCP Access Key', ['invoke', 'manage'])
      if (!created.key) {
        setKeyError('Failed to create API key')
        return
      }
      setCreatedAPIKey(created.key)
    } catch (err) {
      setKeyError(err instanceof Error ? err.message : 'Failed to create API key')
    } finally {
      setCreatingKey(false)
    }
  }

  return (
    <div className="mx-auto max-w-4xl">
      <div className="mb-6 text-center">
        <h1 className="text-3xl font-semibold text-gray-900">Get Started</h1>
        <p className="mt-2 text-sm text-gray-500">
          Choose your platform and copy the MCP config. API key is required.
        </p>
      </div>

      <div className="mb-6 flex flex-wrap justify-center gap-2">
        {PLATFORMS.map((platform) => (
          <button
            key={platform.id}
            type="button"
            onClick={() => setActivePlatform(platform.id)}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              activePlatform === platform.id
                ? 'bg-gray-900 text-white'
                : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
            }`}
          >
            {platform.label}
          </button>
        ))}
      </div>

      <div className="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm">
        <div className="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-sm text-gray-600">
            Add to <code className="rounded bg-gray-100 px-1 py-0.5 text-xs">{activePlatformConfig.configPath}</code>
          </p>
          <button
            type="button"
            onClick={handleGenerateKey}
            disabled={creatingKey || loading || !!error}
            className="rounded-lg bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-50"
          >
            {creatingKey ? 'Generating...' : 'Generate API Key'}
          </button>
        </div>

        {createdAPIKey && (
          <div className="mb-4 rounded-md border border-green-200 bg-green-50 p-3">
            <p className="mb-2 text-sm font-medium text-green-800">
              New API key created. Copy it now — it will never be shown again.
            </p>
            <CopyBox value={createdAPIKey} />
          </div>
        )}

        {keyError && (
          <div className="mb-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            {keyError}
          </div>
        )}

        {loading ? (
          <p className="text-sm text-gray-500">Loading MCP configuration...</p>
        ) : error ? (
          <div className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{error}</div>
        ) : (
          <div className="space-y-5">
            <div>
              <div className="mb-2">
                <p className="text-xs font-medium uppercase tracking-wide text-gray-500">
                  {activePlatformConfig.snippetLabel}
                </p>
                <p className="mt-1 text-xs text-gray-500">{activePlatformConfig.description}</p>
              </div>
              <div className="relative rounded-xl bg-gray-950 px-4 py-3">
                <pre className="overflow-x-auto text-xs leading-relaxed text-gray-100">{configSnippet}</pre>
                <button
                  type="button"
                  className="absolute right-2 top-2 rounded p-1 text-gray-400 hover:text-gray-200"
                  onClick={() => copyToClipboard(configSnippet, 'config')}
                >
                  {copied === 'config' ? (
                    <Check className="h-4 w-4 text-green-400" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </button>
              </div>
            </div>

            {['claude-desktop', 'codex', 'gemini'].includes(activePlatform) && (
              <p className="text-xs text-gray-500">
                This config uses <code className="rounded bg-gray-100 px-1">mcp-remote</code>, which runs through{' '}
                <code className="rounded bg-gray-100 px-1">npx</code> and forwards your Velane API key as an
                Authorization header.
              </p>
            )}

            <div>
              <p className="mb-2 text-xs font-medium uppercase tracking-wide text-gray-500">Raw MCP URL</p>
              <div className="flex items-center gap-2 rounded-md border border-gray-200 bg-gray-50 px-3 py-2">
                <code className="flex-1 text-xs text-gray-700">{mcpURL}</code>
                <button
                  type="button"
                  className="rounded p-1 text-gray-500 hover:bg-gray-200 hover:text-gray-700"
                  onClick={() => copyToClipboard(mcpURL, 'url')}
                >
                  {copied === 'url' ? (
                    <Check className="h-4 w-4 text-green-600" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </button>
              </div>
            </div>

            {!createdAPIKey && (
              <p className="text-xs text-gray-500">
                Using placeholder token <code className="rounded bg-gray-100 px-1">{API_KEY_PLACEHOLDER}</code> until
                you generate a key.
              </p>
            )}
          </div>
        )}
      </div>

      <div className="mt-6 rounded-2xl border border-gray-200 bg-white p-6 shadow-sm">
        <h2 className="text-base font-semibold text-gray-900">What agents can access</h2>
        <div className="mt-4 grid gap-3 sm:grid-cols-3">
          <div className="rounded-xl border border-gray-100 bg-gray-50 p-4">
            <p className="text-sm font-medium text-gray-900">Tools</p>
            <p className="mt-1 text-xs leading-5 text-gray-500">
              Actions such as creating workflows, updating drafts, invoking runs, listing connections, and reading metrics.
            </p>
          </div>
          <div className="rounded-xl border border-gray-100 bg-gray-50 p-4">
            <p className="text-sm font-medium text-gray-900">Resources</p>
            <p className="mt-1 text-xs leading-5 text-gray-500">
              Bounded context such as the runtime contract, compact workflow catalog, and connected integrations.
            </p>
          </div>
          <div className="rounded-xl border border-gray-100 bg-gray-50 p-4">
            <p className="text-sm font-medium text-gray-900">Prompts</p>
            <p className="mt-1 text-xs leading-5 text-gray-500">
              Guided flows for creating integration workflows, debugging failed invocations, and publishing after validation.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
