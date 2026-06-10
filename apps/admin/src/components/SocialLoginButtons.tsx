import { useEffect, useState } from 'react'
import { Github } from 'lucide-react'
import { api } from '../lib/api'

const PROVIDER_LABELS: Record<string, string> = {
  google: 'Continue with Google',
  github: 'Continue with GitHub',
}

function ProviderIcon({ provider }: { provider: string }) {
  if (provider === 'github') {
    return <Github size={16} />
  }
  if (provider === 'google') {
    return (
      <svg width="16" height="16" viewBox="0 0 18 18" aria-hidden="true">
        <path fill="#4285F4" d="M17.64 9.2c0-.64-.06-1.25-.16-1.84H9v3.48h4.84a4.14 4.14 0 0 1-1.8 2.72v2.26h2.92c1.7-1.57 2.68-3.88 2.68-6.62Z" />
        <path fill="#34A853" d="M9 18c2.43 0 4.47-.8 5.96-2.18l-2.92-2.26c-.8.54-1.84.86-3.04.86-2.34 0-4.32-1.58-5.03-3.7H.96v2.33A9 9 0 0 0 9 18Z" />
        <path fill="#FBBC05" d="M3.97 10.72A5.4 5.4 0 0 1 3.68 9c0-.6.1-1.18.29-1.72V4.95H.96A9 9 0 0 0 0 9c0 1.45.35 2.82.96 4.05l3.01-2.33Z" />
        <path fill="#EA4335" d="M9 3.58c1.32 0 2.5.46 3.44 1.35l2.58-2.58C13.46.89 11.43 0 9 0A9 9 0 0 0 .96 4.95l3.01 2.33C4.68 5.16 6.66 3.58 9 3.58Z" />
      </svg>
    )
  }
  return null
}

export default function SocialLoginButtons() {
  const [providers, setProviders] = useState<string[]>([])

  useEffect(() => {
    let cancelled = false
    api.listOAuthProviders().then((list) => {
      if (!cancelled) setProviders(list)
    })
    return () => {
      cancelled = true
    }
  }, [])

  if (providers.length === 0) return null

  return (
    <div className="mb-6">
      <div className="space-y-2">
        {providers.map((provider) => (
          <a
            key={provider}
            href={api.oauthStartUrl(provider)}
            className="flex w-full items-center justify-center gap-2 rounded-lg border border-gray-300 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50"
          >
            <ProviderIcon provider={provider} />
            {PROVIDER_LABELS[provider] ?? `Continue with ${provider}`}
          </a>
        ))}
      </div>
      <div className="mt-6 flex items-center gap-3">
        <div className="h-px flex-1 bg-gray-200" />
        <span className="text-xs text-gray-400">or continue with email</span>
        <div className="h-px flex-1 bg-gray-200" />
      </div>
    </div>
  )
}
