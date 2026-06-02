import { useEffect, useState, type ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { api } from '../lib/api'

interface Props {
  children: ReactNode
}

export default function ProtectedRoute({ children }: Props) {
  const apiKey = localStorage.getItem('apiKey')
  const [status, setStatus] = useState<'checking' | 'allowed' | 'blocked'>(apiKey ? 'allowed' : 'checking')

  useEffect(() => {
    if (apiKey) {
      setStatus('allowed')
      return
    }

    let cancelled = false
    api.me()
      .then(() => {
        if (!cancelled) setStatus('allowed')
      })
      .catch(() => {
        if (!cancelled) setStatus('blocked')
      })

    return () => {
      cancelled = true
    }
  }, [apiKey])

  if (status === 'checking') {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50">
        <p className="text-sm text-gray-500">Loading workspace...</p>
      </div>
    )
  }

  if (status === 'blocked') {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}
