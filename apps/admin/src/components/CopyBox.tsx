import { useState } from 'react'
import { Copy, Check } from 'lucide-react'
import clsx from 'clsx'

interface Props {
  value: string
  className?: string
}

export default function CopyBox({ value, className }: Props) {
  const [copied, setCopied] = useState(false)

  const handleCopy = () => {
    navigator.clipboard.writeText(value).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  return (
    <div className={clsx('flex items-center gap-2 rounded border border-gray-300 bg-gray-50 p-2', className)}>
      <input
        type="text"
        readOnly
        value={value}
        className="flex-1 bg-transparent font-mono text-sm text-gray-700 outline-none"
      />
      <button
        onClick={handleCopy}
        className="flex-shrink-0 rounded p-1 text-gray-500 hover:bg-gray-200 hover:text-gray-700"
        title="Copy to clipboard"
      >
        {copied ? <Check size={16} className="text-green-600" /> : <Copy size={16} />}
      </button>
    </div>
  )
}
