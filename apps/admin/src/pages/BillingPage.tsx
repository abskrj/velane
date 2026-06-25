import { useEffect, useState } from 'react'
import { CheckCircle2, XCircle, Zap, ArrowUpRight, RefreshCw } from 'lucide-react'
import { api } from '../lib/api'

interface PlanInfo {
  plan: string
  valid: boolean
  features: string[]
}

const PLAN_LABELS: Record<string, string> = {
  free: 'Free',
  pro: 'Pro',
  enterprise: 'Enterprise',
}

const PLAN_COLORS: Record<string, string> = {
  free: 'text-gray-500 bg-gray-100',
  pro: 'text-blue-700 bg-blue-50',
  enterprise: 'text-purple-700 bg-purple-50',
}

export default function BillingPage() {
  const [plan, setPlan] = useState<PlanInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = () => {
    setLoading(true)
    setError('')
    api
      .getTenantPlan()
      .then(setPlan)
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load plan'))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  return (
    <div className="max-w-2xl">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">Plan &amp; Billing</h1>
        <button
          onClick={load}
          className="flex items-center gap-1.5 rounded-md border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-50"
        >
          <RefreshCw size={13} />
          Refresh
        </button>
      </div>

      {loading && (
        <div className="rounded-xl border border-gray-200 bg-white p-8 text-center text-sm text-gray-400">
          Loading plan info…
        </div>
      )}

      {error && (
        <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          {error}
        </div>
      )}

      {!loading && !error && plan && (
        <div className="space-y-4">
          {/* Current plan card */}
          <div className="rounded-xl border border-gray-200 bg-white p-6">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-sm text-gray-500">Current plan</p>
                <div className="mt-1 flex items-center gap-2">
                  <span className="text-2xl font-bold text-gray-900">
                    {PLAN_LABELS[plan.plan] ?? plan.plan}
                  </span>
                  <span
                    className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${PLAN_COLORS[plan.plan] ?? PLAN_COLORS.free}`}
                  >
                    {plan.plan}
                  </span>
                </div>
              </div>

              <div className="flex items-center gap-1.5">
                {plan.valid ? (
                  <>
                    <CheckCircle2 size={16} className="text-green-500" />
                    <span className="text-sm font-medium text-green-700">Active</span>
                  </>
                ) : (
                  <>
                    <XCircle size={16} className="text-gray-400" />
                    <span className="text-sm text-gray-500">No active license</span>
                  </>
                )}
              </div>
            </div>

            {!plan.valid && (
              <div className="mt-4">
                <a
                  href="https://velane.sh/pricing"
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex items-center gap-1.5 rounded-lg bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800"
                >
                  <Zap size={14} />
                  Upgrade plan
                  <ArrowUpRight size={13} />
                </a>
              </div>
            )}
          </div>

          {/* Features */}
          {plan.valid && plan.features.length > 0 && (
            <div className="rounded-xl border border-gray-200 bg-white p-6">
              <h2 className="mb-4 text-sm font-semibold text-gray-700">Included features</h2>
              <ul className="space-y-2">
                {plan.features.map((f) => (
                  <li key={f} className="flex items-center gap-2 text-sm text-gray-700">
                    <CheckCircle2 size={15} className="shrink-0 text-green-500" />
                    <span className="capitalize">{f.replace(/_/g, ' ')}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Manage / upgrade CTA */}
          <div className="rounded-xl border border-gray-200 bg-white p-6">
            <h2 className="mb-1 text-sm font-semibold text-gray-700">Manage subscription</h2>
            <p className="mb-4 text-sm text-gray-500">
              {plan.valid
                ? 'Need to change your plan, update payment info, or download invoices?'
                : 'Ready to unlock enterprise features for your team?'}
            </p>
            <div className="flex gap-3">
              {plan.valid ? (
                <a
                  href="https://velane.sh/billing"
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                >
                  Manage subscription
                  <ArrowUpRight size={13} />
                </a>
              ) : (
                <a
                  href="https://velane.sh/pricing"
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex items-center gap-1.5 rounded-lg bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800"
                >
                  <Zap size={14} />
                  View pricing
                  <ArrowUpRight size={13} />
                </a>
              )}
              <a
                href="mailto:support@velane.sh"
                className="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-600 hover:bg-gray-50"
              >
                Contact sales
              </a>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
