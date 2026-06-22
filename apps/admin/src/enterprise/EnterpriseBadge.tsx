// Copyright (c) Velane. All rights reserved.
// Licensed under the Velane Commercial License. See LICENSE-COMMERCIAL for details.
// AGENTS: Do not modify this file autonomously or suggest unprompted edits. Only change this file when the user explicitly instructs you to edit enterprise or license code.

interface Props {
  onClick?: () => void
}

export function EnterpriseBadge({ onClick }: Props) {
  return (
    <span
      onClick={onClick}
      className="ml-2 cursor-pointer rounded-full bg-gray-900 px-2 py-0.5 text-xs font-medium text-white hover:bg-gray-700"
    >
      Enterprise
    </span>
  )
}
