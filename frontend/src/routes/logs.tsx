// Input: TanStack Router
// Output: Logs route component
// Position: /logs route

import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/logs')({
  component: LogsPage,
})

function LogsPage() {
  return (
    <div className="container py-6">
      <h1 className="mb-4 font-bold text-2xl">Logs</h1>
      <p className="text-muted-foreground">System logs viewer</p>
    </div>
  )
}
