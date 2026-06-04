import { useEffect, useMemo, useState } from 'react'
import { Check, Copy } from 'lucide-react'
import { api } from '../lib/api'
import CopyBox from '../components/CopyBox'

type PlatformID = 'cursor' | 'vscode' | 'claude-code' | 'claude-desktop' | 'codex' | 'gemini'

interface Platform {
  id: PlatformID
  label: string
  configPath: string
}

const PLATFORMS: Platform[] = [
  { id: 'cursor', label: 'Cursor', configPath: '.cursor/mcp.json' },
  { id: 'vscode', label: 'VS Code', configPath: '.vscode/mcp.json' },
  { id: 'claude-code', label: 'Claude Code', configPath: '~/.claude/mcp.json' },
  {
    id: 'claude-desktop',
    label: 'Claude Desktop',
    configPath: '~/Library/Application Support/Claude/claude_desktop_config.json',
  },
  { id: 'codex', label: 'Codex', configPath: '.cursor/mcp.json' },
  { id: 'gemini', label: 'Gemini', configPath: '~/.gemini/mcp.json' },
]

const API_KEY_PLACEHOLDER = 'vl_YOUR_API_KEY'

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

  const configJSON = useMemo(() => {
    if (!mcpURL) return ''
    return JSON.stringify(
      {
        mcpServers: {
          velane: {
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
  }, [mcpURL, token])

  const activePlatformConfig = PLATFORMS.find((p) => p.id === activePlatform) ?? PLATFORMS[0]

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
              <p className="mb-2 text-xs font-medium uppercase tracking-wide text-gray-500">Config JSON</p>
              <div className="relative rounded-xl bg-gray-950 px-4 py-3">
                <pre className="overflow-x-auto text-xs leading-relaxed text-gray-100">{configJSON}</pre>
                <button
                  type="button"
                  className="absolute right-2 top-2 rounded p-1 text-gray-400 hover:text-gray-200"
                  onClick={() => copyToClipboard(configJSON, 'json')}
                >
                  {copied === 'json' ? (
                    <Check className="h-4 w-4 text-green-400" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </button>
              </div>
            </div>

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
    </div>
  )
}
