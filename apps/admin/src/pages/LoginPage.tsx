import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { api } from '../lib/api'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import SocialLoginButtons from '../components/SocialLoginButtons'

const OAUTH_ERROR_MESSAGES: Record<string, string> = {
  access_denied: 'Sign-in was cancelled.',
  invalid_state: 'Sign-in session expired. Please try again.',
  missing_code: 'Sign-in failed. Please try again.',
  exchange_failed: 'Could not complete sign-in with the provider.',
  account_error: 'Could not sign you in. The provider may not have shared a verified email.',
  session_error: 'Could not start your session. Please try again.',
}

export default function LoginPage() {
  useDocumentTitle('Sign in')
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const oauthError = searchParams.get('auth_error')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState(oauthError ? OAUTH_ERROR_MESSAGES[oauthError] ?? 'Sign-in failed.' : '')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await api.login(email, password)
      localStorage.removeItem('apiKey')
      navigate('/dashboard/overview')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="w-full max-w-md rounded-lg border border-gray-200 bg-white p-8 shadow-sm">
        <h1 className="mb-2 text-2xl font-bold text-gray-900">Sign in to Velane</h1>
        <p className="mb-6 text-sm text-gray-500">
          Admin portal for managing your tenant.
        </p>

        {error && (
          <div className="mb-4 rounded-md bg-red-50 p-3 text-sm text-red-700">{error}</div>
        )}

        <SocialLoginButtons />

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-gray-400 focus:outline-none"
              placeholder="you@example.com"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-gray-400 focus:outline-none"
              placeholder="••••••••"
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="w-full rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 disabled:opacity-50"
          >
            {loading ? 'Signing in...' : 'Sign in'}
          </button>
        </form>

        <p className="mt-4 text-center text-sm text-gray-500">
          No account?{' '}
          <Link to="/register" className="font-medium text-gray-900 hover:text-gray-700">
            Register
          </Link>
        </p>
      </div>
    </div>
  )
}
