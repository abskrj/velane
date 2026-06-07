import { useEffect } from 'react'
import { matchPath, useLocation } from 'react-router-dom'
import { setDocumentTitle } from '../lib/documentTitle'

const dashboardRoutes: { path: string; title: string }[] = [
  { path: '/dashboard/snippets/:id', title: 'Snippets' },
  { path: '/dashboard/overview', title: 'Overview' },
  { path: '/dashboard/snippets', title: 'Snippets' },
  { path: '/dashboard/integrations', title: 'Integrations' },
  { path: '/dashboard/mcp', title: 'MCP' },
  { path: '/dashboard/variables', title: 'Variables' },
  { path: '/dashboard/api-keys', title: 'API Keys' },
  { path: '/dashboard/team', title: 'Team' },
  { path: '/dashboard/branding', title: 'Branding' },
  { path: '/dashboard/usage', title: 'Usage' },
  { path: '/dashboard/egress', title: 'Egress Policy' },
  { path: '/dashboard/embed', title: 'Embed' },
]

function titleForPath(pathname: string): string | null {
  for (const route of dashboardRoutes) {
    if (matchPath(route.path, pathname)) return route.title
  }
  return null
}

/** Sets document.title from the current dashboard route. Entity pages may override via useDocumentTitle. */
export default function RouteDocumentTitle() {
  const { pathname } = useLocation()

  useEffect(() => {
    const title = titleForPath(pathname)
    if (title) setDocumentTitle(title)
  }, [pathname])

  return null
}
