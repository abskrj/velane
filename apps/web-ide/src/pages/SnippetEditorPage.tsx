import Editor from '@monaco-editor/react'
import { Trash2 } from 'lucide-react'
import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import EnvStatusBadge from '../components/EnvStatusBadge'
import LanguageBadge from '../components/LanguageBadge'
import PublishDropdown from '../components/PublishDropdown'
import { useSnippet } from '../hooks/useSnippet'
import { useVersions } from '../hooks/useVersions'
import { api } from '../lib/api'
import type { InvocationResult, SnippetVersion } from '../types'

type ActiveTab = 'test' | 'logs'

export default function SnippetEditorPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const isNew = !id

  const { snippet, loading: snippetLoading } = useSnippet(id)
  const { versions, reload: reloadVersions } = useVersions(id)

  const [code, setCode] = useState('')
  const [snippetName, setSnippetName] = useState('')
  const [language, setLanguage] = useState<'bun' | 'python'>('bun')
  const [saving, setSaving] = useState(false)
  const [activeTab, setActiveTab] = useState<ActiveTab>('test')
  const [testInput, setTestInput] = useState('{}')
  const [invokeResult, setInvokeResult] = useState<InvocationResult | null>(null)
  const [invoking, setInvoking] = useState(false)
  const [selectedVersion, setSelectedVersion] = useState<SnippetVersion | null>(null)
  const [snippetId, setSnippetId] = useState<string | undefined>(id)

  const autosaveTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Populate form when snippet loads.
  useEffect(() => {
    if (snippet) {
      setSnippetName(snippet.name)
      setLanguage(snippet.language)
    }
  }, [snippet])

  // Show latest version code on load.
  useEffect(() => {
    if (versions.length > 0 && !selectedVersion) {
      const latest = versions[versions.length - 1]
      setCode(latest.code)
    }
  }, [versions, selectedVersion])

  const monacoLanguage = language === 'python' ? 'python' : 'typescript'

  // Debounced auto-save for existing snippets.
  const handleCodeChange = useCallback(
    (value: string | undefined) => {
      const newCode = value ?? ''
      setCode(newCode)
      if (!snippetId) return

      if (autosaveTimer.current) clearTimeout(autosaveTimer.current)
      autosaveTimer.current = setTimeout(async () => {
        try {
          await api.createVersion(snippetId, newCode)
          reloadVersions()
        } catch {
          // Auto-save failures are silent; user can save manually.
        }
      }, 2000)
    },
    [snippetId, reloadVersions],
  )

  async function handleSaveDraft() {
    setSaving(true)
    try {
      if (isNew) {
        const sn = await api.createSnippet({ name: snippetName, language })
        setSnippetId(sn.id)
        await api.createVersion(sn.id, code)
        navigate(`/snippets/${sn.id}`, { replace: true })
      } else if (snippetId) {
        await api.createVersion(snippetId, code)
        reloadVersions()
      }
    } catch (err) {
      alert(String(err))
    } finally {
      setSaving(false)
    }
  }

  async function handlePublish(env: 'dev' | 'staging' | 'prod') {
    if (!snippetId) {
      await handleSaveDraft()
      return
    }
    try {
      const allVersions = await api.listVersions(snippetId)
      const latest = allVersions[allVersions.length - 1]
      if (!latest) return
      await api.publishVersion(snippetId, latest.version_number, env)
      reloadVersions()
    } catch (err) {
      alert(String(err))
    }
  }

  async function handleDelete() {
    if (!snippetId) return
    if (!confirm('Delete this snippet?')) return
    try {
      await api.deleteSnippet(snippetId)
      navigate('/')
    } catch (err) {
      alert(String(err))
    }
  }

  async function handleRun() {
    if (!snippetId) return
    setInvoking(true)
    setInvokeResult(null)
    try {
      const result = await api.invokeSnippet(snippetId ?? snippet?.id ?? '', testInput)
      setInvokeResult(result)
    } catch (err) {
      setInvokeResult({
        output: null,
        invocation_id: '',
        duration_ms: 0,
        status: 'error',
        error: String(err),
        stderr: '',
      })
    } finally {
      setInvoking(false)
    }
  }

  if (snippetLoading) {
    return <div className="flex min-h-screen items-center justify-center">Loading...</div>
  }

  const displayCode = selectedVersion ? selectedVersion.code : code
  const isReadOnly = !!selectedVersion

  return (
    <div className="flex h-screen flex-col bg-gray-50">
      {/* Header */}
      <header className="flex shrink-0 items-center justify-between border-b border-gray-200 bg-white px-4 py-2">
        <div className="flex items-center gap-3">
          <button
            className="text-sm text-gray-500 hover:text-gray-900"
            onClick={() => navigate('/')}
          >
            &larr; Snippets
          </button>
          {isNew ? (
            <input
              className="rounded border border-gray-300 px-2 py-1 text-sm"
              placeholder="Snippet name"
              value={snippetName}
              onChange={(e) => setSnippetName(e.target.value)}
            />
          ) : (
            <span className="font-medium text-gray-900">{snippet?.name}</span>
          )}
          <LanguageBadge language={language} />
        </div>
        <div className="flex items-center gap-2">
          <button
            className="rounded-md border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
            onClick={handleSaveDraft}
            disabled={saving}
          >
            {saving ? 'Saving...' : 'Save Draft'}
          </button>
          <PublishDropdown onPublish={handlePublish} disabled={!snippetId && !snippetName} />
          {snippetId && (
            <button
              className="rounded-md border border-red-200 px-2 py-1.5 text-sm text-red-600 hover:bg-red-50"
              onClick={handleDelete}
              title="Delete snippet"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          )}
        </div>
      </header>

      {/* Body: editor + panel */}
      <div className="flex min-h-0 flex-1">
        {/* Monaco editor */}
        <div className="flex-1 border-r border-gray-200">
          <Editor
            height="100%"
            language={monacoLanguage}
            value={displayCode}
            onChange={isReadOnly ? undefined : handleCodeChange}
            options={{
              readOnly: isReadOnly,
              minimap: { enabled: false },
              fontSize: 14,
              lineNumbers: 'on',
              scrollBeyondLastLine: false,
            }}
            theme="vs"
          />
        </div>

        {/* Right panel */}
        <div className="flex w-96 shrink-0 flex-col">
          {/* Tabs */}
          <div className="flex border-b border-gray-200 bg-white">
            {(['test', 'logs'] as ActiveTab[]).map((tab) => (
              <button
                key={tab}
                className={`px-4 py-2 text-sm font-medium capitalize ${
                  activeTab === tab
                    ? 'border-b-2 border-indigo-600 text-indigo-600'
                    : 'text-gray-500 hover:text-gray-700'
                }`}
                onClick={() => setActiveTab(tab)}
              >
                {tab}
              </button>
            ))}
          </div>

          {activeTab === 'test' && (
            <div className="flex flex-1 flex-col gap-3 overflow-auto p-4">
              <div>
                <label className="mb-1 block text-xs font-medium text-gray-700">
                  Input JSON
                </label>
                <textarea
                  className="h-32 w-full rounded-md border border-gray-300 p-2 font-mono text-xs focus:outline-none focus:ring-1 focus:ring-indigo-500"
                  value={testInput}
                  onChange={(e) => setTestInput(e.target.value)}
                  spellCheck={false}
                />
              </div>
              <button
                className="w-full rounded-md bg-indigo-600 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
                onClick={handleRun}
                disabled={invoking || !snippetId}
              >
                {invoking ? 'Running...' : 'Run'}
              </button>
              {invokeResult && (
                <div className="rounded-md bg-gray-900 p-3">
                  <pre className="overflow-auto text-xs text-green-300">
                    {JSON.stringify(invokeResult, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          )}

          {activeTab === 'logs' && (
            <div className="flex flex-1 items-center justify-center text-sm text-gray-500">
              Live logs coming in Phase 5
            </div>
          )}
        </div>
      </div>

      {/* Version history footer */}
      {versions.length > 0 && (
        <footer className="flex shrink-0 items-center gap-2 border-t border-gray-200 bg-white px-4 py-2">
          <span className="text-xs text-gray-500">Versions:</span>
          {selectedVersion && (
            <button
              className="rounded px-2 py-0.5 text-xs text-gray-500 hover:bg-gray-100"
              onClick={() => setSelectedVersion(null)}
            >
              Latest (editable)
            </button>
          )}
          {versions.map((v) => (
            <button
              key={v.id}
              className={`rounded px-2 py-0.5 text-xs ${
                selectedVersion?.id === v.id
                  ? 'bg-indigo-100 font-medium text-indigo-700'
                  : 'text-gray-600 hover:bg-gray-100'
              }`}
              onClick={() => setSelectedVersion(selectedVersion?.id === v.id ? null : v)}
            >
              v{v.version_number}
              {v.status === 'published' && (
                <EnvStatusBadge env="prod" version={v} />
              )}
            </button>
          ))}
        </footer>
      )}
    </div>
  )
}
