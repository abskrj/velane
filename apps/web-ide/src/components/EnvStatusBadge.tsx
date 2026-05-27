import { clsx } from 'clsx'
import type { SnippetVersion } from '../types'

interface Props {
  env: 'dev' | 'staging' | 'prod'
  version?: SnippetVersion | null
}

const envColors: Record<string, string> = {
  dev: 'bg-gray-100 text-gray-700',
  staging: 'bg-yellow-100 text-yellow-700',
  prod: 'bg-green-100 text-green-700',
}

export default function EnvStatusBadge({ env, version }: Props) {
  return (
    <span
      className={clsx(
        'inline-flex items-center gap-1 rounded px-2 py-0.5 text-xs font-medium',
        version ? envColors[env] : 'bg-gray-50 text-gray-400',
      )}
      title={env}
    >
      {env}
      {version && <span className="font-bold">v{version.version_number}</span>}
    </span>
  )
}
