import { useCallback, useEffect, useState } from 'react'
import { api } from '../lib/api'
import type { SnippetVersion } from '../types'

export function useVersions(snippetId: string | undefined) {
  const [versions, setVersions] = useState<SnippetVersion[]>([])
  const [loading, setLoading] = useState(false)

  const load = useCallback(async () => {
    if (!snippetId) return
    setLoading(true)
    try {
      const vs = await api.listVersions(snippetId)
      setVersions(vs ?? [])
    } catch {
      setVersions([])
    } finally {
      setLoading(false)
    }
  }, [snippetId])

  useEffect(() => {
    load()
  }, [load])

  return { versions, loading, reload: load }
}
