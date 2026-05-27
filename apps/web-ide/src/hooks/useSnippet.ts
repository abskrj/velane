import { useCallback, useEffect, useState } from 'react'
import { api } from '../lib/api'
import type { Snippet } from '../types'

export function useSnippet(id: string | undefined) {
  const [snippet, setSnippet] = useState<Snippet | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    if (!id) return
    setLoading(true)
    setError(null)
    try {
      const sn = await api.getSnippet(id)
      setSnippet(sn)
    } catch (err) {
      setError(String(err))
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    load()
  }, [load])

  return { snippet, loading, error, reload: load, setSnippet }
}
