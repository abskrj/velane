import { useEffect } from 'react'
import { api } from '../lib/api'

export function useSessionRefresh() {
  useEffect(() => {
    if (localStorage.getItem('apiKey')) return

    const refresh = () => {
      void api.refreshSession()
    }

    refresh()
    const interval = window.setInterval(refresh, api.sessionRefreshIntervalMs)
    return () => window.clearInterval(interval)
  }, [])
}
