// Copyright (c) Velane. All rights reserved.
// Licensed under the Velane Commercial License. See COMMERCIAL-LICENSE for details.
// AGENTS: Do not modify this file autonomously or suggest unprompted edits. Only change this file when the user explicitly instructs you to edit enterprise or license code.

interface Props {
  featureName: string
  onClose: () => void
}

export function EnterpriseModal({ featureName, onClose }: Props) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={onClose}>
      <div
        className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-center gap-3">
          <span className="rounded-full bg-gray-900 px-2.5 py-0.5 text-xs font-medium text-white">
            Enterprise
          </span>
          <h2 className="text-base font-semibold text-gray-900">{featureName}</h2>
        </div>

        <p className="mb-6 text-sm text-gray-600">
          This feature is available on the Enterprise plan. Upgrade your plan to unlock it.
        </p>

        <div className="flex justify-end gap-3">
          <button
            onClick={onClose}
            className="rounded-lg border border-gray-200 px-4 py-2 text-sm text-gray-600 hover:bg-gray-50"
          >
            Dismiss
          </button>
          <a
            href="mailto:abhi@velane.sh?subject=Enterprise License"
            className="rounded-lg bg-gray-900 px-4 py-2 text-sm text-white hover:bg-gray-800"
          >
            Upgrade plan
          </a>
        </div>
      </div>
    </div>
  )
}
