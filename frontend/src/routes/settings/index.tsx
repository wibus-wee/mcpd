// Input: TanStack Router
// Output: Settings index that redirects to runtime
// Position: /settings index route

import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/settings/')({
  beforeLoad: () => {
    throw redirect({ to: '/settings/runtime' })
  },
})
