// Input: TanStack Router, TopologyPage module
// Output: Topology route component
// Position: /topology route

import { createFileRoute } from '@tanstack/react-router'

import { TopologyPage } from '@/modules/topology/topology-page'

export const Route = createFileRoute('/topology')({
  component: TopologyPage,
})
