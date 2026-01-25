// Input: TanStack Router, ServersPage module
// Output: Servers route component with URL-synced tab and server selection
// Position: /servers route - unified servers management page

import { createFileRoute } from '@tanstack/react-router'

import { ServersPage } from '@/modules/servers/servers-page'

export const Route = createFileRoute('/servers')({
  validateSearch: (search: Record<string, unknown>) => {
    const tab = typeof search.tab === 'string' && ['overview', 'tools', 'configuration'].includes(search.tab)
      ? search.tab as 'overview' | 'tools' | 'configuration'
      : 'overview'
    const server = typeof search.server === 'string' && search.server.length > 0
      ? search.server
      : undefined
    return { tab, server }
  },
  component: ServersRoute,
})

function ServersRoute() {
  return <ServersPage />
}
