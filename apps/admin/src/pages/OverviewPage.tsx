import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { MousePointer2, Sparkles, Terminal } from 'lucide-react'
import { api } from '../lib/api'

const FEATURED_AGENTS = [
  { label: 'Cursor', icon: MousePointer2 },
  { label: 'Claude Code', icon: Sparkles },
  { label: 'Codex', icon: Terminal },
]

interface Stats {
  apiKeyCount: number
  memberCount: number
  invocations24h: number
  snippetCount: number
  loading: boolean
  error: string
}

export default function OverviewPage() {
  const [stats, setStats] = useState<Stats>({
    apiKeyCount: 0,
    memberCount: 0,
    invocations24h: 0,
    snippetCount: 0,
    loading: true,
    error: '',
  })

  useEffect(() => {
    const load = async () => {
      try {
        const [keys, members, usage, snippets] = await Promise.allSettled([
          api.listAPIKeys(),
          api.listMembers(),
          api.getUsage('24h'),
          api.listSnippets(),
        ])
        setStats({
          apiKeyCount: keys.status === 'fulfilled' ? keys.value.length : 0,
          memberCount: members.status === 'fulfilled' ? members.value.length : 0,
          invocations24h:
            usage.status === 'fulfilled' ? usage.value.total_invocations : 0,
          snippetCount: snippets.status === 'fulfilled' ? snippets.value.length : 0,
          loading: false,
          error: '',
        })
      } catch {
        setStats((s) => ({ ...s, loading: false, error: 'Failed to load stats' }))
      }
    }
    load()
  }, [])

  return (
    <div>
      <h1 className="mb-8 text-2xl font-bold text-gray-900">Overview</h1>

      <div className="mb-8 overflow-hidden rounded-2xl border border-gray-200 bg-gray-900 p-6 text-white shadow-sm sm:p-8">
        <div className="flex flex-col gap-6 sm:flex-row sm:items-center sm:justify-between">
          <div className="max-w-xl">
            <h2 className="text-xl font-semibold">Velane works best inside your coding agent</h2>
            <p className="mt-2 text-sm leading-6 text-gray-300">
              Use this dashboard to manage settings and review usage. Day to day, connect Velane
              directly to your coding agent over MCP and let it create, run, and debug workflows — no
              need to work from here.
            </p>
            <p className="mt-3 text-xs text-gray-400">
              Works with any MCP-compatible agent. Popular setups:
            </p>
            <div className="mt-2 flex flex-wrap gap-2">
              {FEATURED_AGENTS.map(({ label, icon: Icon }) => (
                <span
                  key={label}
                  className="inline-flex items-center gap-1.5 rounded-full border border-gray-700 bg-gray-800 px-3 py-1 text-xs font-medium text-gray-200"
                >
                  <Icon className="h-3.5 w-3.5" />
                  {label}
                </span>
              ))}
              <span className="rounded-full border border-dashed border-gray-600 px-3 py-1 text-xs font-medium text-gray-400">
                More
              </span>
            </div>
          </div>
          <Link
            to="/dashboard/mcp"
            className="inline-flex shrink-0 rounded-lg bg-white px-4 py-2 text-sm font-medium text-gray-900 hover:bg-gray-100"
          >
            Connect your agent
          </Link>
        </div>
      </div>

      {stats.error && (
        <div className="mb-6 rounded-md bg-red-50 p-3 text-sm text-red-700">{stats.error}</div>
      )}

      <div className="grid grid-cols-1 gap-6 sm:grid-cols-4">
        <StatCard label="Workflows" value={stats.loading ? '...' : String(stats.snippetCount)} />
        <StatCard label="API Keys" value={stats.loading ? '...' : String(stats.apiKeyCount)} />
        <StatCard label="Team Members" value={stats.loading ? '...' : String(stats.memberCount)} />
        <StatCard label="Invocations (24h)" value={stats.loading ? '...' : String(stats.invocations24h)} />
      </div>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
      <p className="text-sm font-medium text-gray-500">{label}</p>
      <p className="mt-2 text-3xl font-bold text-gray-900">{value}</p>
    </div>
  )
}
