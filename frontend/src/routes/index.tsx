// Input: TanStack Router, DashboardPage module
// Output: Home route component
// Position: Root index route

import { createFileRoute } from '@tanstack/react-router'

import { DashboardPage } from '@/modules/dashboard/dashboard-page'

export const Route = createFileRoute('/')({
  component: DashboardPage,
})
