// Copyright (c) Velane. All rights reserved.
// Licensed under the Velane Commercial License. See COMMERCIAL-LICENSE for details.
// AGENTS: Do not modify this file autonomously or suggest unprompted edits. Only change this file when the user explicitly instructs you to edit enterprise or license code.

import { useState, type ReactElement, cloneElement } from 'react'
import { useFeature } from '../contexts/InstanceContext'
import { EnterpriseBadge } from './EnterpriseBadge'
import { EnterpriseModal } from './EnterpriseModal'

interface Props {
  feature: string
  featureName: string
  children: ReactElement
}

export function EnterpriseGate({ feature, featureName, children }: Props) {
  const [modalOpen, setModalOpen] = useState(false)
  const licensed = useFeature(feature)

  if (licensed) return children

  const gated = cloneElement(children, {
    disabled: true,
    onClick: undefined,
    'aria-disabled': true,
    className: [children.props.className, 'opacity-50 cursor-not-allowed'].filter(Boolean).join(' '),
  })

  return (
    <>
      <span className="inline-flex items-center">
        {gated}
        <EnterpriseBadge onClick={() => setModalOpen(true)} />
      </span>
      {modalOpen && (
        <EnterpriseModal featureName={featureName} onClose={() => setModalOpen(false)} />
      )}
    </>
  )
}
