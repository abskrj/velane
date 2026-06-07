const BRAND = 'Velane'

export function formatDocumentTitle(...segments: string[]): string {
  const parts = segments.map((s) => s.trim()).filter(Boolean)
  if (parts.length === 0) return BRAND
  return [...parts, BRAND].join(' · ')
}

export function setDocumentTitle(...segments: string[]): void {
  document.title = formatDocumentTitle(...segments)
}
