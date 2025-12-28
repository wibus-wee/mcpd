// Input: TanStack Router, ConfigPage module
// Output: Configuration route component
// Position: /config route

import { createFileRoute } from '@tanstack/react-router'

import { ConfigPage } from '@/modules/config/config-page'

export const Route = createFileRoute('/config')({
  component: ConfigPage,
})
