import { useEffect } from 'react'
import { setDocumentTitle } from '../lib/documentTitle'

export function useDocumentTitle(...segments: (string | undefined | null)[]) {
  const key = segments.map((s) => s ?? '').join('\0')

  useEffect(() => {
    const parts = segments.filter((s): s is string => Boolean(s?.trim()))
    if (parts.length === 0) return
    setDocumentTitle(...parts)
  }, [key])
}
